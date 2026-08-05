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

// MapReasonCode returns a typed error for known IBM MQ reason codes.
func MapReasonCode(code int) error {
	if mapped, ok := knownReasons[code]; ok {
		return &ReasonError{Code: code, Symbol: mapped.symbol, Message: mapped.message, Action: mapped.action}
	}
	return &ReasonError{
		Code:    code,
		Symbol:  "MQRC_UNKNOWN",
		Message: fmt.Sprintf("unmapped IBM MQ reason code %d", code),
		Action:  "check IBM MQ documentation or open INS-003 diagnostics when available",
	}
}

type reasonMapping struct {
	symbol  string
	message string
	action  string
}

var knownReasons = map[int]reasonMapping{
	2009: {
		symbol:  "MQRC_CONNECTION_BROKEN",
		message: "connection to the queue manager was lost",
		action:  "verify mqweb and the queue manager are running and reachable",
	},
	2035: {
		symbol:  "MQRC_NOT_AUTHORIZED",
		message: "the caller is not authorized for the requested operation",
		action: "grant the mqweb user read authority on the object or add inspect capability " +
			"only after MQ auth is fixed",
	},
	2085: {
		symbol:  "MQRC_UNKNOWN_OBJECT_NAME",
		message: "the named object does not exist",
		action:  "check the object name and queue manager profile",
	},
	2195: {
		symbol:  "MQRC_OBJECT_IN_USE",
		message: "the object is in use and the request could not be satisfied",
		action:  "retry after the exclusive operation completes",
	},
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
	case 401, 403:
		return MapReasonCode(2035)
	case 404:
		return MapReasonCode(2085)
	default:
		return fmt.Errorf("mqweb request failed with HTTP %d", status)
	}
}
