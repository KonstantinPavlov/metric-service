package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// RoundTripperFunc — удобный тип-заглушка для мока интерфейса http.RoundTripper в тестах
type RoundTripperFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Хелпер для создания пары ключей RSA и сохранения ПУБЛИЧНОГО во временный файл
func generateTestPublicKeyFile(t *testing.T) (string, *rsa.PrivateKey) {
	t.Helper()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	assert.NoError(t, err)

	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}
	pemBytes := pem.EncodeToMemory(pubBlock)

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "public_test.pem")
	err = os.WriteFile(keyPath, pemBytes, 0600)
	assert.NoError(t, err)

	return keyPath, privKey
}

func TestAsymmetricEncryptTransport_Success(t *testing.T) {
	// 1. Создаем тестовые ключи (для шифрования нужен публичный, приватный сохраняем для проверки)
	keyPath, privKey := generateTestPublicKeyFile(t)

	// Переменные для перехвата данных внутри цепочки транспорта
	var interceptedBytes []byte
	var interceptedContentLength int64

	// 2. Описываем мок-транспорт, который примет запрос ПОСЛЕ нашего шифровальщика
	mockNext := RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		var err error
		interceptedBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		interceptedContentLength = req.ContentLength

		// Возвращаем пустой успешный ответ
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
		}, nil
	})

	// 3. Инициализируем наш кастомный транспорт
	transport, err := NewAsymmetricEncryptTransport(keyPath, mockNext)
	assert.NoError(t, err)

	// 4. Формируем исходный запрос с открытым текстом
	plainText := []byte("metrics-payload-data")
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/updates/", bytes.NewReader(plainText))
	assert.NoError(t, err)

	// 5. Выполняем круг (RoundTrip)
	resp, err := transport.RoundTrip(req)
	assert.NoError(t, err)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close() // Исправлено: закрываем тело ответа
	}

	// 6. ПРОВЕРКИ:
	assert.NotEqual(t, plainText, interceptedBytes)
	assert.Equal(t, int64(len(interceptedBytes)), interceptedContentLength)

	// Расшифровываем приватным ключом то, что улетело в сеть
	decryptedBytes, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, interceptedBytes, nil)
	assert.NoError(t, err)
	assert.Equal(t, plainText, decryptedBytes)
}

func TestAsymmetricEncryptTransport_BypassIfNoKey(t *testing.T) {
	mockCalled := false

	mockNext := RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		mockCalled = true
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
		}, nil
	})

	// Если путь к ключу пустой, должен вернуться оригинальный транспорт
	transport, err := NewAsymmetricEncryptTransport("", mockNext)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/ping", nil)
	resp, err := transport.RoundTrip(req)
	assert.NoError(t, err)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, mockCalled, "Оригинальный mockNext транспорт должен быть вызван напрямую")
}

func TestAsymmetricEncryptTransport_NilOrEmptyBody(t *testing.T) {
	keyPath, _ := generateTestPublicKeyFile(t)

	mockNext := RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
		}, nil
	})

	transport, err := NewAsymmetricEncryptTransport(keyPath, mockNext)
	assert.NoError(t, err)

	// Сценарий 1: req.Body == nil (например, GET запрос)
	reqNil, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/ping", nil)
	respNil, err := transport.RoundTrip(reqNil)
	assert.NoError(t, err)
	if respNil != nil && respNil.Body != nil {
		defer respNil.Body.Close()
	}

	// Сценарий 2: req.Body != nil, но длина сообщения 0 байт
	reqEmpty, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/updates/", bytes.NewReader([]byte{}))
	respEmpty, err := transport.RoundTrip(reqEmpty)
	assert.NoError(t, err)
	if respEmpty != nil && respEmpty.Body != nil {
		defer respEmpty.Body.Close()
	}
}

func TestNewAsymmetricEncryptTransport_Errors(t *testing.T) {
	mockNext := RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
		}, nil
	})

	// Сценарий 1: Файл ключа отсутствует
	_, err := NewAsymmetricEncryptTransport("non_existent_public.pem", mockNext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read public key file")

	// Сценарий 2: Невалидный формат PEM-блока (подсовываем мусор)
	tmpDir := t.TempDir()
	badKeyPath := filepath.Join(tmpDir, "invalid.pem")
	err = os.WriteFile(badKeyPath, []byte("NOT A VALID PEM BLOCK"), 0600)
	assert.NoError(t, err)

	_, err = NewAsymmetricEncryptTransport(badKeyPath, mockNext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key format")
}
