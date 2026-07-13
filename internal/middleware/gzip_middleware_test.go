package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestGzipMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		acceptEncoding string
		contentType    string
		responseBody   string
		wantGzip       bool
	}{
		{
			name:           "No gzip support in request",
			acceptEncoding: "",
			contentType:    "application/json",
			responseBody:   `{"status":"ok"}`,
			wantGzip:       false,
		},
		{
			name:           "Gzip supported and correct Content-Type (JSON)",
			acceptEncoding: "gzip, deflate",
			contentType:    "application/json; charset=UTF-8",
			responseBody:   `{"message":"hello world","data":[1,2,3]}`,
			wantGzip:       true,
		},
		{
			name:           "Gzip supported but skipped Content-Type (Plain text)",
			acceptEncoding: "gzip",
			contentType:    "text/plain",
			responseBody:   "plain text response",
			wantGzip:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Инициализация Echo и запроса
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// 2. Фейковый хендлер, который имитирует ответ сервера
			handler := func(ctx echo.Context) error {
				ctx.Response().Header().Set("Content-Type", tt.contentType)
				return ctx.String(http.StatusOK, tt.responseBody)
			}

			// 3. Вызов middleware
			middleware := GzipMiddleware()
			err := middleware(handler)(c)

			if err != nil {
				t.Fatalf("Middleware returned an error: %v", err)
			}

			// 4. Проверка результатов
			resHeader := rec.Header()
			contentEncoding := resHeader.Get("Content-Encoding")

			if tt.wantGzip {
				if contentEncoding != "gzip" {
					t.Errorf("Expected Content-Encoding 'gzip', got '%s'", contentEncoding)
				}

				// Проверяем, что Content-Length удален, как заложено в логике
				if len(resHeader.Values("Content-Length")) > 0 {
					t.Error("Expected Content-Length header to be deleted")
				}

				// Распаковываем обратно для проверки целостности данных
				reader, err := gzip.NewReader(rec.Body)
				if err != nil {
					t.Fatalf("Failed to create gzip reader: %v", err)
				}
				defer reader.Close()

				unzippedData, err := io.ReadAll(reader)
				if err != nil {
					t.Fatalf("Failed to read unzipped data: %v", err)
				}

				if string(unzippedData) != tt.responseBody {
					t.Errorf("Expected body %q, got %q", tt.responseBody, string(unzippedData))
				}
			} else {
				if contentEncoding == "gzip" {
					t.Error("Content should not be gzipped")
				}

				if rec.Body.String() != tt.responseBody {
					t.Errorf("Expected body %q, got %q", tt.responseBody, rec.Body.String())
				}
			}
		},
		)
	}
}
