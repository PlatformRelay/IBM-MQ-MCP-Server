package mqadmin_test

import (
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func TestExplainReasonCodeKnown(t *testing.T) {
	ex := mqadmin.ExplainReasonCode(2035)
	if !ex.Known {
		t.Fatal("expected known entry for 2035")
	}
	if ex.Symbol != "MQRC_NOT_AUTHORIZED" {
		t.Fatalf("symbol = %q", ex.Symbol)
	}
	if ex.Summary == "" || ex.Action == "" {
		t.Fatalf("expected summary and action: %+v", ex)
	}
	if ex.DocURL == "" || !strings.HasPrefix(ex.DocURL, "https://www.ibm.com/docs/") {
		t.Fatalf("docUrl = %q", ex.DocURL)
	}
}

func TestExplainReasonCodeUnknown(t *testing.T) {
	ex := mqadmin.ExplainReasonCode(99999)
	if ex.Known {
		t.Fatal("expected unknown fallback")
	}
	if ex.Code != 99999 {
		t.Fatalf("code = %d", ex.Code)
	}
	if ex.Symbol != "MQRC_UNKNOWN" {
		t.Fatalf("symbol = %q", ex.Symbol)
	}
	if ex.Summary == "" {
		t.Fatal("expected generic summary")
	}
	if ex.DocURL == "" {
		t.Fatal("expected IBM reason-code documentation URL")
	}
}

func TestExplainReasonCodeDoesNotCopyIBMProse(t *testing.T) {
	ex := mqadmin.ExplainReasonCode(2009)
	if strings.Contains(strings.ToLower(ex.Summary), "ibm") {
		t.Fatalf("summary should be original prose, got %q", ex.Summary)
	}
}
