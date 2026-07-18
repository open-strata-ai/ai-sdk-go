package adapter

import (
	"context"
	"fmt"

	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

// LocalAuth is the default Auth adapter: accepts any non-blank token as tenant "local".
type LocalAuth struct{}

// NewLocalAuth builds a LocalAuth.
func NewLocalAuth() *LocalAuth { return &LocalAuth{} }

func (a *LocalAuth) ValidateToken(ctx context.Context, token string) (domain.TenantContext, error) {
	if token == "" {
		return domain.TenantContext{}, fmt.Errorf("openstrata: blank token")
	}
	return domain.TenantContext{TenantID: "local", Roles: []string{"developer"}, Claims: map[string]string{}}, nil
}
