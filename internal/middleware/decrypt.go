package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/idudko/go-musthave-metrics/pkg/crypto"
)

// DecryptionMiddleware creates a middleware that decrypts request bodies using RSA.
//
// This middleware decrypts request bodies that were encrypted using the agent's
// public key. It uses the server's private key to decrypt the data before
// passing it to the next handler.
//
// Parameters:
//   - privateKeyPath: Path to the file containing the RSA private key (empty string to disable decryption)
//
// Returns:
//   - func(http.Handler) http.Handler: Middleware function for use with HTTP router
//
// Behavior:
//   - If privateKeyPath is empty: Skips decryption and passes request through
//   - If Content-Encoding header is "encrypt": Decrypts request body using RSA private key
//   - On decryption failure: Returns 400 Bad Request with error message
//   - On success: Replaces request body with decrypted data and passes to next handler
//
// HTTP Headers:
//   - Content-Encoding: "encrypt" indicates that request body is encrypted
//
// Response Codes:
//   - Next handler continues on successful decryption
//   - 400 Bad Request: Failed to read body or decrypt data
//
// Example:
//
//	// Create middleware with private key
//	middleware := DecryptionMiddleware("private.pem")
//	r.Use(middleware)
//
//	// Agent sends encrypted request:
//	POST /update/
//	Content-Type: application/json
//	Content-Encoding: encrypt
//
//	[encrypted bytes]
func DecryptionMiddleware(privateKeyPath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKeyPath == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if request body is encrypted
			if r.Header.Get("Content-Encoding") != "encrypt" {
				next.ServeHTTP(w, r)
				return
			}

			// Load private key
			privKey, err := crypto.LoadPrivateKey(privateKeyPath)
			if err != nil {
				log.Printf("Failed to load private key: %v", err)
				http.Error(w, "Failed to load private key", http.StatusInternalServerError)
				return
			}

			// Read encrypted request body
			encryptedBody, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Failed to read request body: %v", err)
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			// Decrypt the request body
			decryptedBody, err := crypto.Decrypt(encryptedBody, privKey)
			if err != nil {
				log.Printf("Failed to decrypt request body: %v", err)
				http.Error(w, "Failed to decrypt request body", http.StatusBadRequest)
				return
			}

			// Replace request body with decrypted data
			r.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))
			r.ContentLength = int64(len(decryptedBody))

			// Remove encryption header as content is now decrypted
			r.Header.Del("Content-Encoding")

			next.ServeHTTP(w, r)
		})
	}
}
