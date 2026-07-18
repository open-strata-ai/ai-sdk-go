# ai-sdk-go · Architecture (ARCH)

> **Source**: docs/DESIGN.md §1 (Positioning) · §2 (Core Abstraction) · §6 (SPI Mapping 15 Port)
> **Audience**: Platform architects, SDK maintainers, AI coding agents
> **Collaboration**: docs/SKILLS.md (extension points and coding rules) · docs/SPECS.md (contracts and versions) · docs/adr/ (major decisions)
> **Platform version**: strata v1.0.0

---

## 1. Positioning and target users

### 1.1 What is SDK

`ai-sdk-go` is the official developer library of the OpenStrata platform for the **Go ecosystem**, published to Go Module Proxy (`github.com/open-strata-ai/ai-sdk-go`). It is a three-layer lightweight library of "client + builder + extension port", and is not a deployable service.

Core Commitments:
- **Minimal Invasion**: Only relies on Go standard library + very few necessary dependencies (`google/wire`, `gin`, `httpx`, `yaml`, `otel`), can be embedded in any Go host application
- **Declarative AgentSpec**: Agent behavior converges to `apiVersion: openstrata.cc/v1` AgentSpec, language/runtime independent (§4.3.5)
- **Dependency Inversion**: All platform SPIs are exposed through the `pkg/domain` interface, and the SDK core is zero-modified when the host injects custom adapters

### 1.2 What problem is solved?

Let Go developers work within existing Go applications (microservices, CLI, background tasks) in a minimally intrusive way:
1. Build a conversational Agent/Client and access platform runtime capabilities (gateway, Agent runtime, tool registration, memory, RAG, cache, observable)
2. Encapsulate internal capabilities into platform tools and connect business data to RAG
3. Implement customized `LLMProvider`/`VectorStore`/`Cache`/`AgentRuntime` adapter based on SDK’s SPI port

### 1.3 Target user portrait

| Character | Scene | Contact |
|------|------|--------|
| Go backend/platform engineer | Embed Agent in existing Go service | `pkg/client`, `pkg/agent` |
| Platform secondary developer | Implement custom SPI Adapter | `pkg/domain` Port interface |
| AI Coding Agent | Generate adaptation code according to skills/ rules | docs/SKILLS.md |

### 1.4 Boundary with platform

```
Host Go application ──→ ai-sdk-go pkg/ ──→ OpenStrata platform SPI
   (Already)          (Main repository)              (GateWay wait)
```

- The SDK only relies on the Go standard library + a very small number of necessary dependencies, ensuring that it can be embedded in any host (§15.5.1 Framework Convergence)
- The SDK itself does not bind any web framework; the host can be Gin/Hertz, and the SDK is adapted through the `adapter` layer
- SDK does not hold third-party Provider Key; the key is routed indirectly via the platform Secret Vault + Gateway (§4.4.6)

---

## 2. Core abstraction

### 2.1 Package structure (DDD four layers)

```
ai-sdk-go/
├── cmd/                     #Optional CLI samples (non-publishing subject)
├── pkg/                     #★ Release body (go module root package)
│   ├── client/              #① Access layer: GatewayClient (OpenAI-compatible)
│   ├── agent/               #② Application layer: AgentBuilder/Runner orchestration
│   ├── domain/              #③ Domain layer: AgentSpec / Tool / Session entity + Port interface
│   ├── adapter/             #④ Infrastructure layer: SPI Adapter (LLMProvider/VectorStore/Cache…)
│   └── config/              #Configuration loading (infrastructure/config fragment rendering)
├── infrastructure/config/   #This repository SPI adapter local configuration fragment
└── arch/ design/ skills/ specs/  #Evolutionary AI Coding Source of Truth
```

Hierarchical dependency: ①→②→③←④. The domain layer (③) does not depend on any framework/external components, and the dependence direction points from ①② inward to ③.

### 2.2 First-class citizen objects (5 external beans)

| Abstract | Package path | Corresponding platform SPI | Description |
|------|--------|-------------|------|
| `Client` | `pkg/client` | **Gateway** SPI (§4.4.1) | OpenAI-compatible call entry; holds tenant Token |
| `AgentBuilder` / `AgentSpec` | `pkg/agent` | **AgentRuntime** SPI (§4.3.5) | Declarative assembly AgentSpec (apiVersion/kind/metadata/model_binding/tool_bindings/memory_bindings/state_machine/guardrails) |
| `Tool` | `pkg/agent` | **ToolRegistry** (§4.3.2) | Declare JSON Schema, support MCP stdio/SSE/HTTP |
| `Session` | `pkg/client` | **Cache** SPI (§4.3.4) | Multi-round session/memory handle, binding memory_bindings |
| `Retriever` | `pkg/domain` | **VectorStore** SPI (§10.4) + **RAG** ​​SPI | Knowledge base retrieval, return hit fragments |

### 2.3 Domain layer Port interface (dependency inversion)

The domain layer defines the following SPI port interface in `pkg/domain`, and the name is strictly consistent with the platform §10.4 canonical port name**:

```go
package domain

import (
    "context"
    "time"
)

//LLMProvider - corresponding platform LLMProvider SPI (§4.4.4), interface_versions: 1.0.0
type LLMProvider interface {
    Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
    Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error)
    Rerank(ctx context.Context, req RerankRequest) (RerankResponse, error)
    Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
}

//VectorStore - corresponding platform VectorStore SPI (§10.4), interface_versions: 1.0.0
type VectorStore interface {
    Upsert(ctx context.Context, col string, docs []Doc) error
    Search(ctx context.Context, col string, vec []float32, topK int) ([]Hit, error)
    Delete(ctx context.Context, col string, ids []string) error
}

//AgentRuntime —— Corresponding platform AgentRuntime SPI (§10.6 / §4.3.5), interface_versions: 1.0.0
type AgentRuntime interface {
    Load(ctx context.Context, spec *AgentSpec) (AgentHandle, error)
    Run(ctx context.Context, h AgentHandle, input map[string]any) (map[string]any, error)
}

//Cache - corresponding platform Cache SPI (§4.3.4), interface_versions: 1.0.0
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, bool, error)
    Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

//Gateway - corresponding platform Gateway SPI (§4.4.1), interface_versions: 1.0.0
type Gateway interface {
    Invoke(ctx context.Context, req GatewayRequest) (GatewayResponse, error)
}

//Tracing - corresponding platform Tracing SPI (§4.8), interface_versions: 1.0.0
type Tracing interface {
    StartSpan(ctx context.Context, name string) (context.Context, Span)
}

//Auth——Corresponding platform Auth SPI, interface_versions: 1.0.0
type Auth interface {
    ValidateToken(ctx context.Context, token string) (TenantContext, error)
}

//RAG ——Corresponding platform RAG SPI, interface_versions: 1.0.0
type RAG interface {
    Retrieve(ctx context.Context, query string, kbID string, topK int) ([]RAGHit, error)
}

//LowCode —— Corresponding platform LowCode SPI, interface_versions: 1.0.0
type LowCode interface {
    ExportSpec(ctx context.Context, canvasID string) (*AgentSpec, error)
}

//Workflow——Corresponding platform Workflow SPI, interface_versions: 1.0.0
type Workflow interface {
    Submit(ctx context.Context, wf WorkflowSpec) (WorkflowHandle, error)
}

//Sandbox - corresponding platform Sandbox SPI, interface_versions: 1.0.0
type Sandbox interface {
    Execute(ctx context.Context, code string, lang string) (SandboxResult, error)
}

//CICD - corresponding platform CICD SPI, interface_versions: 1.0.0
type CICD interface {
    Deploy(ctx context.Context, spec DeploySpec) (DeployHandle, error)
}

//MultiTenancy - corresponding platform MultiTenancy SPI, interface_versions: 1.0.0
type MultiTenancy interface {
    ResolveTenant(ctx context.Context, tenantID string) (TenantConfig, error)
}

//MLOps - corresponding platform MLOps SPI, interface_versions: 1.0.0
type MLOps interface {
    SubmitFineTune(ctx context.Context, spec FineTuneSpec) (FineTuneHandle, error)
}

//Eval——Corresponding platform Eval SPI, interface_versions: 1.0.0
type Eval interface {
    RunEval(ctx context.Context, spec EvalSpec) (EvalReport, error)
}
```

> **Key Convention**: All port interfaces do not contain any specific framework/external component dependencies (§15.5.2 Domain layer prohibition). Interface method signatures are semantically consistent with Java/Python SDK (cross-language contract).

### 2.4 Assembly model

SDK adopts **Wire compile-time injection** + manual injection dual mode:

```go
//Method 1: Wire automatic assembly (recommendation, production)
c := client.New(
    client.WithGateway(adapter.MustHigressFromEnv()),
    client.WithLLMProvider(adapter.MustQwenFromEnv()),
    client.WithCache(adapter.NewRedisCache(cfg.Cache)),
)

//Method 2: Manually inject custom implementation (secondary development)
c := client.New(
    client.WithLLMProvider(&MySelfHosted{endpoint: "http://vllm:8000"}),
    client.WithVectorStore(&MyCustomVS{}),
)
```

---

## 3. Mapping with platform SPI (15 port→canonical→interface_versions)

> Corresponds to docs/DESIGN.md §6, strictly aligned with Platform §10.4.

### 3.1 Mapping table

| SDK Domain Port (pkg/domain) | Platform SPI Port (§10.4 canonical) | interface_versions | SDK Roles | Default Adapter (bom.yaml) |
| --- | --- | --- | --- | --- |
| `Gateway` | **Gateway** | 1.0.0 | Production(invoke) | Higress(core) |
| `AgentRuntime` | **AgentRuntime** | 1.0.0 | Production (Load/Run) | LangGraph/Spring AI binding execution (core) |
| `LLMProvider` | **LLMProvider** | 1.0.0 | Production (chat/embed/rerank/stream) | Qwen-Cloud / OpenAI / Claude (core third party) |
| `VectorStore` | **VectorStore** | 1.0.0 | Production (upsert/search/delete) | Qdrant (core)/Milvus (optional) |
| `Cache` | **Cache** | 1.0.0 | Production (semantic/accurate caching) | Redis (core)/Valkey (optional, OSI) |
| `Auth` | **Auth** | 1.0.0 | Consumption (tenant Token verification) | Keycloak (core) |
| `Tracing` | **Tracing** | 1.0.0 | Production (span) | Langfuse (core) / OTel (core baseline) |
| `RAG` | **RAG** ​​| 1.0.0 | Consumption (retrieval backfill) | RAGFlow (core) |
| `LowCode` | **LowCode** | 1.0.0 | Consumption (canvas→AgentSpec) | Self-developed React Flow (core)/Dify (reference) |
| `Workflow` | **Workflow** | 1.0.0 | Consumption (long task orchestration) | Temporal (optional) |
| `Sandbox` | **Sandbox** | 1.0.0 | Consumption (code execution isolation) | Kata / E2B (optional) |
| `CICD` | **CICD** | 1.0.0 | Consumption (canary Release) | ArgoCD/Istio (optional) |
| `MultiTenancy` | **MultiTenancy** | 1.0.0 | Consumption (Tenant Isolation Context) | Capsule (optional) |
| `MLOps` | **MLOps** | 1.0.0 | Consume (fine-tuning/distillation) | MLflow (optional) |
| `Eval` | **Eval** | 1.0.0 | Consumption (evaluation reflow) | Promptfoo / DeepEval / Ragas (core/optional) |

### 3.2 Role classification

| SDK role | Meaning | Ports involved |
|----------|------|----------|
| **Producer** | SDK actively calls platform services, generates data or performs operations | Gateway, AgentRuntime, LLMProvider, VectorStore, Cache, Tracing |
| **Consumer** | The context/configuration issued by the SDK consumer platform and not actively modified | Auth, RAG, LowCode, Workflow, Sandbox, CICD, MultiTenancy, MLOps, Eval |

### 3.3 Key constraints

1. **The port name is word-for-word**: The SDK port name is exactly the same as the `bom.yaml` `spi` field (§16.2); new/replacement implementations never need to change the core (§10.6 Registry model)
2. **Same-level failover**: Self-hosting and multiple third-party implementations coexist behind the same `LLMProvider` SPI. Cross-implementation failover is handled by the platform `ModelRouter` (§4.4.4/§4.4.5)
3. **VectorStore silent switching**: When switching the VectorStore provider, the platform migrates the Job double-write verification (§10.4), the SDK obtains the corresponding Adapter through the SPI factory, and the application code is zero-changed
4. **Configuration override priority**: Explicit Option > Environment variables > Configuration file > Platform Manifest delivery (§12)

### 3.4 Relationship between ports (data flow)

```
Gateway ──→ LLMProvider ──→ (chat/embed/rerank/stream)
Gateway ──→ VectorStore ──→ RAG ──→ (Search backfill prompt)
Gateway ──→ Cache ──→ Session ──→ (session memory)
AgentRuntime ──→ Tracing ──→ (span Report)
Gateway ──→ Auth ──→ MultiTenancy ──→ (Tenant context)
Workflow ──→ Sandbox ──→ (Long task code isolation)
MLOps ──→ Eval ──→ LowCode ──→ (fine-tuning→Review→Canvas export)
```

### 3.5 Go-specific constraints

- All Port interfaces use `context.Context` as the first parameter (Go convention)
- `Stream` returns `<-chan StreamChunk` (Go channel, non-blocking pipe)
- Injection via `client.WithXxx(...)` Option pattern
- Wire compile-time injection: dependencies are declared in `wire.go`, and `go generate` generates initialization code

---

> **Associated documents**: docs/SKILLS.md (extension points and coding rules) · docs/SPECS.md (SPI version contract and configuration keys) · docs/DESIGN.md (complete design)
