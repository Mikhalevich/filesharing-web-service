package handler

import (
	"net/http"

	"github.com/Mikhalevich/filesharing/httpcode"
)

// GetFileHandler get single file from storage
func (h *Handler) GetFileHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to get request parametes"), w, "GetFileHandler")
		return
	}

	req, err := h.makeGatewayRequest(sp.StorageName, r)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to make request"), w, "GetFileHandler")
		return
	}
	http.Redirect(w, req, req.URL.String(), http.StatusFound)
}
