package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// parseYAML parses a minimal YAML subset: nested mappings (by 2-space indent),
// scalars (string/bool/number), inline flow maps {a: b}, inline flow sequences
// [a, b], and ${VAR} / ${VAR:-default} environment substitution.
func parseYAML(data []byte) (map[string]any, error) {
	root := map[string]any{}
	stack := []map[string]any{root}
	indents := []int{-1}

	lines := strings.Split(string(data), "\n")
	for n, raw := range lines {
		line := stripComment(raw)
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		key, value, hasValue := splitKeyValue(line)
		if key == "" {
			return nil, fmt.Errorf("config: cannot parse line %d: %q", n+1, raw)
		}
		for len(indents) > 1 && indent <= indents[len(indents)-1] {
			stack = stack[:len(stack)-1]
			indents = indents[:len(indents)-1]
		}
		parent := stack[len(stack)-1]

		if !hasValue {
			child := map[string]any{}
			parent[key] = child
			stack = append(stack, child)
			indents = append(indents, indent)
			continue
		}

		trimmed := strings.TrimSpace(value)
		var v any
		switch {
		case strings.HasPrefix(trimmed, "{"):
			m, err := parseFlowMap(trimmed)
			if err != nil {
				return nil, fmt.Errorf("config: line %d: %w", n+1, err)
			}
			v = m
		case strings.HasPrefix(trimmed, "["):
			s, err := parseFlowSeq(trimmed)
			if err != nil {
				return nil, fmt.Errorf("config: line %d: %w", n+1, err)
			}
			v = s
		default:
			v = parseScalar(trimmed)
		}
		parent[key] = v
	}
	return root, nil
}

func stripComment(line string) string {
	var inSingle, inDouble bool
	for i, r := range line {
		switch r {
		case '\'':
			inSingle = !inSingle
		case '"':
			inDouble = !inDouble
		case '#':
			if !inSingle && !inDouble && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t') {
				return line[:i]
			}
		}
	}
	return line
}

func splitKeyValue(line string) (key, value string, hasValue bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	rest := strings.TrimSpace(line[idx+1:])
	return key, rest, rest != ""
}

func parseScalar(s string) any {
	s = envSubst(s)
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	switch s {
	case "true":
		return true
	case "false":
		return false
	case "null", "~", "":
		return nil
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

func envSubst(s string) string {
	return os.Expand(s, func(name string) string {
		if idx := strings.Index(name, ":-"); idx >= 0 {
			key := name[:idx]
			def := name[idx+2:]
			if v, ok := os.LookupEnv(key); ok && v != "" {
				return v
			}
			return def
		}
		return os.Getenv(name)
	})
}

func parseFlowMap(s string) (map[string]any, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	m := map[string]any{}
	if strings.TrimSpace(s) == "" {
		return m, nil
	}
	for _, part := range splitTopLevel(s, ',') {
		key, value, _ := splitKeyValue(part)
		if key == "" {
			return nil, fmt.Errorf("invalid flow map entry %q", part)
		}
		m[key] = parseScalar(strings.TrimSpace(value))
	}
	return m, nil
}

func parseFlowSeq(s string) ([]any, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	var out []any
	if strings.TrimSpace(s) == "" {
		return out, nil
	}
	for _, part := range splitTopLevel(s, ',') {
		out = append(out, parseScalar(strings.TrimSpace(part)))
	}
	return out, nil
}

// splitTopLevel splits on sep that is not nested inside {} or [].
func splitTopLevel(s string, sep byte) []string {
	var parts []string
	var buf strings.Builder
	var depth int
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		}
		if c == sep && depth == 0 {
			parts = append(parts, buf.String())
			buf.Reset()
			continue
		}
		buf.WriteByte(c)
	}
	if buf.Len() > 0 || len(parts) > 0 {
		parts = append(parts, buf.String())
	}
	return parts
}

func applyFile(c *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	m, err := parseYAML(data)
	if err != nil {
		return err
	}
	if v, ok := getString(m, "gateway", "provider"); ok {
		c.Gateway.Provider = v
	}
	if v, ok := getString(m, "gateway", "baseUrl"); ok {
		c.Gateway.BaseURL = v
	}
	if v, ok := getString(m, "model", "qwen", "type"); ok {
		c.Model.Qwen.Type = v
	}
	if v, ok := getBool(m, "model", "qwen", "default"); ok {
		c.Model.Qwen.Default = v
	}
	if v, ok := getString(m, "model", "openai", "type"); ok {
		c.Model.OpenAI.Type = v
	}
	if v, ok := getString(m, "model", "claude", "type"); ok {
		c.Model.Claude.Type = v
	}
	if v, ok := getString(m, "cache", "provider"); ok {
		c.Cache.Provider = v
	}
	if v, ok := getString(m, "vectorStore", "preference"); ok {
		c.VectorStore.Preference = v
	}
	if v, ok := getBool(m, "observability", "otelTraces"); ok {
		c.Observability.OTelTraces = v
	}
	if v, ok := getBool(m, "observability", "auditLog"); ok {
		c.Observability.AuditLog = v
	}
	return nil
}

func getString(m map[string]any, keys ...string) (string, bool) {
	cur := any(m)
	for _, k := range keys {
		mm, ok := cur.(map[string]any)
		if !ok {
			return "", false
		}
		cur, ok = mm[k]
		if !ok {
			return "", false
		}
	}
	s, ok := cur.(string)
	if !ok {
		return "", false
	}
	return s, true
}

func getBool(m map[string]any, keys ...string) (bool, bool) {
	cur := any(m)
	for _, k := range keys {
		mm, ok := cur.(map[string]any)
		if !ok {
			return false, false
		}
		cur, ok = mm[k]
		if !ok {
			return false, false
		}
	}
	b, ok := cur.(bool)
	if !ok {
		return false, false
	}
	return b, true
}
