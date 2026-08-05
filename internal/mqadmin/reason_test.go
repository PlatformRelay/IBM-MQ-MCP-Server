package mqadmin_test

import (
	"errors"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func TestMapReasonCodeKnown(t *testing.T) {
	err := mqadmin.MapReasonCode(2035)
	var re *mqadmin.ReasonError
	if !errors.As(err, &re) {
		t.Fatalf("expected ReasonError, got %T", err)
	}
	if re.Symbol != "MQRC_NOT_AUTHORIZED" {
		t.Fatalf("symbol = %q", re.Symbol)
	}
	if re.Action == "" {
		t.Fatal("expected actionable guidance")
	}
}

func TestMapReasonCodeUnknown(t *testing.T) {
	err := mqadmin.MapReasonCode(99999)
	var re *mqadmin.ReasonError
	if !errors.As(err, &re) {
		t.Fatalf("expected ReasonError, got %T", err)
	}
	if re.Symbol != "MQRC_UNKNOWN" {
		t.Fatalf("symbol = %q", re.Symbol)
	}
}

func TestReasonCodeFromHTTPStatus(t *testing.T) {
	err := mqadmin.ReasonCodeFromHTTPStatus(403)
	var re *mqadmin.ReasonError
	if !errors.As(err, &re) || re.Code != 2035 {
		t.Fatalf("unexpected error: %v", err)
	}
}
