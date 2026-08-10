package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KonstantinPavlov/metric-service/internal/crypto"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var testKey = []byte("my-secret-key-32-bytes-long!!!!")

func TestSha256RequestMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		headerHash     string
		expectedStatus int
	}{
		{
			name:           "Valid hash passes",
			body:           "hello world",
			headerHash:     crypto.CalculateSha256(testKey, []byte("hello world")),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No header skips check",
			body:           "hello world",
			headerHash:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid hex format returns 400 Bad Request",
			body:           "hello world",
			headerHash:     "not-a-hex-string",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrong hex length returns 400 Bad Request",
			body:           "hello world",
			headerHash:     "a1b2c3",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			if tt.headerHash != "" {
				req.Header.Set("HashSHA256", tt.headerHash)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				body, _ := io.ReadAll(c.Request().Body)
				assert.Equal(t, tt.body, string(body))
				return c.String(http.StatusOK, "success")
			}

			mw := Sha256RequestMiddleware(testKey)(handler)
			err := mw(c)

			if err != nil {
				if echoErr, ok := err.(*echo.HTTPError); ok {
					assert.Equal(t, tt.expectedStatus, echoErr.Code)
				} else {
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestSha256ResponseMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		handlerBody  string
		expectHeader bool
	}{
		{
			name:         "Adds hash header for non-empty response",
			handlerBody:  "response data from server",
			expectHeader: true,
		},
		{
			name:         "Does not add header for empty response",
			handlerBody:  "",
			expectHeader: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, tt.handlerBody)
			}

			mw := Sha256ResponseMiddleware(testKey)(handler)
			err := mw(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.handlerBody, rec.Body.String())

			headerHash := rec.Header().Get("HashSHA256")

			if tt.expectHeader {
				expectedHash := crypto.CalculateSha256(testKey, []byte(tt.handlerBody))
				assert.Equal(t, expectedHash, headerHash)
			} else {
				assert.Empty(t, headerHash)
			}
		})
	}
}
