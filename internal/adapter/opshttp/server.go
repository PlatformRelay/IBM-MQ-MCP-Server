// Package opshttp serves operational endpoints separately from MCP transport.
package opshttp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/metrics"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
)

// Server binds health, readiness, and metrics on a dedicated HTTP listener.
type Server struct {
	addr    string
	handler http.Handler
}

// NewServer returns an ops HTTP server for addr (for example ":9090").
func NewServer(addr string, rt *runtime.Runtime, reg *metrics.Registry) *Server {
	return &Server{
		addr:    addr,
		handler: NewHandler(rt, reg),
	}
}

// NewHandler returns the mux for health, readiness, and metrics routes.
func NewHandler(rt *runtime.Runtime, reg *metrics.Registry) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		if rt.Healthy() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}
		http.Error(w, "unhealthy", http.StatusServiceUnavailable)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if rt.Ready() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
			return
		}
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	})
	mux.Handle("/metrics", reg.Handler())
	return mux
}

// ListenAndServe starts the ops listener until an error occurs.
func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return srv.ListenAndServe()
}

// Addr returns the bind address for logging.
func (s *Server) Addr() string {
	return s.addr
}

// String describes the listener for startup logs.
func (s *Server) String() string {
	return fmt.Sprintf("ops-http@%s", s.addr)
}
