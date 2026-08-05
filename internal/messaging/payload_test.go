package messaging_test

import (
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
)

func TestNormalizeBrowseCountDefaults(t *testing.T) {
	if got := messaging.NormalizeBrowseCount(0); got != messaging.DefaultBrowseCount {
		t.Fatalf("got %d want %d", got, messaging.DefaultBrowseCount)
	}
}

func TestNormalizeBrowseCountClampsMax(t *testing.T) {
	if got := messaging.NormalizeBrowseCount(500); got != messaging.MaxBrowseCount {
		t.Fatalf("got %d want %d", got, messaging.MaxBrowseCount)
	}
}

func TestValidateBrowseCountRejectsOverMax(t *testing.T) {
	if err := messaging.ValidateBrowseCount(messaging.MaxBrowseCount + 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeMaxPayloadBytesDefaults(t *testing.T) {
	if got := messaging.NormalizeMaxPayloadBytes(0); got != messaging.DefaultMaxPayloadBytes {
		t.Fatalf("got %d want %d", got, messaging.DefaultMaxPayloadBytes)
	}
}

func TestValidateMaxPayloadBytesRejectsOverHardMax(t *testing.T) {
	if err := messaging.ValidateMaxPayloadBytes(messaging.HardMaxPayloadBytes + 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestRedactSecretsPatterns(t *testing.T) {
	in := "user password=sekret Authorization: Bearer abc.def api_key=xyz"
	out := messaging.RedactSecrets(in)
	for _, forbidden := range []string{"sekret", "Bearer abc", "xyz"} {
		if contains(out, forbidden) {
			t.Fatalf("redacted output still contains %q: %q", forbidden, out)
		}
	}
}

func TestFormatPayloadUTF8(t *testing.T) {
	text, enc, truncated := messaging.FormatPayload([]byte("hello"), 4096)
	if enc != messaging.EncodingUTF8 || text != "hello" || truncated {
		t.Fatalf("text=%q enc=%q truncated=%v", text, enc, truncated)
	}
}

func TestFormatPayloadBinaryBase64(t *testing.T) {
	raw := []byte{0x00, 0x01, 0x02, 0xff}
	_, enc, _ := messaging.FormatPayload(raw, 4096)
	if enc != messaging.EncodingBase64 {
		t.Fatalf("encoding = %q", enc)
	}
}

func TestFormatPayloadTruncates(t *testing.T) {
	raw := []byte("0123456789")
	_, _, truncated := messaging.FormatPayload(raw, 5)
	if !truncated {
		t.Fatal("expected truncation")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexSubstring(s, sub))
}

func indexSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
