package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func CreateMiddleChain(mws ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			mw := mws[i]
			next = mw(next)
		}
		return next
	}
}

type WrappedWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *WrappedWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.StatusCode = status
}
