# ai-sdk-go · AI Coding Skills (SKILLS)

> **Source**: design/DESIGN.md §5 (Extension Points) · §4 (Key Usage and Code Examples) · §9 (Error Handling and Observability)
> **Audience**: AI coding Agent, SDK secondary developers
> **Collaboration**: arch/ARCH.md (architecture) · specs/SPECS.md (contract) · design/adr/ (decision record)
> **Platform version**: strata v1.0.0

---

## 1. Extension points (SPI port implementation guide)

> Corresponds to design/DESIGN.md §5. The extension point of the SDK is the platform SPI port; the name is consistent with the §10.4 canonical port name.
> The host application only needs to implement the corresponding `pkg/domain` interface and inject it through `client.WithXxx(...)` to replace the default implementation - **zero changes to the SDK core** (dependency inversion + anti-corrosion layer ACL).

### 1.1 Access custom LLMProvider

```go
type MySelfHosted struct {
    endpoint string
    client   *http.Client
}

func (m *MySelfHosted) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
    //Tune your own vLLM/TGI OpenAI-compatible endpoint
    reqBody, _ := json.Marshal(req.ToOpenAICompat())
    resp, err := m.client.Post(m.endpoint+"/v1/chat/completions", "application/json",
        bytes.NewReader(reqBody))
    //... Parse the response and return domain.ChatResponse
}
func (m *MySelfHosted) Embed(ctx context.Context, req domain.EmbedRequest) (domain.EmbedResponse, error) { /* ... */ }
func (m *MySelfHosted) Rerank(ctx context.Context, req domain.RerankRequest) (domain.RerankResponse, error) { /* ... */ }
func (m *MySelfHosted) Stream(ctx context.Context, req domain.ChatRequest) (<-chan domain.StreamChunk, error) {
    ch := make(chan domain.StreamChunk, 64)
    go func() {
        defer close(ch)
        //SSE streaming read → send chunks one by one
    }()
    return ch, nil
}

//injection
c := client.New(client.WithLLMProvider(&MySelfHosted{
    endpoint: "http://vllm:8000",
    client:   &http.Client{Timeout: 60 * time.Second},
}))
```

### 1.2 Access custom VectorStore / Cache

```go
//Implement domain.VectorStore interface
type MyChromaStore struct {
    endpoint string
}

func (m *MyChromaStore) Upsert(ctx context.Context, col string, docs []domain.Doc) error {
    //Call ChromaDB API
}
func (m *MyChromaStore) Search(ctx context.Context, col string, vec []float32, topK int) ([]domain.Hit, error) {
    //ANN search
}
func (m *MyChromaStore) Delete(ctx context.Context, col string, ids []string) error {
    //delete vector
}

//injection
c := client.New(
    client.WithVectorStore(&MyChromaStore{endpoint: "http://chroma:8000"}),
    client.WithCache(&MyCustomCache{}),
)
```

Multiple implementations coexisting (Qdrant/Milvus, Redis/Valkey) are routed by the SDK factory according to `tenant.vector_store_preference` (§10.4 Runtime routing).

### 1.3 Access custom AgentRuntime / Tracing

```go
//Implement domain.AgentRuntime interface (local map executor)
type MyLocalRuntime struct {
    graph *Graph
}

func (r *MyLocalRuntime) Load(ctx context.Context, spec *domain.AgentSpec) (domain.AgentHandle, error) {
    //Parse AgentSpec → Build local execution graph
}
func (r *MyLocalRuntime) Run(ctx context.Context, h domain.AgentHandle, input map[string]any) (map[string]any, error) {
    //This map executes
}

//Implement domain.Tracing interface
type MyTracer struct {
    exporter trace.SpanExporter
}

func (t *MyTracer) StartSpan(ctx context.Context, name string) (context.Context, domain.Span) {
    //Custom trace implementation
}

c := client.New(
    client.WithAgentRuntime(&MyLocalRuntime{}),
    client.WithTracing(&MyTracer{}),
)
```

The SDK default implementation binds platform LangGraph (Python)/Spring AI (Java) instances, but the host can change the local map executor without changing the spec - because AgentSpec is declarative and has nothing to do with runtime (§4.3.5).

### 1.4 Complete list of SPI ports (strictly aligned with §10.4)

The SDK exposes all 15 types of port interfaces:

| Port | Interface Signature Summary | Version |
|------|-------------|------|
| `Gateway` | `Invoke(context.Context, GatewayRequest) (GatewayResponse, error)` | 1.0.0 |
| `AgentRuntime` | `Load(...), Run(...)` | 1.0.0 |
| `LLMProvider` | `Chat(...), Embed(...), Rerank(...), Stream(...)` | 1.0.0 |
| `VectorStore` | `Upsert(...), Search(...), Delete(...)` | 1.1.0 |
| `Cache` | `Get(...), Set(...)` | 1.0.0 |
| `Auth` | `ValidateToken(...)` | 1.0.0 |
| `Tracing` | `StartSpan(...)` | 1.0.0 |
| `RAG` | `Retrieve(...)` | 1.0.0 |
| `LowCode` | `ExportSpec(...)` | 1.0.0 |
| `Workflow` | `Submit(...)` | 1.0.0 |
| `Sandbox` | `Execute(...)` | 1.0.0 |
| `CICD` | `Deploy(...)` | 1.0.0 |
| `MultiTenancy` | `ResolveTenant(...)` | 1.0.0 |
| `MLOps` | `SubmitFineTune(...)` | 1.0.0 |
| `Eval` | `RunEval(...)` | 1.0.0 |

---

## 2. Key usage and code examples

> Corresponds to design/DESIGN.md §4. All examples can be compiled and run directly (Go 1.22+).

### 2.1 Quickstart: Run through conversational Agent in 30 minutes

```go
package main

import (
    "context"
    "github.com/openstrata/ai-sdk-go/pkg/client"
    "github.com/openstrata/ai-sdk-go/pkg/agent"
    "github.com/openstrata/ai-sdk-go/pkg/adapter"
)

func main() {
    c := client.New(
        client.WithGateway(adapter.MustHigressFromEnv()),
        client.WithLLMProvider(adapter.MustQwenFromEnv()),
    )

    //Declarative AgentSpec (§4.3.5 Convergence Contract)
    spec := agent.NewSpec("customer-service-v2").
        WithTenant("tenant-b").
        WithModelBinding(agent.ModelBinding{
            Preferred:     "cloud-qwen-max",
            FallbackChain: []string{"local-qwen2.5-72b", "cloud-gpt-4o"},
        }).
        WithInputSchema(agent.ObjectSchema("query", "string")).
        WithOutputSchema(agent.ObjectSchema("answer", "string")).
        WithGuardrails(agent.Guardrails{
            Basic: []string{"injection_scan", "pii_scan", "rate_limit"},
        }).
        Build()

    //Load and run via AgentRuntime SPI (declarative, runtime-independent)
    h, _ := c.AgentRuntime().Load(context.Background(), spec)
    out, _ := c.AgentRuntime().Run(context.Background(), h,
        map[string]any{"query": "Where's my order?"})
    println(out["answer"].(string))
}
```

### 2.2 Register Custom Tool

```go
tool := agent.NewTool("order_lookup").
    WithDescription("Check order logistics status").
    WithInputSchema(agent.ObjectSchema("order_id", "string")).
    WithHandler(func(ctx context.Context, in map[string]any) (map[string]any, error) {
        orderID := in["order_id"].(string)
        //Adjust business database query
        return map[string]any{"status": "shipped", "order_id": orderID}, nil
    })

//Method 1: Bind to AgentSpec
spec = spec.WithToolBindings("order_lookup")

//Method 2: Register to the platform ToolRegistry (using Gateway SPI)
err := c.Tools().Register(context.Background(), tool)

//Method 3: Expose using MCP protocol
tool = tool.WithMCP(agent.MCPConfig{Transport: "stdio"})
```

### 2.3 RAG retrieval

```go
import "github.com/openstrata/ai-sdk-go/pkg/domain"

//Get the query vector (embed via LLMProvider)
emb, _ := c.LLMProvider().Embed(ctx, domain.EmbedRequest{
    Texts: []string{"Where's my order?"},
})

//Retrieve through VectorStore SPI; SDK factory is routed according to tenant.vector_store_preference
hits, _ := c.VectorStore().Search(ctx, "kb-tenant-b", emb.Vectors[0], 5)

//Hit fragment backfill to LLMProvider's prompt
for _, hit := range hits {
    fmt.Printf("[%s] score=%.3f chunk=%s\n", hit.ID, hit.Score, hit.Content)
}
//The RAG pipeline is composed by the platform, the SDK only exposes the retrieval port (§4.5)
```

### 2.4 Multi-round Session/Memory

```go
sess := c.Session("user-123")
sess.SetWorkingMemoryTTL(1 * time.Hour) // working memory（§4.3.5 memory_bindings）

//First round: setting up memory
resp, _ := c.Chat(ctx, client.ChatInput{
    Session: sess,
    Message: "Remember my last name is Zhang, VIP level 5",
})

//Second round: automatically inject memory_bindings context
resp, _ = c.Chat(ctx, client.ChatInput{
    Session: sess,
    Message: "What is my VIP level?",
})
fmt.Println(resp.Content) //"You are VIP 5, Mr. Zhang"
```

### 2.5 Streaming call

```go
ch, err := c.LLMProvider().Stream(ctx, domain.ChatRequest{
    Messages: []domain.Message{{Role: "user", Content: "write a poem"}},
})
if err != nil {
    log.Fatal(err)
}
for chunk := range ch {
    fmt.Print(chunk.Delta)
}
```

---

## 3. Error handling and observability → Coding rules

> Corresponds to design/DESIGN.md §9. The following rules are derived from the error model and observability design, and must be strictly implemented by the AI ​​coding agent.

### Rule R1: Use unified error types

```go
//domain.Error - the only error type in the SDK
type Error struct {
    Code      string //Platform error code (corresponding to Gateway response)
    Port      string //Error SPI port name
    Retryable bool   //Is it possible to retry
    Cause     error  //underlying error
}

func (e *Error) Error() string {
    return fmt.Sprintf("[%s] %s: retryable=%v: %v", e.Port, e.Code, e.Retryable, e.Cause)
}
```

**Rule**: All SDK internal errors must be wrapped as `domain.Error`, annotated with `Port` and `Retryable`. Direct `return err` is not allowed to pass through transparent errors.

### Rule R2: Do not retry by yourself, prompt the upper layer to use fallback_chain

```go
func (c *Client) Chat(ctx context.Context, input ChatInput) (*ChatResponse, error) {
    resp, err := c.gateway.Invoke(ctx, req)
    if err != nil {
        return nil, &domain.Error{
            Code:      "LLM_TIMEOUT",
            Port:      "LLMProvider",
            Retryable: true,
            Cause:     err,
        }
        //Do not retry in this loop - the upper ModelRouter is responsible for failover
    }
    return resp, nil
}
```

**Rule**: The SDK never does exponential backoff retries internally. Third-party LLM timeout/quota exceeded → `Retryable=true` → Return to the caller → `fallback_chain` (§4.4.5) by platform ModelRouter. Avoid the SDK layer avalanche.

### Rule R3: All Port operations must be reported to Tracing span

```go
//Each Port call must be wrapped in span
func (c *Client) doWithTrace(ctx context.Context, port string, fn func(context.Context) error) error {
    ctx, span := c.tracing.StartSpan(ctx, port)
    defer span.End()
    err := fn(ctx)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
    }
    return err
}
```

**Rule**: Each SPI port call must create a Tracing span (port name span name). When an error occurs, set the span status to error and log the `domain.Error` field.

### Rule R4: Audit log is enabled by default

```go
type AuditLog struct {
    Timestamp time.Time
    TenantID  string
    Port      string
    Operation string
    Latency   time.Duration
    Error     *string
}

func (c *Client) emitAudit(ctx context.Context, log AuditLog) {
    if !c.config.Observability.AuditLog {
        return
    }
    //The core baseline must enable auditing (§4.8), here only the switch is checked
    c.auditWriter.Write(ctx, log)
}
```

**Rule**: All Gateway calls, AgentRuntime Load/Run, and Tool registration/execution must output audit logs. AgentSpec's `observability_hooks.audit: enabled` is the core baseline (§4.3.5).

### Rule R5: Metrics Hook reports key indicators

```go
type Metric struct {
    Port          string
    Operation     string
    TokenUsage    TokenUsage    // prompt/completion tokens
    Latency       time.Duration
    CacheHitRate  float64
    ErrorRate     float64
}

//SDK exposes hook callbacks
c := client.New(
    client.WithMetricsHook(func(m Metric) {
        //Exported via host OTel/Micrometer
        otel.RecordMetric(ctx, m)
    }),
)
```

**Rule**: The `MetricsHook` callback is triggered after each SPI port call, including the four dimensions of token usage, call delay, cache hit rate, and error rate.

### Rule R6: Configure coverage priority

```
explicit Option > environment variables > Configuration file > platform Manifest
```

**Rule**: All configuration reads must be merged at this priority. `config.Load()` is a unified entry that prohibits private reading of environment variables inside the adapter.

### Rule R7: Cross-language API semantic alignment

| Concepts | Go | Java | Python | Semantic Consistency |
|------|-----|------|--------|-----------|
| Error type | `domain.Error` | `OpenStrataException` | `OpenStrataError` | code + port + retryable |
| Streaming return | `<-chan StreamChunk` | `Flow<StreamChunk>` | `AsyncIterator[StreamChunk]` | Chunk-by-chunk push, 3 primitives are semantically equivalent |
| AgentSpec parsing | `agent.ParseSpec(yaml)` | `AgentSpec.parse(yaml)` | `AgentSpec.parse_yaml(yaml)` | The same YAML, isomorphic parsing |
| Configuration injection | `client.WithXxx(...)` | `@Bean` / Builder | Construction parameters | Dependency inversion, implementation replacement without changing the core |

**Rule**: Method signature semantics for the same SPI port are consistent across languages. AgentSpec YAML schema and the three SDKs share the same fixture.

---

## 4. Coding rules cheat sheet

| Rule Number | One Sentence | Consequences of Violation |
|----------|--------|----------|
| R1 | All errors are packaged as `domain.Error` | The upper layer cannot determine retryability, and failover fails |
| R2 | Does not retry, returns `Retryable=true` | Avalanche; ModelRouter fallback_chain is short-circuited |
| R3 | Each Port calls the package Tracing span | trace faults, unable to be traced end-to-end |
| R4 | Audit log is enabled by default | Compliance audit is missing (§4.8 core baseline) |
| R5 | Trigger MetricsHook | No token usage/latency/hit rate monitoring |
| R6 | Configuration coverage is merged according to priority | Configuration sources are confusing and debugging is difficult |
| R7 | Cross-language API semantic alignment | AgentSpec parsing is inconsistent in different SDKs |

---

> **Associated documents**: arch/ARCH.md (architecture and port list) · specs/SPECS.md (SPI version contract and configuration keys) · design/DESIGN.md (complete design)
