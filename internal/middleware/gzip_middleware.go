package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

func GzipMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Path(), "/debug/pprof") {
				return next(c)
			}

			if !strings.Contains(c.Request().Header.Get("Accept-Encoding"), "gzip") {
				return next(c)
			}

			// Берем буфер из пула вместо new(bytes.Buffer)
			buf := bufferPool.Get().(*bytes.Buffer)
			buf.Reset()
			defer bufferPool.Put(buf)

			originalWriter := c.Response().Writer
			bufferedProxy := &bodyBufferedResponseWriter{
				ResponseWriter: originalWriter,
				body:           buf,
				statusCode:     http.StatusOK,
			}

			c.Response().Writer = bufferedProxy
			err := next(c)
			c.Response().Writer = originalWriter

			if err != nil {
				return err
			}

			contentType := c.Response().Header().Get("Content-Type")
			if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
				c.Response().Header().Set("Content-Encoding", "gzip")
				c.Response().Header().Del("Content-Length")
				originalWriter.WriteHeader(bufferedProxy.statusCode)

				gz := gzipPool.Get().(*gzip.Writer)
				gz.Reset(originalWriter)

				_, writeErr := gz.Write(buf.Bytes())
				gz.Close()
				gzipPool.Put(gz)

				return writeErr
			}
			originalWriter.WriteHeader(bufferedProxy.statusCode)
			_, writeErr := originalWriter.Write(buf.Bytes())
			return writeErr
		}
	}
}
