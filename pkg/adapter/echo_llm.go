package adapter

import (
	"context"

	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

// EchoLLMProvider is the default LLMProvider adapter: offline echo.
// Hosts override it with a real provider (OpenAI/Qwen/Claude).
type EchoLLMProvider struct{}

// NewEchoLLMProvider builds an EchoLLMProvider.
func NewEchoLLMProvider() *EchoLLMProvider { return &EchoLLMProvider{} }

func (e *EchoLLMProvider) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
	last := ""
	if len(req.Messages) > 0 {
		last = req.Messages[len(req.Messages)-1].Content
	}
	return domain.ChatResponse{Content: "echo: " + last, Model: "local", Usage: map[string]any{}}, nil
}

func (e *EchoLLMProvider) Embed(ctx context.Context, req domain.EmbedRequest) (domain.EmbedResponse, error) {
	vecs := make([][]float32, len(req.Inputs))
	for i := range req.Inputs {
		vecs[i] = []float32{}
	}
	return domain.EmbedResponse{Vectors: vecs}, nil
}

func (e *EchoLLMProvider) Rerank(ctx context.Context, req domain.RerankRequest) (domain.RerankResponse, error) {
	items := make([]domain.RerankItem, len(req.Documents))
	for i := range req.Documents {
		items[i] = domain.RerankItem{Index: i, Score: 1.0 - float64(i)*0.01}
	}
	return domain.RerankResponse{Results: items}, nil
}

func (e *EchoLLMProvider) Stream(ctx context.Context, req domain.ChatRequest) (<-chan domain.StreamChunk, error) {
	ch := make(chan domain.StreamChunk, 1)
	go func() {
		defer close(ch)
		last := ""
		if len(req.Messages) > 0 {
			last = req.Messages[len(req.Messages)-1].Content
		}
		ch <- domain.StreamChunk{Delta: "echo: " + last, Done: true}
	}()
	return ch, nil
}
