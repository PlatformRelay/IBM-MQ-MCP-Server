package mqadmin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
)

var (
	// ErrMQSCCommandEmpty indicates an empty MQSC command string.
	ErrMQSCCommandEmpty = errors.New("mqsc command must not be empty")
	// ErrMQSCMultipleCommands indicates more than one MQSC command was supplied.
	ErrMQSCMultipleCommands = errors.New("mqsc command must not contain multiple statements")
	// ErrMQSCVerbDenied indicates the command verb is outside the v0 allowlist.
	ErrMQSCVerbDenied = errors.New("mqsc verb is not allowlisted for raw execution")
)

var allowedRawMQSCVerbs = map[string]struct{}{
	"DISPLAY": {},
	"DIS":     {},
	"PING":    {},
}

// MQSCCompletion captures mqweb overall completion metadata.
type MQSCCompletion struct {
	OverallCompletionCode int `json:"overallCompletionCode"`
	OverallReasonCode     int `json:"overallReasonCode"`
}

// RawMQSCResult records one exceptional raw MQSC execution.
type RawMQSCResult struct {
	Profile      string         `json:"profile"`
	QueueManager string         `json:"queueManager"`
	Command      string         `json:"command"`
	Completion   MQSCCompletion `json:"completion"`
	CompletedAt  time.Time      `json:"completedAt"`
}

// ValidateRawMQSCCommand parses and validates a raw MQSC command before network I/O.
func ValidateRawMQSCCommand(command string) error {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return ErrMQSCCommandEmpty
	}
	if containsStatementSeparator(trimmed) {
		return ErrMQSCMultipleCommands
	}
	if strings.Contains(trimmed, ";") {
		return ErrMQSCMultipleCommands
	}
	verb, err := parseMQSCVerb(trimmed)
	if err != nil {
		return err
	}
	if _, ok := allowedRawMQSCVerbs[strings.ToUpper(verb)]; !ok {
		return fmt.Errorf("%w: %q", ErrMQSCVerbDenied, verb)
	}
	return nil
}

// RedactMQSCCommandText redacts secret-like substrings from command text for audit output.
func RedactMQSCCommandText(command string) string {
	return messaging.RedactSecrets(command)
}

func parseMQSCVerb(command string) (string, error) {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "", ErrMQSCCommandEmpty
	}
	return fields[0], nil
}

func containsStatementSeparator(command string) bool {
	return strings.ContainsAny(command, "\r\n")
}
