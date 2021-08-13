package handler

import (
	"io"
	"net/http"

	"github.com/Mikhalevich/filesharing/httpcode"
)

// IndexHTMLHandler process index.html file
func (h *Handler) IndexHTMLHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to get request parametes"), w, "IndexHTMLHandler")
		return
	}

	req, err := h.makeGatewayRequest(sp.StorageName, r)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to make request"), w, "IndexHTMLHandler")
		return
	}

	client := http.Client{}

	rsp, err := client.Do(req)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to make request"), w, "IndexHTMLHandler")
		return
	}

	defer rsp.Body.Close()

	if err := convertStatusCode(rsp.StatusCode); err != nil {
		h.handleError(err, sp.StorageName, w, r, "IndexHTMLHandler")
		return
	}

	w.Header().Set("Content-type", "text/html")
	if _, err := io.Copy(w, rsp.Body); err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "can't copy file"), w, "IndexHTMLHandler")
		return
	}
}
