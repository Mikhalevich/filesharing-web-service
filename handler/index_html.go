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

	rsp, httpErr := h.processGWRequest(r, sp.StorageName)
	if httpErr != nil {
		h.handleError(httpErr, w, r, "IndexHTMLHandler")
		return
	}

	defer rsp.Body.Close()

	w.Header().Set("Content-type", "text/html")
	if _, err := io.Copy(w, rsp.Body); err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "can't copy file"), w, "IndexHTMLHandler")
		return
	}
}
