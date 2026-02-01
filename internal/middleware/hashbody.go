package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/idudko/go-musthave-metrics/pkg/hash"
)

// HashValidationMiddleware creates a middleware that validates HMAC-SHA256 signatures on request bodies.
//
// This middleware verifies that request bodies have been signed with a shared secret key
// to ensure data integrity and authenticity. It compares the received hash from the
// "HashSHA256" header with a computed hash of the request body.
//
// Parameters:
//   - key: Secret key used for HMAC-SHA256 signature generation (empty string to disable validation)
//
// Returns:
//   - func(http.Handler) http.Handler: Middleware function for use with HTTP router
//
// Behavior:
//   - If key is empty: Skips validation and passes request through
//   - If "HashSHA256" header is missing or "none": Skips validation
//   - If header exists: Reads request body, validates hash signature
//   - On validation failure: Returns 400 Bad Request with "Invalid hash signature" error
//   - On success: Passes request to next handler
//
// HTTP Headers:
//   - HashSHA256: Expected HMAC-SHA256 hash of the request body (hexadecimal string)
//
// Response Codes:
//   - Next handler continues on valid hash
//   - 400 Bad Request: Failed to read body or invalid hash signature
//
// Example:
//
//	// Create middleware with secret key
//	middleware := HashValidationMiddleware("my-secret-key")
//	r.Use(middleware)
//
//	// Client should send:
//	POST /update/
//	Content-Type: application/json
//	HashSHA256: abc123...
//
//	{"id": "metric", "type": "gauge", "value": 75.5}
func HashValidationMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			receivedHash := r.Header.Get("HashSHA256")

			if receivedHash == "" || receivedHash == "none" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))

			if !hash.ValidateHash(body, key, receivedHash) {
				log.Printf("Invalid hash signature")
				http.Error(w, "Invalid hash signature", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
