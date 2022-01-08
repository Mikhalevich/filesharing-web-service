package router

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/Mikhalevich/filesharing-web-service/internal/template"
	"github.com/Mikhalevich/filesharing/pkg/ctxinfo"
)

type route struct {
	Pattern       string
	IsPrefix      bool
	Methods       string
	Public        bool
	PermanentPath bool
	Handler       http.Handler
}

type handler interface {
	RegisterHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	IndexHTMLHandler(w http.ResponseWriter, r *http.Request)
	ViewHandler(w http.ResponseWriter, r *http.Request)
	UploadHandler(w http.ResponseWriter, r *http.Request)
	RemoveHandler(w http.ResponseWriter, r *http.Request)
	GetFileHandler(w http.ResponseWriter, r *http.Request)
	ShareTextHandler(w http.ResponseWriter, r *http.Request)
	RecoverMiddleware(next http.Handler) http.Handler
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
}

func configure(h handler) []route {
	return []route{
		{
			Pattern: "/",
			Methods: "GET",
			Public:  true,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/common/", http.StatusMovedPermanently)
			}),
		},
		{
			Pattern:  "/res/",
			IsPrefix: true,
			Methods:  "GET",
			Public:   true,
			Handler:  http.FileServer(http.FS(template.Resources())),
		},
		{
			Pattern: "/register/",
			Methods: "GET,POST",
			Public:  true,
			Handler: http.HandlerFunc(h.RegisterHandler),
		},
		{
			Pattern: "/login/{storage}/",
			Methods: "GET,POST",
			Public:  true,
			Handler: http.HandlerFunc(h.LoginHandler),
		},
		{
			Pattern: "/{storage}/index.html",
			Methods: "GET",
			Handler: http.HandlerFunc(h.IndexHTMLHandler),
		},
		{
			Pattern:       "/{storage}/permanent/index.html",
			Methods:       "GET",
			PermanentPath: true,
			Handler:       http.HandlerFunc(h.IndexHTMLHandler),
		},
		{
			Pattern:       "/{storage}/permanent/{file}/",
			Methods:       "GET",
			PermanentPath: true,
			Handler:       http.HandlerFunc(h.GetFileHandler),
		},
		{
			Pattern:       "/{storage}/permanent/",
			Methods:       "GET",
			PermanentPath: true,
			Handler:       http.HandlerFunc(h.ViewHandler),
		},
		{
			Pattern: "/{storage}/{file}/",
			Methods: "GET",
			Handler: http.HandlerFunc(h.GetFileHandler),
		},
		{
			Pattern: "/{storage}/",
			Methods: "GET",
			Handler: http.HandlerFunc(h.ViewHandler),
		},
		{
			Pattern: "/{storage}/upload/",
			Methods: "POST",
			Handler: http.HandlerFunc(h.UploadHandler),
		},
		{
			Pattern:       "/{storage}/permanent/upload/",
			Methods:       "POST",
			PermanentPath: true,
			Handler:       http.HandlerFunc(h.UploadHandler),
		},
		{
			Pattern: "/{storage}/remove/",
			Methods: "POST",
			Handler: http.HandlerFunc(h.RemoveHandler),
		},
		{
			Pattern:       "/{storage}/permanent/remove/",
			Methods:       "POST",
			PermanentPath: true,
			Handler:       http.HandlerFunc(h.RemoveHandler),
		},
		{
			Pattern: "/{storage}/shareText/",
			Methods: "POST",
			Handler: http.HandlerFunc(h.ShareTextHandler),
		},
		{
			Pattern:       "/{storage}/permanent/shareText/",
			Methods:       "POST",
			PermanentPath: true,
			Handler:       http.HandlerFunc(h.ShareTextHandler),
		},
	}
}

func storeRouterParametes(isPublic bool, isPermanent bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		ctx := ctxinfo.WithPublicStorage(request.Context(), isPublic)

		storage := mux.Vars(request)["storage"]
		if storage != "" {
			ctx = ctxinfo.WithUserName(ctx, storage)
			ctx = ctxinfo.WithPermanentStorage(ctx, isPermanent)
		}

		fileName := mux.Vars(request)["file"]
		if fileName != "" {
			ctx = ctxinfo.WithFileName(ctx, fileName)
		}
		request = request.WithContext(ctx)

		next.ServeHTTP(w, request)
	})
}

func MakeRoutes(router *mux.Router, authEnabled bool, h handler, l Logger) {
	for _, route := range configure(h) {
		muxRoute := router.NewRoute()
		if route.IsPrefix {
			muxRoute.PathPrefix(route.Pattern)
		} else {
			muxRoute.Path(route.Pattern)
		}

		muxRoute.Methods(strings.Split(route.Methods, ",")...)

		handler := route.Handler
		if authEnabled && !route.Public {
			//handler = r.h.CheckAuthMiddleware(handler)
		}
		handler = storeRouterParametes(route.Public, route.PermanentPath, handler)

		handler = h.RecoverMiddleware(handler)

		muxRoute.Handler(handler)
	}
}
