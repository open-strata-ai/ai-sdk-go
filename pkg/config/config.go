// Package config loads the SDK configuration from defaults, environment
// variables, an optional YAML file, and explicit overrides (priority:
// explicit Option > env > file > defaults). It depends only on the standard
// library so the SDK stays embeddable in any host.
package config

import (
	"os"
)

// GatewayConfig configures the Gateway SPI adapter.
type GatewayConfig struct {
	Provider string
	BaseURL  string
	Token    string
}

// ModelEndpoint configures a single LLM provider endpoint.
type ModelEndpoint struct {
	Type    string
	Default bool
	APIKey  string
}

// ModelConfig configures LLM provider endpoints.
type ModelConfig struct {
	Qwen   ModelEndpoint
	OpenAI ModelEndpoint
	Claude ModelEndpoint
}

// CacheConfig configures the Cache SPI adapter.
type CacheConfig struct {
	Provider string
}

// VectorStoreConfig configures the VectorStore SPI adapter.
type VectorStoreConfig struct {
	Preference string
}

// ObservabilityConfig configures tracing/audit.
type ObservabilityConfig struct {
	OTelTraces bool
	AuditLog   bool
}

// Config is the assembled SDK configuration.
type Config struct {
	Gateway       GatewayConfig
	Model         ModelConfig
	Cache         CacheConfig
	VectorStore   VectorStoreConfig
	Observability ObservabilityConfig
}

type loadState struct {
	filePath string
}

// LoadOption mutates a Config during Load.
type LoadOption func(*Config, *loadState)

// FromEnv applies environment-variable overrides.
func FromEnv() LoadOption {
	return func(c *Config, _ *loadState) {
		if v := os.Getenv("GATEWAY_BASE_URL"); v != "" {
			c.Gateway.BaseURL = v
		}
		if v := os.Getenv("OPENSTRATA_GATEWAY_PROVIDER"); v != "" {
			c.Gateway.Provider = v
		}
		if v := os.Getenv("OPENSTRATA_TOKEN"); v != "" {
			c.Gateway.Token = v
		}
		if v := os.Getenv("DASHSCOPE_API_KEY"); v != "" {
			c.Model.Qwen.APIKey = v
		}
		if v := os.Getenv("OPENAI_API_KEY"); v != "" {
			c.Model.OpenAI.APIKey = v
		}
		if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
			c.Model.Claude.APIKey = v
		}
		if v := os.Getenv("OPENSTRATA_CACHE_PROVIDER"); v != "" {
			c.Cache.Provider = v
		}
		if v := os.Getenv("OPENSTRATA_VECTORSTORE_PREFERENCE"); v != "" {
			c.VectorStore.Preference = v
		}
		if v := os.Getenv("OPENSTRATA_OTEL_TRACES"); v != "" {
			c.Observability.OTelTraces = v == "true"
		}
		if v := os.Getenv("OPENSTRATA_AUDIT_LOG"); v != "" {
			c.Observability.AuditLog = v == "true"
		}
	}
}

// FromFile records a YAML config file to apply during Load.
func FromFile(path string) LoadOption {
	return func(_ *Config, s *loadState) {
		s.filePath = path
	}
}

func defaultConfig() *Config {
	return &Config{
		Gateway: GatewayConfig{Provider: "higress"},
		Model: ModelConfig{
			Qwen:   ModelEndpoint{Type: "dashscope", Default: true},
			OpenAI: ModelEndpoint{Type: "openai"},
			Claude: ModelEndpoint{Type: "anthropic"},
		},
		Cache:         CacheConfig{Provider: "redis"},
		VectorStore:   VectorStoreConfig{Preference: "qdrant"},
		Observability: ObservabilityConfig{OTelTraces: true, AuditLog: true},
	}
}

// Load assembles a Config from defaults then the supplied options.
func Load(opts ...LoadOption) (*Config, error) {
	c := defaultConfig()
	st := &loadState{}
	for _, o := range opts {
		o(c, st)
	}
	if st.filePath != "" {
		if err := applyFile(c, st.filePath); err != nil {
			return nil, err
		}
	}
	return c, nil
}
