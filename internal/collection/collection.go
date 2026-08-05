// Package collection defines the shared bounded-result contract for inspection tools.
package collection

import "fmt"

const (
	// DefaultLimit is applied when callers omit or pass a non-positive limit.
	DefaultLimit = 50
	// MaxLimit caps every collection request; results are never unbounded.
	MaxLimit = 200
)

// TruncationReason explains why Truncated is true.
type TruncationReason string

const (
	// TruncationLimitReached means more items matched than returned under limit.
	TruncationLimitReached TruncationReason = "limit_reached"
	// TruncationBackendCap means the downstream API returned a partial page.
	TruncationBackendCap TruncationReason = "backend_cap"
	// TruncationMidBatchFailure means consume removed some messages then failed.
	TruncationMidBatchFailure TruncationReason = "mid_batch_failure"
)

// Page is the shared envelope for list-style tool results (ADR-0005).
type Page[T any] struct {
	Items            []T              `json:"items"`
	Limit            int              `json:"limit"`
	Cursor           string           `json:"cursor,omitempty"`
	NextCursor       string           `json:"nextCursor,omitempty"`
	Truncated        bool             `json:"truncated"`
	TruncationReason TruncationReason `json:"truncationReason,omitempty"`
}

// NormalizeLimit clamps requested to [1, MaxLimit], defaulting to DefaultLimit.
func NormalizeLimit(requested int) int {
	if requested <= 0 {
		return DefaultLimit
	}
	if requested > MaxLimit {
		return MaxLimit
	}
	return requested
}

// ValidateLimit returns an error when requested exceeds MaxLimit without clamping.
func ValidateLimit(requested int) error {
	if requested > MaxLimit {
		return fmt.Errorf("limit %d exceeds maximum %d", requested, MaxLimit)
	}
	return nil
}
