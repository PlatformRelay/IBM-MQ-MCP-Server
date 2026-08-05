// Package metrics exposes low-cardinality Prometheus counters and histograms.
package metrics

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

const noneProfile = "_none"

// Registry holds MCP request and policy-denial metrics keyed by profile name.
type Registry struct {
	requests       *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
	policyDenials  *prometheus.CounterVec
	gatherer       prometheus.Gatherer
}

// New registers metrics on a private registry for the ops HTTP handler.
func New() *Registry {
	reg := prometheus.NewRegistry()

	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ibm_mq_mcp_requests_total",
		Help: "Total MCP tool requests handled, labeled by connection profile name.",
	}, []string{"profile"})

	requestLatency := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ibm_mq_mcp_request_duration_seconds",
		Help:    "MCP tool request latency in seconds, labeled by connection profile name.",
		Buckets: prometheus.DefBuckets,
	}, []string{"profile"})

	policyDenials := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ibm_mq_mcp_policy_denials_total",
		Help: "Total policy denials, labeled by connection profile name.",
	}, []string{"profile"})

	reg.MustRegister(requests, requestLatency, policyDenials)

	return &Registry{
		requests:       requests,
		requestLatency: requestLatency,
		policyDenials:  policyDenials,
		gatherer:       reg,
	}
}

// RecordRequest increments the request counter and observes latency for profile.
func (r *Registry) RecordRequest(profile string, seconds float64) {
	label := normalizeProfile(profile)
	r.requests.WithLabelValues(label).Inc()
	r.requestLatency.WithLabelValues(label).Observe(seconds)
}

// RecordPolicyDenial increments the policy denial counter for profile.
func (r *Registry) RecordPolicyDenial(profile string) {
	r.policyDenials.WithLabelValues(normalizeProfile(profile)).Inc()
}

// Handler returns the Prometheus HTTP handler backed by this registry.
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.gatherer, promhttp.HandlerOpts{})
}

// Gather returns metric families for tests.
func (r *Registry) Gather() ([]*dto.MetricFamily, error) {
	return r.gatherer.Gather()
}

func normalizeProfile(profile string) string {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return noneProfile
	}
	return profile
}
