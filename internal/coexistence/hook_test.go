package coexistence_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/coexistence"
)

func TestPreMutationHookAllowsUnmanagedObject(t *testing.T) {
	hook := coexistence.NewPreMutationHook(coexistence.ProfileConfig{})
	result := hook.Evaluate(coexistence.MutationTarget{
		Profile:      "prod",
		QueueManager: "QM1",
		Kind:         coexistence.ObjectQueue,
		Name:         "APP.IN",
	}, nil)
	if result.Outcome != coexistence.OutcomeAllow {
		t.Fatalf("outcome = %q", result.Outcome)
	}
	if err := hook.Enforce(result); err != nil {
		t.Fatal(err)
	}
}

func TestPreMutationHookWarnsManagedObjectByDefault(t *testing.T) {
	hook := coexistence.NewPreMutationHook(coexistence.ProfileConfig{
		ManagedObjects: []coexistence.ManagedObjectRef{{
			Kind: coexistence.ObjectQueue,
			Name: "APP.*",
		}},
	})
	result := hook.Evaluate(coexistence.MutationTarget{
		Profile:      "prod",
		QueueManager: "QM1",
		Kind:         coexistence.ObjectQueue,
		Name:         "APP.OUT",
	}, nil)
	if result.Outcome != coexistence.OutcomeWarn {
		t.Fatalf("outcome = %q", result.Outcome)
	}
	if !result.Ownership.Managed {
		t.Fatal("expected managed ownership")
	}
	if result.Ownership.Source != coexistence.OwnershipCatalog {
		t.Fatalf("source = %q", result.Ownership.Source)
	}
	if err := hook.Enforce(result); err != nil {
		t.Fatal("warn should not block")
	}
}

func TestPreMutationHookBlocksWhenPolicyBlock(t *testing.T) {
	hook := coexistence.NewPreMutationHook(coexistence.ProfileConfig{
		MutationPolicy: coexistence.PolicyBlock,
		ManagedObjects: []coexistence.ManagedObjectRef{{
			Kind: coexistence.ObjectQueue,
			Name: "ORDERS",
		}},
	})
	result := hook.Evaluate(coexistence.MutationTarget{
		Profile: "prod",
		Kind:    coexistence.ObjectQueue,
		Name:    "ORDERS",
	}, nil)
	if result.Outcome != coexistence.OutcomeBlock {
		t.Fatalf("outcome = %q", result.Outcome)
	}
	err := hook.Enforce(result)
	if err == nil {
		t.Fatal("expected block error")
	}
	var block *coexistence.BlockError
	if !errors.As(err, &block) {
		t.Fatalf("error type = %T", err)
	}
	if !strings.Contains(block.Error(), "MKurator") {
		t.Fatalf("error = %v", err)
	}
}

func TestPreMutationHookMatchesObjectTag(t *testing.T) {
	hook := coexistence.NewPreMutationHook(coexistence.ProfileConfig{})
	result := hook.Evaluate(coexistence.MutationTarget{
		Profile: "prod",
		Kind:    coexistence.ObjectQueue,
		Name:    "FREE.STANDING",
	}, map[string]string{
		coexistence.TagManagedByMKurator: "MkQueue/prod/free",
	})
	if result.Outcome != coexistence.OutcomeWarn {
		t.Fatalf("outcome = %q", result.Outcome)
	}
	if result.Ownership.Source != coexistence.OwnershipObjectTag {
		t.Fatalf("source = %q", result.Ownership.Source)
	}
	if result.Ownership.ResourceRef != "MkQueue/prod/free" {
		t.Fatalf("ref = %q", result.Ownership.ResourceRef)
	}
}

func TestPreMutationHookCatalogOverridesTagWhenBothMatch(t *testing.T) {
	hook := coexistence.NewPreMutationHook(coexistence.ProfileConfig{
		ManagedObjects: []coexistence.ManagedObjectRef{{
			Kind: coexistence.ObjectQueue,
			Name: "APP.IN",
		}},
	})
	result := hook.Evaluate(coexistence.MutationTarget{
		Profile: "prod",
		Kind:    coexistence.ObjectQueue,
		Name:    "APP.IN",
	}, map[string]string{
		coexistence.TagManagedByMKurator: "MkQueue/prod/other",
	})
	if result.Ownership.Source != coexistence.OwnershipCatalog {
		t.Fatalf("source = %q", result.Ownership.Source)
	}
}

func TestDefaultMutationPolicyIsWarn(t *testing.T) {
	if coexistence.DefaultMutationPolicy() != coexistence.PolicyWarn {
		t.Fatalf("default = %q", coexistence.DefaultMutationPolicy())
	}
}
