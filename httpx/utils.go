package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var _ http.Handler = (*HandlerWrapper)(nil)

type responseWriterWrapper struct {
	http.ResponseWriter
}

func isBodyAllowed(code int) bool {
	switch {
	case code >= 100 && code < 200,
		code == http.StatusNoContent,
		code == http.StatusNotModified:
		return false
	}

	return true
}

func newWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{
		ResponseWriter: w,
	}
}

func renderJSON(w http.ResponseWriter, code int, obj any) {
	w.WriteHeader(code)

	if !isBodyAllowed(code) || obj == nil {
		return
	}

	data, err := json.Marshal(obj)
	if err != nil {
		panic(err) // Let recovery middleware handles the panic.
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	w.Write(data)
}

type HandlerFunc func(ctx *Context, w http.ResponseWriter, r *http.Request)

// HandlerWrapper implements http.Handler that can be used to wrap another handler to
// form a chain of handlers to handle an incoming request.
type HandlerWrapper struct {
	Server      *Server
	HandlerFunc HandlerFunc
}

type Context struct{}

func (hw HandlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w = newWriterWrapper(w)

	ctx := &Context{}

	hw.HandlerFunc(ctx, w, r)
}

func APIHandler(handlerFn HandlerFunc) http.Handler {
	hw := HandlerWrapper{
		Server:      nil,
		HandlerFunc: handlerFn,
	}

	return &hw
}

func WrapHandlerFunc(hf http.HandlerFunc) HandlerFunc {
	return func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		hf(w, r)
	}
}

func WrapHandler(h http.Handler) HandlerFunc {
	return func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}
