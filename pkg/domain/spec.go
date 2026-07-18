package domain

import (
	"context"
	"fmt"
	"time"
)

// AgentAPIVersion is the AgentSpec apiVersion (§4.3.5).
const AgentAPIVersion = "openstrata.cc/v1"

// AgentKind is the AgentSpec kind.
const AgentKind = "Agent"

// ObjectSchema is a minimal JSON-Schema-style field descriptor.
type ObjectSchema struct {
	Name string
	Type string
}

// NewObjectSchema builds an ObjectSchema.
func NewObjectSchema(name, typ string) *ObjectSchema {
	return &ObjectSchema{Name: name, Type: typ}
}

// ModelBinding is a declarative model binding (primary + ordered fallbacks).
type ModelBinding struct {
	Primary  string
	Fallback []string
}

// Guardrails are guardrail checks attached to an AgentSpec.
type Guardrails struct {
	Checks []string
}

// ToolBinding references a bound tool for tool_bindings.
type ToolBinding struct {
	ToolName string
	Ref      string
}

// Metadata is the AgentSpec metadata (name/tenant/labels).
type Metadata struct {
	Name     string
	TenantID string
	Labels   map[string]string
}

// AgentSpec is the declarative Agent specification (apiVersion: openstrata.cc/v1).
// Language/runtime independent; bound to any AgentRuntime instance for execution.
type AgentSpec struct {
	APIVersion     string
	Kind           string
	Metadata       Metadata
	ModelBinding   *ModelBinding
	InputSchema    *ObjectSchema
	OutputSchema   *ObjectSchema
	ToolBindings   []ToolBinding
	MemoryBindings []string
	StateMachine   string
	Guardrails     *Guardrails
}

// AgentSpecBuilder is the fluent builder for AgentSpec (DESIGN §4.1).
type AgentSpecBuilder struct {
	name           string
	tenantID       string
	modelBinding   *ModelBinding
	inputSchema    *ObjectSchema
	outputSchema   *ObjectSchema
	toolBindings   []ToolBinding
	memoryBindings []string
	stateMachine   string
	guardrails     *Guardrails
}

// NewAgentSpec starts building an AgentSpec with the given name.
func NewAgentSpec(name string) *AgentSpecBuilder {
	return &AgentSpecBuilder{name: name, guardrails: &Guardrails{}}
}

// Tenant sets the tenant id.
func (b *AgentSpecBuilder) Tenant(tenantID string) *AgentSpecBuilder {
	b.tenantID = tenantID
	return b
}

// ModelBinding sets the model binding.
func (b *AgentSpecBuilder) ModelBinding(mb *ModelBinding) *AgentSpecBuilder {
	b.modelBinding = mb
	return b
}

// InputSchema sets the input schema.
func (b *AgentSpecBuilder) InputSchema(s *ObjectSchema) *AgentSpecBuilder {
	b.inputSchema = s
	return b
}

// OutputSchema sets the output schema.
func (b *AgentSpecBuilder) OutputSchema(s *ObjectSchema) *AgentSpecBuilder {
	b.outputSchema = s
	return b
}

// ToolBindings sets the tool bindings.
func (b *AgentSpecBuilder) ToolBindings(tb []ToolBinding) *AgentSpecBuilder {
	b.toolBindings = tb
	return b
}

// MemoryBindings sets the memory bindings.
func (b *AgentSpecBuilder) MemoryBindings(mb []string) *AgentSpecBuilder {
	b.memoryBindings = mb
	return b
}

// StateMachine sets the state machine reference.
func (b *AgentSpecBuilder) StateMachine(sm string) *AgentSpecBuilder {
	b.stateMachine = sm
	return b
}

// Guardrails sets the guardrails.
func (b *AgentSpecBuilder) Guardrails(g *Guardrails) *AgentSpecBuilder {
	b.guardrails = g
	return b
}

// Build materializes the AgentSpec.
func (b *AgentSpecBuilder) Build() *AgentSpec {
	return &AgentSpec{
		APIVersion:     AgentAPIVersion,
		Kind:           AgentKind,
		Metadata:       Metadata{Name: b.name, TenantID: b.tenantID, Labels: map[string]string{}},
		ModelBinding:   b.modelBinding,
		InputSchema:    b.inputSchema,
		OutputSchema:   b.outputSchema,
		ToolBindings:   b.toolBindings,
		MemoryBindings: b.memoryBindings,
		StateMachine:   b.stateMachine,
		Guardrails:     b.guardrails,
	}
}

// ToolContext is the execution context handed to a ToolHandler.
type ToolContext struct {
	TenantID string
}

// ToolHandler is the local tool execution contract.
type ToolHandler func(ctx ToolContext, input map[string]any) (map[string]any, error)

// Tool is a tool declaration (MCP stdio/SSE/HTTP) bound to tool_bindings,
// executed locally or registered to the ToolRegistry.
type Tool struct {
	Name        string
	Description string
	InputSchema *ObjectSchema
	Handler     ToolHandler
}

// ToolBuilder is the fluent builder for Tool.
type ToolBuilder struct {
	name        string
	description string
	inputSchema *ObjectSchema
	handler     ToolHandler
}

// NewTool starts building a Tool with the given name.
func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{name: name, handler: func(ToolContext, map[string]any) (map[string]any, error) {
		return map[string]any{}, nil
	}}
}

// Description sets the tool description.
func (b *ToolBuilder) Description(d string) *ToolBuilder {
	b.description = d
	return b
}

// InputSchema sets the tool input schema.
func (b *ToolBuilder) InputSchema(s *ObjectSchema) *ToolBuilder {
	b.inputSchema = s
	return b
}

// Handler sets the tool handler.
func (b *ToolBuilder) Handler(h ToolHandler) *ToolBuilder {
	b.handler = h
	return b
}

// Build materializes the Tool.
func (b *ToolBuilder) Build() *Tool {
	return &Tool{Name: b.name, Description: b.description, InputSchema: b.inputSchema, Handler: b.handler}
}

// Run executes the tool's handler.
func (t *Tool) Run(ctx ToolContext, input map[string]any) (map[string]any, error) {
	if t.Handler == nil {
		return nil, fmt.Errorf("openstrata: tool %q has no handler", t.Name)
	}
	return t.Handler(ctx, input)
}

// Session is a multi-turn conversation/memory handle backed by the Cache SPI.
// Keys are namespaced per session id.
type Session struct {
	sessionID string
	cache     Cache
	ttl       time.Duration
}

// NewSession creates a Session bound to a Cache.
func NewSession(sessionID string, cache Cache) *Session {
	return &Session{sessionID: sessionID, cache: cache, ttl: time.Hour}
}

// SessionID returns the session id.
func (s *Session) SessionID() string { return s.sessionID }

// SetWorkingMemoryTTL sets the TTL for working-memory entries.
func (s *Session) SetWorkingMemoryTTL(ttl time.Duration) { s.ttl = ttl }

// Put stores a value under the session-namespaced key.
func (s *Session) Put(key string, value []byte) error {
	return s.cache.Set(context.Background(), s.prefix(key), value, s.ttl)
}

// Get retrieves a value under the session-namespaced key.
func (s *Session) Get(key string) ([]byte, bool, error) {
	return s.cache.Get(context.Background(), s.prefix(key))
}

func (s *Session) prefix(key string) string {
	return "session:" + s.sessionID + ":" + key
}

// Retriever is a RAG retrieval facade over VectorStore and RAG SPIs.
type Retriever struct {
	vs  VectorStore
	rag RAG
}

// NewRetriever builds a Retriever.
func NewRetriever(vs VectorStore, rag RAG) *Retriever {
	return &Retriever{vs: vs, rag: rag}
}

// Search runs a vector search.
func (r *Retriever) Search(collection string, vec []float32, topK int) ([]Hit, error) {
	return r.vs.Search(context.Background(), collection, vec, topK)
}

// Retrieve runs a RAG retrieval.
func (r *Retriever) Retrieve(query string, kbID string, topK int) ([]RAGHit, error) {
	return r.rag.Retrieve(context.Background(), query, kbID, topK)
}
