package mqadmin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// FailureCause classifies connectivity failures without leaking credentials.
type FailureCause string

// FailureCause values classify connectivity failures without leaking credentials.
const (
	FailureDNS            FailureCause = "dns"
	FailureTLS            FailureCause = "tls"
	FailureAuthentication FailureCause = "authentication"
	FailureAuthorization  FailureCause = "authorization"
	FailureTimeout        FailureCause = "timeout"
	FailureUnreachable    FailureCause = "unreachable"
)

// ConnectivityReport is the side-effect-free result of a profile connectivity check.
type ConnectivityReport struct {
	Profile       string       `json:"profile"`
	Endpoint      string       `json:"endpoint"`
	Reachable     bool         `json:"reachable"`
	Identity      Identity     `json:"identity"`
	IdentityMatch bool         `json:"identityMatch"`
	LatencyMs     int64        `json:"latencyMs"`
	FailureCause  FailureCause `json:"failureCause,omitempty"`
	Detail        string       `json:"detail,omitempty"`
	CheckedAt     time.Time    `json:"checkedAt"`
}

// ClassifyConnectivityError maps downstream errors to typed failure causes.
func ClassifyConnectivityError(err error) (FailureCause, string) {
	if err == nil {
		return "", ""
	}
	if re, ok := AsReasonError(err); ok {
		switch re.Code {
		case 2035:
			return FailureAuthorization, re.Error()
		}
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return FailureTimeout, msg
	case isTimeout(err):
		return FailureTimeout, msg
	case isDNSError(err):
		return FailureDNS, msg
	case isTLSError(err):
		return FailureTLS, msg
	case strings.Contains(lower, "resolve basic credentials"),
		strings.Contains(lower, "basic credentials"),
		strings.Contains(lower, "resolve mtls material"),
		strings.Contains(lower, "authentication type"):
		return FailureAuthentication, msg
	case strings.Contains(lower, "http 401"):
		return FailureAuthentication, msg
	case strings.Contains(lower, "http 403"):
		return FailureAuthorization, msg
	default:
		return FailureUnreachable, msg
	}
}

// SanitizeConnectivityDetail removes credential-like substrings from error text.
func SanitizeConnectivityDetail(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	out := raw
	out = userinfoURLPattern.ReplaceAllString(out, "://***@")
	out = basicSecretPattern.ReplaceAllString(out, ":***")
	out = credentialPairPattern.ReplaceAllString(out, "$1:***")
	return strings.TrimSpace(out)
}

var (
	userinfoURLPattern    = regexp.MustCompile(`://[^@\s]+@`)
	basicSecretPattern    = regexp.MustCompile(`(?i)(user(?:name)?|password|passphrase|secret)[=:]\S+`)
	credentialPairPattern = regexp.MustCompile(`(?i)([A-Za-z0-9._-]+):[^\s]+`)
)

func isDNSError(err error) bool {
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func isTLSError(err error) bool {
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "tls") ||
		strings.Contains(lower, "x509") ||
		strings.Contains(lower, "certificate")
}

// BuildConnectivityReport assembles a report from queue manager status and timing.
func BuildConnectivityReport(
	profileName, endpoint, configuredQM string,
	start time.Time,
	status QueueManagerStatus,
	err error,
) ConnectivityReport {
	report := ConnectivityReport{
		Profile:   profileName,
		Endpoint:  endpoint,
		CheckedAt: time.Now().UTC(),
		LatencyMs: time.Since(start).Milliseconds(),
		Identity: Identity{
			Configured: configuredQM,
		},
	}
	if status.Identity.Configured != "" {
		report.Identity = status.Identity
	}
	if err != nil {
		cause, detail := ClassifyConnectivityError(err)
		report.FailureCause = cause
		report.Detail = SanitizeConnectivityDetail(detail)
		return report
	}
	report.Reachable = status.Availability == Available || status.Availability == Stale
	report.IdentityMatch = strings.EqualFold(status.Identity.Observed, status.Identity.Configured) &&
		status.Identity.Observed != ""
	if status.Availability == Stale {
		report.Detail = "observed queue manager name differs from configured profile"
	}
	if status.Error != "" && !report.Reachable {
		cause, detail := ClassifyConnectivityError(fmt.Errorf("%s", status.Error))
		if report.FailureCause == "" {
			report.FailureCause = cause
		}
		if report.Detail == "" {
			report.Detail = SanitizeConnectivityDetail(detail)
		}
	}
	return report
}

// RedactEndpointCredentials removes userinfo from an endpoint URL for safe display.
func RedactEndpointCredentials(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil || u.User == nil {
		return endpoint
	}
	u.User = url.UserPassword("***", "***")
	return u.String()
}
