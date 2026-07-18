// Package client is the access layer of ai-sdk-go: the OpenAI-compatible
// Client facade that holds the tenant token and exposes the wired platform SPI
// ports. Hosts inject ports via options or use NewDefault for offline wiring.
package client

import (
	"context"

	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

// Client is the first-class citizen facade over the platform SPI.
type Client struct {
	gateway      domain.Gateway
	agentRuntime domain.AgentRuntime
	vectorStore  domain.VectorStore
	cache        domain.Cache
	llm          domain.LLMProvider
	rag          domain.RAG
	tracing      domain.Tracing
	auth         domain.Auth
}

// Option configures a Client.
type Option func(*Client)

// WithGateway sets the Gateway port.
func WithGateway(g domain.Gateway) Option { return func(c *Client) { c.gateway = g } }

// WithAgentRuntime sets the AgentRuntime port.
func WithAgentRuntime(a domain.AgentRuntime) Option { return func(c *Client) { c.agentRuntime = a } }

// WithVectorStore sets the VectorStore port.
func WithVectorStore(v domain.VectorStore) Option { return func(c *Client) { c.vectorStore = v } }

// WithCache sets the Cache port.
func WithCache(ca domain.Cache) Option { return func(c *Client) { c.cache = ca } }

// WithLLMProvider sets the LLMProvider port.
func WithLLMProvider(l domain.LLMProvider) Option { return func(c *Client) { c.llm = l } }

// WithRAG sets the RAG port.
func WithRAG(r domain.RAG) Option { return func(c *Client) { c.rag = r } }

// WithTracing sets the Tracing port.
func WithTracing(t domain.Tracing) Option { return func(c *Client) { c.tracing = t } }

// WithAuth sets the Auth port.
func WithAuth(a domain.Auth) Option { return func(c *Client) { c.auth = a } }

// New assembles a Client from the given options (manual injection).
func New(opts ...Option) *Client {
	c := &Client{}
	for _, o := range opts {
		o(c)
	}
	return c
}

// AgentRuntime exposes the AgentRuntime port.
func (c *Client) AgentRuntime() domain.AgentRuntime { return c.agentRuntime }

// VectorStore exposes the VectorStore port.
func (c *Client) VectorStore() domain.VectorStore { return c.vectorStore }

// Retriever builds a RAG retrieval facade over VectorStore + RAG.
func (c *Client) Retriever() *domain.Retriever { return domain.NewRetriever(c.vectorStore, c.rag) }

// Session builds a multi-turn session handle backed by the Cache port.
func (c *Client) Session(id string) *domain.Session { return domain.NewSession(id, c.cache) }

// Chat runs a chat completion.
func (c *Client) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
	return c.llm.Chat(ctx, req)
}

// Embed runs an embedding request.
func (c *Client) Embed(ctx context.Context, req domain.EmbedRequest) (domain.EmbedResponse, error) {
	return c.llm.Embed(ctx, req)
}

// Rerank runs a rerank request.
func (c *Client) Rerank(ctx context.Context, req domain.RerankRequest) (domain.RerankResponse, error) {
	return c.llm.Rerank(ctx, req)
}

// Invoke calls the platform Gateway.
func (c *Client) Invoke(ctx context.Context, req domain.GatewayRequest) (domain.GatewayResponse, error) {
	return c.gateway.Invoke(ctx, req)
}

// Trace starts a tracing span.
func (c *Client) Trace(ctx context.Context, name string) (context.Context, domain.Span) {
	return c.tracing.StartSpan(ctx, name)
}

// Auth validates a tenant token.
func (c *Client) Auth(ctx context.Context, token string) (domain.TenantContext, error) {
	return c.auth.ValidateToken(ctx, token)
}

// ChatInput is a convenience input for ChatByInput.
type ChatInput struct {
	SessionID string
	Message   string
}

// ChatByInput runs a chat from a ChatInput and returns the content string.
func (c *Client) ChatByInput(ctx context.Context, in ChatInput) (string, error) {
	resp, err := c.llm.Chat(ctx, domain.ChatRequest{
		Messages: []domain.ChatMessage{{Role: "user", Content: in.Message}},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
