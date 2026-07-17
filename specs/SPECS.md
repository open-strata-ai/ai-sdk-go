# ai-sdk-go · 契约（SPECS）

> **来源**: design/DESIGN.md §6（SPI 映射表）· §7（配置键）· §8（版本/SemVer）
> **受众**: 平台集成工程师、CI/CD 校验、AI 编码 Agent
> **协同**: arch/ARCH.md（架构）· skills/SKILLS.md（编码规则）· bom.yaml（§16）
> **平台版本**: strata v1.4.0

---

## 1. SPI 映射表（端口 → 版本 → SDK 角色）

> 对应 design/DESIGN.md §6。本表是 SDK 与平台 SPI 之间的**唯一契约来源**。
> 端口名与 `bom.yaml` `spi` 字段逐字一致（§16.2）。CI 在发版时自动校验一致性。

### 1.1 完整映射

| SDK 端口（pkg/domain） | 平台 canonical | interface_versions | SDK 角色 | 默认 Adapter | optional |
| --- | --- | --- | --- | --- | --- |
| `Gateway` | **Gateway** | 1.2.0 | 生产 | Higress | false |
| `AgentRuntime` | **AgentRuntime** | 1.3.0 | 生产 | LangGraph 绑定 | false |
| `LLMProvider` | **LLMProvider** | 1.0.0 | 生产 | Qwen/OpenAI/Claude | false |
| `VectorStore` | **VectorStore** | 1.1.0 | 生产 | Qdrant | false |
| `Cache` | **Cache** | 1.0.0 | 生产 | Redis | false |
| `Auth` | **Auth** | 1.0.0 | 消费 | Keycloak | false |
| `Tracing` | **Tracing** | 1.0.0 | 生产 | OTel / Langfuse | false |
| `RAG` | **RAG** | 1.0.0 | 消费 | RAGFlow | false |
| `LowCode` | **LowCode** | 1.0.0 | 消费 | React Flow | false |
| `Workflow` | **Workflow** | 1.0.0 | 消费 | Temporal | true |
| `Sandbox` | **Sandbox** | 1.0.0 | 消费 | Kata / E2B | true |
| `CICD` | **CICD** | 1.0.0 | 消费 | ArgoCD / Istio | true |
| `MultiTenancy` | **MultiTenancy** | 1.0.0 | 消费 | Capsule | true |
| `MLOps` | **MLOps** | 1.0.0 | 消费 | MLflow | true |
| `Eval` | **Eval** | 1.0.0 | 消费 | Promptfoo / DeepEval / Ragas | false |

### 1.2 角色语义

| 角色 | 含义 | 实现约束 |
|------|------|----------|
| **生产（producer）** | SDK 主动调平台服务 | Adapter 必须实现；接口签名冻结 |
| **消费（consumer）** | SDK 消费平台下发上下文 | 默认 pass-through；宿主可自定义 |

---

## 2. 配置键（Config Keys）

> 对应 design/DESIGN.md §7。配置文件位于 `infrastructure/config/client.yaml`，须可被元仓渲染（§15.7.3）。

### 2.1 配置键完整列表

```yaml
# === Gateway（必填，core） ===
gateway:
  provider: higress              # str: "higress" | "apigw"
  base_url: ${GATEWAY_BASE_URL}  # env var：网关地址
  timeout_sec: 30                # int: 调用超时（秒）
  max_retries: 0                 # int: SDK 层不重试（由 ModelRouter 负责）

# === Model / LLMProvider（必填，core） ===
model:
  qwen:
    type: dashscope              # str: "dashscope" | "local_vllm"
    default: true                # bool: 是否默认 provider
  openai:
    type: openai                 # str: "openai" | "local_vllm"
  claude:
    type: anthropic              # str: "anthropic" | "bedrock"

# === Cache（必填，core） ===
cache:
  provider: redis                # str: "redis" | "valkey"
  ttl_sec: 3600                  # int: 默认缓存 TTL（秒）
  endpoints: ["redis:6379"]      # list[str]: 集群地址

# === VectorStore（必填，core） ===
vector_store:
  preference: qdrant             # str: "qdrant" | "milvus"
  collection_prefix: "kb-"       # str: 集合前缀
  dimension: 1536                # int: 默认向量维度

# === Observability（必填，core 基线） ===
observability:
  otel_traces: true              # bool: OTel traces（core 基线）
  audit_log: true                # bool: 审计日志（core 基线）
  metrics_hook_enabled: true     # bool: 是否触发 MetricsHook
  langfuse_enabled: false        # bool: Langfuse 额外层（optional）

# === Auth（必填，core） ===
auth:
  provider: keycloak             # str: "keycloak"
  issuer_url: ${AUTH_ISSUER_URL}
  token_ttl_sec: 3600

# === Sandbox（optional） ===
sandbox:
  provider: kata                 # str: "kata" | "e2b"
  image: "sandbox:latest"
  timeout_sec: 300
  memory_mb: 512

# === MLOps（optional） ===
mlops:
  provider: mlflow               # str: "mlflow"
  tracking_uri: ${MLFLOW_TRACKING_URI}
```

### 2.2 配置优先级

```
显式 client.Option > 环境变量 > 配置文件片段 > 平台 Manifest 下发
```

1. **显式 Option**: `client.WithGateway(...)` 等代码构造参数
2. **环境变量**: `GATEWAY_BASE_URL`、`DASHSCOPE_API_KEY` 等（仅用于调试；生产走平台 Gateway + 租户 Token）
3. **配置文件**: `infrastructure/config/client.yaml` 片段
4. **平台 Manifest**: 控制面下发（运行时可覆盖配置片段）

### 2.3 凭证安全（§4.4.6）

| 约束 | 说明 |
|------|------|
| 租户 Token 持有者 | SDK Client，经 `Auth` 端口校验 |
| 第三方 Provider Key | 仅存于平台 Secret Vault，SDK 永不经手 |
| 调试环境变量 | `DASHSCOPE_API_KEY` 等仅本地调试用；生产禁用 |
| 数据出境管控 | 第三方调用前网关做 PII 脱敏；`no_egress` 策略强制自托管 |

---

## 3. 版本与兼容策略

> 对应 design/DESIGN.md §8。

### 3.1 SDK 自身版本（SemVer）

- 当前版本：`v1.4.0`（对齐 `strata v1.4.0`，§16.1）
- MAJOR：破坏性 API 变更（如接口方法签名修改）
- MINOR：新增端口/功能/Adapter，向后兼容
- PATCH：Bug 修复、性能优化、文档更新
- 破坏性变更必须附 ADR（`design/adr/`）
- Module path: `github.com/openstrata/ai-sdk-go`

### 3.2 SPI 端口版本契约

SDK 承诺与以下 `interface_versions` 兼容：

| 端口 | 最低兼容 version | 约束 | 变更策略 |
|------|-----------------|------|----------|
| Gateway | 1.2.0 | OpenAI-compatible 协议不变 | 新增可选字段→MINOR；新增必填字段→MAJOR |
| AgentRuntime | 1.3.0 | AgentSpec `apiVersion: openstrata.io/v1` | Spec 新增可选字段向后兼容 |
| LLMProvider | 1.0.0 | chat/embed/rerank/stream 签名稳定 | 签名变更→MAJOR；新增方法→MINOR |
| VectorStore | 1.1.0 | upsert/search/delete 稳定 | 返回结构字段变更→MAJOR |
| Cache | 1.0.0 | get/set 稳定 | 同上 |

### 3.3 跨语言一致性约束

- `ai-sdk-go` / `ai-sdk-java` / `ai-sdk-python` 三件套对同一 SPI 端口的方法签名语义一致
- AgentSpec YAML schema 三 SDK 共用同一份 fixture（`specs/fixtures/`）
- CI 跨语言契约测试验证：同一份 AgentSpec 可被任一语言 SDK 构建、任一运行时实例绑定执行

### 3.4 平台版本冻结

| SDK 版本 | 平台 strata | bom.yaml tag | Go module |
|----------|------------|-------------|-----------|
| v1.4.0 | v1.4.0 | v1.4.0 | `github.com/openstrata/ai-sdk-go v1.4.0` |

变更流程：
1. SDK 仓打 tag（如 `v1.5.0`）
2. CI 校验 `bom.yaml` 中 `interface_versions` 一致性（§16.4 / §15.7.4）
3. 元仓 `repos.yaml` / `bom.yaml` 钉版本 → Dependency Resolver 装配
4. 控制面 `GET /v1/release/manifest` 可审计偏离认证清单

---

> **关联文档**: arch/ARCH.md（架构与端口清单）· skills/SKILLS.md（编码规则）· design/DESIGN.md（完整设计）· bom.yaml §16
