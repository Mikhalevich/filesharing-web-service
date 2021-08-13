package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Mikhalevich/filesharing-web-service/template"
	"github.com/Mikhalevich/filesharing/httpcode"
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
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to get request parametes"), w, "LoginHandler")
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

	rsp, err := http.PostForm(h.convertToGatewayURL(r.URL), url.Values{"password": {userInfo.Password}})
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to make request"), w, "LoginHandler")
		return
	}

	defer rsp.Body.Close()

	if err := convertStatusCode(rsp.StatusCode); err != nil {
		if errors.Is(err, ErrNotExist) {
			userInfo.AddError("common", "Invalid storage name or password")
		} else if errors.Is(err, ErrNotMatch) {
			userInfo.AddError("common", "Invalid storage name or password")
		} else {
			userInfo.AddError("common", err.Error())
		}
		return
	}

	token, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "invalid session token"), w, "LoginHandler")
		return
	}

	h.session.SetToken(w, &Token{Value: string(token)}, sp.StorageName)

	renderTemplate = false
	http.Redirect(w, r, fmt.Sprintf("/%s", sp.StorageName), http.StatusFound)
}
