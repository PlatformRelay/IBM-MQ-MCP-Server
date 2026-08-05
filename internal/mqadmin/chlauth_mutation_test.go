package mqadmin_test

import (
	"errors"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func TestValidateDefineCHLAUTHRequiresExactIdentity(t *testing.T) {
	err := mqadmin.ValidateDefineCHLAUTHRequest(mqadmin.DefineCHLAUTHRequest{
		Target: mqadmin.CHLAUTHTarget{
			ChannelName: "DEV.SVRCONN",
			RuleType:    mqadmin.CHLAUTHTypeAddressMap,
		},
	})
	if !errors.Is(err, mqadmin.ErrCHLAUTHIdentityIncomplete) {
		t.Fatalf("addressmap without address: error = %v", err)
	}
}

func TestValidateDefineCHLAUTHAcceptsAddressMap(t *testing.T) {
	err := mqadmin.ValidateDefineCHLAUTHRequest(mqadmin.DefineCHLAUTHRequest{
		Target: mqadmin.CHLAUTHTarget{
			ChannelName: "DEV.SVRCONN",
			RuleType:    mqadmin.CHLAUTHTypeAddressMap,
			Address:     "127.0.0.1",
		},
		UserSource: mqadmin.CHLAUTHUserSourceNoAccess,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidateDeleteCHLAUTHRequiresExactIdentity(t *testing.T) {
	err := mqadmin.ValidateDeleteCHLAUTHRequest(mqadmin.CHLAUTHTarget{
		ChannelName: "DEV.SVRCONN",
		RuleType:    mqadmin.CHLAUTHTypeUserMap,
	})
	if !errors.Is(err, mqadmin.ErrCHLAUTHIdentityIncomplete) {
		t.Fatalf("usermap without clientUser: error = %v", err)
	}
}
