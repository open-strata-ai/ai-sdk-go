package adapter

import (
	"context"
	"fmt"

	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

// LocalAgentRuntime is the default AgentRuntime adapter: local in-process executor.
type LocalAgentRuntime struct{}

// NewLocalAgentRuntime builds a LocalAgentRuntime.
func NewLocalAgentRuntime() *LocalAgentRuntime { return &LocalAgentRuntime{} }

func (a *LocalAgentRuntime) Load(ctx context.Context, spec *domain.AgentSpec) (domain.AgentHandle, error) {
	if spec == nil {
		return domain.AgentHandle{}, fmt.Errorf("openstrata: nil AgentSpec")
	}
	return domain.AgentHandle{ID: fmt.Sprintf("local-%s", spec.Metadata.Name), Spec: spec}, nil
}

func (a *LocalAgentRuntime) Run(ctx context.Context, handle domain.AgentHandle, input map[string]any) (map[string]any, error) {
	out := map[string]any{}
	for k, v := range input {
		out[k] = v
	}
	name := ""
	if handle.Spec != nil {
		name = handle.Spec.Metadata.Name
	}
	out["agent"] = name
	return out, nil
}
