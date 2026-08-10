package middleware

import (
	"bytes"
	"net/http"
)

type bodyBufferedResponseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *bodyBufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *bodyBufferedResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}