package mqadmin_test

import (
	"errors"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func TestUnsupportedFamilyError(t *testing.T) {
	err := mqadmin.UnsupportedFamily("listener")
	var ue *mqadmin.UnsupportedError
	if !errors.As(err, &ue) {
		t.Fatalf("expected UnsupportedError, got %T", err)
	}
	if ue.Family != "listener" || ue.Action == "" {
		t.Fatalf("unexpected error: %+v", ue)
	}
}
