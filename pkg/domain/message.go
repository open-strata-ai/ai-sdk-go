package domain

// ChatMessage is a single chat turn passed to LLMProvider.
type ChatMessage struct {
	Role    string
	Content string
}

// ChatRequest is an OpenAI-compatible chat completion request.
type ChatRequest struct {
	Messages    []ChatMessage
	Model       string
	Temperature *float64
	Extra       map[string]any
}

// ChatResponse is a chat completion response.
type ChatResponse struct {
	Content string
	Model   string
	Usage   map[string]any
}

// EmbedRequest is an embedding request.
type EmbedRequest struct {
	Inputs []string
	Model  string
}

// EmbedResponse is one float vector per input.
type EmbedResponse struct {
	Vectors [][]float32
}

// RerankRequest is a rerank request.
type RerankRequest struct {
	Query     string
	Documents []string
	Model     string
}

// RerankItem is a single reranked document.
type RerankItem struct {
	Index int
	Score float64
}

// RerankResponse is a rerank response.
type RerankResponse struct {
	Results []RerankItem
}

// GatewayRequest is a raw invoke against the platform Gateway.
type GatewayRequest struct {
	Path    string
	Method  string
	Body    any
	Headers map[string]string
}

// GatewayResponse is the Gateway response.
type GatewayResponse struct {
	Status  int
	Body    string
	Headers map[string]string
}

// Doc is a document upserted into a VectorStore.
type Doc struct {
	ID     string
	Text   string
	Vector []float32
}

// Hit is a vector search hit.
type Hit struct {
	ID    string
	Text  string
	Score float64
}

// RAGHit is a retrieval hit from the RAG SPI.
type RAGHit struct {
	KbID  string
	DocID string
	Text  string
	Score float64
}

// StreamChunk is one streaming delta from LLMProvider.Stream.
type StreamChunk struct {
	Delta string
	Done  bool
}

// Span is a tracing span handle from Tracing.
type Span struct {
	ID         string
	Name       string
	StartNanos int64
}

// AgentHandle is a handle for a loaded Agent, bound to its AgentSpec.
type AgentHandle struct {
	ID   string
	Spec *AgentSpec
}

// TenantContext is the resolved tenant context (Auth SPI output).
type TenantContext struct {
	TenantID string
	Roles    []string
	Claims   map[string]string
}

// TenantConfig is the per-tenant configuration resolved by MultiTenancy.
type TenantConfig struct {
	TenantID              string
	VectorStorePreference string
	CacheProvider         string
}

// DeploySpec is a canary/cicd deployment specification (CICD SPI).
type DeploySpec struct {
	Name   string
	Image  string
	Labels map[string]string
}

// DeployHandle is a handle for a submitted deployment.
type DeployHandle struct {
	DeployID string
	Status   string
}

// FineTuneSpec is a fine-tuning specification (MLOps SPI).
type FineTuneSpec struct {
	BaseModel   string
	Dataset     string
	Hyperparams map[string]any
}

// FineTuneHandle is a handle for a submitted fine-tune job.
type FineTuneHandle struct {
	JobID  string
	Status string
}

// WorkflowSpec is a long-running workflow specification (Workflow SPI).
type WorkflowSpec struct {
	WorkflowID string
	Spec       map[string]any
}

// WorkflowHandle is a handle for a submitted workflow run.
type WorkflowHandle struct {
	RunID  string
	Status string
}

// EvalSpec is an evaluation specification (Eval SPI).
type EvalSpec struct {
	Name     string
	Dataset  string
	Criteria map[string]any
}

// EvalReport is an evaluation report (Eval SPI output).
type EvalReport struct {
	Name    string
	Score   float64
	Details map[string]any
}

// SandboxResult is the result of an isolated code execution (Sandbox SPI).
type SandboxResult struct {
	Stdout   string
	ExitCode int
	TimedOut bool
}
