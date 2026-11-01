package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// helper для генерации пары RSA ключей и записи приватного во временный файл
func generateTestKeyFile(t *testing.T) (string, *rsa.PublicKey) {
	t.Helper()

	// Генерируем ключ (2048 бит для скорости тестов)
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Кодируем приватный ключ в PEM
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}
	pemBytes := pem.EncodeToMemory(privBlock)

	// Создаем временную директорию и файл
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "private_test.pem")
	err = os.WriteFile(keyPath, pemBytes, 0600)
	assert.NoError(t, err)

	return keyPath, &privKey.PublicKey
}

func TestPrivateKeyMiddleware_Success(t *testing.T) {
	// 1. Готовим ключи и middleware
	keyPath, pubKey := generateTestKeyFile(t)
	decryptor, err := NewPrivateKeyMiddleware(keyPath)
	assert.NoError(t, err)

	// 2. Инициализируем Echo
	e := echo.New()

	// Финальный хэндлер, который проверяет, что до него дошли уже ЧИСТЫЕ данные
	targetHandler := func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	}

	// 3. Шифруем тестовое сообщение публичным ключом (как это делает агент)
	secretMessage := []byte("secret-metrics-payload")
	encryptedMessage, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, secretMessage, nil)
	assert.NoError(t, err)

	// 4. Формируем HTTP-запрос
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(encryptedMessage))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 5. Запускаем цепочку: Middleware -> Хэндлер
	middleware := decryptor.Middleware()
	err = middleware(targetHandler)(c)

	// 6. Проверяем результаты
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "secret-metrics-payload", rec.Body.String())
}

func TestPrivateKeyMiddleware_BypassIfNoKey(t *testing.T) {
	// Если путь пустой, middleware должен прозрачно пропускать данные
	decryptor, err := NewPrivateKeyMiddleware("")
	assert.NoError(t, err)

	e := echo.New()
	targetHandler := func(c echo.Context) error {
		body, _ := io.ReadAll(c.Request().Body)
		return c.String(http.StatusOK, string(body))
	}

	rawMessage := "plain-text-data"
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString(rawMessage))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := decryptor.Middleware()
	err = middleware(targetHandler)(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, rawMessage, rec.Body.String())
}

func TestPrivateKeyMiddleware_EmptyBody(t *testing.T) {
	// Middleware не должен падать или выдавать ошибку на пустой POST/GET запрос
	keyPath, _ := generateTestKeyFile(t)
	decryptor, err := NewPrivateKeyMiddleware(keyPath)
	assert.NoError(t, err)

	e := echo.New()
	targetHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "empty-handled")
	}

	req := httptest.NewRequest(http.MethodPost, "/updates/", nil) // nil body
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := decryptor.Middleware()
	err = middleware(targetHandler)(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "empty-handled", rec.Body.String())
}

func TestPrivateKeyMiddleware_DecryptionFailed(t *testing.T) {
	// Передаем мусор вместо зашифрованных RSA байт. Ожидаем статус 400 Bad Request
	keyPath, _ := generateTestKeyFile(t)
	decryptor, err := NewPrivateKeyMiddleware(keyPath)
	assert.NoError(t, err)

	e := echo.New()
	targetHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "should not be reached")
	}

	invalidEncryptedData := []byte("this is not a valid encrypted rsa cyphertext")
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(invalidEncryptedData))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := decryptor.Middleware()
	err = middleware(targetHandler)(c)

	assert.NoError(t, err) // Echo обрабатывает ошибку возвратом c.String(), поэтому сам метод возвращает nil
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "decryption failed")
}

func TestNewPrivateKeyMiddleware_FileNotFound(t *testing.T) {
	// Проверяем обработку ошибки, если файла физически нет
	_, err := NewPrivateKeyMiddleware("non_existent_file.pem")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read private key file")
}
