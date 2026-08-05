package mcpserver

// ServerOption configures optional MCP server behaviour.
type ServerOption func(*serverConfig)

type serverConfig struct {
	enableMQSC bool
}

// WithEnableMQSC registers the exceptional raw MQSC tool when true (ADR-0008).
func WithEnableMQSC(enabled bool) ServerOption {
	return func(cfg *serverConfig) {
		cfg.enableMQSC = enabled
	}
}

func applyServerOptions(opts []ServerOption) serverConfig {
	cfg := serverConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
