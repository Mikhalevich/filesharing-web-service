package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Mikhalevich/filesharing-web-service/internal/template"
	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// RegisterHandler register a new storage(user)
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := template.NewTemplateRegister()
	renderTemplate := true

	defer func() {
		if renderTemplate {
			if err := userInfo.Execute(w); err != nil {
				h.logger.Error(err)
			}
		}
	}()

	if r.Method != http.MethodPost {
		return
	}

	userInfo.StorageName = r.FormValue("name")
	userInfo.Password = r.FormValue("password")

	if userInfo.StorageName == "" {
		userInfo.AddError("name", "please specify storage name")
		return
	}

	values := url.Values{}
	values.Add("storage", userInfo.StorageName)
	values.Add("password", userInfo.Password)

	rsp, httpErr := h.makePostRequest(r, w, userInfo.StorageName, "register", values)
	if httpErr != nil {
		switch httpErr.Code {
		case httperror.CodeAlreadyExist:
			userInfo.AddError("common", "storage with this name already exists")
		default:
			userInfo.AddError("common", httpErr.Description)
		}
		return
	}

	defer rsp.Body.Close()

	token, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		h.Error(httperror.NewInternalError("invalid session token").WithError(err), w, "RegisterHandler")
		return
	}

	h.session.SetToken(w, &Token{Value: string(token)}, userInfo.StorageName)

	renderTemplate = false
	http.Redirect(w, r, fmt.Sprintf("/%s", userInfo.StorageName), http.StatusFound)
}
