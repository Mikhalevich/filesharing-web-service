package handler

import (
	"fmt"
	"io"
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

	rsp, httpErr := h.processGWRequest(r, sp.StorageName)
	if httpErr != nil {
		h.handleError(httpErr, w, r, "GetFileHandler")
		return
	}

	defer rsp.Body.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", sp.FileName))

	if _, err := io.Copy(w, rsp.Body); err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "failed to transfer bytes"), w, "GetFileHandler")
		return
	}
}
