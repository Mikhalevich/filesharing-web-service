package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Mikhalevich/filesharing-web-service/internal/template"
	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// ViewHandler executes view.html template for view files in requested folder
func (h *Handler) ViewHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request parametes").WithError(err), w, "ViewHandler")
		return
	}

	fileResp, httpErr := h.makeGetRequest(r, w, sp.StorageName, "list", sp.Values())
	if httpErr != nil {
		h.Error(httpErr, w, "ViewHandler")
		return
	}
	defer fileResp.Body.Close()

	type fileList struct {
		Name    string `json:"name"`
		Size    int64  `json:"size"`
		ModTime int64  `json:"mod_time"`
	}

	var files []fileList
	if err := json.NewDecoder(fileResp.Body).Decode(&files); err != nil {
		h.Error(httperror.NewInternalError("files json decode error").WithError(err), w, "ViewHandler")
		return
	}

	fileInfos := make([]template.FileInfo, 0, len(files))
	for _, f := range files {
		fileInfos = append(fileInfos, template.FileInfo{
			Name:    f.Name,
			Size:    f.Size,
			ModTime: f.ModTime,
		})
	}

	viewPermanentLink := !sp.IsPermanent && !sp.IsPublic
	viewTemplate := template.NewTemplateView(Title, viewPermanentLink, fileInfos)

	if err := viewTemplate.Execute(w); err != nil {
		h.Error(httperror.NewInternalError("view error").WithError(err), w, "ViewHandler")
		return
	}
}
