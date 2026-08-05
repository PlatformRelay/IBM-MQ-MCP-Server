// Package remotemcp serves opt-in Streamable HTTP MCP with bearer auth and abuse limits (ADR-0006).
package remotemcp

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/time/rate"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
)

const (
	// EnvRemoteAddr is the listen address for Streamable HTTP MCP.
	EnvRemoteAddr = "IBM_MQ_MCP_REMOTE_ADDR"
	// EnvRemoteAuthTokenRef is the env:/file: reference for the MCP gate bearer token.
	EnvRemoteAuthTokenRef = "IBM_MQ_MCP_REMOTE_AUTH_TOKEN_REF" //nolint:gosec // G101: env var name, not a secret
)

// Config configures the remote MCP HTTP listener.
type Config struct {
	Addr          string
	AuthTokenRef  string
	AuthToken     string // resolved token; required when Addr is set
	Limits        Limits
	TransportName string
}

// Limits defines abuse controls for remote MCP (ADR-0006).
type Limits struct {
	MaxBodyBytes      int64
	RequestsPerSecond float64
	Burst             int
	MaxConcurrency    int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

// DefaultLimits returns conservative defaults for remote MCP.
func DefaultLimits() Limits {
	return Limits{
		MaxBodyBytes:      1 << 20, // 1 MiB
		RequestsPerSecond: 20,
		Burst:             40,
		MaxConcurrency:    32,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

// Validate ensures remote configuration is safe to start.
func Validate(cfg Config) error {
	if strings.TrimSpace(cfg.Addr) == "" {
		return nil
	}
	if strings.TrimSpace(cfg.AuthTokenRef) == "" && strings.TrimSpace(cfg.AuthToken) == "" {
		return errors.New("remote MCP requires auth token ref when listen address is set")
	}
	if cfg.Limits.MaxBodyBytes <= 0 {
		return errors.New("remote MCP max body bytes must be positive")
	}
	if cfg.Limits.MaxConcurrency <= 0 {
		return errors.New("remote MCP max concurrency must be positive")
	}
	return nil
}

// ResolveAuthToken loads the gate bearer token from an env: or file: reference.
func ResolveAuthToken(ref string) (string, error) {
	parsed, err := secrets.Parse(ref)
	if err != nil {
		return "", err
	}
	value, err := secrets.NewResolver().Resolve(parsed)
	if err != nil {
		return "", fmt.Errorf("resolve remote auth token: %w", err)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("remote auth token must not be empty")
	}
	return value, nil
}

// Server binds Streamable HTTP MCP separately from ops endpoints.
type Server struct {
	addr    string
	http    *http.Server
	handler http.Handler
	rt      *runtime.Runtime
	name    string
}

// NewServer constructs a remote MCP listener. Call ListenAndServe or Serve.
func NewServer(cfg Config, mcpServer *mcp.Server, rt *runtime.Runtime) (*Server, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	token := strings.TrimSpace(cfg.AuthToken)
	if token == "" && cfg.AuthTokenRef != "" {
		var err error
		token, err = ResolveAuthToken(cfg.AuthTokenRef)
		if err != nil {
			return nil, err
		}
	}
	if token == "" {
		return nil, errors.New("remote MCP auth token is required")
	}

	transportName := cfg.TransportName
	if transportName == "" {
		transportName = "streamable-http"
	}

	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return mcpServer
	}, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})

	var chain http.Handler = mcpHandler
	chain = BearerAuth(token, chain)
	chain = LimitHandler(cfg.Limits, chain)

	limits := cfg.Limits
	if limits.ReadHeaderTimeout == 0 {
		limits = DefaultLimits()
	}

	return &Server{
		addr:    cfg.Addr,
		handler: chain,
		rt:      rt,
		name:    transportName,
		http: &http.Server{
			Addr:              cfg.Addr,
			Handler:           chain,
			ReadHeaderTimeout: limits.ReadHeaderTimeout,
			ReadTimeout:       limits.ReadTimeout,
			WriteTimeout:      limits.WriteTimeout,
			IdleTimeout:       limits.IdleTimeout,
		},
	}, nil
}

// TestListener serves remote MCP on an ephemeral port for tests.
type TestListener struct {
	Server *http.Server
	URL    string
}

// NewTestServer listens on an ephemeral port for tests.
func NewTestServer(cfg Config, mcpServer *mcp.Server, rt *runtime.Runtime) (*TestListener, error) {
	cfg.Addr = "127.0.0.1:0"
	srv, err := NewServer(cfg, mcpServer, rt)
	if err != nil {
		return nil, err
	}
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, err
	}
	testSrv := &http.Server{
		Addr:              ln.Addr().String(),
		Handler:           srv.handler,
		ReadHeaderTimeout: srv.http.ReadHeaderTimeout,
		ReadTimeout:       srv.http.ReadTimeout,
		WriteTimeout:      srv.http.WriteTimeout,
		IdleTimeout:       srv.http.IdleTimeout,
	}
	go func() { _ = testSrv.Serve(ln) }()
	if rt != nil {
		rt.SetTransportReady(true, srv.name)
	}
	return &TestListener{
		Server: testSrv,
		URL:    "http://" + ln.Addr().String(),
	}, nil
}

// Close shuts down the test listener.
func (t *TestListener) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return t.Server.Shutdown(ctx)
}

// ListenAndServe starts the remote MCP listener until an error occurs.
func (s *Server) ListenAndServe() error {
	if s.rt != nil {
		s.rt.SetTransportReady(true, s.name)
		defer s.rt.SetTransportReady(false, "")
	}
	return s.http.ListenAndServe()
}

// Addr returns the bind address for logging.
func (s *Server) Addr() string {
	return s.addr
}

// BearerAuth validates a server-configured bearer token and strips Authorization
// before forwarding to the MCP handler so client tokens never reach downstream MQ.
func BearerAuth(expectedToken string, next http.Handler) http.Handler {
	expected := []byte(expectedToken)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := parseBearer(r.Header.Get("Authorization"))
		if !ok || subtle.ConstantTimeCompare([]byte(token), expected) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		r.Header.Del("Authorization")
		next.ServeHTTP(w, r)
	})
}

func parseBearer(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}

// LimitHandler enforces body size, rate, and concurrency limits.
func LimitHandler(limits Limits, next http.Handler) http.Handler {
	if limits.MaxBodyBytes <= 0 {
		limits.MaxBodyBytes = DefaultLimits().MaxBodyBytes
	}
	if limits.RequestsPerSecond <= 0 {
		limits.RequestsPerSecond = DefaultLimits().RequestsPerSecond
	}
	if limits.Burst <= 0 {
		limits.Burst = DefaultLimits().Burst
	}
	if limits.MaxConcurrency <= 0 {
		limits.MaxConcurrency = DefaultLimits().MaxConcurrency
	}

	limiter := rate.NewLimiter(rate.Limit(limits.RequestsPerSecond), limits.Burst)
	sem := make(chan struct{}, limits.MaxConcurrency)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		if r.ContentLength > limits.MaxBodyBytes {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		select {
		case sem <- struct{}{}:
		default:
			http.Error(w, "too many concurrent requests", http.StatusServiceUnavailable)
			return
		}
		defer func() { <-sem }()

		r.Body = http.MaxBytesReader(w, r.Body, limits.MaxBodyBytes)
		next.ServeHTTP(w, r)
	})
}

// Wait blocks until ctx is cancelled — used for remote-only mode without stdio.
func Wait(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
