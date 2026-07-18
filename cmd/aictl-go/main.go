// Command aictl-go is an optional demo of ai-sdk-go (not a release subject).
// It runs fully offline using the stdlib default adapters.
package main

import (
	"context"
	"fmt"

	"github.com/open-strata-ai/ai-sdk-go/pkg/agent"
	"github.com/open-strata-ai/ai-sdk-go/pkg/client"
	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

func main() {
	c := client.NewDefault(nil)

	resp, err := c.Chat(context.Background(), domain.ChatRequest{
		Messages: []domain.ChatMessage{{Role: "user", Content: "Where is my order?"}},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("chat:", resp.Content)

	spec := agent.New(c.AgentRuntime(), "customer-service-v2").
		Tenant("tenant-b").
		ModelBinding(&domain.ModelBinding{Primary: "cloud-qwen-max"}).
		InputSchema(domain.NewObjectSchema("query", "string")).
		OutputSchema(domain.NewObjectSchema("answer", "string")).
		Guardrails(&domain.Guardrails{Checks: []string{"injection_scan", "pii_scan", "rate_limit"}}).
		Build()

	h, err := c.AgentRuntime().Load(context.Background(), spec)
	if err != nil {
		panic(err)
	}
	out, err := c.AgentRuntime().Run(context.Background(), h, map[string]any{"query": "Where's my order?"})
	if err != nil {
		panic(err)
	}
	fmt.Println("agent run:", out)

	tool := domain.NewTool("order_lookup").
		Description("Check order logistics status").
		InputSchema(domain.NewObjectSchema("order_id", "string")).
		Handler(func(ctx domain.ToolContext, in map[string]any) (map[string]any, error) {
			return map[string]any{"status": "shipped"}, nil
		}).Build()
	res, _ := tool.Run(domain.ToolContext{TenantID: "tenant-b"}, map[string]any{"order_id": "123"})
	fmt.Println("tool:", res)
}
