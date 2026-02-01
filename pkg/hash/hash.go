package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// ComputeHash calculates the HMAC-SHA256 hash of the provided data using the given key.
//
// This function computes a hash-based message authentication code (HMAC) using
// the SHA-256 algorithm. HMAC provides both data integrity and authenticity verification.
//
// Parameters:
//   - data: Byte slice containing the data to be hashed
//   - key: Secret key used for HMAC computation (empty string returns empty hash)
//
// Returns:
//   - string: Hexadecimal representation of the HMAC-SHA256 hash (64 characters)
//   - Empty string if key is empty
//
// Use Cases:
//   - Signing HTTP request bodies for integrity verification
//   - Signing HTTP response bodies for authenticity
//   - Generating secure checksums for data
//
// Example:
//
//	data := []byte(`{"id": "metric", "type": "gauge", "value": 75.5}`)
//	key := "my-secret-key"
//	hash := ComputeHash(data, key)
//	fmt.Println(hash) // Output: abc123... (64 hex characters)
func ComputeHash(data []byte, key string) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateHash verifies that the received HMAC-SHA256 hash matches the computed hash of the data.
//
// This function is used to verify data integrity and authenticity by comparing
// a received hash with a newly computed hash of the same data using the shared secret key.
//
// Parameters:
//   - data: Byte slice containing the original data
//   - key: Secret key used for HMAC computation (empty string always returns true)
//   - receivedHash: Hexadecimal string containing the hash to verify (empty returns false)
//
// Returns:
//   - bool: true if hashes match (or validation is skipped), false otherwise
//
// Validation Rules:
//   - If key is empty: Returns true (validation disabled)
//   - If receivedHash is empty: Returns false (no hash provided to verify)
//   - Otherwise: Compares received hash with computed hash using constant-time comparison
//
// Security:
//   - Uses hmac.Equal for constant-time comparison to prevent timing attacks
//
// Example:
//
//	data := []byte(`{"id": "metric", "type": "gauge", "value": 75.5}`)
//	key := "my-secret-key"
//	receivedHash := ComputeHash(data, key) // In real scenario, this comes from client
//
//	if ValidateHash(data, key, receivedHash) {
//	    fmt.Println("Hash is valid - data is authentic")
//	} else {
//	    fmt.Println("Hash is invalid - data may be tampered")
//	}
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
