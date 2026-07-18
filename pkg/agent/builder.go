// Package agent is the application layer of ai-sdk-go: it orchestrates the
// declarative AgentSpec (pkg/domain) and binds it to the AgentRuntime SPI.
package agent

import (
	"context"

	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

// AgentBuilder declaratively assembles an AgentSpec and binds it to an
// AgentRuntime for load/run (runtime independent).
type AgentBuilder struct {
	rt domain.AgentRuntime
	sb *domain.AgentSpecBuilder
}

// New creates an AgentBuilder bound to the given runtime.
func New(rt domain.AgentRuntime, name string) *AgentBuilder {
	return &AgentBuilder{rt: rt, sb: domain.NewAgentSpec(name)}
}

// Tenant sets the tenant id.
func (b *AgentBuilder) Tenant(tenantID string) *AgentBuilder {
	b.sb.Tenant(tenantID)
	return b
}

// ModelBinding sets the model binding.
func (b *AgentBuilder) ModelBinding(mb *domain.ModelBinding) *AgentBuilder {
	b.sb.ModelBinding(mb)
	return b
}

// InputSchema sets the input schema.
func (b *AgentBuilder) InputSchema(s *domain.ObjectSchema) *AgentBuilder {
	b.sb.InputSchema(s)
	return b
}

// OutputSchema sets the output schema.
func (b *AgentBuilder) OutputSchema(s *domain.ObjectSchema) *AgentBuilder {
	b.sb.OutputSchema(s)
	return b
}

// Guardrails sets the guardrails.
func (b *AgentBuilder) Guardrails(g *domain.Guardrails) *AgentBuilder {
	b.sb.Guardrails(g)
	return b
}

// ToolBindings sets the tool bindings.
func (b *AgentBuilder) ToolBindings(tb []domain.ToolBinding) *AgentBuilder {
	b.sb.ToolBindings(tb)
	return b
}

// MemoryBindings sets the memory bindings.
func (b *AgentBuilder) MemoryBindings(mb []string) *AgentBuilder {
	b.sb.MemoryBindings(mb)
	return b
}

// StateMachine sets the state machine reference.
func (b *AgentBuilder) StateMachine(sm string) *AgentBuilder {
	b.sb.StateMachine(sm)
	return b
}

// Build materializes the AgentSpec.
func (b *AgentBuilder) Build() *domain.AgentSpec {
	return b.sb.Build()
}

// Load builds and loads the AgentSpec onto the runtime.
func (b *AgentBuilder) Load(ctx context.Context) (domain.AgentHandle, error) {
	return b.rt.Load(ctx, b.Build())
}

// Run builds, loads, and runs the agent with the given input.
func (b *AgentBuilder) Run(ctx context.Context, input map[string]any) (map[string]any, error) {
	h, err := b.Load(ctx)
	if err != nil {
		return nil, err
	}
	return b.rt.Run(ctx, h, input)
}
