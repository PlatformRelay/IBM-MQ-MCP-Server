// Command ibm-mq-mcp runs the IBM MQ MCP server over stdio.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/opshttp"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/logging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/metrics"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
)

const envOpsAddr = "IBM_MQ_MCP_OPS_ADDR"

func main() {
	opsAddrFlag := flag.String("ops-addr", "", "optional ops HTTP listen address (health, readiness, metrics)")
	flag.Parse()

	logger := slog.New(logging.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	rt := runtime.New()
	rt.SetConfigValid(validateBootstrapConfig())

	metricReg := metrics.New()

	opsAddr := resolveOpsAddr(*opsAddrFlag)
	if opsAddr != "" {
		opsServer := opshttp.NewServer(opsAddr, rt, metricReg)
		go func() {
			slog.Info("ops listener starting", slog.String("addr", opsServer.Addr()))
			if err := opsServer.ListenAndServe(); err != nil {
				slog.Error("ops listener failed", slog.String("error", err.Error()))
				os.Exit(1)
			}
		}()
	}

	server := mcpserver.New()
	if err := mcpserver.RunStdio(ctx, server, rt); err != nil {
		slog.Error("mcp server stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func resolveOpsAddr(flagValue string) string {
	if v := strings.TrimSpace(flagValue); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(envOpsAddr))
}

// validateBootstrapConfig checks static configuration without contacting queue managers.
// Profiles and config files arrive in CON-001; until then an empty bootstrap is valid.
func validateBootstrapConfig() bool {
	return true
}
