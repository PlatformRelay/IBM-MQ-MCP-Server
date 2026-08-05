package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/coexistence"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
)

const securityAdminProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_ADMIN_SECRET
    capabilities:
      - administer
    mkurator:
      mutationPolicy: block
      managedObjects:
        - kind: channel
          name: DEV.*
        - kind: chlauth
          name: DEV.*
        - kind: authrec
          name: APP.*
`

func TestAdministratorDeleteCHLAUTHBlocksBeforeAdapter(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_ADMIN_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(securityAdminProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := fake.New("prod")
	pool := newAdminPool(t, cat, fakeClient)
	admin := application.NewAdministrator(pool)

	_, err = admin.DeleteCHLAUTH(context.Background(), "prod", mqadmin.CHLAUTHTarget{
		ChannelName: "DEV.SVRCONN",
		RuleType:    mqadmin.CHLAUTHTypeAddressMap,
		Address:     "*",
	})
	if err == nil {
		t.Fatal("expected block")
	}
	var block *coexistence.BlockError
	if !errors.As(err, &block) {
		t.Fatalf("error type = %T: %v", err, err)
	}
	if fakeClient.TotalCalls() != 0 {
		t.Fatalf("adapter invoked on block: total=%d", fakeClient.TotalCalls())
	}
}

func TestAdministratorDeleteChannelBlocksBeforeAdapter(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_ADMIN_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(securityAdminProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := fake.New("prod")
	pool := newAdminPool(t, cat, fakeClient)
	admin := application.NewAdministrator(pool)

	_, err = admin.DeleteChannel(context.Background(), "prod", "DEV.SVRCONN")
	if err == nil {
		t.Fatal("expected block")
	}
	var block *coexistence.BlockError
	if !errors.As(err, &block) {
		t.Fatalf("error type = %T: %v", err, err)
	}
	if fakeClient.TotalCalls() != 0 {
		t.Fatalf("adapter invoked on block: total=%d", fakeClient.TotalCalls())
	}
}

func TestAdministratorDefineAuthrecNotFound(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_ADMIN_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(`
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_ADMIN_SECRET
    capabilities:
      - administer
`))
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := fake.New("prod")
	fakeClient.DefineAuthrecErr = mqadmin.MapReasonCode(2085)
	pool := newAdminPool(t, cat, fakeClient)
	admin := application.NewAdministrator(pool)

	_, err = admin.DefineAuthrec(context.Background(), "prod", mqadmin.DefineAuthrecRequest{
		Target: mqadmin.AuthrecTarget{
			Profile:    "MISSING.Q",
			ObjectType: mqadmin.AuthrecObjectQueue,
			Entity:     "mqm",
			EntityType: mqadmin.AuthrecEntityPrincipal,
		},
		Authorities: []mqadmin.AuthrecAuthority{mqadmin.AuthrecAuthorityGet},
	})
	if err == nil {
		t.Fatal("expected not-found")
	}
	var reason *mqadmin.ReasonError
	if !errors.As(err, &reason) || reason.Code != 2085 {
		t.Fatalf("error = %v", err)
	}
}

func TestAdministratorDefineChannelConflict(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_ADMIN_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(`
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_ADMIN_SECRET
    capabilities:
      - administer
`))
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := fake.New("prod")
	fakeClient.DefineChannelErr = mqadmin.MapReasonCode(2192)
	pool := newAdminPool(t, cat, fakeClient)
	admin := application.NewAdministrator(pool)

	_, err = admin.DefineChannel(context.Background(), "prod", "DEV.SVRCONN", mqadmin.DefineChannelRequest{
		ChannelType: mqadmin.ChannelTypeServerConnection,
	})
	if err == nil {
		t.Fatal("expected conflict")
	}
	var reason *mqadmin.ReasonError
	if !errors.As(err, &reason) || reason.Code != 2192 {
		t.Fatalf("error = %v", err)
	}
}
