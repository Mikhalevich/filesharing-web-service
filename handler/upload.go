package handler

import (
	"net/http"

	"github.com/Mikhalevich/filesharing/httpcode"
)

// UploadHandler upload file to storage
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to get request parametes"), w, "UploadHandler")
		return
	}

	rsp, httpErr := h.processMultipartRequest(r, sp.StorageName, w)
	if httpErr != nil {
		h.handleError(httpErr, w, r, "GetFileHandler")
		return
	}

	defer rsp.Body.Close()
	w.WriteHeader(http.StatusOK)
}
