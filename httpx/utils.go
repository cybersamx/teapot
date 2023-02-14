package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

var _ http.Handler = (*HandlerWrapper)(nil)

type responseWriterWrapper struct {
	http.ResponseWriter
	status int
	size   int
}

func (rww *responseWriterWrapper) Status() int {
	return rww.status
}

func (rww *responseWriterWrapper) Size() int {
	return rww.size
}

func (rww *responseWriterWrapper) Write(data []byte) (int, error) {
	if rww.status > 0 {
		rww.WriteHeader(rww.status)
	}

	n, err := rww.ResponseWriter.Write(data)
	rww.size += n

	return n, err
}

func (rww *responseWriterWrapper) WriteHeader(status int) {
	if rww.status > 0 {
		// Status code already written.
		return
	}

	rww.status = status
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

type HandlerFunc func(ctx *Context, w *responseWriterWrapper, r *http.Request)

// HandlerWrapper implements http.Handler that can be used to wrap another handler to
// form a chain of handlers to handle an incoming request.
type HandlerWrapper struct {
	Factory     *HandlerWrapperFactory
	HandlerFunc HandlerFunc
}

func (hw HandlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rww := newWriterWrapper(w)

	ctx := Context{
		startTime: time.Now(),
	}

	for _, mw := range hw.Factory.preMiddleware {
		mw(&ctx, rww, r)
	}

	hw.HandlerFunc(&ctx, rww, r)

	for _, mw := range hw.Factory.postMiddleware {
		mw(&ctx, rww, r)
	}
}

func WrapHandlerFunc(hf http.HandlerFunc) HandlerFunc {
	return func(ctx *Context, w *responseWriterWrapper, r *http.Request) {
		hf(w, r)
	}
}

func WrapHandler(h http.Handler) HandlerFunc {
	return func(ctx *Context, w *responseWriterWrapper, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}
