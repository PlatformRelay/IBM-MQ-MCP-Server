package mqadmin_test

import (
	"errors"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func TestValidateDefineChannelRequestRejectsEmptyName(t *testing.T) {
	err := mqadmin.ValidateDefineChannelRequest("", mqadmin.DefineChannelRequest{
		ChannelType: mqadmin.ChannelTypeServerConnection,
	})
	if !errors.Is(err, mqadmin.ErrChannelNameRequired) {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateDefineChannelRequestRejectsInvalidType(t *testing.T) {
	err := mqadmin.ValidateDefineChannelRequest("DEV.SVRCONN", mqadmin.DefineChannelRequest{
		ChannelType: mqadmin.ChannelType("BOGUS"),
	})
	if !errors.Is(err, mqadmin.ErrInvalidChannelType) {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateAlterChannelRequestRejectsNoChanges(t *testing.T) {
	err := mqadmin.ValidateAlterChannelRequest("DEV.SVRCONN", mqadmin.AlterChannelRequest{})
	if !errors.Is(err, mqadmin.ErrAlterNoChanges) {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateDeleteChannelRequestRequiresName(t *testing.T) {
	err := mqadmin.ValidateDeleteChannelRequest("")
	if !errors.Is(err, mqadmin.ErrChannelNameRequired) {
		t.Fatalf("error = %v", err)
	}
}
