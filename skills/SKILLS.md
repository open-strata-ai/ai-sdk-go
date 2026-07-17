# ai-sdk-go · AI 编码技能（SKILLS）

> **来源**: design/DESIGN.md §5（扩展点）· §4（关键用法与代码示例）· §9（错误处理与可观测性）
> **受众**: AI 编码 Agent、SDK 二次开发者
> **协同**: arch/ARCH.md（架构）· specs/SPECS.md（契约）· design/adr/（决策记录）
> **平台版本**: strata v1.4.0

---

## 1. 扩展点（SPI 端口实现指南）

> 对应 design/DESIGN.md §5。 SDK 的扩展点即平台 SPI 端口；名称与 §10.4 canonical 端口名一致。
> 宿主应用只需实现对应 `pkg/domain` 接口并通过 `client.WithXxx(...)` 注入，即可替换默认实现——**零改动 SDK 核心**（依赖倒置 + 防腐层 ACL）。

### 1.1 接入自定义 LLMProvider

```go
type MySelfHosted struct {
    endpoint string
    client   *http.Client
}

func (m *MySelfHosted) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
    // 调自有 vLLM/TGI OpenAI-compatible 端点
    reqBody, _ := json.Marshal(req.ToOpenAICompat())
    resp, err := m.client.Post(m.endpoint+"/v1/chat/completions", "application/json",
        bytes.NewReader(reqBody))
    // ... 解析响应，返回 domain.ChatResponse
}
func (m *MySelfHosted) Embed(ctx context.Context, req domain.EmbedRequest) (domain.EmbedResponse, error) { /* ... */ }
func (m *MySelfHosted) Rerank(ctx context.Context, req domain.RerankRequest) (domain.RerankResponse, error) { /* ... */ }
func (m *MySelfHosted) Stream(ctx context.Context, req domain.ChatRequest) (<-chan domain.StreamChunk, error) {
    ch := make(chan domain.StreamChunk, 64)
    go func() {
        defer close(ch)
        // SSE 流式读取 → 逐个发送 chunk
    }()
    return ch, nil
}

// 注入
c := client.New(client.WithLLMProvider(&MySelfHosted{
    endpoint: "http://vllm:8000",
    client:   &http.Client{Timeout: 60 * time.Second},
}))
```

### 1.2 接入自定义 VectorStore / Cache

```go
// 实现 domain.VectorStore 接口
type MyChromaStore struct {
    endpoint string
}

func (m *MyChromaStore) Upsert(ctx context.Context, col string, docs []domain.Doc) error {
    // 调 ChromaDB API
}
func (m *MyChromaStore) Search(ctx context.Context, col string, vec []float32, topK int) ([]domain.Hit, error) {
    // ANN 检索
}
func (m *MyChromaStore) Delete(ctx context.Context, col string, ids []string) error {
    // 删除向量
}

// 注入
c := client.New(
    client.WithVectorStore(&MyChromaStore{endpoint: "http://chroma:8000"}),
    client.WithCache(&MyCustomCache{}),
)
```

多实现并存（Qdrant/Milvus、Redis/Valkey）由 SDK 工厂按 `tenant.vector_store_preference` 路由（§10.4 运行时路由）。

### 1.3 接入自定义 AgentRuntime / Tracing

```go
// 实现 domain.AgentRuntime 接口（本地图执行器）
type MyLocalRuntime struct {
    graph *Graph
}

func (r *MyLocalRuntime) Load(ctx context.Context, spec *domain.AgentSpec) (domain.AgentHandle, error) {
    // 解析 AgentSpec → 构建本地执行图
}
func (r *MyLocalRuntime) Run(ctx context.Context, h domain.AgentHandle, input map[string]any) (map[string]any, error) {
    // 本地图执行
}

// 实现 domain.Tracing 接口
type MyTracer struct {
    exporter trace.SpanExporter
}

func (t *MyTracer) StartSpan(ctx context.Context, name string) (context.Context, domain.Span) {
    // 自定义 trace 实现
}

c := client.New(
    client.WithAgentRuntime(&MyLocalRuntime{}),
    client.WithTracing(&MyTracer{}),
)
```

SDK 默认实现绑定平台 LangGraph（Python）/Spring AI（Java）实例，但宿主可换本地图执行器而不改 spec——因为 AgentSpec 声明式、与运行时无关（§4.3.5）。

### 1.4 SPI 端口完整清单（与 §10.4 严格对齐）

SDK 暴露全部 15 类端口接口：

| 端口 | 接口签名摘要 | 版本 |
|------|-------------|------|
| `Gateway` | `Invoke(context.Context, GatewayRequest) (GatewayResponse, error)` | 1.2.0 |
| `AgentRuntime` | `Load(...), Run(...)` | 1.3.0 |
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

## 2. 关键用法与代码示例

> 对应 design/DESIGN.md §4。所有示例可直接编译运行（Go 1.22+）。

### 2.1 Quickstart: 30 分钟跑通对话式 Agent

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

    // 声明式 AgentSpec（§4.3.5 收敛契约）
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

    // 经 AgentRuntime SPI 加载并运行（声明式、与运行时无关）
    h, _ := c.AgentRuntime().Load(context.Background(), spec)
    out, _ := c.AgentRuntime().Run(context.Background(), h,
        map[string]any{"query": "我的订单到哪了？"})
    println(out["answer"].(string))
}
```

### 2.2 注册自定义 Tool

```go
tool := agent.NewTool("order_lookup").
    WithDescription("查询订单物流状态").
    WithInputSchema(agent.ObjectSchema("order_id", "string")).
    WithHandler(func(ctx context.Context, in map[string]any) (map[string]any, error) {
        orderID := in["order_id"].(string)
        // 调业务数据库查询
        return map[string]any{"status": "shipped", "order_id": orderID}, nil
    })

// 方式一：绑定到 AgentSpec
spec = spec.WithToolBindings("order_lookup")

// 方式二：注册到平台 ToolRegistry（走 Gateway SPI）
err := c.Tools().Register(context.Background(), tool)

// 方式三：使用 MCP 协议暴露
tool = tool.WithMCP(agent.MCPConfig{Transport: "stdio"})
```

### 2.3 RAG 检索

```go
import "github.com/openstrata/ai-sdk-go/pkg/domain"

// 获取查询向量（经 LLMProvider embed）
emb, _ := c.LLMProvider().Embed(ctx, domain.EmbedRequest{
    Texts: []string{"我的订单到哪了？"},
})

// 检索走 VectorStore SPI；SDK 工厂按 tenant.vector_store_preference 路由
hits, _ := c.VectorStore().Search(ctx, "kb-tenant-b", emb.Vectors[0], 5)

// 命中片段回填给 LLMProvider 的 prompt
for _, hit := range hits {
    fmt.Printf("[%s] score=%.3f chunk=%s\n", hit.ID, hit.Score, hit.Content)
}
// RAG 管线由平台组合，SDK 仅暴露检索端口（§4.5）
```

### 2.4 多轮 Session / 记忆

```go
sess := c.Session("user-123")
sess.SetWorkingMemoryTTL(1 * time.Hour) // working memory（§4.3.5 memory_bindings）

// 首轮：设置记忆
resp, _ := c.Chat(ctx, client.ChatInput{
    Session: sess,
    Message: "记住我姓张，VIP 等级 5",
})

// 次轮：自动注入 memory_bindings 上下文
resp, _ = c.Chat(ctx, client.ChatInput{
    Session: sess,
    Message: "我的 VIP 等级是多少？",
})
fmt.Println(resp.Content) // "您是 VIP 5，张先生"
```

### 2.5 流式调用

```go
ch, err := c.LLMProvider().Stream(ctx, domain.ChatRequest{
    Messages: []domain.Message{{Role: "user", Content: "写一首诗"}},
})
if err != nil {
    log.Fatal(err)
}
for chunk := range ch {
    fmt.Print(chunk.Delta)
}
```

---

## 3. 错误处理与可观测性 → 编码规则

> 对应 design/DESIGN.md §9。以下规则由错误模型和可观测性设计推导，AI 编码 Agent 须严格执行。

### 规则 R1: 使用统一错误类型

```go
// domain.Error —— SDK 唯一错误类型
type Error struct {
    Code      string // 平台错误码（对应 Gateway 响应）
    Port      string // 出错 SPI 端口名
    Retryable bool   // 是否可重试
    Cause     error  // 底层错误
}

func (e *Error) Error() string {
    return fmt.Sprintf("[%s] %s: retryable=%v: %v", e.Port, e.Code, e.Retryable, e.Cause)
}
```

**规则**: 所有 SDK 内部错误必须包装为 `domain.Error`，标注 `Port` 和 `Retryable`。不允许直接 `return err` 穿透明细错误。

### 规则 R2: 不自行重试，提示上层走 fallback_chain

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
        // 不在此循环重试 —— 上层 ModelRouter 负责故障转移
    }
    return resp, nil
}
```

**规则**: SDK 永远不在内部做指数退避重试。第三方 LLM 超时/配额超限 → `Retryable=true` → 返回调用方 → 由平台 ModelRouter 走 `fallback_chain`（§4.4.5）。避免 SDK 层雪崩。

### 规则 R3: 所有 Port 操作须上报 Tracing span

```go
// 每个 Port 调用须包裹在 span 中
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

**规则**: 每个 SPI 端口调用必须创建一个 Tracing span（端口名作 span name）。出错时设置 span 状态为 error 并记录 `domain.Error` 字段。

### 规则 R4: 审计日志默认开启

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
    // core 基线必须开启审计（§4.8），这里仅检查开关
    c.auditWriter.Write(ctx, log)
}
```

**规则**: 所有 Gateway 调用、AgentRuntime Load/Run、Tool 注册/执行必须输出审计日志。AgentSpec 的 `observability_hooks.audit: enabled` 为 core 基线（§4.3.5）。

### 规则 R5: Metrics Hook 上报关键指标

```go
type Metric struct {
    Port          string
    Operation     string
    TokenUsage    TokenUsage    // prompt/completion tokens
    Latency       time.Duration
    CacheHitRate  float64
    ErrorRate     float64
}

// SDK 暴露 hook 回调
c := client.New(
    client.WithMetricsHook(func(m Metric) {
        // 经宿主 OTel/Micrometer 导出
        otel.RecordMetric(ctx, m)
    }),
)
```

**规则**: 每个 SPI 端口调用结束后触发 `MetricsHook` 回调，包含 token 用量、调用延迟、缓存命中率、错误率四个维度。

### 规则 R6: 配置覆盖优先级

```
显式 Option > 环境变量 > 配置文件 > 平台 Manifest
```

**规则**: 所有配置读取必须按此优先级合并。`config.Load()` 统一入口，禁止在 adapter 内部私读环境变量。

### 规则 R7: 跨语言 API 语义对齐

| 概念 | Go | Java | Python | 语义一致性 |
|------|-----|------|--------|-----------|
| 错误类型 | `domain.Error` | `OpenStrataException` | `OpenStrataError` | code + port + retryable |
| 流式返回 | `<-chan StreamChunk` | `Flow<StreamChunk>` | `AsyncIterator[StreamChunk]` | 逐块推送，3 种原语语义等价 |
| AgentSpec 解析 | `agent.ParseSpec(yaml)` | `AgentSpec.parse(yaml)` | `AgentSpec.parse_yaml(yaml)` | 同一份 YAML，同构解析 |
| 配置注入 | `client.WithXxx(...)` | `@Bean` / Builder | 构造参数 | 依赖倒置，实现替换不改核心 |

**规则**: 同一 SPI 端口的方法签名语义跨语言一致。AgentSpec YAML schema 三 SDK 共用同一份 fixture。

---

## 4. 编码规则速查表

| 规则编号 | 一句话 | 违反后果 |
|----------|--------|----------|
| R1 | 所有错误包装为 `domain.Error` | 上层无法判断可重试性，故障转移失效 |
| R2 | 不自重试，返回 `Retryable=true` | 雪崩；ModelRouter fallback_chain 被短路 |
| R3 | 每个 Port 调用包裹 Tracing span | trace 断层，无法端到端追踪 |
| R4 | 审计日志默认开 | 合规审计缺失（§4.8 core 基线） |
| R5 | 触发 MetricsHook | 无 token 用量/延迟/命中率监控 |
| R6 | 配置覆盖按优先级合并 | 配置源混乱，调试困难 |
| R7 | 跨语言 API 语义对齐 | AgentSpec 在不同 SDK 解析不一致 |

---

> **关联文档**: arch/ARCH.md（架构与端口清单）· specs/SPECS.md（SPI 版本契约与配置键）· design/DESIGN.md（完整设计）
