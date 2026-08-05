package mqadmin_test

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func TestClassifyConnectivityErrorDNS(t *testing.T) {
	cause, _ := mqadmin.ClassifyConnectivityError(&net.DNSError{IsNotFound: true, Name: "mq.example.test"})
	if cause != mqadmin.FailureDNS {
		t.Fatalf("cause = %q", cause)
	}
}

func TestClassifyConnectivityErrorTLS(t *testing.T) {
	cause, _ := mqadmin.ClassifyConnectivityError(
		fmt.Errorf("tls handshake failed: certificate signed by unknown authority"),
	)
	if cause != mqadmin.FailureTLS {
		t.Fatalf("cause = %q", cause)
	}
}

func TestClassifyConnectivityErrorAuthentication(t *testing.T) {
	cause, _ := mqadmin.ClassifyConnectivityError(mqadmin.MapReasonCode(2035))
	if cause != mqadmin.FailureAuthorization {
		t.Fatalf("2035 should map to authorization, got %q", cause)
	}
	err := mqadmin.ReasonCodeFromHTTPStatus(401)
	cause, _ = mqadmin.ClassifyConnectivityError(err)
	if cause != mqadmin.FailureAuthentication {
		t.Fatalf("401 should map to authentication, got %q", cause)
	}
}

func TestClassifyConnectivityErrorTimeout(t *testing.T) {
	cause, _ := mqadmin.ClassifyConnectivityError(context.DeadlineExceeded)
	if cause != mqadmin.FailureTimeout {
		t.Fatalf("cause = %q", cause)
	}
	var netErr net.Error = timeoutError{}
	cause, _ = mqadmin.ClassifyConnectivityError(netErr)
	if cause != mqadmin.FailureTimeout {
		t.Fatalf("cause = %q", cause)
	}
}

func TestSanitizeConnectivityDetailRedactsCredentials(t *testing.T) {
	detail := mqadmin.SanitizeConnectivityDetail("basic auth failed for user:secret@mq.example.test:9443")
	if strings.Contains(detail, "secret") {
		t.Fatalf("detail leaked credential: %q", detail)
	}
}

func TestSanitizeConnectivityDetailRedactsTLSMaterial(t *testing.T) {
	detail := mqadmin.SanitizeConnectivityDetail("tls: x509: certificate has expired")
	if detail == "" {
		t.Fatal("expected sanitized detail")
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return false }

func TestClassifyConnectivityErrorReasonAuthorization(t *testing.T) {
	err := mqadmin.ReasonCodeFromHTTPStatus(403)
	cause, _ := mqadmin.ClassifyConnectivityError(err)
	if cause != mqadmin.FailureAuthorization {
		t.Fatalf("403 should map to authorization, got %q", cause)
	}
}

func TestBuildConnectivityReportRedactsEndpointCredentials(t *testing.T) {
	start := time.Now().Add(-5 * time.Millisecond)
	report := mqadmin.BuildConnectivityReport(
		"prod",
		"https://mquser:mqsecret@mq.example.test:9443",
		"QM1",
		start,
		mqadmin.QueueManagerStatus{},
		fmt.Errorf("resolve basic credentials: missing env"),
	)
	if strings.Contains(report.Endpoint, "mqsecret") || strings.Contains(report.Endpoint, "mquser") {
		t.Fatalf("endpoint leaked credentials: %q", report.Endpoint)
	}
	if strings.Contains(report.Detail, "mqsecret") {
		t.Fatalf("detail leaked credentials: %q", report.Detail)
	}
}

func TestRedactEndpointCredentials(t *testing.T) {
	redacted := mqadmin.RedactEndpointCredentials("https://user:pass@mq.example.test:9443/path")
	if strings.Contains(redacted, "pass") || strings.Contains(redacted, "user@") {
		t.Fatalf("redacted = %q", redacted)
	}
}
