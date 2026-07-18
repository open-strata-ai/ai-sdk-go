package adapter

import (
	"context"
	"fmt"

	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

// LoggingRAG is a default RAG adapter (no-op retrieval).
type LoggingRAG struct{}

// NewLoggingRAG builds a LoggingRAG.
func NewLoggingRAG() *LoggingRAG { return &LoggingRAG{} }

func (r *LoggingRAG) Retrieve(ctx context.Context, query, kbID string, topK int) ([]domain.RAGHit, error) {
	return nil, nil
}

// LoggingGateway is a default Gateway adapter that echoes a 200 (no network).
type LoggingGateway struct{}

// NewLoggingGateway builds a LoggingGateway.
func NewLoggingGateway() *LoggingGateway { return &LoggingGateway{} }

func (g *LoggingGateway) Invoke(ctx context.Context, req domain.GatewayRequest) (domain.GatewayResponse, error) {
	return domain.GatewayResponse{Status: 200, Body: "{}", Headers: map[string]string{}}, nil
}

// LoggingLowCode exports a minimal AgentSpec for a canvas id.
type LoggingLowCode struct{}

// NewLoggingLowCode builds a LoggingLowCode.
func NewLoggingLowCode() *LoggingLowCode { return &LoggingLowCode{} }

func (l *LoggingLowCode) ExportSpec(ctx context.Context, canvasID string) (*domain.AgentSpec, error) {
	return domain.NewAgentSpec(fmt.Sprintf("canvas-%s", canvasID)).Build(), nil
}

// LoggingWorkflow is a default Workflow adapter.
type LoggingWorkflow struct{}

// NewLoggingWorkflow builds a LoggingWorkflow.
func NewLoggingWorkflow() *LoggingWorkflow { return &LoggingWorkflow{} }

func (w *LoggingWorkflow) Submit(ctx context.Context, spec domain.WorkflowSpec) (domain.WorkflowHandle, error) {
	return domain.WorkflowHandle{RunID: "local", Status: "submitted"}, nil
}

// LoggingSandbox is a default Sandbox adapter.
type LoggingSandbox struct{}

// NewLoggingSandbox builds a LoggingSandbox.
func NewLoggingSandbox() *LoggingSandbox { return &LoggingSandbox{} }

func (s *LoggingSandbox) Execute(ctx context.Context, code, language string) (domain.SandboxResult, error) {
	return domain.SandboxResult{ExitCode: 0, TimedOut: false}, nil
}

// LoggingCICD is a default CICD adapter.
type LoggingCICD struct{}

// NewLoggingCICD builds a LoggingCICD.
func NewLoggingCICD() *LoggingCICD { return &LoggingCICD{} }

func (c *LoggingCICD) Deploy(ctx context.Context, spec domain.DeploySpec) (domain.DeployHandle, error) {
	return domain.DeployHandle{DeployID: "local", Status: "deployed"}, nil
}

// LoggingMLOps is a default MLOps adapter.
type LoggingMLOps struct{}

// NewLoggingMLOps builds a LoggingMLOps.
func NewLoggingMLOps() *LoggingMLOps { return &LoggingMLOps{} }

func (m *LoggingMLOps) SubmitFineTune(ctx context.Context, spec domain.FineTuneSpec) (domain.FineTuneHandle, error) {
	return domain.FineTuneHandle{JobID: "local", Status: "submitted"}, nil
}

// LoggingEval is a default Eval adapter.
type LoggingEval struct{}

// NewLoggingEval builds a LoggingEval.
func NewLoggingEval() *LoggingEval { return &LoggingEval{} }

func (e *LoggingEval) RunEval(ctx context.Context, spec domain.EvalSpec) (domain.EvalReport, error) {
	return domain.EvalReport{Name: spec.Name, Score: 0, Details: map[string]any{}}, nil
}

// LoggingMultiTenancy is a default MultiTenancy adapter.
type LoggingMultiTenancy struct{}

// NewLoggingMultiTenancy builds a LoggingMultiTenancy.
func NewLoggingMultiTenancy() *LoggingMultiTenancy { return &LoggingMultiTenancy{} }

func (m *LoggingMultiTenancy) ResolveTenant(ctx context.Context, tenantID string) (domain.TenantConfig, error) {
	return domain.TenantConfig{TenantID: tenantID, VectorStorePreference: "qdrant", CacheProvider: "redis"}, nil
}
