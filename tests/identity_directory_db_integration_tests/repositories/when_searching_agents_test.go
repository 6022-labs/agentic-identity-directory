package repositories_test

import (
	"context"
	"testing"

	domain_repositories "github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenSearchingAgentsTestingSuite struct {
	*databaseTestingSuite
	sut *repositories.AgentRepository
}

func WhenSearchingAgentsBeforeEach() *WhenSearchingAgentsTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenSearchingAgentsTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewAgentRepository(databaseTestingSuite.dbClient),
	}
}

func (s *WhenSearchingAgentsTestingSuite) search(
	t *testing.T,
	params domain_repositories.AgentSearchParams,
) ([]string, int64) {
	if params.Limit == 0 {
		params.Limit = 10
	}
	agents, total, err := s.sut.Search(context.Background(), &params)
	require.NoError(t, err)
	names := []string{}
	for _, agent := range agents {
		names = append(names, agent.Name)
	}
	return names, total
}

func TestWhenSearchingAgents(t *testing.T) {
	t.Parallel()

	// initSuite indexes a published expert scout and an unpublished facilitator trader in one collection.
	initSuite := func(suite *WhenSearchingAgentsTestingSuite) (*models.Agent, *models.Agent) {
		collection := suite.seedCollection(model_builders.NewAgentCollectionBuilder().
			WithName("markets").
			WithDescription("100% on-chain").
			Build())
		scout, trader := suite.seedScoutAndTrader(collection)
		suite.seedEnsRecord(scout, "url", "https://scout.example")
		return scout, trader
	}

	t.Run("Given no filter", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return every agent with its collection", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			initSuite(suite)

			agents, total, err := suite.sut.Search(
				context.Background(),
				&domain_repositories.AgentSearchParams{Limit: 10},
			)

			require.NoError(t, err)
			assert.Equal(t, int64(2), total)
			require.Len(t, agents, 2)
			assert.Equal(t, "markets", agents[0].Collection.Name)
		})
	})

	t.Run("Given a page smaller than the result", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return that page with the full total", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			initSuite(suite)

			names, total := suite.search(t, domain_repositories.AgentSearchParams{Offset: 1, Limit: 1})

			assert.Equal(t, []string{"trader"}, names)
			assert.Equal(t, int64(2), total)
		})
	})

	t.Run("Given a query word in an attribute", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match it case-insensitively", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			initSuite(suite)

			names, _ := suite.search(t, domain_repositories.AgentSearchParams{Query: "BONDS"})

			assert.Equal(t, []string{"trader"}, names)
		})
	})

	t.Run("Given a query with LIKE wildcards", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match a percent sign literally", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			initSuite(suite)

			names, _ := suite.search(t, domain_repositories.AgentSearchParams{Query: "%"})

			assert.Equal(t, []string{"scout", "trader"}, names)
		})

		t.Run("Should match an underscore literally", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			initSuite(suite)

			names, _ := suite.search(t, domain_repositories.AgentSearchParams{Query: "e_b"})

			assert.Equal(t, []string{"trader"}, names)
		})
	})

	t.Run("Given a role", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return only the agents with that role", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			initSuite(suite)

			names, _ := suite.search(t, domain_repositories.AgentSearchParams{Role: "expert"})

			assert.Equal(t, []string{"scout"}, names)
		})
	})

	t.Run("Given another chain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nothing", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			initSuite(suite)

			names, total := suite.search(t, domain_repositories.AgentSearchParams{ChainId: 137})

			assert.Empty(t, names)
			assert.Zero(t, total)
		})
	})

	t.Run("Given a collection address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the agents of that collection only", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			scout, _ := initSuite(suite)
			suite.seedAgent(model_builders.NewAgentBuilder().WithName("stranger").Build())

			names, _ := suite.search(
				t,
				domain_repositories.AgentSearchParams{CollectionAddress: scout.CollectionAddress},
			)

			assert.Equal(t, []string{"scout", "trader"}, names)
		})
	})

	t.Run("Given a collection address no agent belongs to", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nothing", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			initSuite(suite)

			names, total := suite.search(t,
				domain_repositories.AgentSearchParams{CollectionAddress: common_tests.RandomLowerEthAddress()})

			assert.Empty(t, names)
			assert.Zero(t, total)
		})
	})

	t.Run("Given an owner", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the owner's agents only", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			scout, _ := initSuite(suite)

			names, _ := suite.search(t, domain_repositories.AgentSearchParams{Owner: scout.Owner})

			assert.Equal(t, []string{"scout"}, names)
		})
	})

	t.Run("Given one of an agent's addresses", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return that agent", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			_, trader := initSuite(suite)

			names, _ := suite.search(t, domain_repositories.AgentSearchParams{Address: trader.Addresses[0].Value})

			assert.Equal(t, []string{"trader"}, names)
		})
	})

	t.Run("Given the published filter", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the agents that published a url when true", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			initSuite(suite)

			names, _ := suite.search(t, domain_repositories.AgentSearchParams{Published: new(true)})

			assert.Equal(t, []string{"scout"}, names)
		})

		t.Run("Should return the agents that did not when false", func(t *testing.T) {
			t.Parallel()
			suite := WhenSearchingAgentsBeforeEach()
			defer suite.close()
			initSuite(suite)

			names, _ := suite.search(t, domain_repositories.AgentSearchParams{Published: new(false)})

			assert.Equal(t, []string{"trader"}, names)
		})
	})
}
