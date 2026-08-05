package audit

import "context"

type ctxKey int

const (
	correlationIDKey ctxKey = iota
	clientSessionKey
)

// WithCorrelationID stores a correlation identifier on ctx.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, correlationIDKey, id)
}

// CorrelationIDFrom returns the correlation identifier from ctx, if present.
func CorrelationIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(correlationIDKey).(string)
	return id
}

// WithClientSession stores the MCP client/session identity on ctx.
func WithClientSession(ctx context.Context, session string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, clientSessionKey, session)
}

// ClientSessionFrom returns the MCP client/session identity from ctx, if present.
func ClientSessionFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	session, _ := ctx.Value(clientSessionKey).(string)
	return session
}

// EnsureCorrelationID returns ctx with a generated correlation id when absent.
func EnsureCorrelationID(ctx context.Context) context.Context {
	if CorrelationIDFrom(ctx) != "" {
		return ctx
	}
	return WithCorrelationID(ctx, NewCorrelationID())
}
