package messaging_test

import (
	"errors"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
)

func TestNewPartialConsumeErrorReturnsNilWithoutItems(t *testing.T) {
	err := errors.New("boom")
	if got := messaging.NewPartialConsumeError(collection.Page[messaging.MessageRecord]{}, err); got != err {
		t.Fatalf("got %v want underlying err", got)
	}
}

func TestPartialConsumeErrorWrapsPageAndCause(t *testing.T) {
	page := collection.Page[messaging.MessageRecord]{
		Items: []messaging.MessageRecord{{MessageID: "ID:1"}},
	}
	cause := errors.New("status 500")
	err := messaging.NewPartialConsumeError(page, cause)
	var partial *messaging.PartialConsumeError
	if !errors.As(err, &partial) {
		t.Fatalf("expected PartialConsumeError, got %T", err)
	}
	if len(partial.Page.Items) != 1 {
		t.Fatalf("items = %d", len(partial.Page.Items))
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected unwrap to cause")
	}
}
