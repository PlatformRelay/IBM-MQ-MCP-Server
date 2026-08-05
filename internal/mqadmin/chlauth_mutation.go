package mqadmin

import (
	"strings"
	"time"
)

// CHLAUTHType names a supported channel authentication rule type.
type CHLAUTHType string

// Supported CHLAUTH rule types for typed administration.
const (
	CHLAUTHTypeAddressMap CHLAUTHType = "ADDRESSMAP"
	CHLAUTHTypeBlockUser  CHLAUTHType = "BLOCKUSER"
	CHLAUTHTypeUserMap    CHLAUTHType = "USERMAP"
	CHLAUTHTypeQMgrMap    CHLAUTHType = "QMGRMAP"
	CHLAUTHTypeSSLPeerMap CHLAUTHType = "SSLPEERMAP"
)

// CHLAUTHUserSource controls how a matching rule maps asserted users.
type CHLAUTHUserSource string

// CHLAUTH user source values.
const (
	CHLAUTHUserSourceNoAccess CHLAUTHUserSource = "NOACCESS"
	CHLAUTHUserSourceChannel  CHLAUTHUserSource = "CHANNEL"
	CHLAUTHUserSourceMap      CHLAUTHUserSource = "MAP"
)

// CHLAUTHTarget identifies one channel authentication record by exact identity.
type CHLAUTHTarget struct {
	ChannelName string
	RuleType    CHLAUTHType
	Address     string
	ClientUser  string
	SSLPeer     string
	QMgrName    string
}

// DefineCHLAUTHRequest carries typed CHLAUTH creation attributes.
type DefineCHLAUTHRequest struct {
	Target     CHLAUTHTarget
	UserSource CHLAUTHUserSource
	MCAUser    string
}

// AlterCHLAUTHRequest carries typed CHLAUTH alteration attributes.
type AlterCHLAUTHRequest struct {
	Target     CHLAUTHTarget
	UserSource *CHLAUTHUserSource
	MCAUser    *string
}

// CHLAUTHSnapshot captures CHLAUTH identity fields for mutation results.
type CHLAUTHSnapshot struct {
	ChannelName string `json:"channelName"`
	RuleType    string `json:"ruleType"`
	Address     string `json:"address,omitempty"`
	ClientUser  string `json:"clientUser,omitempty"`
	SSLPeer     string `json:"sslPeer,omitempty"`
	QMgrName    string `json:"qMgrName,omitempty"`
	UserSource  string `json:"userSource,omitempty"`
	MCAUser     string `json:"mcaUser,omitempty"`
}

// CHLAUTHMutationResult records before/after identifiers for audited CHLAUTH mutations.
type CHLAUTHMutationResult struct {
	Operation    MutationOperation `json:"operation"`
	Profile      string            `json:"profile"`
	QueueManager string            `json:"queueManager"`
	Target       CHLAUTHSnapshot   `json:"target"`
	Before       *CHLAUTHSnapshot  `json:"before,omitempty"`
	After        *CHLAUTHSnapshot  `json:"after,omitempty"`
	Warning      string            `json:"warning,omitempty"`
	CompletedAt  time.Time         `json:"completedAt"`
}

// ValidateDefineCHLAUTHRequest rejects incomplete identity and unsupported types.
func ValidateDefineCHLAUTHRequest(req DefineCHLAUTHRequest) error {
	if err := validateCHLAUTHType(req.Target.RuleType); err != nil {
		return err
	}
	if err := validateCHLAUTHIdentity(req.Target); err != nil {
		return err
	}
	if req.UserSource == "" {
		return ErrCHLAUTHUserSourceRequired
	}
	return validateCHLAUTHUserSource(req.UserSource)
}

// ValidateAlterCHLAUTHRequest rejects incomplete identity and no-op payloads.
func ValidateAlterCHLAUTHRequest(req AlterCHLAUTHRequest) error {
	if err := validateCHLAUTHType(req.Target.RuleType); err != nil {
		return err
	}
	if err := validateCHLAUTHIdentity(req.Target); err != nil {
		return err
	}
	if req.UserSource == nil && req.MCAUser == nil {
		return ErrAlterNoChanges
	}
	if req.UserSource != nil {
		if err := validateCHLAUTHUserSource(*req.UserSource); err != nil {
			return err
		}
	}
	return nil
}

// ValidateDeleteCHLAUTHRequest rejects incomplete CHLAUTH identity.
func ValidateDeleteCHLAUTHRequest(target CHLAUTHTarget) error {
	if err := validateCHLAUTHType(target.RuleType); err != nil {
		return err
	}
	return validateCHLAUTHIdentity(target)
}

func validateCHLAUTHType(ruleType CHLAUTHType) error {
	switch ruleType {
	case CHLAUTHTypeAddressMap, CHLAUTHTypeBlockUser, CHLAUTHTypeUserMap,
		CHLAUTHTypeQMgrMap, CHLAUTHTypeSSLPeerMap:
		return nil
	default:
		return ErrInvalidCHLAUTHType
	}
}

func validateCHLAUTHIdentity(target CHLAUTHTarget) error {
	if strings.TrimSpace(target.ChannelName) == "" {
		return ErrCHLAUTHChannelRequired
	}
	switch target.RuleType {
	case CHLAUTHTypeAddressMap:
		if strings.TrimSpace(target.Address) == "" {
			return ErrCHLAUTHIdentityIncomplete
		}
	case CHLAUTHTypeBlockUser, CHLAUTHTypeUserMap:
		if strings.TrimSpace(target.ClientUser) == "" {
			return ErrCHLAUTHIdentityIncomplete
		}
	case CHLAUTHTypeSSLPeerMap:
		if strings.TrimSpace(target.SSLPeer) == "" {
			return ErrCHLAUTHIdentityIncomplete
		}
	case CHLAUTHTypeQMgrMap:
		if strings.TrimSpace(target.QMgrName) == "" {
			return ErrCHLAUTHIdentityIncomplete
		}
	}
	return nil
}

func validateCHLAUTHUserSource(src CHLAUTHUserSource) error {
	switch src {
	case CHLAUTHUserSourceNoAccess, CHLAUTHUserSourceChannel, CHLAUTHUserSourceMap:
		return nil
	default:
		return ErrInvalidCHLAUTHUserSource
	}
}

// CHLAUTHSnapshotFromTarget builds a snapshot from a validated target.
func CHLAUTHSnapshotFromTarget(target CHLAUTHTarget) CHLAUTHSnapshot {
	return CHLAUTHSnapshot{
		ChannelName: target.ChannelName,
		RuleType:    string(target.RuleType),
		Address:     target.Address,
		ClientUser:  target.ClientUser,
		SSLPeer:     target.SSLPeer,
		QMgrName:    target.QMgrName,
	}
}
