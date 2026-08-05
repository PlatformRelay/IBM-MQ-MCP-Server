package mqadmin

const (
	ibmMQReasonCodesDoc = "https://www.ibm.com/docs/en/ibm-mq/9.4?topic=constants-reason-codes"
	ibmMQDocBase        = "https://www.ibm.com/docs/en/ibm-mq/9.4?topic=constants-"
)

// ReasonExplanation is offline reference material for one IBM MQ reason code.
type ReasonExplanation struct {
	Code    int    `json:"code"`
	Symbol  string `json:"symbol"`
	Summary string `json:"summary"`
	Action  string `json:"action,omitempty"`
	Known   bool   `json:"known"`
	DocURL  string `json:"docUrl"`
}

// ExplainReasonCode returns bundled reference data for code without network I/O.
// Unknown codes receive a documented generic fallback, not an error.
func ExplainReasonCode(code int) ReasonExplanation {
	if entry, ok := reasonReferences[code]; ok {
		return ReasonExplanation{
			Code:    code,
			Symbol:  entry.symbol,
			Summary: entry.summary,
			Action:  entry.action,
			Known:   true,
			DocURL:  ibmMQDocBase + entry.docSlug,
		}
	}
	return ReasonExplanation{
		Code:   code,
		Symbol: "MQRC_UNKNOWN",
		Summary: "This numeric value is an IBM MQ reason code (MQRC) not included in the " +
			"server's curated offline reference table.",
		Action: "Look up the code in IBM MQ product documentation for the queue manager " +
			"version you are using, then correlate with the failing API call and object name.",
		Known:  false,
		DocURL: ibmMQReasonCodesDoc,
	}
}
