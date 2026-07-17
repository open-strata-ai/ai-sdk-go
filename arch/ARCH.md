# ai-sdk-go · 架构（ARCH）

> **来源**: design/DESIGN.md §1（定位）· §2（核心抽象）· §6（SPI 映射 15 端口）
> **受众**: 平台架构师、SDK 维护者、AI 编码 Agent
> **协同**: skills/SKILLS.md（扩展点与编码规则）· specs/SPECS.md（契约与版本）· design/adr/（重大决策）
> **平台版本**: strata v1.4.0

---

## 1. 定位与目标用户

### 1.1 SDK 是什么

`ai-sdk-go` 是 OpenStrata 平台面向 **Go 生态**的官方开发者库，发布到 Go Module Proxy（`github.com/openstrata/ai-sdk-go`）。它是"客户端 + 构建器 + 扩展端口"三层合一的轻量库，不是可部署服务。

核心承诺：
- **最小侵入**：仅依赖 Go 标准库 + 极少量必要依赖（`google/wire`、`gin`、`httpx`、`yaml`、`otel`），可嵌入任意 Go 宿主应用
- **声明式 AgentSpec**：Agent 行为收敛为 `apiVersion: openstrata.io/v1` AgentSpec，与语言/运行时无关（§4.3.5）
- **依赖倒置**：所有平台 SPI 以 `pkg/domain` 接口暴露，宿主注入自定义适配器时零改动 SDK 核心

### 1.2 解决什么问题

让 Go 开发者在已有 Go 应用（微服务、CLI、后台任务）中以最小侵入方式：
1. 构建对话式 Agent/Client，接入平台运行时能力（网关、Agent 运行时、工具注册、记忆、RAG、缓存、可观测）
2. 把内部能力封装成平台 Tool，将业务数据接入 RAG
3. 基于 SDK 的 SPI 端口实现自定义 `LLMProvider`/`VectorStore`/`Cache`/`AgentRuntime` 适配器

### 1.3 目标用户画像

| 角色 | 场景 | 接触面 |
|------|------|--------|
| Go 后端/平台工程师 | 在现有 Go 服务里嵌入 Agent | `pkg/client`、`pkg/agent` |
| 平台二次开发者 | 实现自定义 SPI Adapter | `pkg/domain` Port 接口 |
| AI 编码 Agent | 按 skills/ 规则生成适配代码 | skills/SKILLS.md |

### 1.4 与平台的边界

```
宿主 Go 应用 ──→ ai-sdk-go pkg/ ──→ OpenStrata 平台 SPI
   (已有)          (本仓)              (GateWay 等)
```

- SDK 仅依赖 Go 标准库 + 极少量必要依赖，保证可嵌入任意宿主（§15.6.1 框架收敛）
- SDK 自身不绑定任何 Web 框架；宿主可选 Gin/Hertz，SDK 通过 `adapter` 层适配
- SDK 不持有第三方 Provider Key；密钥经平台 Secret Vault + Gateway 间接路由（§4.4.6）

---

## 2. 核心抽象

### 2.1 包结构（DDD 四层）

```
ai-sdk-go/
├── cmd/                     # 可选 CLI 样例（非发布主体）
├── pkg/                     # ★ 发布主体（go module 根包）
│   ├── client/              # ① 接入层：GatewayClient（OpenAI-compatible）
│   ├── agent/               # ② 应用层：AgentBuilder / Runner 编排
│   ├── domain/              # ③ 领域层：AgentSpec / Tool / Session 实体 + Port 接口
│   ├── adapter/             # ④ 基础设施层：SPI Adapter（LLMProvider/VectorStore/Cache…）
│   └── config/              # 配置加载（infrastructure/config 片段渲染）
├── infrastructure/config/   # 本仓 SPI 适配器局部配置片段
└── arch/ design/ skills/ specs/  # 演化式 AI 编码事实源
```

层次依赖：①→②→③←④。领域层（③）不依赖任何框架/外部组件，依赖方向自 ①② 向内指向 ③。

### 2.2 一等公民对象（5 个对外 Bean）

| 抽象 | 包路径 | 对应平台 SPI | 说明 |
|------|--------|-------------|------|
| `Client` | `pkg/client` | **Gateway** SPI (§4.4.1) | OpenAI-compatible 调用入口；持有租户 Token |
| `AgentBuilder` / `AgentSpec` | `pkg/agent` | **AgentRuntime** SPI (§4.3.5) | 声明式拼装 AgentSpec（apiVersion/kind/metadata/model_binding/tool_bindings/memory_bindings/state_machine/guardrails） |
| `Tool` | `pkg/agent` | **ToolRegistry** (§4.3.2) | 声明 JSON Schema，支持 MCP stdio/SSE/HTTP |
| `Session` | `pkg/client` | **Cache** SPI (§4.3.4) | 多轮会话/记忆句柄，绑定 memory_bindings |
| `Retriever` | `pkg/domain` | **VectorStore** SPI (§10.4) + **RAG** SPI | 知识库检索，返回命中片段 |

### 2.3 领域层 Port 接口（依赖倒置）

领域层在 `pkg/domain` 中定义以下 SPI 端口接口，名称与平台 §10.4 canonical 端口名**严格一致**：

```go
package domain

import (
    "context"
    "time"
)

// LLMProvider —— 对应平台 LLMProvider SPI（§4.4.4），interface_versions: 1.0.0
type LLMProvider interface {
    Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
    Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error)
    Rerank(ctx context.Context, req RerankRequest) (RerankResponse, error)
    Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
}

// VectorStore —— 对应平台 VectorStore SPI（§10.4），interface_versions: 1.1.0
type VectorStore interface {
    Upsert(ctx context.Context, col string, docs []Doc) error
    Search(ctx context.Context, col string, vec []float32, topK int) ([]Hit, error)
    Delete(ctx context.Context, col string, ids []string) error
}

// AgentRuntime —— 对应平台 AgentRuntime SPI（§10.6 / §4.3.5），interface_versions: 1.3.0
type AgentRuntime interface {
    Load(ctx context.Context, spec *AgentSpec) (AgentHandle, error)
    Run(ctx context.Context, h AgentHandle, input map[string]any) (map[string]any, error)
}

// Cache —— 对应平台 Cache SPI（§4.3.4），interface_versions: 1.0.0
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, bool, error)
    Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

// Gateway —— 对应平台 Gateway SPI（§4.4.1），interface_versions: 1.2.0
type Gateway interface {
    Invoke(ctx context.Context, req GatewayRequest) (GatewayResponse, error)
}

// Tracing —— 对应平台 Tracing SPI（§4.8），interface_versions: 1.0.0
type Tracing interface {
    StartSpan(ctx context.Context, name string) (context.Context, Span)
}

// Auth —— 对应平台 Auth SPI，interface_versions: 1.0.0
type Auth interface {
    ValidateToken(ctx context.Context, token string) (TenantContext, error)
}

// RAG —— 对应平台 RAG SPI，interface_versions: 1.0.0
type RAG interface {
    Retrieve(ctx context.Context, query string, kbID string, topK int) ([]RAGHit, error)
}

// LowCode —— 对应平台 LowCode SPI，interface_versions: 1.0.0
type LowCode interface {
    ExportSpec(ctx context.Context, canvasID string) (*AgentSpec, error)
}

// Workflow —— 对应平台 Workflow SPI，interface_versions: 1.0.0
type Workflow interface {
    Submit(ctx context.Context, wf WorkflowSpec) (WorkflowHandle, error)
}

// Sandbox —— 对应平台 Sandbox SPI，interface_versions: 1.0.0
type Sandbox interface {
    Execute(ctx context.Context, code string, lang string) (SandboxResult, error)
}

// CICD —— 对应平台 CICD SPI，interface_versions: 1.0.0
type CICD interface {
    Deploy(ctx context.Context, spec DeploySpec) (DeployHandle, error)
}

// MultiTenancy —— 对应平台 MultiTenancy SPI，interface_versions: 1.0.0
type MultiTenancy interface {
    ResolveTenant(ctx context.Context, tenantID string) (TenantConfig, error)
}

// MLOps —— 对应平台 MLOps SPI，interface_versions: 1.0.0
type MLOps interface {
    SubmitFineTune(ctx context.Context, spec FineTuneSpec) (FineTuneHandle, error)
}

// Eval —— 对应平台 Eval SPI，interface_versions: 1.0.0
type Eval interface {
    RunEval(ctx context.Context, spec EvalSpec) (EvalReport, error)
}
```

> **关键约定**: 所有端口接口不含任何具体框架/外部组件依赖（§15.6.2 领域层禁令）。接口方法签名与 Java/Python SDK 语义一致（跨语言契约）。

### 2.4 装配模型

SDK 采用 **Wire 编译期注入** + 手动注入双模式：

```go
// 方式一：Wire 自动装配（推荐，生产）
c := client.New(
    client.WithGateway(adapter.MustHigressFromEnv()),
    client.WithLLMProvider(adapter.MustQwenFromEnv()),
    client.WithCache(adapter.NewRedisCache(cfg.Cache)),
)

// 方式二：手动注入自定义实现（二次开发）
c := client.New(
    client.WithLLMProvider(&MySelfHosted{endpoint: "http://vllm:8000"}),
    client.WithVectorStore(&MyCustomVS{}),
)
```

---

## 3. 与平台 SPI 的映射（15 端口→canonical→interface_versions）

> 对应 design/DESIGN.md §6，与平台 §10.4 严格对齐。

### 3.1 映射表

| SDK 领域端口（pkg/domain） | 平台 SPI 端口（§10.4 canonical） | interface_versions | SDK 角色 | 默认 Adapter（bom.yaml） |
| --- | --- | --- | --- | --- |
| `Gateway` | **Gateway** | 1.2.0 | 生产（invoke） | Higress（core） |
| `AgentRuntime` | **AgentRuntime** | 1.3.0 | 生产（Load/Run） | LangGraph/Spring AI 绑定执行（core） |
| `LLMProvider` | **LLMProvider** | 1.0.0 | 生产（chat/embed/rerank/stream） | Qwen-Cloud / OpenAI / Claude（core 第三方） |
| `VectorStore` | **VectorStore** | 1.1.0 | 生产（upsert/search/delete） | Qdrant（core）/ Milvus（optional） |
| `Cache` | **Cache** | 1.0.0 | 生产（语义/精确缓存） | Redis（core）/ Valkey（optional, OSI） |
| `Auth` | **Auth** | 1.0.0 | 消费（租户 Token 校验） | Keycloak（core） |
| `Tracing` | **Tracing** | 1.0.0 | 生产（span） | Langfuse（core）/ OTel（core 基线） |
| `RAG` | **RAG** | 1.0.0 | 消费（检索回填） | RAGFlow（core） |
| `LowCode` | **LowCode** | 1.0.0 | 消费（画布→AgentSpec） | 自研 React Flow（core）/ Dify（reference） |
| `Workflow` | **Workflow** | 1.0.0 | 消费（长任务编排） | Temporal（optional） |
| `Sandbox` | **Sandbox** | 1.0.0 | 消费（代码执行隔离） | Kata / E2B（optional） |
| `CICD` | **CICD** | 1.0.0 | 消费（灰度发布） | ArgoCD / Istio（optional） |
| `MultiTenancy` | **MultiTenancy** | 1.0.0 | 消费（租户隔离上下文） | Capsule（optional） |
| `MLOps` | **MLOps** | 1.0.0 | 消费（微调/蒸馏） | MLflow（optional） |
| `Eval` | **Eval** | 1.0.0 | 消费（评测回流） | Promptfoo / DeepEval / Ragas（core/optional） |

### 3.2 角色分类

| SDK 角色 | 含义 | 涉及端口 |
|----------|------|----------|
| **生产（producer）** | SDK 主动调用平台服务，生成数据或执行操作 | Gateway、AgentRuntime、LLMProvider、VectorStore、Cache、Tracing |
| **消费（consumer）** | SDK 消费平台下发的上下文/配置，不主动修改 | Auth、RAG、LowCode、Workflow、Sandbox、CICD、MultiTenancy、MLOps、Eval |

### 3.3 关键约束

1. **端口名逐字一致**: SDK 端口名与 `bom.yaml` `spi` 字段完全相同（§16.2）；新增/替换实现永不需改核心（§10.6 Registry 模型）
2. **同级故障转移**: 同一 `LLMProvider` SPI 背后并存自托管与多个第三方实现，跨实现故障转移由平台 `ModelRouter` 负责（§4.4.4/§4.4.5）
3. **VectorStore 无感切换**: 切换 VectorStore 提供方时由平台迁移 Job 双写校验（§10.4），SDK 经 SPI 工厂拿对应 Adapter，业务代码零改动
4. **配置覆盖优先级**: 显式 Option > 环境变量 > 配置文件 > 平台 Manifest 下发（§12）

### 3.4 端口间关系（数据流）

```
Gateway ──→ LLMProvider ──→ (chat/embed/rerank/stream)
Gateway ──→ VectorStore ──→ RAG ──→ (检索回填 prompt)
Gateway ──→ Cache ──→ Session ──→ (会话记忆)
AgentRuntime ──→ Tracing ──→ (span 上报)
Gateway ──→ Auth ──→ MultiTenancy ──→ (租户上下文)
Workflow ──→ Sandbox ──→ (长任务代码隔离)
MLOps ──→ Eval ──→ LowCode ──→ (微调→评测→画布导出)
```

### 3.5 Go 特定约束

- 所有 Port 接口使用 `context.Context` 作为第一参数（Go 惯例）
- `Stream` 返回 `<-chan StreamChunk`（Go channel，非阻塞管道）
- 注入通过 `client.WithXxx(...)` Option 模式
- Wire 编译期注入：依赖关系在 `wire.go` 声明，`go generate` 生成初始化代码

---

> **关联文档**: skills/SKILLS.md（扩展点与编码规则）· specs/SPECS.md（SPI 版本契约与配置键）· design/DESIGN.md（完整设计）
