package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseYAML(t *testing.T) {
	src := `
gateway:
  provider: higress
  baseUrl: ${GATEWAY_BASE_URL}
model:
  qwen: { type: dashscope, default: true }
  openai: { type: openai }
cache:
  provider: redis
vectorStore:
  preference: qdrant
observability:
  otelTraces: true
  auditLog: true
`
	m, err := parseYAML([]byte(src))
	if err != nil {
		t.Fatalf("parseYAML: %v", err)
	}
	if v, ok := getString(m, "gateway", "provider"); !ok || v != "higress" {
		t.Fatalf("gateway.provider = %q,%v", v, ok)
	}
	if v, ok := getString(m, "model", "qwen", "type"); !ok || v != "dashscope" {
		t.Fatalf("model.qwen.type = %q,%v", v, ok)
	}
	if v, ok := getBool(m, "model", "qwen", "default"); !ok || !v {
		t.Fatalf("model.qwen.default = %v,%v", v, ok)
	}
	if v, ok := getString(m, "cache", "provider"); !ok || v != "redis" {
		t.Fatalf("cache.provider = %q,%v", v, ok)
	}
	if v, ok := getString(m, "vectorStore", "preference"); !ok || v != "qdrant" {
		t.Fatalf("vectorStore.preference = %q,%v", v, ok)
	}
	if v, ok := getBool(m, "observability", "otelTraces"); !ok || !v {
		t.Fatalf("observability.otelTraces = %v,%v", v, ok)
	}
	if v, ok := getBool(m, "observability", "auditLog"); !ok || !v {
		t.Fatalf("observability.auditLog = %v,%v", v, ok)
	}
}

func TestEnvSubstitution(t *testing.T) {
	if err := os.Setenv("MY_TEST_URL", "http://gw"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	defer os.Unsetenv("MY_TEST_URL")
	m, err := parseYAML([]byte("gateway:\n  baseUrl: ${MY_TEST_URL:-http://default}"))
	if err != nil {
		t.Fatalf("parseYAML: %v", err)
	}
	if v, ok := getString(m, "gateway", "baseUrl"); !ok || v != "http://gw" {
		t.Fatalf("baseUrl = %q,%v", v, ok)
	}

	// default branch
	m2, err := parseYAML([]byte("gateway:\n  baseUrl: ${UNSET_VAR:-http://default}"))
	if err != nil {
		t.Fatalf("parseYAML: %v", err)
	}
	if v, ok := getString(m2, "gateway", "baseUrl"); !ok || v != "http://default" {
		t.Fatalf("baseUrl default = %q,%v", v, ok)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "client.yaml")
	content := `
gateway:
  provider: higress
  baseUrl: ${GATEWAY_BASE_URL}
model:
  qwen: { type: dashscope, default: true }
cache:
  provider: redis
vectorStore:
  preference: milvus
observability:
  otelTraces: true
  auditLog: true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	c, err := Load(FromFile(path))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Cache.Provider != "redis" {
		t.Fatalf("cache.provider = %q", c.Cache.Provider)
	}
	if c.VectorStore.Preference != "milvus" {
		t.Fatalf("vectorStore.preference = %q", c.VectorStore.Preference)
	}
	if c.Model.Qwen.Type != "dashscope" || !c.Model.Qwen.Default {
		t.Fatalf("model.qwen = %+v", c.Model.Qwen)
	}
	if !c.Observability.OTelTraces || !c.Observability.AuditLog {
		t.Fatalf("observability = %+v", c.Observability)
	}
}

func TestLoadFromEnvOverride(t *testing.T) {
	if err := os.Setenv("OPENSTRATA_VECTORSTORE_PREFERENCE", "weaviate"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	defer os.Unsetenv("OPENSTRATA_VECTORSTORE_PREFERENCE")
	c, err := Load(FromEnv())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.VectorStore.Preference != "weaviate" {
		t.Fatalf("preference = %q", c.VectorStore.Preference)
	}
}
