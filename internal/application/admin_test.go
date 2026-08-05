package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/coexistence"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
)

const adminProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_ADMIN_SECRET
    tls:
      insecureSkipVerify: true
    capabilities:
      - administer
    mkurator:
      mutationPolicy: block
      managedObjects:
        - kind: queue
          name: APP.*
`

func TestAdministratorDeniesBeforeAdapter(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_ADMIN_SECRET", "user:pass")
	doc := `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_ADMIN_SECRET
    capabilities:
      - inspect
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := fake.New("prod")
	pool := application.NewProfilePool(cat, cat.Validate(), secrets.NewResolver(), nil,
		application.WithAdminFactory(func(_ catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
			return fakeClient, nil
		}),
	)
	t.Cleanup(func() { _ = pool.Close() })
	admin := application.NewAdministrator(pool)

	_, err = admin.DefineQueue(context.Background(), "prod", "NEW.Q", mqadmin.DefineQueueRequest{
		QueueType: mqadmin.QueueTypeLocal,
	})
	if err == nil {
		t.Fatal("expected policy denial")
	}
	if !strings.Contains(err.Error(), "administer") {
		t.Fatalf("error = %v", err)
	}
	if fakeClient.TotalCalls() != 0 {
		t.Fatalf("adapter invoked on deny: %d", fakeClient.TotalCalls())
	}
}

func TestAdministratorBlocksManagedQueueBeforeAdapter(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_ADMIN_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(adminProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := fake.New("prod")
	pool := newAdminPool(t, cat, fakeClient)
	admin := application.NewAdministrator(pool)

	_, err = admin.DefineQueue(context.Background(), "prod", "APP.IN", mqadmin.DefineQueueRequest{
		QueueType: mqadmin.QueueTypeLocal,
	})
	if err == nil {
		t.Fatal("expected block")
	}
	var block *coexistence.BlockError
	if !errors.As(err, &block) {
		t.Fatalf("error type = %T: %v", err, err)
	}
	if fakeClient.DefineQueueCalls != 0 {
		t.Fatalf("define calls = %d", fakeClient.DefineQueueCalls)
	}
}

func TestAdministratorDefineQueueReturnsWarning(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_ADMIN_SECRET", "user:pass")
	doc := `
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
      managedObjects:
        - kind: queue
          name: APP.*
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := fake.New("prod")
	fakeClient.DefineQueueResult = mqadmin.QueueMutationResult{
		Operation:    mqadmin.MutationDefine,
		Profile:      "prod",
		QueueManager: "QM1",
		QueueName:    "APP.NEW",
		After:        &mqadmin.QueueSnapshot{Name: "APP.NEW", Type: string(mqadmin.QueueTypeLocal)},
	}
	pool := newAdminPool(t, cat, fakeClient)
	admin := application.NewAdministrator(pool)

	result, err := admin.DefineQueue(context.Background(), "prod", "APP.NEW", mqadmin.DefineQueueRequest{
		QueueType: mqadmin.QueueTypeLocal,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Warning == "" {
		t.Fatal("expected coexistence warning")
	}
	if fakeClient.DefineQueueCalls != 1 {
		t.Fatalf("define calls = %d", fakeClient.DefineQueueCalls)
	}
}

func TestAdministratorRejectsInvalidQueueType(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_ADMIN_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(adminProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := fake.New("prod")
	pool := newAdminPool(t, cat, fakeClient)
	admin := application.NewAdministrator(pool)

	_, err = admin.DefineQueue(context.Background(), "prod", "FREE.Q", mqadmin.DefineQueueRequest{
		QueueType: mqadmin.QueueType("Bogus"),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if fakeClient.TotalCalls() != 0 {
		t.Fatalf("adapter invoked on validation error: %d", fakeClient.TotalCalls())
	}
}

func TestAdministratorDeleteNotFound(t *testing.T) {
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
	fakeClient.DeleteQueueErr = mqadmin.MapReasonCode(2085)
	pool := newAdminPool(t, cat, fakeClient)
	admin := application.NewAdministrator(pool)

	_, err = admin.DeleteQueue(context.Background(), "prod", "MISSING.Q")
	if err == nil {
		t.Fatal("expected not-found")
	}
	var reason *mqadmin.ReasonError
	if !errors.As(err, &reason) || reason.Code != 2085 {
		t.Fatalf("error = %v", err)
	}
}

func TestAdministratorDefineConflict(t *testing.T) {
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
	fakeClient.DefineQueueErr = mqadmin.MapReasonCode(2192)
	pool := newAdminPool(t, cat, fakeClient)
	admin := application.NewAdministrator(pool)

	_, err = admin.DefineQueue(context.Background(), "prod", "EXISTING.Q", mqadmin.DefineQueueRequest{
		QueueType: mqadmin.QueueTypeLocal,
	})
	if err == nil {
		t.Fatal("expected conflict")
	}
	var reason *mqadmin.ReasonError
	if !errors.As(err, &reason) || reason.Code != 2192 {
		t.Fatalf("error = %v", err)
	}
}

func newAdminPool(t *testing.T, cat *catalog.Catalog, fakeClient *fake.Client) *application.ProfilePool {
	t.Helper()
	pool := application.NewProfilePool(cat, cat.Validate(), secrets.NewResolver(), nil,
		application.WithAdminFactory(func(_ catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
			return fakeClient, nil
		}),
	)
	t.Cleanup(func() { _ = pool.Close() })
	return pool
}
