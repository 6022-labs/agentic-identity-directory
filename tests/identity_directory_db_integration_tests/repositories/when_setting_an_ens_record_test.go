package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenSettingAnEnsRecordTestingSuite struct {
	*databaseTestingSuite
	sut   *repositories.AgentEnsRecordRepository
	agent *models.Agent
}

func WhenSettingAnEnsRecordBeforeEach() *WhenSettingAnEnsRecordTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenSettingAnEnsRecordTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewAgentEnsRecordRepository(databaseTestingSuite.dbClient),
		agent:                databaseTestingSuite.seedAgent(model_builders.NewAgentBuilder().Build()),
	}
}

func (s *WhenSettingAnEnsRecordTestingSuite) record(key, value string) *models.AgentEnsRecord {
	return &models.AgentEnsRecord{
		ChainId: s.agent.ChainId, CollectionAddress: s.agent.CollectionAddress, TokenId: s.agent.TokenId,
		Key: key, Value: value,
	}
}

func (s *WhenSettingAnEnsRecordTestingSuite) storedRecords(t *testing.T) map[string]string {
	agent, err := repositories.NewAgentRepository(s.dbClient).Get(context.Background(), s.agent.ChainId,
		models.AgentRef{CollectionAddress: s.agent.CollectionAddress, TokenId: s.agent.TokenId})
	require.NoError(t, err)
	records := map[string]string{}
	for _, record := range agent.EnsRecords {
		records[record.Key] = record.Value
	}
	return records
}

func TestWhenSettingAnEnsRecord(t *testing.T) {
	t.Parallel()

	t.Run("Given a key already set", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep the last value", func(t *testing.T) {
			t.Parallel()
			suite := WhenSettingAnEnsRecordBeforeEach()
			defer suite.close()
			require.NoError(t, suite.sut.Set(context.Background(), suite.record("url", "https://a")))

			err := suite.sut.Set(context.Background(), suite.record("url", "https://b"))

			require.NoError(t, err)
			assert.Equal(t, map[string]string{"url": "https://b"}, suite.storedRecords(t))
		})
	})

	t.Run("Given an empty value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should delete only that record", func(t *testing.T) {
			t.Parallel()
			suite := WhenSettingAnEnsRecordBeforeEach()
			defer suite.close()
			require.NoError(t, suite.sut.Set(context.Background(), suite.record("url", "https://a")))
			require.NoError(t, suite.sut.Set(context.Background(), suite.record("com.github", "scout")))

			err := suite.sut.Set(context.Background(), suite.record("com.github", ""))

			require.NoError(t, err)
			assert.Equal(t, map[string]string{"url": "https://a"}, suite.storedRecords(t))
		})
	})
}
