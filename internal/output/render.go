// Package output renders compact deterministic text fallbacks for MCP tool results.
// JSON structuredContent remains canonical per ADR-0005; text blocks supplement only.
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func renderCollectionMeta(
	count, limit int,
	truncated bool,
	reason collection.TruncationReason,
	nextCursor string,
) string {
	var b strings.Builder
	fmt.Fprintf(&b, "count=%d limit=%d", count, limit)
	if truncated {
		fmt.Fprintf(&b, " truncated=%s", reason)
	}
	if nextCursor != "" {
		fmt.Fprintf(&b, " nextCursor=%s", nextCursor)
	}
	return b.String()
}

func joinPairs(pairs ...string) string {
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}

func kv(key, value string) string {
	if value == "" {
		return ""
	}
	return key + "=" + value
}

func kvInt(key string, value int) string {
	if value == 0 {
		return ""
	}
	return fmt.Sprintf("%s=%d", key, value)
}

func kvBool(key string, value bool) string {
	if !value {
		return ""
	}
	return key + "=true"
}

func kvTime(key string, value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return key + "=" + value.UTC().Format(time.RFC3339)
}

func collectionHeader[T any](page collection.Page[T]) string {
	return renderCollectionMeta(
		len(page.Items), page.Limit, page.Truncated, page.TruncationReason, page.NextCursor,
	)
}

// RenderProfilePage renders a compact profile listing.
func RenderProfilePage(page collection.Page[application.ProfileSummary]) string {
	lines := []string{collectionHeader(page)}
	for _, item := range page.Items {
		lines = append(lines, joinPairs(
			kv("name", item.Name),
			kv("qm", item.QueueManager),
			kv("endpoint", item.Endpoint),
			fmt.Sprintf("caps=%d", len(item.Capabilities)),
			fmt.Sprintf("valid=%t", item.Valid),
		))
	}
	return strings.Join(lines, "\n")
}

// RenderQueuePage renders a compact queue listing.
func RenderQueuePage(page collection.Page[mqadmin.QueueSummary]) string {
	lines := []string{collectionHeader(page)}
	for _, item := range page.Items {
		lines = append(lines, joinPairs(kv("name", item.Name), kv("type", item.Type)))
	}
	return strings.Join(lines, "\n")
}

// RenderChannelPage renders a compact channel listing.
func RenderChannelPage(page collection.Page[mqadmin.ChannelSummary]) string {
	lines := []string{collectionHeader(page)}
	for _, item := range page.Items {
		lines = append(lines, joinPairs(kv("name", item.Name), kv("type", item.Type)))
	}
	return strings.Join(lines, "\n")
}

// RenderListenerPage renders a compact listener listing.
func RenderListenerPage(page collection.Page[mqadmin.ListenerSummary]) string {
	lines := []string{collectionHeader(page)}
	for _, item := range page.Items {
		lines = append(lines, kv("name", item.Name))
	}
	return strings.Join(lines, "\n")
}

// RenderSubscriptionPage renders a compact subscription listing.
func RenderSubscriptionPage(page collection.Page[mqadmin.SubscriptionSummary]) string {
	lines := []string{collectionHeader(page)}
	for _, item := range page.Items {
		lines = append(lines, joinPairs(
			kv("name", item.Name),
			kv("id", item.ID),
			kv("topic", item.TopicString),
			kv("type", item.Type),
		))
	}
	return strings.Join(lines, "\n")
}

// RenderMessagePage renders browse/consume message pages without payload bodies.
func RenderMessagePage(page collection.Page[messaging.MessageRecord]) string {
	lines := []string{collectionHeader(page)}
	for _, item := range page.Items {
		lines = append(lines, joinPairs(
			kv("messageId", item.MessageID),
			kv("correlationId", item.CorrelationID),
			kv("format", item.Format),
			kvInt("len", item.MessageLength),
			kv("encoding", string(item.Encoding)),
			kvBool("payloadTruncated", item.PayloadTruncated),
		))
	}
	return strings.Join(lines, "\n")
}

// RenderQueueManagerStatus renders queue manager health compactly.
func RenderQueueManagerStatus(status mqadmin.QueueManagerStatus) string {
	return joinPairs(
		kv("profile", status.Profile),
		kv("availability", string(status.Availability)),
		kvBool("running", status.Running),
		kv("configured", status.Identity.Configured),
		kv("observed", status.Identity.Observed),
		kv("status", status.StatusText),
		kvTime("lastChecked", status.LastChecked),
		kv("error", status.Error),
	)
}

// RenderQueueDetail renders one queue definition and depth.
func RenderQueueDetail(detail mqadmin.QueueDetail) string {
	depth := fmt.Sprintf("depth=%d/%d", detail.CurrentDepth, detail.MaxDepth)
	return joinPairs(
		kv("name", detail.Name),
		kv("type", detail.Type),
		depth,
		kvInt("openIn", detail.OpenInputCount),
		kvInt("openOut", detail.OpenOutputCount),
		kv("inhibitGet", detail.InhibitGet),
		kv("inhibitPut", detail.InhibitPut),
	)
}

// RenderChannelDetail renders channel definition attributes.
func RenderChannelDetail(detail mqadmin.ChannelDetail) string {
	return joinPairs(
		kv("name", detail.Name),
		kv("type", detail.Type),
		kv("connection", detail.ConnectionName),
		kv("xmitQ", detail.TransmissionQueue),
		kv("desc", detail.Description),
	)
}

// RenderChannelStatus renders channel runtime status.
func RenderChannelStatus(status mqadmin.ChannelStatus) string {
	return joinPairs(
		kv("name", status.Name),
		kv("type", status.Type),
		kv("state", status.State),
		kv("availability", string(status.Availability)),
		kv("status", status.StatusText),
		kvTime("lastChecked", status.LastChecked),
		kv("error", status.Error),
	)
}

// RenderListenerDetail renders listener definition attributes.
func RenderListenerDetail(detail mqadmin.ListenerDetail) string {
	return joinPairs(
		kv("name", detail.Name),
		kvInt("port", detail.Port),
		kv("transport", detail.Transport),
		kv("desc", detail.Description),
	)
}

// RenderListenerStatus renders listener runtime status.
func RenderListenerStatus(status mqadmin.ListenerStatus) string {
	return joinPairs(
		kv("name", status.Name),
		kv("state", status.State),
		kv("availability", string(status.Availability)),
		kv("status", status.StatusText),
		kvTime("lastChecked", status.LastChecked),
		kv("error", status.Error),
	)
}

// RenderSubscriptionDetail renders subscription definition attributes.
func RenderSubscriptionDetail(detail mqadmin.SubscriptionDetail) string {
	return joinPairs(
		kv("name", detail.Name),
		kv("id", detail.ID),
		kv("topic", detail.TopicString),
		kv("type", detail.Type),
		kv("dest", detail.Destination),
	)
}

// RenderReasonExplanation renders offline reason-code reference text.
func RenderReasonExplanation(explanation mqadmin.ReasonExplanation) string {
	known := "false"
	if explanation.Known {
		known = "true"
	}
	head := joinPairs(
		fmt.Sprintf("code=%d", explanation.Code),
		kv("symbol", explanation.Symbol),
		"known="+known,
	)
	body := strings.TrimSpace(explanation.Summary)
	if explanation.Action != "" {
		body = body + " | " + strings.TrimSpace(explanation.Action)
	}
	return head + "\n" + body
}

// RenderConnectivityReport renders a side-effect-free connectivity check.
func RenderConnectivityReport(report mqadmin.ConnectivityReport) string {
	return joinPairs(
		kv("profile", report.Profile),
		kv("endpoint", report.Endpoint),
		fmt.Sprintf("reachable=%t", report.Reachable),
		fmt.Sprintf("identityMatch=%t", report.IdentityMatch),
		kvInt("latencyMs", int(report.LatencyMs)),
		kv("configured", report.Identity.Configured),
		kv("observed", report.Identity.Observed),
		kv("failure", string(report.FailureCause)),
		kv("detail", report.Detail),
		kvTime("checkedAt", report.CheckedAt),
	)
}

// RenderPutResult renders message production identifiers only.
func RenderPutResult(result messaging.PutResult) string {
	return joinPairs(
		kv("messageId", result.MessageID),
		kv("correlationId", result.CorrelationID),
		kv("format", result.Format),
	)
}

// RenderQueueMutationResult renders queue mutation before/after identifiers.
func RenderQueueMutationResult(result mqadmin.QueueMutationResult) string {
	pairs := []string{
		kv("operation", string(result.Operation)),
		kv("profile", result.Profile),
		kv("queueManager", result.QueueManager),
		kv("queue", result.QueueName),
		kvTime("completedAt", result.CompletedAt),
	}
	if result.Before != nil {
		pairs = append(pairs, kv("before", result.Before.Name))
	}
	if result.After != nil {
		pairs = append(pairs, kv("after", result.After.Name))
	}
	if result.Warning != "" {
		pairs = append(pairs, kv("warning", result.Warning))
	}
	return joinPairs(pairs...)
}

// RenderChannelMutationResult renders channel mutation before/after identifiers.
func RenderChannelMutationResult(result mqadmin.ChannelMutationResult) string {
	pairs := []string{
		kv("operation", string(result.Operation)),
		kv("profile", result.Profile),
		kv("queueManager", result.QueueManager),
		kv("channel", result.ChannelName),
		kvTime("completedAt", result.CompletedAt),
	}
	if result.Before != nil {
		pairs = append(pairs, kv("before", result.Before.Name))
	}
	if result.After != nil {
		pairs = append(pairs, kv("after", result.After.Name))
	}
	if result.Warning != "" {
		pairs = append(pairs, kv("warning", result.Warning))
	}
	return joinPairs(pairs...)
}

// RenderCHLAUTHMutationResult renders CHLAUTH mutation identifiers.
func RenderCHLAUTHMutationResult(result mqadmin.CHLAUTHMutationResult) string {
	pairs := []string{
		kv("operation", string(result.Operation)),
		kv("profile", result.Profile),
		kv("queueManager", result.QueueManager),
		kv("channel", result.Target.ChannelName),
		kv("ruleType", result.Target.RuleType),
		kvTime("completedAt", result.CompletedAt),
	}
	if result.Warning != "" {
		pairs = append(pairs, kv("warning", result.Warning))
	}
	return joinPairs(pairs...)
}

// RenderAuthrecMutationResult renders authority-record mutation identifiers.
func RenderAuthrecMutationResult(result mqadmin.AuthrecMutationResult) string {
	pairs := []string{
		kv("operation", string(result.Operation)),
		kv("profile", result.Profile),
		kv("queueManager", result.QueueManager),
		kv("authrecProfile", result.Target.Profile),
		kv("objectType", result.Target.ObjectType),
		kv("entity", result.Target.Entity),
		kvTime("completedAt", result.CompletedAt),
	}
	if result.Warning != "" {
		pairs = append(pairs, kv("warning", result.Warning))
	}
	return joinPairs(pairs...)
}

// RenderRawMQSCResult renders exceptional raw MQSC completion metadata.
func RenderRawMQSCResult(result mqadmin.RawMQSCResult) string {
	return joinPairs(
		kv("profile", result.Profile),
		kv("queueManager", result.QueueManager),
		kv("command", mqadmin.RedactMQSCCommandText(result.Command)),
		kv("completionCode", fmt.Sprintf("%d", result.Completion.OverallCompletionCode)),
		kv("reasonCode", fmt.Sprintf("%d", result.Completion.OverallReasonCode)),
		kvTime("completedAt", result.CompletedAt),
	)
}

// RenderMarkdownQueueTable renders a Markdown table for benchmark comparison.
func RenderMarkdownQueueTable(page collection.Page[mqadmin.QueueSummary]) string {
	var b strings.Builder
	b.WriteString("| name | type |\n| --- | --- |\n")
	for _, item := range page.Items {
		fmt.Fprintf(&b, "| %s | %s |\n", item.Name, item.Type)
	}
	if page.Truncated {
		fmt.Fprintf(&b, "_truncated=%s limit=%d_\n", page.TruncationReason, page.Limit)
	}
	return b.String()
}

// RenderMarkdownMessageTable renders a Markdown table for benchmark comparison.
func RenderMarkdownMessageTable(page collection.Page[messaging.MessageRecord]) string {
	var b strings.Builder
	b.WriteString("| messageId | len | encoding |\n| --- | --- | --- |\n")
	for _, item := range page.Items {
		fmt.Fprintf(&b, "| %s | %d | %s |\n", item.MessageID, item.MessageLength, item.Encoding)
	}
	if page.Truncated {
		fmt.Fprintf(&b, "_truncated=%s limit=%d_\n", page.TruncationReason, page.Limit)
	}
	return b.String()
}
