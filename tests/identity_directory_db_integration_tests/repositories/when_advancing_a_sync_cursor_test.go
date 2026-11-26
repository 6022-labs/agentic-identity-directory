package repositories_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenAdvancingASyncCursorTestingSuite struct {
	*databaseTestingSuite
	sut *repositories.SyncCursorRepository
}

func WhenAdvancingASyncCursorBeforeEach() *WhenAdvancingASyncCursorTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenAdvancingASyncCursorTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewSyncCursorRepository(databaseTestingSuite.dbClient),
	}
}

func (s *WhenAdvancingASyncCursorTestingSuite) storedBlock(t *testing.T) *uint64 {
	block, err := s.sut.Get(context.Background(), 80002)
	require.NoError(t, err)
	return block
}

func TestWhenAdvancingASyncCursor(t *testing.T) {
	t.Parallel()

	t.Run("Given a chain never indexed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should store the first cursor", func(t *testing.T) {
			t.Parallel()
			suite := WhenAdvancingASyncCursorBeforeEach()
			defer suite.close()

			err := suite.sut.Advance(context.Background(), 80002, nil, 150)

			require.NoError(t, err)
			assert.Equal(t, uint64(150), *suite.storedBlock(t))
		})
	})

	t.Run("Given the cursor still at the expected block", func(t *testing.T) {
		t.Parallel()

		t.Run("Should move it", func(t *testing.T) {
			t.Parallel()
			suite := WhenAdvancingASyncCursorBeforeEach()
			defer suite.close()
			require.NoError(t, suite.sut.Advance(context.Background(), 80002, nil, 150))

			err := suite.sut.Advance(context.Background(), 80002, new(uint64(150)), 300)

			require.NoError(t, err)
			assert.Equal(t, uint64(300), *suite.storedBlock(t))
		})
	})

	t.Run("Given another replica stored the first cursor", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail with ErrSyncCursorMoved", func(t *testing.T) {
			t.Parallel()
			suite := WhenAdvancingASyncCursorBeforeEach()
			defer suite.close()
			require.NoError(t, suite.sut.Advance(context.Background(), 80002, nil, 150))

			err := suite.sut.Advance(context.Background(), 80002, nil, 160)

			assert.ErrorIs(t, err, apperrors.ErrSyncCursorMoved)
			assert.Equal(t, uint64(150), *suite.storedBlock(t))
		})
	})

	t.Run("Given another replica moved the cursor", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail with ErrSyncCursorMoved", func(t *testing.T) {
			t.Parallel()
			suite := WhenAdvancingASyncCursorBeforeEach()
			defer suite.close()
			require.NoError(t, suite.sut.Advance(context.Background(), 80002, nil, 150))

			err := suite.sut.Advance(context.Background(), 80002, new(uint64(140)), 160)

			assert.ErrorIs(t, err, apperrors.ErrSyncCursorMoved)
			assert.Equal(t, uint64(150), *suite.storedBlock(t))
		})
	})
}
