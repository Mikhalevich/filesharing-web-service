package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Mikhalevich/filesharing-web-service/internal/template"
	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// LoginHandler sign in for the existing storage(user)
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := template.NewTemplatePassword()
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

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request parametes").WithError(err), w, "LoginHandler")
		return
	}

	userInfo.Password = r.FormValue("password")

	if sp.StorageName == "" {
		userInfo.AddError("name", "Please specify storage name to login")
	}

	if userInfo.Password == "" {
		userInfo.AddError("password", "Please enter password to login")
	}

	if len(userInfo.Errors) > 0 {
		return
	}

	values := sp.Values()
	values.Add("password", userInfo.Password)

	rsp, httpErr := h.makePostRequest(r, w, sp.StorageName, "login", values)
	if httpErr != nil {
		switch httpErr.Code {
		case httperror.CodeNotExist:
			userInfo.AddError("common", "Invalid storage name or password")
		case httperror.CodeNotMatch:
			userInfo.AddError("common", "Invalid storage name or password")
		default:
			userInfo.AddError("common", httpErr.Description)
		}
		return
	}

	defer rsp.Body.Close()

	token, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		h.Error(httperror.NewInternalError("invalid session token").WithError(err), w, "LoginHandler")
		return
	}

	h.session.SetToken(w, &Token{Value: string(token)}, sp.StorageName)

	renderTemplate = false
	http.Redirect(w, r, fmt.Sprintf("/%s", sp.StorageName), http.StatusFound)
}
