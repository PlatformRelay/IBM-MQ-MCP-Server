package collection_test

import (
	"encoding/json"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
)

func TestNormalizeLimitDefaults(t *testing.T) {
	if got := collection.NormalizeLimit(0); got != collection.DefaultLimit {
		t.Fatalf("NormalizeLimit(0) = %d, want %d", got, collection.DefaultLimit)
	}
	if got := collection.NormalizeLimit(-1); got != collection.DefaultLimit {
		t.Fatalf("NormalizeLimit(-1) = %d, want %d", got, collection.DefaultLimit)
	}
}

func TestNormalizeLimitClampsMax(t *testing.T) {
	if got := collection.NormalizeLimit(999); got != collection.MaxLimit {
		t.Fatalf("NormalizeLimit(999) = %d, want %d", got, collection.MaxLimit)
	}
}

func TestValidateLimitRejectsOverMax(t *testing.T) {
	if err := collection.ValidateLimit(collection.MaxLimit + 1); err == nil {
		t.Fatal("expected error for limit over max")
	}
}

func TestPageJSONEnvelope(t *testing.T) {
	page := collection.Page[string]{
		Items:            []string{"a", "b"},
		Limit:            50,
		Cursor:           "0",
		NextCursor:       "2",
		Truncated:        true,
		TruncationReason: collection.TruncationLimitReached,
	}
	data, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"items", "limit", "cursor", "nextCursor", "truncated", "truncationReason"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("missing key %q in %s", key, data)
		}
	}
}
