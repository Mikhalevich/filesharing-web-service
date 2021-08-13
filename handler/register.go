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

	rsp, err := http.PostForm(h.makeGatewayURL("/register"), url.Values{"name": {userInfo.StorageName}, "password": {userInfo.Password}})
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "unable to make request"), w, "RegisterHandler")
		return
	}

	defer rsp.Body.Close()

	if err := convertStatusCode(rsp.StatusCode); err != nil {
		if errors.Is(err, ErrAlreadyExist) {
			userInfo.AddError("common", "storage with this name already exists")
		} else {
			userInfo.AddError("common", err.Error())
		}
		return
	}

	token, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "invalid session token"), w, "RegisterHandler")
		return
	}

	h.session.SetToken(w, &Token{Value: string(token)}, userInfo.StorageName)

	renderTemplate = false
	http.Redirect(w, r, fmt.Sprintf("/%s", userInfo.StorageName), http.StatusFound)
}
