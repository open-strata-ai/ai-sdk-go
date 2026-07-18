package domain

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestAgentSpecBuilder(t *testing.T) {
	spec := NewAgentSpec("customer-service-v2").
		Tenant("tenant-b").
		ModelBinding(&ModelBinding{Primary: "cloud-qwen-max"}).
		InputSchema(NewObjectSchema("query", "string")).
		OutputSchema(NewObjectSchema("answer", "string")).
		Guardrails(&Guardrails{Checks: []string{"injection_scan"}}).
		MemoryBindings([]string{"mem1"}).
		ToolBindings([]ToolBinding{{ToolName: "order_lookup"}}).
		Build()

	if spec.APIVersion != AgentAPIVersion {
		t.Fatalf("apiVersion = %q, want %q", spec.APIVersion, AgentAPIVersion)
	}
	if spec.Kind != AgentKind {
		t.Fatalf("kind = %q, want %q", spec.Kind, AgentKind)
	}
	if spec.Metadata.Name != "customer-service-v2" || spec.Metadata.TenantID != "tenant-b" {
		t.Fatalf("metadata = %+v", spec.Metadata)
	}
	if spec.ModelBinding == nil || spec.ModelBinding.Primary != "cloud-qwen-max" {
		t.Fatalf("modelBinding = %+v", spec.ModelBinding)
	}
	if spec.InputSchema == nil || spec.InputSchema.Name != "query" {
		t.Fatalf("inputSchema = %+v", spec.InputSchema)
	}
	if len(spec.MemoryBindings) != 1 || len(spec.ToolBindings) != 1 {
		t.Fatalf("bindings not set: mem=%v tool=%v", spec.MemoryBindings, spec.ToolBindings)
	}
	if spec.Guardrails == nil || len(spec.Guardrails.Checks) != 1 {
		t.Fatalf("guardrails = %+v", spec.Guardrails)
	}
}

func TestToolRun(t *testing.T) {
	tool := NewTool("order_lookup").
		Description("Check order logistics status").
		InputSchema(NewObjectSchema("order_id", "string")).
		Handler(func(ctx ToolContext, in map[string]any) (map[string]any, error) {
			return map[string]any{"status": "shipped"}, nil
		}).Build()

	out, err := tool.Run(ToolContext{TenantID: "t1"}, map[string]any{"order_id": "123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["status"] != "shipped" {
		t.Fatalf("out = %+v", out)
	}
	if tool.Name != "order_lookup" {
		t.Fatalf("name = %q", tool.Name)
	}
}

func TestToolRunNoHandler(t *testing.T) {
	tool := NewTool("broken").Handler(nil).Build()
	if _, err := tool.Run(ToolContext{}, map[string]any{}); err == nil {
		t.Fatalf("expected error for nil handler")
	}
}

func TestGuardrailsAndModelBinding(t *testing.T) {
	g := &Guardrails{Checks: []string{"a", "b"}}
	if len(g.Checks) != 2 {
		t.Fatalf("checks = %v", g.Checks)
	}
	mb := &ModelBinding{Primary: "x", Fallback: []string{"y", "z"}}
	if mb.Primary != "x" || len(mb.Fallback) != 2 {
		t.Fatalf("binding = %+v", mb)
	}
}

func TestErrorWrap(t *testing.T) {
	base := errors.New("boom")
	e := NewError("LLMProvider", "RATE_LIMIT", true, base)
	if !errors.Is(e, base) {
		t.Fatalf("expected wrapped error to match base")
	}
	if !e.Retryable {
		t.Fatalf("expected Retryable = true")
	}
	if e.Port != "LLMProvider" || e.Code != "RATE_LIMIT" {
		t.Fatalf("error = %+v", e)
	}
}

func TestSessionRoundtrip(t *testing.T) {
	// Zero TTL -> entry expires immediately on read.
	s := NewSession("user-1", &fakeCache{})
	s.SetWorkingMemoryTTL(0)
	if err := s.Put("k", []byte("v")); err != nil {
		t.Fatalf("put error: %v", err)
	}
	if _, ok, _ := s.Get("k"); ok {
		t.Fatalf("expected expired entry to be missing")
	}

	// Real TTL -> entry survives.
	s2 := NewSession("user-2", &fakeCache{})
	s2.SetWorkingMemoryTTL(time.Hour)
	if err := s2.Put("k", []byte("v")); err != nil {
		t.Fatalf("put error: %v", err)
	}
	if v, ok, _ := s2.Get("k"); !ok || string(v) != "v" {
		t.Fatalf("expected stored value, got ok=%v v=%q", ok, v)
	}
}

// fakeCache is a minimal TTL-aware Cache used only by tests (avoids an import
// cycle with the adapter package).
type fakeCache struct {
	mu    sync.Mutex
	store map[string]fakeEntry
}

type fakeEntry struct {
	value []byte
	exp   time.Time
}

func (f *fakeCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.store[key]
	if !ok {
		return nil, false, nil
	}
	if time.Now().After(e.exp) {
		delete(f.store, key)
		return nil, false, nil
	}
	cp := append([]byte(nil), e.value...)
	return cp, true, nil
}

func (f *fakeCache) Set(_ context.Context, key string, val []byte, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.store == nil {
		f.store = map[string]fakeEntry{}
	}
	cp := append([]byte(nil), val...)
	f.store[key] = fakeEntry{value: cp, exp: time.Now().Add(ttl)}
	return nil
}
