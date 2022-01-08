package handler

import (
	"io"
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// IndexHTMLHandler process index.html file
func (h *Handler) IndexHTMLHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request parametes").WithError(err), w, "IndexHTMLHandler")
		return
	}

	rsp, httpErr := h.makeGetRequest(r, w, sp.StorageName, "index.html", sp.Values())
	if httpErr != nil {
		h.Error(httpErr, w, "IndexHTMLHandler")
		return
	}

	defer rsp.Body.Close()

	w.Header().Set("Content-type", "text/html")
	if _, err := io.Copy(w, rsp.Body); err != nil {
		h.Error(httperror.NewInternalError("can't copy file").WithError(err), w, "IndexHTMLHandler")
		return
	}
}
