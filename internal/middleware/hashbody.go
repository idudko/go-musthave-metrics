package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/idudko/go-musthave-metrics/pkg/hash"
	"github.com/rs/zerolog/log"
)

func HashValidationMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			receivedHash := r.Header.Get("HashSHA256") // HashSHA256!

			fmt.Println("RHASH =", receivedHash)

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
				//http.Error(w, "Invalid hash signature", http.StatusBadRequest)
				//return
			}

			next.ServeHTTP(w, r)
		})
	}
}
