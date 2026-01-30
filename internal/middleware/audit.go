package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/idudko/go-musthave-metrics/internal/audit"
)

// AuditMiddleware creates a middleware that tracks metrics changes and creates audit events.
//
// This middleware monitors HTTP requests and records which metrics are being modified.
// It creates audit events when metrics are updated, allowing for tracking and logging
// of metric changes across the application.
//
// Parameters:
//   - auditSubject: Subject that will be notified of audit events (nil to disable auditing)
//
// Returns:
//   - func(http.Handler) http.Handler: Middleware function for use with HTTP router
//
// Behavior:
//   - Creates a new audit context for each request
//   - Attaches audit context to request context
//   - Executes the request handler
//   - After handler completes, checks if any metrics were modified
//   - Creates audit event with request details and modified metrics
//   - Notifies audit subject if metrics were modified
//
// Use Cases:
//   - Tracking metric updates for security purposes
//   - Debugging metric modification patterns
//   - Creating audit logs for compliance
//
// Example:
//
//	auditSubject := audit.NewSubject()
//	middleware := AuditMiddleware(auditSubject)
//	r.Use(middleware)
//
//	// Handlers can add metrics to audit context:
//	// auditCtx := middleware.GetAuditContext(r.Context())
//	// auditCtx.AddMetric("cpu_usage")
func AuditMiddleware(auditSubject *audit.Subject) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			ctx := WithAuditContext(r.Context(), NewAuditContext())
			r = r.WithContext(ctx)

			next.ServeHTTP(wrapped, r)

			auditCtx := GetAuditContext(r.Context())

			if auditCtx != nil && len(auditCtx.Metrics) > 0 && auditSubject != nil {
				event := audit.CreateAuditEvent(r, auditCtx.Metrics)
				auditSubject.NotifyAll(event)
			}
		}
		return http.HandlerFunc(fn)
	}
}

// AuditContext stores information about metrics modified during a request.
//
// This context is used to track which metrics were updated, deleted, or accessed
// during a single HTTP request. The collected metrics are then used to create
// audit events for logging and monitoring purposes.
//
// Thread Safety:
//
//	This struct is NOT safe for concurrent use. A new instance should be
//	created for each request.
type AuditContext struct {
	Metrics []string
}

// NewAuditContext creates a new empty AuditContext instance.
//
// Returns:
//   - *AuditContext: Context with empty metrics slice ready for use
//
// Example:
//
//	auditCtx := NewAuditContext()
//	auditCtx.AddMetric("cpu_usage")
//	auditCtx.AddMetric("memory_usage")
func NewAuditContext() *AuditContext {
	return &AuditContext{
		Metrics: make([]string, 0),
	}
}

// AddMetric adds a metric name to the audit context.
//
// This method is typically called by handlers when they modify or access a metric.
// The collected metric names will be included in the audit event after the request completes.
//
// Parameters:
//   - name: Unique identifier of the metric being modified/accessed
//
// Example:
//
//	auditCtx := GetAuditContext(r.Context())
//	if auditCtx != nil {
//	    auditCtx.AddMetric("cpu_usage")
//	    auditCtx.AddMetric("requests_total")
//	}
func (c *AuditContext) AddMetric(name string) {
	c.Metrics = append(c.Metrics, name)
}

// auditContextKey is a private type used as a key for storing AuditContext in request context.
//
// Using a private type prevents collisions with other context values.
type auditContextKey struct{}

// WithAuditContext stores an AuditContext in the request context.
//
// Parameters:
//   - ctx: The request context to modify
//   - auditCtx: The AuditContext to store (can be nil)
//
// Returns:
//   - context.Context: New context with AuditContext attached
//
// Example:
//
//	auditCtx := NewAuditContext()
//	ctx := WithAuditContext(r.Context(), auditCtx)
//	r = r.WithContext(ctx)
func WithAuditContext(ctx context.Context, auditCtx *AuditContext) context.Context {
	return context.WithValue(ctx, auditContextKey{}, auditCtx)
}

// GetAuditContext retrieves the AuditContext from the request context.
//
// This method safely retrieves the audit context, returning nil if none exists
// or if the stored value is not of the correct type.
//
// Parameters:
//   - ctx: The request context to search
//
// Returns:
//   - *AuditContext: The stored audit context, or nil if not found
//
// Example:
//
//	auditCtx := GetAuditContext(r.Context())
//	if auditCtx != nil {
//	    auditCtx.AddMetric("metric_name")
//	}
func GetAuditContext(ctx context.Context) *AuditContext {
	if auditCtx := ctx.Value(auditContextKey{}); auditCtx != nil {
		if ctx, ok := auditCtx.(*AuditContext); ok {
			return ctx
		}
	}
	return nil
}
