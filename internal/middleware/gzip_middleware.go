package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)


func GzipMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !strings.Contains(c.Request().Header.Get("Accept-Encoding"), "gzip") {
				return next(c)
			}
			buf := new(bytes.Buffer)
			originalWriter := c.Response().Writer
			bufferedProxy := &bodyBufferedResponseWriter{
				ResponseWriter: originalWriter,
				body:           buf,
				statusCode:     http.StatusNotFound,
			}
			c.Response().Writer = bufferedProxy
			err := next(c)
			c.Response().Writer = originalWriter
			if err != nil {
				return err
			}			
			contentType := c.Response().Header().Get("Content-Type")

			if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
				c.Response().Header().Add("Content-Encoding", "gzip")
				c.Response().Header().Del("Content-Length")

				var gzipBuf bytes.Buffer
				gz := gzip.NewWriter(&gzipBuf)
				_, err := gz.Write(buf.Bytes())
				if err != nil {
					return err
				}
				err = gz.Close()

				if err != nil {
					return err
				}
				originalWriter.WriteHeader(bufferedProxy.statusCode)
				_, writeErr := originalWriter.Write(gzipBuf.Bytes())
				return writeErr
			} else {
				originalWriter.WriteHeader(bufferedProxy.statusCode)
				_, writeErr := originalWriter.Write(buf.Bytes())
				return writeErr
			}
		}
	}
}
