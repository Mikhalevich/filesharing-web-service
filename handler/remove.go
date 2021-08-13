package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Mikhalevich/filesharing/httpcode"
)

// RemoveHandler removes current file from storage
func (h *Handler) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.FormValue("fileName")
	if fileName == "" {
		h.Error(httpcode.NewBadRequest("file name was not set"), w, "RemoveHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to get request parametes"), w, "RemoveHandler")
		return
	}

	rsp, err := http.PostForm(h.convertToGatewayURL(r.URL), url.Values{"fileName": {fileName}})
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to make request"), w, "RemoveHandler")
		return
	}

	defer rsp.Body.Close()

	if err := convertStatusCode(rsp.StatusCode); err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, fmt.Sprintf("unable to remove file: %s from storage: %s", fileName, sp.StorageName)), w, "RemoveHandler")
		return
	}

	w.WriteHeader(http.StatusOK)
}
