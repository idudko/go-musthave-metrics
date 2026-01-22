package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func ComputeHash(data []byte, key string) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateHash(data []byte, key string, receivedHash string) bool {
	if key == "" {
		return true
	}

	if receivedHash == "" {
		return false
	}

	expectedHash := ComputeHash(data, key)
	return hmac.Equal([]byte(expectedHash), []byte(receivedHash))
}
