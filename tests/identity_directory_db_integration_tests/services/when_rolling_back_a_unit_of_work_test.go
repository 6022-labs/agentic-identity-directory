package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type WhenRollingBackAUnitOfWorkTestingSuite struct {
	*unitOfWorkTestingSuite
}

func WhenRollingBackAUnitOfWorkBeforeEach() *WhenRollingBackAUnitOfWorkTestingSuite {
	return &WhenRollingBackAUnitOfWorkTestingSuite{unitOfWorkTestingSuite: newUnitOfWorkTestingSuite()}
}

func TestWhenRollingBackAUnitOfWork(t *testing.T) {
	t.Parallel()

	t.Run("Given writes made through the scope", func(t *testing.T) {
		t.Parallel()

		t.Run("Should discard them", func(t *testing.T) {
			t.Parallel()
			suite := WhenRollingBackAUnitOfWorkBeforeEach()
			defer suite.close()

			unitOfWork, err := suite.sut.Start(context.Background())
			assert.NoError(t, err)
			collection := suite.createCollection(t, unitOfWork.Context())

			assert.NoError(t, unitOfWork.Rollback())
			assert.False(t, suite.collectionExists(t, collection))
		})
	})

	t.Run("Given writes made outside the scope", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep them", func(t *testing.T) {
			t.Parallel()
			suite := WhenRollingBackAUnitOfWorkBeforeEach()
			defer suite.close()

			unitOfWork, err := suite.sut.Start(context.Background())
			assert.NoError(t, err)
			collection := suite.createCollection(t, context.Background())

			assert.NoError(t, unitOfWork.Rollback())
			assert.True(t, suite.collectionExists(t, collection))
		})
	})
}
