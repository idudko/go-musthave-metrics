package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/idudko/go-musthave-metrics/internal/audit"
)

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

type AuditContext struct {
	Metrics []string
}

func NewAuditContext() *AuditContext {
	return &AuditContext{
		Metrics: make([]string, 0),
	}
}

func (c *AuditContext) AddMetric(name string) {
	c.Metrics = append(c.Metrics, name)
}

type auditContextKey struct{}

func WithAuditContext(ctx context.Context, auditCtx *AuditContext) context.Context {
	return context.WithValue(ctx, auditContextKey{}, auditCtx)
}

func GetAuditContext(ctx context.Context) *AuditContext {
	if auditCtx := ctx.Value(auditContextKey{}); auditCtx != nil {
		if ctx, ok := auditCtx.(*AuditContext); ok {
			return ctx
		}
	}
	return nil
}
