package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Mikhalevich/filesharing-web-service/template"
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
	GetToken(name string, r *http.Request) (*Token, error)
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

func marshalFileInfo(file *File) *template.FileInfo {
	return &template.FileInfo{
		Name:    file.Name,
		Size:    file.Size,
		ModTime: file.ModTime,
	}
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
	gwURL.Host = h.gwh
	return gwURL.String()
}

func (h *Handler) makeGatewayRequest(name string, r *http.Request) (*http.Request, error) {
	req := r.Clone(r.Context())
	req.URL.Host = h.gwh

	token, err := h.session.GetToken(name, r)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.Value))
	return req, nil
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

func (h *Handler) handleError(err error, name string, w http.ResponseWriter, r *http.Request, context string) {
	if errors.Is(err, ErrNotAuthorized) {
		h.ErrorRedirect(httpcode.NewHTTPError(http.StatusFound, err.Error()), w, r, fmt.Sprintf("/login/%s", name), context)
	} else if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "internal server error"), w, context)
	}

	http.Error(w, "empty error", http.StatusInternalServerError)
}