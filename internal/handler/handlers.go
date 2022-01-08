package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/Mikhalevich/filesharing/pkg/ctxinfo"
	"github.com/Mikhalevich/filesharing/pkg/httperror"
	"github.com/Mikhalevich/filesharing/pkg/service"
)

const (
	// Title it's just title for view page
	Title = "Duplo"
)

var (
	// ErrAlreadyExist indicates that storage already exists
	ErrAlreadyExist        = errors.New("alredy exist")
	ErrNotExist            = errors.New("not exist")
	ErrNotMatch            = errors.New("not match")
	ErrExpired             = errors.New("session is expired")
	ErrNotAuthorized       = errors.New("not authorized")
	ErrInternalServerError = errors.New("intrnal server error")
)

// File represents one file from storage
type File struct {
	Name    string
	Size    int64
	ModTime int64
}

type User struct {
	Name string
	Pwd  string
}

type Token struct {
	Value string
}

type Sessioner interface {
	GetToken(name string, r *http.Request) *Token
	SetToken(w http.ResponseWriter, token *Token, name string)
	Remove(w http.ResponseWriter, name string)
}

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	WithField(key string, value interface{}) service.Logger
	WithError(err error) service.Logger
}

// Handler represents gateway handler
type Handler struct {
	gwh     string
	session Sessioner
	logger  Logger
}

// New constructor for Handler
func New(gatewayHost string, ses Sessioner, l Logger) *Handler {
	return &Handler{
		gwh:     gatewayHost,
		session: ses,
		logger:  l,
	}
}

func (h *Handler) Error(err *httperror.Error, w http.ResponseWriter, handler string) {
	h.logger.WithError(err).
		WithField("handler", handler).
		Error("handler error")
	err.WriteJSON(w)
}

type storageParameters struct {
	StorageName string
	IsPublic    bool
	IsPermanent bool
	FileName    string
}

func (sp storageParameters) Values() url.Values {
	values := url.Values{}
	values.Add("storage", sp.StorageName)
	if sp.IsPublic {
		values.Add("public", "true")
	}

	if sp.IsPermanent {
		values.Add("permanent", "true")
	}
	values.Add("file", sp.FileName)

	return values
}

func (h *Handler) requestParameters(r *http.Request) (storageParameters, error) {
	ctx := r.Context()
	storage, err := ctxinfo.UserName(ctx)
	if errors.Is(err, ctxinfo.ErrNotFound) {
		storage = ""
	} else if err != nil {
		return storageParameters{}, fmt.Errorf("unable to get storage name: %w", err)
	}

	isPublic, err := ctxinfo.PublicStorage(ctx)
	if errors.Is(err, ctxinfo.ErrNotFound) {
		isPublic = false
	} else if err != nil {
		return storageParameters{}, fmt.Errorf("unable to get public storage: %w", err)
	}

	isPermanent, err := ctxinfo.PermanentStorage(ctx)
	if errors.Is(err, ctxinfo.ErrNotFound) {
		isPermanent = false
	} else if err != nil {
		return storageParameters{}, fmt.Errorf("unable to get permanent storage: %w", err)
	}

	fileName, err := ctxinfo.FileName(ctx)
	if errors.Is(err, ctxinfo.ErrNotFound) {
		fileName = ""
	} else if err != nil {
		return storageParameters{}, fmt.Errorf("unable to get file name: %w", err)
	}

	return storageParameters{
		StorageName: storage,
		IsPublic:    isPublic,
		IsPermanent: isPermanent,
		FileName:    fileName,
	}, nil
}

// RecoverMiddleware middlewere recover for undefined panic error
func (h *Handler) RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				h.Error(httperror.NewInternalError("recover from panic").WithError(e), w, "RecoverHandler")
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) makeURL(endpoint string) string {
	return fmt.Sprintf("%s/%s/", h.gwh, endpoint)
}

func (h *Handler) sessionToken(r *http.Request, storageName string) string {
	if token := h.session.GetToken(storageName, r); token != nil {
		return token.Value
	}
	return ""
}

func (h *Handler) processRequest(req *http.Request, storageName string, w http.ResponseWriter) (*http.Response, *httperror.Error) {
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, httperror.NewInternalError("do request").WithError(err)
	}

	if rsp.StatusCode != http.StatusOK {
		if rsp.StatusCode == http.StatusBadRequest {
			var httpErr httperror.Error
			if err := json.NewDecoder(rsp.Body).Decode(&httpErr); err != nil {
				return nil, httperror.NewInternalError("json decode").WithError(err)
			}

			// if httpErr.Code == httperror.CodeUnauthorized {
			// 	return nil, httpcode.NewHTTPRedirectFoundError(fmt.Sprintf("/login/%s", storageName), err.Error())
			// }

			return nil, &httpErr
		}

		return nil, httperror.NewInternalError("invalid status code")
	}

	if token := rsp.Header.Get("X-Token"); token != "" {
		h.session.SetToken(w, &Token{Value: string(token)}, storageName)
	}

	return rsp, nil
}

func (h *Handler) makeGetRequest(originReq *http.Request, w http.ResponseWriter, storageName string, endpoint string, values url.Values) (*http.Response, *httperror.Error) {
	req, err := http.NewRequest(http.MethodGet, h.makeURL(endpoint), nil)
	if err != nil {
		return nil, httperror.NewInternalError("make get request").WithError(err)
	}

	req.URL.RawQuery = values.Encode()

	if token := h.sessionToken(originReq, storageName); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return h.processRequest(req, storageName, w)
}

func (h *Handler) makePostRequest(originReq *http.Request, w http.ResponseWriter, storageName string, endpoint string, values url.Values) (*http.Response, *httperror.Error) {
	req, err := http.NewRequest(http.MethodPost, h.makeURL(endpoint), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, httperror.NewInternalError("make post request").WithError(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if token := h.sessionToken(originReq, storageName); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return h.processRequest(req, storageName, w)
}

func (h *Handler) multipartBody(originReq *http.Request) (*bytes.Buffer, string, error) {
	mr, err := originReq.MultipartReader()
	if err != nil {
		return nil, "", fmt.Errorf("multipart reader: %w", err)
	}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)

	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, "", fmt.Errorf("next part: %w", err)
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		filePart, err := mw.CreateFormFile(fileName, fileName)
		if err != nil {
			return nil, "", fmt.Errorf("create form file: %w", err)
		}

		if _, err = io.Copy(filePart, part); err != nil {
			return nil, "", fmt.Errorf("copy data: %w", err)
		}
	}

	if err = mw.Close(); err != nil {
		return nil, "", fmt.Errorf("close: %w", err)
	}

	return body, mw.FormDataContentType(), nil
}

func (h *Handler) makeMultipartRequest(originReq *http.Request, w http.ResponseWriter, storageName string, endpoint string) (*http.Response, *httperror.Error) {
	body, contentType, err := h.multipartBody(originReq)
	if err != nil {
		return nil, httperror.NewInternalError("make body").WithError(err)
	}

	req, err := http.NewRequest(http.MethodPost, h.makeURL(endpoint), body)
	if err != nil {
		return nil, httperror.NewInternalError("make post request").WithError(err)
	}

	req.Header.Set("Content-Type", contentType)
	if token := h.sessionToken(originReq, storageName); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return h.processRequest(req, storageName, w)
}
