package middleware

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/KonstantinPavlov/metric-service/internal/crypto"
	"github.com/labstack/echo/v4"
)

func Sha256RequestMiddleware(keyBytes []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientHashHex := c.Request().Header.Get("HashSHA256")
			if clientHashHex == "" {
				return next(c)
			}

			bodyBytes, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read request body")
			}

			c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			hasher := sha256.New()
			hasher.Write(keyBytes)
			hasher.Write(bodyBytes)
			expectedHash := hasher.Sum(nil)

			clientHashBytes, err := hex.DecodeString(clientHashHex)
			if err != nil || len(clientHashBytes) != 32 {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid HashSHA256 header format")
			}

			if subtle.ConstantTimeCompare(expectedHash, clientHashBytes) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Bad Request Hash")
			}
			return next(c)
		}
	}
}

func Sha256ResponseMiddleware(keyBytes []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			buf := new(bytes.Buffer)
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
			if buf.Len() > 0 {
				c.Response().Header().Add("HashSHA256", crypto.CalculateSha256(keyBytes, buf.Bytes()))
			}
			originalWriter.WriteHeader(bufferedProxy.statusCode)
			_, writeErr := originalWriter.Write(buf.Bytes())
			return writeErr
		}
	}
}
