package client

import (
	"github.com/open-strata-ai/ai-sdk-go/pkg/adapter"
	"github.com/open-strata-ai/ai-sdk-go/pkg/config"
)

// NewDefault wires the SDK with stdlib-only default adapters (offline/test friendly).
// cfg may be nil to use built-in defaults; it is reserved for future provider
// selection based on tenant preferences (§10.4 runtime routing).
func NewDefault(cfg *config.Config) *Client {
	return New(
		WithGateway(adapter.NewLoggingGateway()),
		WithAgentRuntime(adapter.NewLocalAgentRuntime()),
		WithVectorStore(adapter.NewInMemoryVectorStore()),
		WithCache(adapter.NewInMemoryCache()),
		WithLLMProvider(adapter.NewEchoLLMProvider()),
		WithRAG(adapter.NewLoggingRAG()),
		WithTracing(adapter.NewNoOpTracing()),
		WithAuth(adapter.NewLocalAuth()),
	)
}
