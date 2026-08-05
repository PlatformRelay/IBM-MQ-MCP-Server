package mqadmin

import (
	"errors"
	"fmt"
)

// ReasonError is an actionable typed error mapped from an IBM MQ reason code.
type ReasonError struct {
	Code    int    `json:"code"`
	Symbol  string `json:"symbol"`
	Message string `json:"message"`
	Action  string `json:"action"`
}

func (e *ReasonError) Error() string {
	return fmt.Sprintf("MQ reason %d (%s): %s — %s", e.Code, e.Symbol, e.Message, e.Action)
}

type reasonReference struct {
	symbol  string
	summary string
	action  string
	docSlug string
}

// reasonReferences is the curated MQRC table shared by MapReasonCode and ExplainReasonCode.
// IBM owns the MQRC symbol namespace; see docs/NOTICE.md.
var reasonReferences = map[int]reasonReference{
	2009: {
		symbol:  "MQRC_CONNECTION_BROKEN",
		summary: "The active connection to the queue manager was interrupted or reset.",
		action:  "Verify mqweb and the queue manager are running, then retry after network stability returns.",
		docSlug: "mqrc-connection-broken",
	},
	2018: {
		symbol:  "MQRC_HCONN_ERROR",
		summary: "The connection handle is no longer valid for the requested call.",
		action:  "Discard the stale handle, reconnect to the queue manager, and retry the operation.",
		docSlug: "mqrc-hconn-error",
	},
	2019: {
		symbol:  "MQRC_HOBJ_ERROR",
		summary: "The object handle is no longer valid for the requested call.",
		action:  "Close and reopen the object, or reconnect if the queue manager recycled the handle.",
		docSlug: "mqrc-hobj-error",
	},
	2035: {
		symbol:  "MQRC_NOT_AUTHORIZED",
		summary: "The authenticated principal lacks authority for the requested operation or object.",
		action: "Grant the mqweb user read authority on the object or adjust profile " +
			"capabilities after MQ auth is corrected.",
		docSlug: "mqrc-not-authorized",
	},
	2059: {
		symbol:  "MQRC_Q_MGR_NOT_AVAILABLE",
		summary: "The queue manager is not accepting connections or is not in a runnable state.",
		action:  "Confirm the queue manager process is started and listeners are active on the target host.",
		docSlug: "mqrc-q-mgr-not-available",
	},
	2063: {
		symbol:  "MQRC_SECURITY_ERROR",
		summary: "A security subsystem or channel security exit rejected the connection attempt.",
		action:  "Review channel auth records, TLS settings, and any security exits configured on the queue manager.",
		docSlug: "mqrc-security-error",
	},
	2082: {
		symbol:  "MQRC_UNKNOWN_ALIAS",
		summary: "The supplied object name resolves to an alias that does not exist.",
		action:  "Verify alias queue or topic alias definitions and the spelling of the supplied name.",
		docSlug: "mqrc-unknown-alias",
	},
	2085: {
		symbol:  "MQRC_UNKNOWN_OBJECT_NAME",
		summary: "The named queue, channel, or other object was not found on the queue manager.",
		action:  "Check the object name, queue manager profile, and whether the object exists in the target environment.",
		docSlug: "mqrc-unknown-object-name",
	},
	2110: {
		symbol:  "MQRC_Q_FULL",
		summary: "The target queue reached its maximum depth and cannot accept more messages.",
		action:  "Drain or consume messages, increase max depth if appropriate, or redirect traffic temporarily.",
		docSlug: "mqrc-q-full",
	},
	2161: {
		symbol:  "MQRC_Q_DELETED",
		summary: "The queue was deleted while it was open or referenced by the failing call.",
		action:  "Recreate the queue or point producers and consumers at a replacement definition.",
		docSlug: "mqrc-q-deleted",
	},
	2190: {
		symbol:  "MQRC_OBJECT_TYPE_ERROR",
		summary: "The API expected one object type but the supplied name refers to another.",
		action:  "Use the object type that matches the API (for example queue vs channel) and retry.",
		docSlug: "mqrc-object-type-error",
	},
	2192: {
		symbol:  "MQRC_OBJECT_ALREADY_EXISTS",
		summary: "A create request targeted an object name that is already defined.",
		action:  "Choose a different name, alter the existing object, or delete it before recreating.",
		docSlug: "mqrc-object-already-exists",
	},
	2195: {
		symbol:  "MQRC_OBJECT_IN_USE",
		summary: "The object is locked or in exclusive use and cannot satisfy the request now.",
		action:  "Wait for the exclusive operation to finish, then retry with a short backoff.",
		docSlug: "mqrc-object-in-use",
	},
}

// MapReasonCode returns a typed error for known IBM MQ reason codes.
func MapReasonCode(code int) error {
	if entry, ok := reasonReferences[code]; ok {
		return &ReasonError{
			Code:    code,
			Symbol:  entry.symbol,
			Message: entry.summary,
			Action:  entry.action,
		}
	}
	return &ReasonError{
		Code:    code,
		Symbol:  "MQRC_UNKNOWN",
		Message: fmt.Sprintf("unmapped IBM MQ reason code %d", code),
		Action:  "check IBM MQ documentation or open INS-003 diagnostics when available",
	}
}

// AsReasonError extracts a ReasonError from err when present.
func AsReasonError(err error) (*ReasonError, bool) {
	var re *ReasonError
	if errors.As(err, &re) {
		return re, true
	}
	return nil, false
}

// ReasonCodeFromHTTPStatus maps common mqweb HTTP failures when no reason code is present.
func ReasonCodeFromHTTPStatus(status int) error {
	switch status {
	case 401:
		return fmt.Errorf("mqweb request failed with HTTP %d", status)
	case 403:
		return MapReasonCode(2035)
	case 404:
		return MapReasonCode(2085)
	default:
		return fmt.Errorf("mqweb request failed with HTTP %d", status)
	}
}
