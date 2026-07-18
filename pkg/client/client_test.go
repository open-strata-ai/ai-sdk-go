package client

import (
	"context"
	"testing"

	"github.com/open-strata-ai/ai-sdk-go/pkg/adapter"
	"github.com/open-strata-ai/ai-sdk-go/pkg/agent"
	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

func newTestClient() *Client {
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

func TestNewDefaultChat(t *testing.T) {
	c := NewDefault(nil)
	resp, err := c.Chat(context.Background(), domain.ChatRequest{
		Messages: []domain.ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if resp.Content != "echo: hello" {
		t.Fatalf("content = %q", resp.Content)
	}

	content, err := c.ChatByInput(context.Background(), ChatInput{Message: "world"})
	if err != nil || content != "echo: world" {
		t.Fatalf("chatByInput = %q, %v", content, err)
	}
}

func TestRetriever(t *testing.T) {
	c := newTestClient()
	if err := c.VectorStore().Upsert(context.Background(), "kb", []domain.Doc{
		{ID: "a", Text: "alpha", Vector: []float32{1, 0}},
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	r := c.Retriever()
	hits, err := r.Search("kb", []float32{1, 0}, 1)
	if err != nil || len(hits) != 1 || hits[0].ID != "a" {
		t.Fatalf("search = %+v, %v", hits, err)
	}
	rag, err := r.Retrieve("q", "kb", 3)
	if err != nil || rag != nil {
		t.Fatalf("retrieve = %+v, %v", rag, err)
	}
}

func TestSessionRoundtrip(t *testing.T) {
	c := newTestClient()
	s := c.Session("user-1")
	s.SetWorkingMemoryTTL(0)
	if err := s.Put("k", []byte("v")); err != nil {
		t.Fatalf("put: %v", err)
	}
	if _, ok, _ := s.Get("k"); ok {
		t.Fatalf("expected expired entry missing")
	}
}

func TestAgentBuilderRun(t *testing.T) {
	c := NewDefault(nil)
	out, err := agent.New(c.AgentRuntime(), "demo").
		Tenant("tenant-b").
		ModelBinding(&domain.ModelBinding{Primary: "cloud-qwen-max"}).
		InputSchema(domain.NewObjectSchema("query", "string")).
		OutputSchema(domain.NewObjectSchema("answer", "string")).
		Guardrails(&domain.Guardrails{Checks: []string{"injection_scan", "pii_scan"}}).
		Run(context.Background(), map[string]any{"query": "Where's my order?"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out["agent"] != "demo" || out["query"] != "Where's my order?" {
		t.Fatalf("out = %+v", out)
	}
}

func TestAuthAndTrace(t *testing.T) {
	c := NewDefault(nil)
	tc, err := c.Auth(context.Background(), "token")
	if err != nil || tc.TenantID != "local" {
		t.Fatalf("auth = %+v, %v", tc, err)
	}
	ctx, span := c.Trace(context.Background(), "op")
	if span.Name != "op" || ctx == nil {
		t.Fatalf("trace = %+v, %v", span, ctx)
	}
}
