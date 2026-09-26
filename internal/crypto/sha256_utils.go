package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func DecodeKeyString(cryptoKeyString string) ([]byte, error) {
	keyBytes := make([]byte,0)
	key := sha256.Sum256([]byte(cryptoKeyString))	
	keyBytes = append(keyBytes, key[:]...)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("Invalid key length: got %d bytes, want 32 bytes", len(keyBytes))
	}
	return keyBytes, nil
}

func CalculateSha256(cryptoKey []byte, data []byte) string {
	cryptoData := append(cryptoKey, data...)
	hash := sha256.Sum256(cryptoData)
	return hex.EncodeToString(hash[:])
}
