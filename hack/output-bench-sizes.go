// Command output-bench-sizes prints byte sizes for output benchmark docs.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/output"
)

func main() {
	f := output.BenchmarkFixtures
	compactQ := output.RenderQueuePage(f.QueuePage)
	mdQ := output.RenderMarkdownQueueTable(f.QueuePage)
	minQ, _ := json.Marshal(f.QueuePage)
	prettyQ, _ := json.MarshalIndent(f.QueuePage, "", "  ")

	compactM := output.RenderMessagePage(f.MessagePage)
	mdM := output.RenderMarkdownMessageTable(f.MessagePage)
	minM, _ := json.Marshal(f.MessagePage)
	prettyM, _ := json.MarshalIndent(f.MessagePage, "", "  ")

	fmt.Printf("queue_list compact=%d minified_json=%d pretty_json=%d markdown=%d\n",
		len(compactQ), len(minQ), len(prettyQ), len(mdQ))
	fmt.Printf("message_list compact=%d minified_json=%d pretty_json=%d markdown=%d\n",
		len(compactM), len(minM), len(prettyM), len(mdM))
	fmt.Printf("qm_status compact=%d\n", len(output.RenderQueueManagerStatus(f.QMStatus)))
	fmt.Printf("connectivity compact=%d\n", len(output.RenderConnectivityReport(f.Connectivity)))
	fmt.Printf("reason compact=%d\n", len(output.RenderReasonExplanation(f.Reason)))
}
