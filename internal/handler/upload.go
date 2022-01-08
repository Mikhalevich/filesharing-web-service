package handler

import (
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// UploadHandler upload file to storage
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request parametes").WithError(err), w, "UploadHandler")
		return
	}

	rsp, httpErr := h.makeMultipartRequest(r, w, sp.StorageName, "upload")
	if httpErr != nil {
		h.Error(httpErr, w, "GetFileHandler")
		return
	}

	defer rsp.Body.Close()
	w.WriteHeader(http.StatusOK)
}
