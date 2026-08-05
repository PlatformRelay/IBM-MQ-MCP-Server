package mqadmin_test

import (
	"errors"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func TestValidateDefineAuthrecRequiresProfile(t *testing.T) {
	err := mqadmin.ValidateDefineAuthrecRequest(mqadmin.DefineAuthrecRequest{
		Target: mqadmin.AuthrecTarget{
			ObjectType: mqadmin.AuthrecObjectQueue,
			Entity:     "mqm",
			EntityType: mqadmin.AuthrecEntityPrincipal,
		},
		Authorities: []mqadmin.AuthrecAuthority{mqadmin.AuthrecAuthorityGet},
	})
	if !errors.Is(err, mqadmin.ErrAuthrecProfileRequired) {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateDefineAuthrecRequiresAuthorities(t *testing.T) {
	err := mqadmin.ValidateDefineAuthrecRequest(mqadmin.DefineAuthrecRequest{
		Target: mqadmin.AuthrecTarget{
			Profile:    "APP.IN",
			ObjectType: mqadmin.AuthrecObjectQueue,
			Entity:     "mqm",
			EntityType: mqadmin.AuthrecEntityPrincipal,
		},
	})
	if !errors.Is(err, mqadmin.ErrAuthrecAuthoritiesRequired) {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateDeleteAuthrecRequiresExactIdentity(t *testing.T) {
	err := mqadmin.ValidateDeleteAuthrecRequest(mqadmin.AuthrecTarget{
		Profile:    "APP.IN",
		ObjectType: mqadmin.AuthrecObjectQueue,
	})
	if !errors.Is(err, mqadmin.ErrAuthrecIdentityIncomplete) {
		t.Fatalf("error = %v", err)
	}
}
