package mqadmin

import "time"

// ChannelType is a validated channel definition type (no raw MQSC).
type ChannelType string

// Supported channel definition types for typed administration.
const (
	ChannelTypeSender           ChannelType = "SDR"
	ChannelTypeServer           ChannelType = "SVR"
	ChannelTypeReceiver         ChannelType = "RCVR"
	ChannelTypeRequester        ChannelType = "RQSTR"
	ChannelTypeClientConnection ChannelType = "CLNTCONN"
	ChannelTypeServerConnection ChannelType = "SVRCONN"
	ChannelTypeClusterSender    ChannelType = "CLUSSDR"
	ChannelTypeClusterReceiver  ChannelType = "CLUSRCVR"
)

// DefineChannelRequest carries typed channel creation attributes.
type DefineChannelRequest struct {
	ChannelType       ChannelType
	Description       string
	ConnectionName    string
	TransmissionQueue string
}

// AlterChannelRequest carries typed channel alteration attributes.
type AlterChannelRequest struct {
	Description       *string
	ConnectionName    *string
	TransmissionQueue *string
}

// ChannelSnapshot captures channel identity fields for mutation results.
type ChannelSnapshot struct {
	Name              string `json:"name"`
	Type              string `json:"type,omitempty"`
	Description       string `json:"description,omitempty"`
	ConnectionName    string `json:"connectionName,omitempty"`
	TransmissionQueue string `json:"transmissionQueue,omitempty"`
}

// ChannelMutationResult records before/after identifiers for audited channel mutations.
type ChannelMutationResult struct {
	Operation    MutationOperation `json:"operation"`
	Profile      string            `json:"profile"`
	QueueManager string            `json:"queueManager"`
	ChannelName  string            `json:"channelName"`
	Before       *ChannelSnapshot  `json:"before,omitempty"`
	After        *ChannelSnapshot  `json:"after,omitempty"`
	Warning      string            `json:"warning,omitempty"`
	CompletedAt  time.Time         `json:"completedAt"`
}

// ValidateDefineChannelRequest rejects unsupported channel types and empty names.
func ValidateDefineChannelRequest(name string, req DefineChannelRequest) error {
	if err := validateChannelName(name); err != nil {
		return err
	}
	return validateChannelType(req.ChannelType)
}

// ValidateAlterChannelRequest rejects empty names and no-op alter payloads.
func ValidateAlterChannelRequest(name string, req AlterChannelRequest) error {
	if err := validateChannelName(name); err != nil {
		return err
	}
	if req.Description == nil && req.ConnectionName == nil && req.TransmissionQueue == nil {
		return ErrAlterNoChanges
	}
	return nil
}

// ValidateDeleteChannelRequest rejects empty channel names.
func ValidateDeleteChannelRequest(name string) error {
	return validateChannelName(name)
}

func validateChannelName(name string) error {
	if name == "" {
		return ErrChannelNameRequired
	}
	return nil
}

func validateChannelType(chType ChannelType) error {
	switch chType {
	case ChannelTypeSender, ChannelTypeServer, ChannelTypeReceiver, ChannelTypeRequester,
		ChannelTypeClientConnection, ChannelTypeServerConnection,
		ChannelTypeClusterSender, ChannelTypeClusterReceiver:
		return nil
	default:
		return ErrInvalidChannelType
	}
}
