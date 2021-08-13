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

	req, err := h.makeGatewayRequest(sp.StorageName, r)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to make request"), w, "UploadHandler")
		return
	}
	http.Redirect(w, req, req.URL.String(), http.StatusFound)
}
