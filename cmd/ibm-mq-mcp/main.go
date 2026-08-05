// Command ibm-mq-mcp runs the IBM MQ MCP server over stdio (default) with optional remote HTTP.
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

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/mqweb"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/opshttp"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/remotemcp"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/logging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/metrics"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
)

const (
	envOpsAddr            = "IBM_MQ_MCP_OPS_ADDR"
	envConfigPath         = "IBM_MQ_MCP_CONFIG"
	envEnableMQSC         = "IBM_MQ_MCP_ENABLE_MQSC"
	envRemoteAddr         = remotemcp.EnvRemoteAddr
	envRemoteAuthTokenRef = remotemcp.EnvRemoteAuthTokenRef
)

func main() {
	opsAddrFlag := flag.String("ops-addr", "", "optional ops HTTP listen address (health, readiness, metrics)")
	remoteAddrFlag := flag.String(
		"remote-addr", "",
		"optional Streamable HTTP MCP listen address (requires remote auth token ref)",
	)
	remoteAuthRefFlag := flag.String(
		"remote-auth-token-ref", "",
		"env: or file: reference for MCP client bearer token gate",
	)
	stdioFlag := flag.Bool("stdio", true, "serve MCP over stdio (disable for remote-only deployments)")
	configFlag := flag.String("config", "", "path to profile catalog YAML or JSON file")
	strictStartup := flag.Bool("strict-startup", false, "exit on any profile validation failure")
	enableMQSCFlag := flag.Bool(
		"enable-mqsc",
		false,
		"register exceptional raw MQSC tool (ADR-0008; also requires profile execute_mqsc)",
	)
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
			if closeErr := pool.Close(); closeErr != nil {
				slog.Error("profile pool shutdown", slog.String("error", closeErr.Error()))
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
			if serveErr := opsServer.ListenAndServe(); serveErr != nil {
				slog.Error("ops listener failed", slog.String("error", serveErr.Error()))
				os.Exit(1)
			}
		}()
	}

	var server *mcp.Server
	if pool != nil {
		server = mcpserver.NewWithInspector(
			application.NewInspector(pool),
			mcpserver.WithEnableMQSC(resolveEnableMQSC(*enableMQSCFlag)),
		)
	} else {
		server = mcpserver.New()
	}

	remoteCfg, err := resolveRemoteConfig(*remoteAddrFlag, *remoteAuthRefFlag)
	if err != nil {
		slog.Error("remote MCP configuration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if remoteCfg.Addr != "" {
		remoteServer, err := remotemcp.NewServer(remoteCfg, server, rt)
		if err != nil {
			slog.Error("remote MCP setup failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		go func() {
			slog.Info("remote MCP listener starting", slog.String("addr", remoteServer.Addr()))
			if serveErr := remoteServer.ListenAndServe(); serveErr != nil {
				slog.Error("remote MCP listener failed", slog.String("error", serveErr.Error()))
				os.Exit(1)
			}
		}()
	}

	if *stdioFlag {
		if err := mcpserver.RunStdio(ctx, server, rt); err != nil {
			slog.Error("mcp server stopped", slog.String("error", err.Error()))
			os.Exit(1)
		}
		return
	}

	if remoteCfg.Addr == "" {
		slog.Error("stdio disabled but no remote MCP listen address configured")
		os.Exit(1)
	}
	if err := remotemcp.Wait(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("remote MCP wait failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func resolveOpsAddr(flagValue string) string {
	if v := strings.TrimSpace(flagValue); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(envOpsAddr))
}

func resolveRemoteConfig(addrFlag, authRefFlag string) (remotemcp.Config, error) {
	addr := strings.TrimSpace(addrFlag)
	if addr == "" {
		addr = strings.TrimSpace(os.Getenv(envRemoteAddr))
	}
	authRef := strings.TrimSpace(authRefFlag)
	if authRef == "" {
		authRef = strings.TrimSpace(os.Getenv(envRemoteAuthTokenRef))
	}
	cfg := remotemcp.Config{
		Addr:          addr,
		AuthTokenRef:  authRef,
		Limits:        remotemcp.DefaultLimits(),
		TransportName: "streamable-http",
	}
	if err := remotemcp.Validate(cfg); err != nil {
		return remotemcp.Config{}, err
	}
	return cfg, nil
}

func resolveConfigPath(flagValue string) string {
	if v := strings.TrimSpace(flagValue); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(envConfigPath))
}

func resolveEnableMQSC(flagValue bool) bool {
	if flagValue {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv(envEnableMQSC))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
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
	pool := application.NewProfilePool(cat, validation, nil, nil,
		application.WithAdminFactory(mqweb.NewAdminClient),
		application.WithMessagingFactory(mqweb.NewMessagingClient),
	)
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
