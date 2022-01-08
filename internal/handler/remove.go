package handler

import (
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// RemoveHandler removes current file from storage
func (h *Handler) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.FormValue("fileName")
	if fileName == "" {
		h.Error(httperror.NewInvalidParams("file name was not set"), w, "RemoveHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInternalError("request parametes").WithError(err), w, "RemoveHandler")
		return
	}

	rsp, httpErr := h.makePostRequest(r, w, sp.StorageName, "remove", sp.Values())
	if httpErr != nil {
		h.Error(httpErr, w, "RemoveHandler")
		return
	}

	defer rsp.Body.Close()
	w.WriteHeader(http.StatusOK)
}
