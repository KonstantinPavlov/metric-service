package crypto

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
)

type AsymmetricEncryptTransport struct {
	publicKey *rsa.PublicKey
	next      http.RoundTripper
}

func NewAsymmetricEncryptTransport(publicKeyPath string, next http.RoundTripper) (http.RoundTripper, error) {
	if publicKeyPath == "" {
		return next, nil
	}

	keyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid public key format (expected PUBLIC KEY)")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	if next == nil {
		next = http.DefaultTransport
	}

	return &AsymmetricEncryptTransport{
		publicKey: rsaKey,
		next:      next,
	}, nil
}

func (a *AsymmetricEncryptTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body == nil {
		return a.next.RoundTrip(req)
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body for encryption: %w", err)
	}
	req.Body.Close()

	if len(bodyBytes) == 0 {
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		return a.next.RoundTrip(req)
	}

	encryptedBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, a.publicKey, bodyBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt request body: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewReader(encryptedBytes))
	req.ContentLength = int64(len(encryptedBytes))

	return a.next.RoundTrip(req)
}
