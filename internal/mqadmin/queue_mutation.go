package mqadmin

import "time"

// QueueType is a validated local queue definition type (no raw MQSC).
type QueueType string

// Supported queue definition types for typed administration.
const (
	QueueTypeLocal  QueueType = "LOCAL"
	QueueTypeAlias  QueueType = "ALIAS"
	QueueTypeRemote QueueType = "REMOTE"
	QueueTypeModel  QueueType = "MODEL"
)

// MutationOperation names a queue administration verb.
type MutationOperation string

// Queue mutation operations exposed by ADM-001.
const (
	MutationDefine MutationOperation = "define"
	MutationAlter  MutationOperation = "alter"
	MutationDelete MutationOperation = "delete"
)

// DefineQueueRequest carries typed queue creation attributes.
type DefineQueueRequest struct {
	QueueType   QueueType
	MaxDepth    *int
	Description string
}

// AlterQueueRequest carries typed queue alteration attributes.
type AlterQueueRequest struct {
	MaxDepth    *int
	Description *string
}

// QueueSnapshot captures queue identity fields for mutation results.
type QueueSnapshot struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	MaxDepth    int    `json:"maxDepth,omitempty"`
	Description string `json:"description,omitempty"`
}

// QueueMutationResult records before/after identifiers for audited mutations.
type QueueMutationResult struct {
	Operation    MutationOperation `json:"operation"`
	Profile      string            `json:"profile"`
	QueueManager string            `json:"queueManager"`
	QueueName    string            `json:"queueName"`
	Before       *QueueSnapshot    `json:"before,omitempty"`
	After        *QueueSnapshot    `json:"after,omitempty"`
	Warning      string            `json:"warning,omitempty"`
	CompletedAt  time.Time         `json:"completedAt"`
}

// ValidateDefineQueueRequest rejects unsupported queue types and empty names.
func ValidateDefineQueueRequest(name string, req DefineQueueRequest) error {
	if err := validateQueueName(name); err != nil {
		return err
	}
	return validateQueueType(req.QueueType)
}

// ValidateAlterQueueRequest rejects empty names and no-op alter payloads.
func ValidateAlterQueueRequest(name string, req AlterQueueRequest) error {
	if err := validateQueueName(name); err != nil {
		return err
	}
	if req.MaxDepth == nil && req.Description == nil {
		return ErrAlterNoChanges
	}
	return nil
}

// ValidateDeleteQueueRequest rejects empty queue names.
func ValidateDeleteQueueRequest(name string) error {
	return validateQueueName(name)
}

func validateQueueName(name string) error {
	if name == "" {
		return ErrQueueNameRequired
	}
	return nil
}

func validateQueueType(qtype QueueType) error {
	switch qtype {
	case QueueTypeLocal, QueueTypeAlias, QueueTypeRemote, QueueTypeModel:
		return nil
	default:
		return ErrInvalidQueueType
	}
}
