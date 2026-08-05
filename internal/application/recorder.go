// Package application orchestrates use cases across policy and domain ports.
package application

// Recorder captures MCP request and policy-denial activity for observability.
// Implementations must use low-cardinality profile labels only.
type Recorder interface {
	RecordRequest(profile string, seconds float64)
	RecordPolicyDenial(profile string)
}
