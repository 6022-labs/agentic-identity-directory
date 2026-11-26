package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenGettingASyncCursorTestingSuite struct {
	*databaseTestingSuite
	sut *repositories.SyncCursorRepository
}

func WhenGettingASyncCursorBeforeEach() *WhenGettingASyncCursorTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenGettingASyncCursorTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewSyncCursorRepository(databaseTestingSuite.dbClient),
	}
}

func TestWhenGettingASyncCursor(t *testing.T) {
	t.Parallel()

	t.Run("Given a chain never indexed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nil without an error", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingASyncCursorBeforeEach()
			defer suite.close()

			block, err := suite.sut.Get(context.Background(), 80002)

			require.NoError(t, err)
			assert.Nil(t, block)
		})
	})

	t.Run("Given a chain indexed up to a block", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return that block", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingASyncCursorBeforeEach()
			defer suite.close()
			require.NoError(t, suite.sut.Advance(context.Background(), 80002, nil, 150))

			block, err := suite.sut.Get(context.Background(), 80002)

			require.NoError(t, err)
			assert.Equal(t, uint64(150), *block)
		})
	})
}
