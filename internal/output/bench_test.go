package output_test

import (
	"encoding/json"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/output"
)

func TestBenchmarkFixturesRepresentative(t *testing.T) {
	f := output.BenchmarkFixtures
	if len(f.QueuePage.Items) < 8 {
		t.Fatalf("queue fixture too small: %d items", len(f.QueuePage.Items))
	}
	if len(f.MessagePage.Items) < 2 {
		t.Fatal("message fixture too small")
	}
	if len(f.ProfilePage.Items) < 2 {
		t.Fatal("profile fixture too small")
	}
}

func TestCompactTextShorterThanPrettyJSONForQueues(t *testing.T) {
	f := output.BenchmarkFixtures.QueuePage
	compact := output.RenderQueuePage(f)
	pretty, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if len(compact) >= len(pretty) {
		t.Fatalf("compact text should be shorter than pretty JSON: compact=%d pretty=%d", len(compact), len(pretty))
	}
}

func TestMarkdownTableShorterThanPrettyJSONForQueues(t *testing.T) {
	f := output.BenchmarkFixtures.QueuePage
	markdown := output.RenderMarkdownQueueTable(f)
	pretty, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if len(markdown) >= len(pretty) {
		t.Fatalf("markdown should beat pretty JSON: md=%d pretty=%d", len(markdown), len(pretty))
	}
}
