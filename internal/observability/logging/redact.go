// Package logging provides structured JSON logs with central redaction.
package logging

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"unicode"
)

const (
	// RedactedValue replaces secret-like attribute values in logs.
	RedactedValue = "[REDACTED]"
)

var secretKeys = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"apikey":        {},
	"api_key":       {},
	"secret":        {},
	"credential":    {},
	"credentials":   {},
	"authorization": {},
	"bearer":        {},
}

// NewJSONHandler returns a JSON slog handler that redacts secrets and sanitizes values.
func NewJSONHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	base := slog.NewJSONHandler(w, opts)
	return &redactingHandler{next: base}
}

type redactingHandler struct {
	next slog.Handler
}

func (h *redactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *redactingHandler) Handle(ctx context.Context, record slog.Record) error {
	attrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, sanitizeAttr(a))
		return true
	})
	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	newRecord.AddAttrs(attrs...)
	return h.next.Handle(ctx, newRecord)
}

func (h *redactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	sanitized := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		sanitized[i] = sanitizeAttr(a)
	}
	return &redactingHandler{next: h.next.WithAttrs(sanitized)}
}

func (h *redactingHandler) WithGroup(name string) slog.Handler {
	return &redactingHandler{next: h.next.WithGroup(sanitizeString(name))}
}

func sanitizeAttr(a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)
	if _, secret := secretKeys[key]; secret {
		return slog.String(a.Key, RedactedValue)
	}
	switch a.Value.Kind() {
	case slog.KindString:
		return slog.String(a.Key, sanitizeString(a.Value.String()))
	case slog.KindGroup:
		attrs := a.Value.Group()
		sanitized := make([]any, 0, len(attrs)*2)
		for _, ga := range attrs {
			s := sanitizeAttr(ga)
			sanitized = append(sanitized, s.Key, s.Value)
		}
		return slog.Group(a.Key, sanitized...)
	default:
		return a
	}
}

func sanitizeString(value string) string {
	if value == "" {
		return value
	}
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if r == '\n' || r == '\r' || unicode.IsControl(r) {
			b.WriteRune(' ')
			continue
		}
		b.WriteRune(r)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
