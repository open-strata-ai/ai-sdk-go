package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

// NoOpTracing is the default Tracing adapter: a no-op span (OTel/Langfuse wired by the host).
type NoOpTracing struct{}

// NewNoOpTracing builds a NoOpTracing.
func NewNoOpTracing() *NoOpTracing { return &NoOpTracing{} }

func (t *NoOpTracing) StartSpan(ctx context.Context, name string) (context.Context, domain.Span) {
	now := time.Now().UnixNano()
	return ctx, domain.Span{ID: fmt.Sprintf("%d", now), Name: name, StartNanos: now}
}
