package handler

import (
	"fmt"
	"net/http"

	"github.com/Mikhalevich/filesharing/httpcode"
)

// ShareTextHandler crate file from share text request
func (h *Handler) ShareTextHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")

	if title == "" || body == "" {
		h.Error(httpcode.NewBadRequest(fmt.Sprintf("title or body was not set; title = %s body = %s", title, body)), w, "ShareTextHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to get request parametes"), w, "ShareTextHandler")
		return
	}

	rsp, httpErr := h.processURLEncodedRequest(r, sp.StorageName)
	if httpErr != nil {
		h.handleError(httpErr, w, r, "ShareTextHandler")
		return
	}

	defer rsp.Body.Close()
	w.WriteHeader(http.StatusOK)
}
