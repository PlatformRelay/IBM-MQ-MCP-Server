package application_test

import (
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
)

func newPool(
	t *testing.T,
	cat *catalog.Catalog,
	gate *application.PolicyGate,
	factory application.AdminClientFactory,
) *application.ProfilePool {
	t.Helper()
	pool := application.NewProfilePool(
		cat,
		cat.Validate(),
		secrets.NewResolver(),
		gate,
		application.WithAdminFactory(factory),
	)
	t.Cleanup(func() { _ = pool.Close() })
	return pool
}
