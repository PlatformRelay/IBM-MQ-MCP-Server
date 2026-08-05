// Package coexistence implements INT-001 MKurator pre-mutation policy (ADR-0007).
package coexistence

import (
	"fmt"
	"path"
	"strings"
)

const (
	// TagManagedByMKurator is the object-tag key for declarative ownership metadata.
	TagManagedByMKurator = "mkurator.platformrelay.io/managed"
)

// ObjectKind names a supported MQ object family for ownership checks.
type ObjectKind string

// Supported object kinds for ownership metadata.
const (
	ObjectQueue   ObjectKind = "queue"
	ObjectChannel ObjectKind = "channel"
	ObjectCHLAUTH ObjectKind = "chlauth"
	ObjectAuthrec ObjectKind = "authrec"
)

// MutationPolicy controls behaviour when a managed object is mutated.
type MutationPolicy string

// Mutation policy values from ADR-0007.
const (
	PolicyWarn  MutationPolicy = "warn"
	PolicyBlock MutationPolicy = "block"
)

// PreMutationOutcome is the hook decision before mqweb mutation I/O.
type PreMutationOutcome string

// Pre-mutation hook outcomes.
const (
	OutcomeAllow PreMutationOutcome = "allow"
	OutcomeWarn  PreMutationOutcome = "warn"
	OutcomeBlock PreMutationOutcome = "block"
)

// OwnershipSource records where ownership metadata was resolved.
type OwnershipSource string

// Ownership metadata sources for INT-001.
const (
	OwnershipCatalog   OwnershipSource = "catalog"
	OwnershipObjectTag OwnershipSource = "object_tag"
)

// ProfileConfig holds per-profile MKurator coexistence settings from the catalog.
type ProfileConfig struct {
	MutationPolicy MutationPolicy     `yaml:"mutationPolicy,omitempty" json:"mutationPolicy,omitempty"`
	ManagedObjects []ManagedObjectRef `yaml:"managedObjects,omitempty" json:"managedObjects,omitempty"`
}

// ManagedObjectRef declares catalog ownership for one object name or glob pattern.
type ManagedObjectRef struct {
	Kind ObjectKind `yaml:"kind" json:"kind"`
	Name string     `yaml:"name" json:"name"`
}

// MutationTarget identifies the object about to be mutated.
type MutationTarget struct {
	Profile      string
	QueueManager string
	Kind         ObjectKind
	Name         string
}

// OwnershipMetadata describes declarative ownership when present.
type OwnershipMetadata struct {
	Managed     bool            `json:"managed"`
	Source      OwnershipSource `json:"source,omitempty"`
	ResourceRef string          `json:"resourceRef,omitempty"`
	Pattern     string          `json:"pattern,omitempty"`
}

// PreMutationResult is returned by Evaluate before adapter mutation I/O.
type PreMutationResult struct {
	Outcome   PreMutationOutcome `json:"outcome"`
	Message   string             `json:"message,omitempty"`
	Ownership OwnershipMetadata  `json:"ownership"`
}

// BlockError indicates the hook blocked a mutation under block policy.
type BlockError struct {
	Target  MutationTarget
	Result  PreMutationResult
	Message string
}

func (e *BlockError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf(
		"mutation of MKurator-managed %s %q blocked by profile policy",
		e.Target.Kind,
		e.Target.Name,
	)
}

// PreMutationHook evaluates ownership metadata and mutation policy (INT-001).
type PreMutationHook struct {
	policy  MutationPolicy
	managed []ManagedObjectRef
}

// DefaultMutationPolicy returns warn, the ADR-0007 v0 default.
func DefaultMutationPolicy() MutationPolicy {
	return PolicyWarn
}

// NewPreMutationHook constructs a hook from profile catalog settings.
func NewPreMutationHook(cfg ProfileConfig) *PreMutationHook {
	policy := cfg.MutationPolicy
	if policy == "" {
		policy = DefaultMutationPolicy()
	}
	managed := append([]ManagedObjectRef(nil), cfg.ManagedObjects...)
	return &PreMutationHook{policy: policy, managed: managed}
}

// Evaluate resolves ownership from catalog patterns and optional object tags.
func (h *PreMutationHook) Evaluate(target MutationTarget, tags map[string]string) PreMutationResult {
	ownership := h.resolveOwnership(target, tags)
	if !ownership.Managed {
		return PreMutationResult{Outcome: OutcomeAllow, Ownership: ownership}
	}
	msg := fmt.Sprintf(
		"%s %q is managed by MKurator (%s); direct mutation may be reconciled away",
		target.Kind,
		target.Name,
		ownership.ResourceRef,
	)
	switch h.policy {
	case PolicyBlock:
		return PreMutationResult{
			Outcome:   OutcomeBlock,
			Message:   msg,
			Ownership: ownership,
		}
	default:
		return PreMutationResult{
			Outcome:   OutcomeWarn,
			Message:   msg,
			Ownership: ownership,
		}
	}
}

// Enforce returns BlockError when the result requires fail-closed behaviour.
func (h *PreMutationHook) Enforce(result PreMutationResult) error {
	if result.Outcome != OutcomeBlock {
		return nil
	}
	return &BlockError{
		Message: result.Message,
		Result:  result,
	}
}

func (h *PreMutationHook) resolveOwnership(target MutationTarget, tags map[string]string) OwnershipMetadata {
	if ref, pattern, ok := h.matchCatalog(target); ok {
		return OwnershipMetadata{
			Managed:     true,
			Source:      OwnershipCatalog,
			ResourceRef: catalogResourceRef(target, ref),
			Pattern:     pattern,
		}
	}
	if ref := strings.TrimSpace(tags[TagManagedByMKurator]); ref != "" {
		return OwnershipMetadata{
			Managed:     true,
			Source:      OwnershipObjectTag,
			ResourceRef: ref,
		}
	}
	return OwnershipMetadata{Managed: false}
}

func (h *PreMutationHook) matchCatalog(target MutationTarget) (resourceSuffix string, pattern string, ok bool) {
	name := strings.TrimSpace(target.Name)
	for _, ref := range h.managed {
		if ref.Kind != "" && ref.Kind != target.Kind {
			continue
		}
		pattern = strings.TrimSpace(ref.Name)
		if pattern == "" {
			continue
		}
		if matchObjectName(pattern, name) {
			return pattern, pattern, true
		}
	}
	return "", "", false
}

func catalogResourceRef(target MutationTarget, pattern string) string {
	qm := strings.TrimSpace(target.QueueManager)
	if qm == "" {
		qm = "unknown"
	}
	kind := string(target.Kind)
	if kind == "" {
		kind = "object"
	}
	return path.Join("Mk"+objectKindResourcePrefix(kind), qm, pattern)
}

func objectKindResourcePrefix(kind string) string {
	switch ObjectKind(kind) {
	case ObjectQueue:
		return "Queue"
	case ObjectChannel:
		return "Channel"
	case ObjectCHLAUTH:
		return "CHLAUTH"
	case ObjectAuthrec:
		return "Authrec"
	default:
		if kind == "" {
			return "Object"
		}
		return strings.ToUpper(kind[:1]) + kind[1:]
	}
}

func matchObjectName(pattern, name string) bool {
	pattern = strings.TrimSpace(pattern)
	name = strings.TrimSpace(name)
	if pattern == "" || name == "" {
		return false
	}
	if !strings.Contains(pattern, "*") {
		return strings.EqualFold(pattern, name)
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(strings.ToUpper(name), strings.ToUpper(prefix))
	}
	return false
}
