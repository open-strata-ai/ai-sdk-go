# ai-sdk-go · Contract (SPECS)

> **Source**: design/DESIGN.md §6 (SPI Mapping Table) · §7 (Configuration Key) · §8 (Version/SemVer)
> **Audience**: Platform integration engineers, CI/CD verification, AI coding agents
> **Collaboration**: arch/ARCH.md (architecture) · skills/SKILLS.md (coding rules) · bom.yaml (§16)
> **Platform version**: strata v1.4.0

---

## 1. SPI mapping table (port → version → SDK role)

> Corresponds to design/DESIGN.md §6. This table is the **sole source of contract** between the SDK and the platform SPI.
> The port name corresponds verbatim to the `bom.yaml` `spi` field (§16.2). CI automatically verifies consistency when releasing a version.

### 1.1 Complete mapping

| SDK port (pkg/domain) | platform canonical | interface_versions | SDK role | default Adapter | optional |
| --- | --- | --- | --- | --- | --- |
| `Gateway` | **Gateway** | 1.2.0 | Production | Higress | false |
| `AgentRuntime` | **AgentRuntime** | 1.3.0 | Production | LangGraph Binding | false |
| `LLMProvider` | **LLMProvider** | 1.0.0 | Production | Qwen/OpenAI/Claude | false |
| `VectorStore` | **VectorStore** | 1.1.0 | Production | Qdrant | false |
| `Cache` | **Cache** | 1.0.0 | Production | Redis | false |
| `Auth` | **Auth** | 1.0.0 | Consume | Keycloak | false |
| `Tracing` | **Tracing** | 1.0.0 | Production | OTel/Langfuse | false |
| `RAG` | **RAG** ​​| 1.0.0 | Consumption | RAGFlow | false |
| `LowCode` | **LowCode** | 1.0.0 | Consume | React Flow | false |
| `Workflow` | **Workflow** | 1.0.0 | Consume | Temporal | true |
| `Sandbox` | **Sandbox** | 1.0.0 | Consumption | Kata / E2B | true |
| `CICD` | **CICD** | 1.0.0 | Consume | ArgoCD/Istio | true |
| `MultiTenancy` | **MultiTenancy** | 1.0.0 | Consume | Capsule | true |
| `MLOps` | **MLOps** | 1.0.0 | Consume | MLflow | true |
| `Eval` | **Eval** | 1.0.0 | Consume | Promptfoo / DeepEval / Ragas | false |

### 1.2 Role semantics

| Role | Meaning | Implementation Constraints |
|------|------|----------|
| **Producer** | SDK actively adjusts platform services | Adapter must be implemented; interface signature frozen |
| **Consumer** | SDK consumption platform delivers context | Default pass-through; host can be customized |

---

## 2. Configuration Keys

> Corresponds to design/DESIGN.md §7. The configuration file is located in `infrastructure/config/client.yaml` and must be renderable by Metacang (§15.6.3).

### 2.1 Complete list of configuration keys

```yaml
# === Gateway（Required，core） ===
gateway:
  provider: higress              # str: "higress" | "apigw"
  base_url: ${GATEWAY_BASE_URL}  #env var: gateway address
  timeout_sec: 30                #int: call timeout (seconds)
  max_retries: 0                 #int: SDK layer does not retry (responsible by ModelRouter)

# === Model / LLMProvider（Required，core） ===
model:
  qwen:
    type: dashscope              # str: "dashscope" | "local_vllm"
    default: true                #bool: whether to be the default provider
  openai:
    type: openai                 # str: "openai" | "local_vllm"
  claude:
    type: anthropic              # str: "anthropic" | "bedrock"

# === Cache（Required，core） ===
cache:
  provider: redis                # str: "redis" | "valkey"
  ttl_sec: 3600                  #int: default cache TTL (seconds)
  endpoints: ["redis:6379"]      #list[str]: cluster address

# === VectorStore（Required，core） ===
vector_store:
  preference: qdrant             # str: "qdrant" | "milvus"
  collection_prefix: "kb-"       #str: collection prefix
  dimension: 1536                #int: default vector dimension

# === Observability（Required，core baseline） ===
observability:
  otel_traces: true              #bool: OTel traces (core baseline)
  audit_log: true                #bool: audit log (core baseline)
  metrics_hook_enabled: true     #bool: whether to trigger MetricsHook
  langfuse_enabled: false        #bool: Langfuse extra layer (optional)

# === Auth（Required，core） ===
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

### 2.2 Configure priority

```
explicit client.Option > environment variables > Configuration file snippet > platform Manifest Issue
```

1. **Explicit Option**: `client.WithGateway(...)` and other code construction parameters
2. **Environment variables**: `GATEWAY_BASE_URL`, `DASHSCOPE_API_KEY`, etc. (only for debugging; production uses platform Gateway + tenant Token)
3. **Configuration file**: `infrastructure/config/client.yaml` fragment
4. **Platform Manifest**: Control plane delivery (configuration fragments can be overwritten during runtime)

### 2.3 Credential security (§4.4.6)

| Constraints | Description |
|------|------|
| Tenant Token holder | SDK Client, verified by `Auth` port |
| Third-party Provider Key | Only exists in the platform Secret Vault, and the SDK will never handle it |
| Debugging environment variables | `DASHSCOPE_API_KEY` and others are only used for local debugging; production is disabled |
| Data outbound control | Gateway desensitizes PII before third-party calls; `no_egress` policy forces self-hosting |

---

## 3. Version and compatibility strategy

> Corresponds to design/DESIGN.md §8.

### 3.1 SDK own version (SemVer)

- Current version: `v1.4.0` (aligned with `strata v1.4.0`, §16.1)
- MAJOR: destructive API changes (such as interface method signature modification)
- MINOR: New port/function/Adapter, backward compatible
- PATCH: Bug fixes, performance optimization, document updates
- Destructive changes must be accompanied by ADR (`design/adr/`)
- Module path: `github.com/openstrata/ai-sdk-go`

### 3.2 SPI port version contract

The SDK promises compatibility with the following `interface_versions`:

| Port | Minimum compatible version | Constraints | Change policy |
|------|-----------------|------|----------|
| Gateway | 1.2.0 | OpenAI-compatible protocol remains unchanged | Add optional field → MINOR; Add required field → MAJOR |
| AgentRuntime | 1.3.0 | AgentSpec `apiVersion: openstrata.io/v1` | Spec added optional field for backward compatibility |
| LLMProvider | 1.0.0 | chat/embed/rerank/stream Signature stable | Signature change→MAJOR; New method→MINOR |
| VectorStore | 1.1.0 | upsert/search/delete stable | Return structure field changes→MAJOR |
| Cache | 1.0.0 | get/set stable | Same as above |

### 3.3 Cross-language consistency constraints

- `ai-sdk-go` / `ai-sdk-java` / `ai-sdk-python` three-piece set have consistent semantics for the method signatures of the same SPI port
- AgentSpec YAML schema three SDKs share the same fixture (`specs/fixtures/`)
- CI cross-language contract test verification: the same AgentSpec can be built by any language SDK and bound to any runtime instance for execution

### 3.4 Platform version frozen

| SDK version | platform strata | bom.yaml tag | Go module |
|----------|------------|-------------|-----------|
| v1.4.0 | v1.4.0 | v1.4.0 | `github.com/openstrata/ai-sdk-go v1.4.0` |

Change process:
1. SDK repository tag (such as `v1.5.0`)
2. CI verifies the consistency of `interface_versions` in `bom.yaml` (§16.4 / §15.6.4)
3. meta repository `repos.yaml` / `bom.yaml` nail version → Dependency Resolver assembly
4. Control plane `GET /v1/release/manifest` auditable deviation certification list

---

> **Associated documents**: arch/ARCH.md (architecture and port list) · skills/SKILLS.md (coding rules) · design/DESIGN.md (complete design) · bom.yaml §16
