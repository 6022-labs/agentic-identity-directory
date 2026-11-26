package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/clients"
	metrics_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/metrics/interfaces"
	"github.com/samber/lo"
)

// AgentCardRequester reads an agent's 6022 card through the agent-node gateway client.
type AgentCardRequester struct {
	client           clients.ClientWithResponsesInterface
	agentCardMetrics metrics_interfaces.AgentCardMetricsInterface
}

func NewAgentCardRequester(
	client clients.ClientWithResponsesInterface,
	agentCardMetrics metrics_interfaces.AgentCardMetricsInterface,
) *AgentCardRequester {
	return &AgentCardRequester{client: client, agentCardMetrics: agentCardMetrics}
}

func (r *AgentCardRequester) GetCard(ctx context.Context, agentUrl string) (*models.AgentCard, error) {
	start := time.Now()
	agentCard, err := r.getCard(ctx, agentUrl)
	r.agentCardMetrics.RecordFetch(ctx, time.Since(start), err)
	return agentCard, err
}

func (r *AgentCardRequester) getCard(ctx context.Context, agentUrl string) (*models.AgentCard, error) {
	target, err := url.Parse(agentUrl)
	if err != nil || target.Host == "" {
		return nil, fmt.Errorf("agent url %q is not an absolute URL", agentUrl)
	}

	response, err := r.client.GetWellKnown6022WithResponse(ctx, r.toAgent(target))
	if err != nil {
		return nil, fmt.Errorf("fetch card: %w", err)
	}
	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", response.StatusCode())
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("the answer is not a JSON card")
	}
	return r.toModel(response.JSON200, response.Body), nil
}

// toAgent sends the request to the agent's own origin: the shared client only knows a placeholder server.
func (r *AgentCardRequester) toAgent(target *url.URL) clients.RequestEditorFn {
	return func(_ context.Context, request *http.Request) error {
		request.URL.Scheme = target.Scheme
		request.URL.Host = target.Host
		request.URL.Path = strings.TrimSuffix(target.Path, "/") + request.URL.Path
		request.Host = target.Host
		return nil
	}
}

// toModel keeps the raw body: the signatures cover those exact bytes, not a re-encoding.
func (r *AgentCardRequester) toModel(card *clients.ResponsesWellKnownAgent6022, raw []byte) *models.AgentCard {
	agentCard := &models.AgentCard{
		Protocol:        lo.FromPtr(card.Protocol),
		ProtocolVersion: lo.FromPtr(card.ProtocolVersion),
		Name:            lo.FromPtr(card.Name),
		EnsDomain:       lo.FromPtr(card.EnsDomain),
		Endpoints:       map[string]string{},
		PaymentMethods:  lo.FromPtr(card.PaymentMethods),
		Raw:             raw,
	}
	if card.Endpoints != nil {
		if card.Endpoints.Responses != nil {
			agentCard.Endpoints["responses"] = *card.Endpoints.Responses
		}
		if card.Endpoints.ChatCompletions != nil {
			agentCard.Endpoints["chatCompletions"] = *card.Endpoints.ChatCompletions
		}
	}
	for _, signature := range lo.FromPtr(card.Signatures) {
		agentCard.Signatures = append(agentCard.Signatures, models.AgentCardSignature{
			Protected: lo.FromPtr(signature.Protected),
			Signature: lo.FromPtr(signature.Signature),
		})
	}
	return agentCard
}
