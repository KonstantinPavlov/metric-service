package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

type PrivateKeyMiddleware struct {
	privateKey *rsa.PrivateKey
}

func NewPrivateKeyMiddleware(keyPath string) (*PrivateKeyMiddleware, error) {
	if keyPath == "" {
		return &PrivateKeyMiddleware{}, nil
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid private key file format (expected RSA PRIVATE KEY)")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &PrivateKeyMiddleware{privateKey: privateKey}, nil
}

func (p *PrivateKeyMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if p.privateKey == nil {
				return next(c)
			}

			encryptedBody, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return c.String(http.StatusBadRequest, "failed to read encrypted body")
			}
			defer c.Request().Body.Close()

			if len(encryptedBody) == 0 {
				return next(c)
			}

			decryptedBody, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, p.privateKey, encryptedBody, nil)
			if err != nil {
				return c.String(http.StatusBadRequest, "decryption failed: "+err.Error())
			}
			c.Request().Body = io.NopCloser(bytes.NewReader(decryptedBody))
			return next(c)
		}
	}
}
