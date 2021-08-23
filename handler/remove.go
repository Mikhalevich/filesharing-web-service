package handler

import (
	"net/http"

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

	rsp, httpErr := h.processURLEncodedRequest(r, sp.StorageName, w)
	if httpErr != nil {
		h.handleError(httpErr, w, r, "RemoveHandler")
		return
	}

	defer rsp.Body.Close()
	w.WriteHeader(http.StatusOK)
}
