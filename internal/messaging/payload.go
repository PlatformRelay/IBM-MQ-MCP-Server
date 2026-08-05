package messaging

import (
	"encoding/base64"
	"regexp"
	"strings"
)

const redactedPlaceholder = "[REDACTED]"

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)password\s*=\s*\S+`),
	regexp.MustCompile(`(?i)authorization\s*:\s*\S+`),
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-._~+/]+=*`),
	regexp.MustCompile(`(?i)api[_-]?key\s*[=:]\s*\S+`),
}

// RedactSecrets replaces secret-like substrings in text before tool results.
func RedactSecrets(text string) string {
	if text == "" {
		return text
	}
	out := text
	for _, pattern := range secretPatterns {
		out = pattern.ReplaceAllString(out, redactedPlaceholder)
	}
	return out
}

// FormatPayload maps raw bytes to display text, encoding label, and truncation flag.
func FormatPayload(raw []byte, maxBytes int) (text string, encoding PayloadEncoding, truncated bool) {
	if len(raw) == 0 {
		return "", EncodingOmitted, false
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxPayloadBytes
	}
	if len(raw) > maxBytes {
		raw = raw[:maxBytes]
		truncated = true
	}
	if isValidUTF8(raw) {
		return RedactSecrets(string(raw)), EncodingUTF8, truncated
	}
	return RedactSecrets(base64.StdEncoding.EncodeToString(raw)), EncodingBase64, truncated
}

func isValidUTF8(b []byte) bool {
	return strings.ToValidUTF8(string(b), "\uFFFD") == string(b)
}
