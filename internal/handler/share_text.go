package handler

import (
	"fmt"
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// ShareTextHandler crate file from share text request
func (h *Handler) ShareTextHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")

	if title == "" || body == "" {
		h.Error(httperror.NewInvalidParams(fmt.Sprintf("title or body was not set; title = %s body = %s", title, body)), w, "ShareTextHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request parametes").WithError(err), w, "ShareTextHandler")
		return
	}

	values := sp.Values()
	values.Add("title", title)
	values.Add("body", body)

	rsp, httpErr := h.makePostRequest(r, w, sp.StorageName, "shareText", values)
	if httpErr != nil {
		h.Error(httpErr, w, "ShareTextHandler")
		return
	}

	defer rsp.Body.Close()
	w.WriteHeader(http.StatusOK)
}
