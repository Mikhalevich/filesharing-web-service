package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// GetFileHandler get single file from storage
func (h *Handler) GetFileHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request parametes").WithError(err), w, "GetFileHandler")
		return
	}

	rsp, httpErr := h.makeGetRequest(r, w, sp.StorageName, "file", sp.Values())
	if httpErr != nil {
		h.Error(httpErr, w, "GetFileHandler")
		return
	}

	defer rsp.Body.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", sp.FileName))

	if _, err := io.Copy(w, rsp.Body); err != nil {
		h.Error(httperror.NewInternalError("failed to transfer bytes").WithError(err), w, "GetFileHandler")
		return
	}
}
