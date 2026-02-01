package handler

import (
	"context"
	"log"
	"net/http"
	"time"
)

// DBPinger defines the interface for database connectivity health checks.
//
// Implementations of this interface should provide a mechanism to verify
// that the database connection is active and can respond to queries.
//
// Use Cases:
//   - Health check endpoints
//   - Connection pool monitoring
//   - Database readiness checks during application startup
type DBPinger interface {
	// Ping verifies the database connection is healthy and responsive.
	//
	// Parameters:
	//   - ctx: Context for request cancellation and timeout handling
	//
	// Returns:
	//   - error: nil if database is healthy, error otherwise
	//
	// Example:
	//   err := pinger.Ping(ctx)
	//   if err != nil {
	//       log.Printf("Database check failed: %v", err)
	//       return err
	//   }
	//   log.Println("Database connection is healthy")
	Ping(ctx context.Context) error
}

// PingHandler provides an HTTP handler for database health check endpoints.
//
// This handler responds to GET requests and indicates whether the
// database connection is healthy (200 OK) or has issues (500 Internal Server Error).
//
// Typical Usage:
//
//	r.Get("/ping", pingHandler.PingHandler)
//
// Response Codes:
//   - 200 OK: Database is accessible and responding
//   - 500 Internal Server Error: Database is not initialized or connection check failed
//
// Thread Safety:
//
//	The handler is safe for concurrent requests as it only calls
//	the pinger interface method without maintaining any state.
type PingHandler struct {
	pinger DBPinger // Database pinger instance for health checks (can be nil if DB not configured)
}

// NewPingHandler creates a new PingHandler instance with the provided database pinger.
//
// Parameters:
//   - pinger: Database pinger implementation (can be nil if no database is configured)
//
// Returns:
//   - *PingHandler: Configured handler instance ready for use with HTTP router
//
// Example:
//
//	// With database storage
//	dbStorage := repository.NewDBStorage("postgres://...")
//	pingHandler := handler.NewPingHandler(dbStorage)
//	r.Get("/ping", pingHandler.PingHandler)
//
//	// With in-memory storage (no DB ping)
//	memStorage := repository.NewMemStorage()
//	pingHandler := handler.NewPingHandler(nil)
//	// Handler will return 200 OK without actual DB check
func NewPingHandler(pinger DBPinger) *PingHandler {
	return &PingHandler{pinger: pinger}
}

// PingHandler handles GET requests to check database connectivity health.
//
// Endpoint: GET /ping
//
// Behavior:
//   - If pinger is nil: Returns 200 OK (database not configured/initialized)
//   - If pinger is not nil: Calls Ping() with 1-second timeout
//   - On success: Returns 200 OK (database is healthy)
//   - On failure: Returns 500 Internal Server Error with error message
//
// Timeout:
//   - A 1-second timeout is enforced on the database ping operation
//   - This prevents the handler from hanging on slow/unresponsive databases
//
// Response:
//   - Body: Empty (status code indicates health)
//   - Headers: None
//
// Example curl command:
//
//	curl http://localhost:8080/ping
//	HTTP/1.1 200 OK (on success)
//	HTTP/1.1 500 Internal Server Error (on failure)
func (h *PingHandler) PingHandler(w http.ResponseWriter, r *http.Request) {
	// Check if pinger is initialized
	if h.pinger == nil {
		log.Println("Ping failed: pinger is nil")
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	// Create context with 1-second timeout for database ping
	// This prevents the handler from hanging on slow/unresponsive databases
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	// Execute database health check
	if err := h.pinger.Ping(ctx); err != nil {
		log.Printf("Ping failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Database is healthy
	w.WriteHeader(http.StatusOK)
}
