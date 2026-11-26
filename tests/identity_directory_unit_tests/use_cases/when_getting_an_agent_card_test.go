package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenGettingAnAgentCardTestingSuite struct {
	mockAgentRepository            *repositories_mocks.MockAgentRepositoryInterface
	mockAgentCardRequester         *interfaces_mocks.MockAgentCardRequesterInterface
	mockAgentCardSignatureVerifier *interfaces_mocks.MockAgentCardSignatureVerifierInterface
	sut                            *use_cases.GetAgentCard
}

func WhenGettingAnAgentCardBeforeEach(t *testing.T) *WhenGettingAnAgentCardTestingSuite {
	controller := gomock.NewController(t)
	suite := &WhenGettingAnAgentCardTestingSuite{
		mockAgentRepository:            repositories_mocks.NewMockAgentRepositoryInterface(controller),
		mockAgentCardRequester:         interfaces_mocks.NewMockAgentCardRequesterInterface(controller),
		mockAgentCardSignatureVerifier: interfaces_mocks.NewMockAgentCardSignatureVerifierInterface(controller),
	}
	suite.sut = use_cases.NewGetAgentCard(
		suite.mockAgentRepository,
		suite.mockAgentCardRequester,
		suite.mockAgentCardSignatureVerifier,
	)
	return suite
}

func (s *WhenGettingAnAgentCardTestingSuite) assertUnavailable(t *testing.T, err error, code string) {
	var unavailableErr *apperrors.UnavailableError
	require.ErrorAs(t, err, &unavailableErr)
	assert.Equal(t, code, unavailableErr.Code)
}

func TestWhenGettingAnAgentCard(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail with a validation error on ensName", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)

			response, err := suite.sut.Execute(context.Background(), &requests.EnsNameRequest{})

			assert.Nil(t, response)
			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "ensName", validationErr.Field)
		})
	})

	request := &requests.EnsNameRequest{EnsName: "scout.agents.80002.6022.eth"}

	t.Run("Given the index cannot be read", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				GetByEnsName(gomock.Any(), request.EnsName).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should fail with get_agent_failed", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			require.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "get_agent_failed", internalErr.Code)
		})
	})

	t.Run("Given no agent has the ENS name", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(nil, nil)
		}

		t.Run("Should fail with agent_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var notFoundErr *apperrors.NotFoundError
			require.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "agent_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given an agent that has not published its url", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(agent, nil)
		}

		t.Run("Should fail with agent_url_not_published", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var notFoundErr *apperrors.NotFoundError
			require.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "agent_url_not_published", notFoundErr.Code)
		})
	})

	t.Run("Given an agent whose url is not http", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().WithEnsRecord("url", "file:///etc/passwd").Build()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(agent, nil)
		}

		t.Run("Should fail with agent_url_invalid without fetching it", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "agent_url_invalid")
		})
	})

	agent := model_builders.NewAgentBuilder().WithEnsRecord("url", "https://scout.example").Build()

	t.Run("Given the card cannot be fetched", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(agent, nil)
			suite.mockAgentCardRequester.EXPECT().
				GetCard(gomock.Any(), "https://scout.example").
				Return(nil, errors.New("connection refused"))
		}

		t.Run("Should fail with agent_card_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "agent_card_unavailable")
		})
	})

	t.Run("Given a card of another protocol", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(agent, nil)
			suite.mockAgentCardRequester.EXPECT().
				GetCard(gomock.Any(), "https://scout.example").
				Return(&models.AgentCard{Protocol: "a2a"}, nil)
		}

		t.Run("Should fail with agent_card_invalid", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "agent_card_invalid")
		})
	})

	agentCard := &models.AgentCard{
		Protocol:        "6022",
		ProtocolVersion: "0.1",
		Name:            "scout",
		EnsDomain:       agent.EnsName,
		Endpoints:       map[string]string{"responses": "https://scout.example/api/responses"},
		PaymentMethods:  []string{"x402"},
	}

	t.Run("Given a card not signed by the agent", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(agent, nil)
			suite.mockAgentCardRequester.EXPECT().GetCard(gomock.Any(), "https://scout.example").Return(agentCard, nil)
			suite.mockAgentCardSignatureVerifier.EXPECT().
				Verify(agentCard, agent.EvmAddresses()).
				Return(false, errors.New("signed by someone else"))
		}

		t.Run("Should fail with agent_card_signature_invalid", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "agent_card_signature_invalid")
		})
	})

	t.Run("Given a card signed by the agent's address but naming another agent", func(t *testing.T) {
		t.Parallel()

		replayedCard := &models.AgentCard{Protocol: "6022", EnsDomain: "victim.agents.80002.6022.eth"}

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(agent, nil)
			suite.mockAgentCardRequester.EXPECT().
				GetCard(gomock.Any(), "https://scout.example").
				Return(replayedCard, nil)
			suite.mockAgentCardSignatureVerifier.EXPECT().Verify(replayedCard, agent.EvmAddresses()).Return(true, nil)
		}

		t.Run("Should fail with agent_card_invalid", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			suite.assertUnavailable(t, err, "agent_card_invalid")
		})
	})

	t.Run("Given an unsigned card", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(agent, nil)
			suite.mockAgentCardRequester.EXPECT().GetCard(gomock.Any(), "https://scout.example").Return(agentCard, nil)
			suite.mockAgentCardSignatureVerifier.EXPECT().Verify(agentCard, agent.EvmAddresses()).Return(false, nil)
		}

		t.Run("Should return the card as not verified", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			require.NoError(t, err)
			assert.False(t, response.Verified)
			assert.Equal(t, agentCard.Endpoints, response.Endpoints)
		})
	})

	t.Run("Given a card signed by the agent", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(agent, nil)
			suite.mockAgentCardRequester.EXPECT().GetCard(gomock.Any(), "https://scout.example").Return(agentCard, nil)
			suite.mockAgentCardSignatureVerifier.EXPECT().Verify(agentCard, agent.EvmAddresses()).Return(true, nil)
		}

		t.Run("Should return the card as verified", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			require.NoError(t, err)
			assert.True(t, response.Verified)
			assert.Equal(t, agent.EnsName, response.EnsName)
			assert.Equal(t, "https://scout.example", response.Url)
			assert.Equal(t, agentCard.Endpoints, response.Endpoints)
			assert.Equal(t, []string{"x402"}, response.PaymentMethods)
		})
	})
}
