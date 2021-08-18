package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/Mikhalevich/filesharing/ctxinfo"
	"github.com/Mikhalevich/filesharing/httpcode"
	"github.com/sirupsen/logrus"
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

// Handler represents gateway handler
type Handler struct {
	gwh     string
	session Sessioner
	logger  *logrus.Logger
}

// NewHandler constructor for Handler
func NewHandler(gateway string, ses Sessioner, l *logrus.Logger) *Handler {
	return &Handler{
		gwh:     gateway,
		session: ses,
		logger:  l,
	}
}

func (h *Handler) Error(err httpcode.Error, w http.ResponseWriter, context string) {
	if err == nil {
		h.logger.Error(fmt.Errorf("[%s] empty error", context))
		http.Error(w, "empty error", http.StatusInternalServerError)
		return
	}

	h.logger.Error(fmt.Errorf("[%s] %s: %w", context, err.Description(), err))
	http.Error(w, err.Description(), err.StatusCode())
}

func (h *Handler) ErrorRedirect(err httpcode.Error, w http.ResponseWriter, r *http.Request, url string, context string) {
	if err == nil {
		h.logger.Error(fmt.Errorf("[%s] empty error", context))
		http.Error(w, "empty error", http.StatusInternalServerError)
		return
	}

	h.logger.Error(fmt.Errorf("[%s] %s: %w", context, err.Description(), err))
	http.Redirect(w, r, url, err.StatusCode())
}

type storageParameters struct {
	StorageName string
	IsPublic    bool
	IsPermanent bool
	FileName    string
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
				h.Error(httpcode.NewWrapInternalServerError(e, "internal server error"), w, "RecoverHandler")
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func convertStatusCode(statusCode int) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return ErrNotAuthorized
	case httpcode.HTTPStatusNotExist:
		return ErrNotExist
	case httpcode.HTTPStatusNotMatch:
		return ErrNotMatch
	case httpcode.HTTPStatusAlreadyExist:
		return ErrAlreadyExist
	}
	return ErrInternalServerError
}

func (h *Handler) handleError(err httpcode.Error, w http.ResponseWriter, r *http.Request, context string) {
	if redirect, ok := err.(*httpcode.HTTPRedirectError); ok {
		h.ErrorRedirect(redirect, w, r, redirect.Location(), context)
		return
	}

	if err != nil {
		h.Error(err, w, context)
		return
	}

	http.Error(w, "empty error", http.StatusInternalServerError)
}

func (h *Handler) makeGatewayURL(path string) string {
	u, err := url.Parse(h.gwh)
	if err != nil {
		return ""
	}
	u.Path = path
	return u.String()
}

func (h *Handler) convertToGatewayURL(u *url.URL) string {
	gwURL := *u
	if gwURL.Scheme == "" {
		gwURL.Scheme = "http"
	}
	gwURL.Host = h.gwh
	return gwURL.String()
}

func (h *Handler) makeURLEncodedRequest(originReq *http.Request, storageName string) (*http.Request, error) {
	if originReq.Form == nil {
		originReq.ParseForm()
	}

	if originReq.Form == nil {
		return nil, errors.New("invalid form values")
	}

	req, err := http.NewRequest(http.MethodPost, h.convertToGatewayURL(originReq.URL), strings.NewReader(originReq.Form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if token := h.session.GetToken(storageName, originReq); token != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Value))
	}

	return req, nil
}

func (h *Handler) makeMultipartRequest(originReq *http.Request, storageName string) (*http.Request, error) {
	mr, err := originReq.MultipartReader()
	if err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)

	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		filePart, err := mw.CreateFormFile(fileName, fileName)
		if err != nil {
			return nil, err
		}

		if _, err = io.Copy(filePart, part); err != nil {
			return nil, err
		}
	}

	if err = mw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, h.convertToGatewayURL(originReq.URL), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())
	if token := h.session.GetToken(storageName, originReq); token != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Value))
	}

	return req, nil
}

func (h *Handler) processGWRequest(req *http.Request, storageName string) (*http.Response, httpcode.Error) {
	client := http.Client{}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, httpcode.NewWrapInternalServerError(err, "unable to execute request")
	}

	defer func() {
		if err != nil {
			rsp.Body.Close()
		}
	}()

	err = convertStatusCode(rsp.StatusCode)
	if errors.Is(err, ErrNotAuthorized) {
		return nil, httpcode.NewHTTPRedirectFoundError(fmt.Sprintf("/login/%s", storageName), err.Error())
	} else if err != nil {
		return nil, httpcode.NewWrapInternalServerError(err, "internal server error")
	}

	return rsp, nil
}

func (h *Handler) processGetRequest(originReq *http.Request, storageName string) (*http.Response, httpcode.Error) {
	req, err := http.NewRequest(http.MethodGet, h.convertToGatewayURL(originReq.URL), nil)
	if err != nil {
		return nil, httpcode.NewWrapInternalServerError(err, "unable to make request")
	}

	if ct := originReq.Header.Get("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	}

	if token := h.session.GetToken(storageName, originReq); token != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Value))
	}

	return h.processGWRequest(req, storageName)
}

func (h *Handler) processURLEncodedRequest(originReq *http.Request, storageName string) (*http.Response, httpcode.Error) {
	req, err := h.makeURLEncodedRequest(originReq, storageName)
	if err != nil {
		return nil, httpcode.NewWrapInternalServerError(err, "unable to make request")
	}

	return h.processGWRequest(req, storageName)
}

func (h *Handler) processMultipartRequest(originReq *http.Request, storageName string) (*http.Response, httpcode.Error) {
	req, err := h.makeMultipartRequest(originReq, storageName)
	if err != nil {
		return nil, httpcode.NewWrapInternalServerError(err, "unable to make request")
	}

	return h.processGWRequest(req, storageName)
}
