package messaging

import (
	"fmt"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
)

// PartialConsumeError reports a mid-batch consume failure after one or more
// messages were already removed from the queue. Page holds those records;
// Err is the underlying mqweb or transport error.
type PartialConsumeError struct {
	Page collection.Page[MessageRecord]
	Err  error
}

func (e *PartialConsumeError) Error() string {
	if e == nil || e.Err == nil {
		return "partial consume failure"
	}
	n := 0
	if e.Page.Items != nil {
		n = len(e.Page.Items)
	}
	return fmt.Sprintf("consume stopped after %d message(s): %v", n, e.Err)
}

func (e *PartialConsumeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewPartialConsumeError builds a partial failure when page already has items.
func NewPartialConsumeError(page collection.Page[MessageRecord], err error) error {
	if err == nil || len(page.Items) == 0 {
		return err
	}
	page.Truncated = true
	if page.TruncationReason == "" {
		page.TruncationReason = collection.TruncationMidBatchFailure
	}
	return &PartialConsumeError{Page: page, Err: err}
}
