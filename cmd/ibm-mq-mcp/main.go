// Command ibm-mq-mcp runs the IBM MQ MCP server over stdio.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/opshttp"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/logging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/metrics"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
)

const (
	envOpsAddr    = "IBM_MQ_MCP_OPS_ADDR"
	envConfigPath = "IBM_MQ_MCP_CONFIG"
)

func main() {
	opsAddrFlag := flag.String("ops-addr", "", "optional ops HTTP listen address (health, readiness, metrics)")
	configFlag := flag.String("config", "", "path to profile catalog YAML or JSON file")
	strictStartup := flag.Bool("strict-startup", false, "exit on any profile validation failure")
	flag.Parse()

	logger := slog.New(logging.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	rt := runtime.New()
	pool, configValid, err := loadProfiles(resolveConfigPath(*configFlag), *strictStartup)
	if err != nil {
		slog.Error("configuration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if pool != nil {
		defer func() {
			if err := pool.Close(); err != nil {
				slog.Error("profile pool shutdown", slog.String("error", err.Error()))
			}
		}()
	}
	rt.SetConfigValid(configValid)

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

func resolveConfigPath(flagValue string) string {
	if v := strings.TrimSpace(flagValue); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(envConfigPath))
}

func loadProfiles(path string, strictStartup bool) (*application.ProfilePool, bool, error) {
	if path == "" {
		return nil, true, nil
	}
	cat, err := application.LoadCatalogFromFile(path)
	if err != nil {
		return nil, false, err
	}
	validation := cat.Validate()
	if strictStartup && !validation.AllValid() {
		return nil, false, firstValidationError(validation)
	}
	pool := application.NewProfilePool(cat, validation, nil, nil)
	return pool, application.ConfigReady(cat, validation), nil
}

func firstValidationError(validation catalog.ValidationResult) error {
	for _, status := range validation.Statuses {
		if !status.Valid && status.Err != nil {
			return fmt.Errorf("profile %q: %w", status.Name, status.Err)
		}
	}
	return errors.New("strict startup: invalid profile catalog")
}
