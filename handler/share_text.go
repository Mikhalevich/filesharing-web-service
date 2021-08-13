package handler

import (
	"fmt"
	"net/http"
	"net/url"

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

	rsp, err := http.PostForm(h.convertToGatewayURL(r.URL), url.Values{"title": {title}, "body": {body}})
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to make request"), w, "ShareTextHandler")
		return
	}

	defer rsp.Body.Close()

	if err := convertStatusCode(rsp.StatusCode); err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, fmt.Sprintf("unable to store text file: %s for storage: %s", title, sp.StorageName)), w, "ShareTextHandler")
		return
	}

	w.WriteHeader(http.StatusOK)
}
