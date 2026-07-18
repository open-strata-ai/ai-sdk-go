// Package domain defines the OpenStrata SDK's framework-free domain layer:
// the 15 platform SPI port interfaces (dependency inversion) plus the value
// types they exchange. It has no dependency on any external component, so the
// SDK core is zero-modified when a host injects a custom adapter.
package domain

import (
	"context"
	"time"
)

// InterfaceVersion is the SPI contract version this SDK aligns with
// (platform bom.yaml §16; strata v1.0.0).
const InterfaceVersion = "1.0.0"

// LLMProvider - platform LLMProvider SPI (§4.4.4), interface_versions 1.0.0.
type LLMProvider interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error)
	Rerank(ctx context.Context, req RerankRequest) (RerankResponse, error)
	Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
}

// VectorStore - platform VectorStore SPI (§10.4), interface_versions 1.0.0.
type VectorStore interface {
	Upsert(ctx context.Context, collection string, docs []Doc) error
	Search(ctx context.Context, collection string, vec []float32, topK int) ([]Hit, error)
	Delete(ctx context.Context, collection string, ids []string) error
}

// AgentRuntime - platform AgentRuntime SPI (§4.3.5/§10.6), interface_versions 1.0.0.
type AgentRuntime interface {
	Load(ctx context.Context, spec *AgentSpec) (AgentHandle, error)
	Run(ctx context.Context, handle AgentHandle, input map[string]any) (map[string]any, error)
}

// Cache - platform Cache SPI (§4.3.4), interface_versions 1.0.0.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

// Gateway - platform Gateway SPI (§4.4.1), interface_versions 1.0.0.
type Gateway interface {
	Invoke(ctx context.Context, req GatewayRequest) (GatewayResponse, error)
}

// Tracing - platform Tracing SPI (§4.8), interface_versions 1.0.0.
type Tracing interface {
	StartSpan(ctx context.Context, name string) (context.Context, Span)
}

// Auth - platform Auth SPI, interface_versions 1.0.0.
type Auth interface {
	ValidateToken(ctx context.Context, token string) (TenantContext, error)
}

// RAG - platform RAG SPI, interface_versions 1.0.0.
type RAG interface {
	Retrieve(ctx context.Context, query string, kbID string, topK int) ([]RAGHit, error)
}

// LowCode - platform LowCode SPI, interface_versions 1.0.0.
type LowCode interface {
	ExportSpec(ctx context.Context, canvasID string) (*AgentSpec, error)
}

// Workflow - platform Workflow SPI, interface_versions 1.0.0.
type Workflow interface {
	Submit(ctx context.Context, spec WorkflowSpec) (WorkflowHandle, error)
}

// Sandbox - platform Sandbox SPI, interface_versions 1.0.0.
type Sandbox interface {
	Execute(ctx context.Context, code string, language string) (SandboxResult, error)
}

// CICD - platform CICD SPI, interface_versions 1.0.0.
type CICD interface {
	Deploy(ctx context.Context, spec DeploySpec) (DeployHandle, error)
}

// MultiTenancy - platform MultiTenancy SPI, interface_versions 1.0.0.
type MultiTenancy interface {
	ResolveTenant(ctx context.Context, tenantID string) (TenantConfig, error)
}

// MLOps - platform MLOps SPI, interface_versions 1.0.0.
type MLOps interface {
	SubmitFineTune(ctx context.Context, spec FineTuneSpec) (FineTuneHandle, error)
}

// Eval - platform Eval SPI, interface_versions 1.0.0.
type Eval interface {
	RunEval(ctx context.Context, spec EvalSpec) (EvalReport, error)
}

// Error is the SDK's unified error model (DESIGN §9.1).
type Error struct {
	Code      string // platform error code (Gateway response)
	Port      string // the SPI port name that produced the error
	Retryable bool   // whether the upper layer may fall back (ModelRouter)
	Err       error  // underlying error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return "openstrata: port=" + e.Port + " code=" + e.Code + ": " + e.Err.Error()
	}
	return "openstrata: port=" + e.Port + " code=" + e.Code
}

func (e *Error) Unwrap() error { return e.Err }

// NewError builds an *Error.
func NewError(port, code string, retryable bool, err error) *Error {
	return &Error{Port: port, Code: code, Retryable: retryable, Err: err}
}
