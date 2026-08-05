package mqadmin

import (
	"strings"
	"time"
)

// AuthrecObjectType names the IBM MQ object class for an authority record.
type AuthrecObjectType string

// Supported authority-record object types.
const (
	AuthrecObjectQueue    AuthrecObjectType = "QUEUE"
	AuthrecObjectQMgr     AuthrecObjectType = "QMGRC"
	AuthrecObjectChannel  AuthrecObjectType = "CHANNEL"
	AuthrecObjectProcess  AuthrecObjectType = "PROCESS"
	AuthrecObjectTopic    AuthrecObjectType = "TOPIC"
	AuthrecObjectNamelist AuthrecObjectType = "NAMELIST"
)

// AuthrecEntityType names whether the entity is a principal or group.
type AuthrecEntityType string

// Authority record entity types.
const (
	AuthrecEntityPrincipal AuthrecEntityType = "PRINCIPAL"
	AuthrecEntityGroup     AuthrecEntityType = "GROUP"
)

// AuthrecAuthority names one grantable authority for typed authrec mutations.
type AuthrecAuthority string

// Common authority record grants exposed by ADM-002.
const (
	AuthrecAuthorityAll     AuthrecAuthority = "ALL"
	AuthrecAuthorityAlter   AuthrecAuthority = "ALTER"
	AuthrecAuthorityBrowse  AuthrecAuthority = "BROWSE"
	AuthrecAuthorityChange  AuthrecAuthority = "CHG"
	AuthrecAuthorityClear   AuthrecAuthority = "CLR"
	AuthrecAuthorityConnect AuthrecAuthority = "CONNECT"
	AuthrecAuthorityCreate  AuthrecAuthority = "CRT"
	AuthrecAuthorityDelete  AuthrecAuthority = "DLT"
	AuthrecAuthorityDisplay AuthrecAuthority = "DSP"
	AuthrecAuthorityGet     AuthrecAuthority = "GET"
	AuthrecAuthorityInquire AuthrecAuthority = "INQ"
	AuthrecAuthorityPut     AuthrecAuthority = "PUT"
	AuthrecAuthoritySet     AuthrecAuthority = "SET"
	AuthrecAuthoritySetAll  AuthrecAuthority = "SETALL"
	AuthrecAuthoritySetID   AuthrecAuthority = "SETID"
	AuthrecAuthorityControl AuthrecAuthority = "CTRL"
	AuthrecAuthorityPassAll AuthrecAuthority = "PASSALL"
	AuthrecAuthorityPassID  AuthrecAuthority = "PASSID"
)

// AuthrecTarget identifies one authority record by exact identity.
type AuthrecTarget struct {
	Profile    string
	ObjectType AuthrecObjectType
	Entity     string
	EntityType AuthrecEntityType
}

// DefineAuthrecRequest carries typed authority-record creation attributes.
type DefineAuthrecRequest struct {
	Target      AuthrecTarget
	Authorities []AuthrecAuthority
}

// AlterAuthrecRequest carries typed authority-record alteration attributes.
type AlterAuthrecRequest struct {
	Target      AuthrecTarget
	AddAuths    []AuthrecAuthority
	RemoveAuths []AuthrecAuthority
}

// AuthrecSnapshot captures authority-record identity for mutation results.
type AuthrecSnapshot struct {
	Profile     string   `json:"profile"`
	ObjectType  string   `json:"objectType"`
	Entity      string   `json:"entity"`
	EntityType  string   `json:"entityType"`
	Authorities []string `json:"authorities,omitempty"`
}

// AuthrecMutationResult records before/after identifiers for audited authrec mutations.
type AuthrecMutationResult struct {
	Operation    MutationOperation `json:"operation"`
	Profile      string            `json:"profile"`
	QueueManager string            `json:"queueManager"`
	Target       AuthrecSnapshot   `json:"target"`
	Before       *AuthrecSnapshot  `json:"before,omitempty"`
	After        *AuthrecSnapshot  `json:"after,omitempty"`
	Warning      string            `json:"warning,omitempty"`
	CompletedAt  time.Time         `json:"completedAt"`
}

// ValidateDefineAuthrecRequest rejects incomplete identity and empty authority lists.
func ValidateDefineAuthrecRequest(req DefineAuthrecRequest) error {
	if err := validateAuthrecTarget(req.Target); err != nil {
		return err
	}
	if len(req.Authorities) == 0 {
		return ErrAuthrecAuthoritiesRequired
	}
	for _, auth := range req.Authorities {
		if err := validateAuthrecAuthority(auth); err != nil {
			return err
		}
	}
	return nil
}

// ValidateAlterAuthrecRequest rejects incomplete identity and no-op payloads.
func ValidateAlterAuthrecRequest(req AlterAuthrecRequest) error {
	if err := validateAuthrecTarget(req.Target); err != nil {
		return err
	}
	if len(req.AddAuths) == 0 && len(req.RemoveAuths) == 0 {
		return ErrAlterNoChanges
	}
	for _, auth := range append(append([]AuthrecAuthority{}, req.AddAuths...), req.RemoveAuths...) {
		if err := validateAuthrecAuthority(auth); err != nil {
			return err
		}
	}
	return nil
}

// ValidateDeleteAuthrecRequest rejects incomplete authority-record identity.
func ValidateDeleteAuthrecRequest(target AuthrecTarget) error {
	return validateAuthrecTarget(target)
}

func validateAuthrecTarget(target AuthrecTarget) error {
	if strings.TrimSpace(target.Profile) == "" {
		return ErrAuthrecProfileRequired
	}
	if err := validateAuthrecObjectType(target.ObjectType); err != nil {
		return err
	}
	if strings.TrimSpace(target.Entity) == "" {
		return ErrAuthrecIdentityIncomplete
	}
	return validateAuthrecEntityType(target.EntityType)
}

func validateAuthrecObjectType(objType AuthrecObjectType) error {
	switch objType {
	case AuthrecObjectQueue, AuthrecObjectQMgr, AuthrecObjectChannel,
		AuthrecObjectProcess, AuthrecObjectTopic, AuthrecObjectNamelist:
		return nil
	default:
		return ErrInvalidAuthrecObjectType
	}
}

func validateAuthrecEntityType(entityType AuthrecEntityType) error {
	switch entityType {
	case AuthrecEntityPrincipal, AuthrecEntityGroup:
		return nil
	default:
		return ErrInvalidAuthrecEntityType
	}
}

func validateAuthrecAuthority(auth AuthrecAuthority) error {
	switch auth {
	case AuthrecAuthorityAll, AuthrecAuthorityAlter, AuthrecAuthorityBrowse,
		AuthrecAuthorityChange, AuthrecAuthorityClear, AuthrecAuthorityConnect,
		AuthrecAuthorityCreate, AuthrecAuthorityDelete, AuthrecAuthorityDisplay,
		AuthrecAuthorityGet, AuthrecAuthorityInquire, AuthrecAuthorityPut,
		AuthrecAuthoritySet, AuthrecAuthoritySetAll, AuthrecAuthoritySetID,
		AuthrecAuthorityControl, AuthrecAuthorityPassAll, AuthrecAuthorityPassID:
		return nil
	default:
		return ErrInvalidAuthrecAuthority
	}
}

// AuthrecSnapshotFromTarget builds a snapshot from a validated target.
func AuthrecSnapshotFromTarget(target AuthrecTarget) AuthrecSnapshot {
	return AuthrecSnapshot{
		Profile:    target.Profile,
		ObjectType: string(target.ObjectType),
		Entity:     target.Entity,
		EntityType: string(target.EntityType),
	}
}
