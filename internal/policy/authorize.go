package policy

import (
	"fmt"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
)

// DenialError is returned when a profile lacks a required capability.
type DenialError struct {
	Profile  string
	Required Capability
}

func (e *DenialError) Error() string {
	return fmt.Sprintf(
		`profile %q lacks capability %q required for this operation; `+
			`add the grant under capabilities in the profile catalog (see ADR-0003)`,
		e.Profile,
		e.Required,
	)
}

// HasGrant reports whether profile explicitly grants required.
func HasGrant(profile catalog.Profile, required Capability) bool {
	if !IsKnown(required) {
		return false
	}
	for _, grant := range profile.Capabilities {
		if Capability(grant) == required {
			return true
		}
	}
	return false
}

// Authorize denies by default when required is absent from profile grants.
func Authorize(profile catalog.Profile, required Capability) error {
	if !IsKnown(required) {
		return &DenialError{Profile: profile.Name, Required: required}
	}
	if !HasGrant(profile, required) {
		return &DenialError{Profile: profile.Name, Required: required}
	}
	return nil
}
