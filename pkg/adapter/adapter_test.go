package adapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

func TestInMemoryCacheTTL(t *testing.T) {
	c := NewInMemoryCache()
	if err := c.Set(context.Background(), "k", []byte("v"), time.Hour); err != nil {
		t.Fatalf("set: %v", err)
	}
	v, ok, err := c.Get(context.Background(), "k")
	if err != nil || !ok || string(v) != "v" {
		t.Fatalf("get = %q,%v,%v", v, ok, err)
	}
	// expired entry
	if err := c.Set(context.Background(), "e", []byte("x"), -time.Second); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, ok, _ := c.Get(context.Background(), "e"); ok {
		t.Fatalf("expected expired entry missing")
	}
}

func TestInMemoryVectorStore(t *testing.T) {
	v := NewInMemoryVectorStore()
	docs := []domain.Doc{
		{ID: "a", Text: "alpha", Vector: []float32{1, 0, 0}},
		{ID: "b", Text: "beta", Vector: []float32{0, 1, 0}},
		{ID: "c", Text: "gamma", Vector: []float32{0, 0, 1}},
	}
	if err := v.Upsert(context.Background(), "kb", docs); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	hits, err := v.Search(context.Background(), "kb", []float32{1, 0, 0}, 2)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 2 || hits[0].ID != "a" {
		t.Fatalf("hits = %+v", hits)
	}
	if err := v.Delete(context.Background(), "kb", []string{"a"}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	hits, _ = v.Search(context.Background(), "kb", []float32{1, 0, 0}, 5)
	for _, h := range hits {
		if h.ID == "a" {
			t.Fatalf("deleted doc still present")
		}
	}
}

func TestEchoLLM(t *testing.T) {
	e := NewEchoLLMProvider()
	resp, err := e.Chat(context.Background(), domain.ChatRequest{
		Messages: []domain.ChatMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if resp.Content != "echo: hi" {
		t.Fatalf("content = %q", resp.Content)
	}
	emb, err := e.Embed(context.Background(), domain.EmbedRequest{Inputs: []string{"x", "y"}})
	if err != nil || len(emb.Vectors) != 2 {
		t.Fatalf("embed = %+v, %v", emb, err)
	}
	rr, err := e.Rerank(context.Background(), domain.RerankRequest{Documents: []string{"d1", "d2", "d3"}})
	if err != nil || len(rr.Results) != 3 || rr.Results[0].Index != 0 {
		t.Fatalf("rerank = %+v, %v", rr, err)
	}
	ch, err := e.Stream(context.Background(), domain.ChatRequest{Messages: []domain.ChatMessage{{Role: "user", Content: "hey"}}})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	chunk, ok := <-ch
	if !ok || !chunk.Done || chunk.Delta != "echo: hey" {
		t.Fatalf("chunk = %+v ok=%v", chunk, ok)
	}
}

func TestLocalAgentRuntime(t *testing.T) {
	a := NewLocalAgentRuntime()
	spec := domain.NewAgentSpec("demo").Tenant("t").Build()
	h, err := a.Load(context.Background(), spec)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if h.Spec == nil || h.ID == "" {
		t.Fatalf("handle = %+v", h)
	}
	out, err := a.Run(context.Background(), h, map[string]any{"query": "q"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out["agent"] != "demo" || out["query"] != "q" {
		t.Fatalf("out = %+v", out)
	}
	if _, err := a.Load(context.Background(), nil); err == nil {
		t.Fatalf("expected error on nil spec")
	}
}

func TestLocalAuth(t *testing.T) {
	au := NewLocalAuth()
	tc, err := au.ValidateToken(context.Background(), "tok")
	if err != nil || tc.TenantID != "local" {
		t.Fatalf("tc = %+v, %v", tc, err)
	}
	if _, err := au.ValidateToken(context.Background(), ""); err == nil {
		t.Fatalf("expected error on blank token")
	}
}

func TestNoOpTracing(t *testing.T) {
	tr := NewNoOpTracing()
	ctx, span := tr.StartSpan(context.Background(), "op")
	if span.Name != "op" || span.StartNanos == 0 {
		t.Fatalf("span = %+v", span)
	}
	if ctx == nil {
		t.Fatalf("expected non-nil context")
	}
}

func TestLoggingRAG(t *testing.T) {
	r := NewLoggingRAG()
	hits, err := r.Retrieve(context.Background(), "q", "kb", 5)
	if err != nil || hits != nil {
		t.Fatalf("hits = %+v, %v", hits, err)
	}
}

func TestHTTPGatewayInvoke(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	g, err := NewHTTPGateway(srv.URL, "secret").Invoke(context.Background(), domain.GatewayRequest{
		Path:   "/v1/chat",
		Method: http.MethodPost,
		Body:   map[string]any{"msg": "hi"},
	})
	if err != nil {
		t.Fatalf("invoke: %v", err)
	}
	if g.Status != http.StatusOK || g.Body != `{"ok":true}` {
		t.Fatalf("resp = %+v", g)
	}
}
