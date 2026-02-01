package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// LoggingMiddleware is an HTTP middleware that logs request and response information.
//
// This middleware captures and logs the following information for each request:
//   - HTTP method (GET, POST, etc.)
//   - Request URI
//   - HTTP status code
//   - Response size in bytes
//   - Request duration
//
// The middleware wraps the http.ResponseWriter to capture the status code
// and response size, then logs this information after the handler completes.
//
// The log entries are structured using zerolog for easy parsing and filtering.
//
// Example:
//
//	r := chi.NewRouter()
//	r.Use(middleware.LoggingMiddleware)
//	r.Get("/", handler)
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		log.Info().
			Str("method", r.Method).
			Str("uri", r.RequestURI).
			Int("status", lrw.statusCode).
			Int("size", lrw.size).
			Dur("duration", time.Since(start)).
			Msg("handled request")
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture response status and size.
//
// This struct is used by LoggingMiddleware to capture the HTTP status code
// and response body size that would otherwise not be accessible from middleware.
//
// Thread Safety:
//
//	This struct is NOT safe for concurrent use. A new instance should be
//	created for each request.
type loggingResponseWriter struct {
	http.ResponseWriter     // Embedded standard ResponseWriter
	statusCode          int // HTTP status code returned by the handler
	size                int // Total response size in bytes
}

// WriteHeader captures the HTTP status code and forwards it to the underlying ResponseWriter.
//
// This method ensures that the middleware can capture the actual status code
// returned by the handler, even if the handler doesn't explicitly call WriteHeader
// (in which case the default status code 200 OK is used).
//
// Parameters:
//   - code: HTTP status code to return
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write captures the response data size and forwards it to the underlying ResponseWriter.
//
// This method accumulates the total size of all writes to capture the complete
// response size for logging purposes.
//
// Parameters:
//   - b: Data to write to the response
//
// Returns:
//   - int: Number of bytes written
//   - error: Any error that occurred during writing
func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}
