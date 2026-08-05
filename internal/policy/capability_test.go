package policy_test

import (
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

func TestAuthorizeGrantsEachKnownCapability(t *testing.T) {
	t.Parallel()

	for _, cap := range policy.AllCapabilities() {
		t.Run(string(cap), func(t *testing.T) {
			t.Parallel()
			profile := catalog.Profile{
				Name:         "prod",
				Capabilities: []string{string(cap)},
			}
			if err := policy.Authorize(profile, cap); err != nil {
				t.Fatalf("Authorize(%q) = %v, want nil", cap, err)
			}
		})
	}
}

func TestAuthorizeDeniesMissingGrant(t *testing.T) {
	t.Parallel()

	profile := catalog.Profile{
		Name:         "prod",
		Capabilities: []string{string(policy.Inspect)},
	}
	err := policy.Authorize(profile, policy.Administer)
	if err == nil {
		t.Fatal("expected denial")
	}
	var denial *policy.DenialError
	if !errorsAsDenial(err, &denial) {
		t.Fatalf("expected DenialError, got %T: %v", err, err)
	}
	if denial.Profile != "prod" || denial.Required != policy.Administer {
		t.Fatalf("denial = %+v", denial)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("error must not mention secrets: %v", err)
	}
}

func TestAuthorizeDeniesEmptyCapabilities(t *testing.T) {
	t.Parallel()

	profile := catalog.Profile{Name: "prod"}
	if err := policy.Authorize(profile, policy.Inspect); err == nil {
		t.Fatal("expected denial for empty grants")
	}
}

func TestAuthorizeDeniesUnknownRequiredCapability(t *testing.T) {
	t.Parallel()

	profile := catalog.Profile{
		Name:         "prod",
		Capabilities: []string{string(policy.Inspect)},
	}
	if err := policy.Authorize(profile, policy.Capability("destroy")); err == nil {
		t.Fatal("expected denial for unknown capability")
	}
}

func errorsAsDenial(err error, target **policy.DenialError) bool {
	if err == nil {
		return false
	}
	if d, ok := err.(*policy.DenialError); ok {
		*target = d
		return true
	}
	return false
}
