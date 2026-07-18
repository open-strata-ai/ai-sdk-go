package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

// HTTPGateway is a Gateway adapter that calls the OpenStrata platform Gateway
// over HTTP (OpenAI-compatible transport). It uses only the standard library.
type HTTPGateway struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewHTTPGateway builds an HTTPGateway for the given base URL and tenant token.
func NewHTTPGateway(baseURL, token string) *HTTPGateway {
	return &HTTPGateway{baseURL: strings.TrimRight(baseURL, "/"), token: token, client: http.DefaultClient}
}

// MustHTTPGatewayFromEnv builds an HTTPGateway from GATEWAY_BASE_URL and
// OPENSTRATA_TOKEN; it returns an error if GATEWAY_BASE_URL is unset.
func MustHTTPGatewayFromEnv() (*HTTPGateway, error) {
	base := os.Getenv("GATEWAY_BASE_URL")
	if base == "" {
		return nil, fmt.Errorf("openstrata: GATEWAY_BASE_URL not set")
	}
	return NewHTTPGateway(base, os.Getenv("OPENSTRATA_TOKEN")), nil
}

func (g *HTTPGateway) Invoke(ctx context.Context, req domain.GatewayRequest) (domain.GatewayResponse, error) {
	method := req.Method
	if method == "" {
		method = http.MethodPost
	}
	var bodyReader io.Reader
	if req.Body != nil {
		b, err := json.Marshal(req.Body)
		if err != nil {
			return domain.GatewayResponse{}, err
		}
		bodyReader = bytes.NewReader(b)
	}
	url := g.baseURL + req.Path
	httpReq, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return domain.GatewayResponse{}, err
	}
	if g.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+g.token)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	resp, err := g.client.Do(httpReq)
	if err != nil {
		return domain.GatewayResponse{}, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.GatewayResponse{}, err
	}
	return domain.GatewayResponse{Status: resp.StatusCode, Body: string(respBody), Headers: map[string]string{}}, nil
}
