package httpapi

import (
	"net/http"
	"strings"
)

type errorResponseWriter struct {
	http.ResponseWriter
	intercepted bool
}

func (w *errorResponseWriter) WriteHeader(status int) {
	header := w.Header().Get("Content-Type")

	if (status == http.StatusNotFound || status == http.StatusMethodNotAllowed) && strings.HasPrefix(header, "text/plain") {
		w.intercepted = true
		if status == http.StatusNotFound {
			writeError(w.ResponseWriter, status, "not found")
		}
		if status == http.StatusMethodNotAllowed {
			writeError(w.ResponseWriter, status, "method not allowed")
		}
	} else {
		w.ResponseWriter.WriteHeader(status)
	}

}

func (w *errorResponseWriter) Write(p []byte) (int, error) {
	if w.intercepted {
		return len(p), nil
	}
	return w.ResponseWriter.Write(p)
}

func JSONErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := errorResponseWriter{
			ResponseWriter: w,
			intercepted:    false,
		}
		next.ServeHTTP(&wrapper, r)
	})
}
