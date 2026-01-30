package middleware

import (
	"compress/gzip"
	"net/http"
)

// GzipRequestMiddleware is an HTTP middleware that handles gzip-compressed request bodies.
//
// This middleware checks if the request body is gzip-compressed by examining
// the "Content-Encoding" header. If compressed, it decompresses the body
// before passing it to the next handler.
//
// Supported Content Types:
//   - application/json
//   - text/html
//
// Behavior:
//   - Checks for "Content-Encoding: gzip" header
//   - Validates that Content-Type is supported (application/json or text/html)
//   - Decompresses the request body using gzip.NewReader
//   - Removes the Content-Encoding header after decompression
//   - Returns 400 Bad Request if decompression fails or content type is unsupported
//
// Example:
//
//	r := chi.NewRouter()
//	r.Use(middleware.GzipRequestMiddleware)
//	r.Post("/update/", handler)
func GzipRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			ct := r.Header.Get("Content-Type")
			if ct == "application/json" || ct == "text/html" {
				g, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Failed to read gzip body", http.StatusBadRequest)
					return
				}
				defer g.Close()
				r.Body = g
				r.Header.Del("Content-Encoding")
			} else {
				http.Error(w, "Unsupported content type", http.StatusBadRequest)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
