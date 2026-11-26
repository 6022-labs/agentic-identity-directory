package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	metrics_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/metrics/interfaces"
	http_models "github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/models"
)

const (
	wellKnown6022Path = "/.well-known/6022"

	// Bounds what a hostile agent can make us buffer and canonicalize.
	maxAgentCardBytes = 1 << 20
)

type AgentCardRequester struct {
	httpClient       *http.Client
	agentCardMetrics metrics_interfaces.AgentCardMetricsInterface
}

func NewAgentCardRequester(
	httpClient *http.Client,
	agentCardMetrics metrics_interfaces.AgentCardMetricsInterface,
) *AgentCardRequester {
	return &AgentCardRequester{httpClient: httpClient, agentCardMetrics: agentCardMetrics}
}

func (r *AgentCardRequester) GetCard(ctx context.Context, agentUrl string) (*models.AgentCard, error) {
	start := time.Now()
	agentCard, err := r.getCard(ctx, agentUrl)
	r.agentCardMetrics.RecordFetch(ctx, time.Since(start), err)
	return agentCard, err
}

func (r *AgentCardRequester) getCard(ctx context.Context, agentUrl string) (*models.AgentCard, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		strings.TrimRight(agentUrl, "/")+wellKnown6022Path,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := r.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch card: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxAgentCardBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read card: %w", err)
	}
	if len(body) > maxAgentCardBytes {
		return nil, fmt.Errorf("card exceeds %d bytes", maxAgentCardBytes)
	}

	var card http_models.ExternalAgentCard
	if err := json.Unmarshal(body, &card); err != nil {
		return nil, fmt.Errorf("decode card: %w", err)
	}
	return card.ToModel(body), nil
}
