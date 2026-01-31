package middleware

import (
	"compress/gzip"
	"net/http"
)

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
