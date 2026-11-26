package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type WhenCommittingAUnitOfWorkTestingSuite struct {
	*unitOfWorkTestingSuite
}

func WhenCommittingAUnitOfWorkBeforeEach() *WhenCommittingAUnitOfWorkTestingSuite {
	return &WhenCommittingAUnitOfWorkTestingSuite{unitOfWorkTestingSuite: newUnitOfWorkTestingSuite()}
}

func TestWhenCommittingAUnitOfWork(t *testing.T) {
	t.Parallel()

	t.Run("Given writes made through the scope", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep them", func(t *testing.T) {
			t.Parallel()
			suite := WhenCommittingAUnitOfWorkBeforeEach()
			defer suite.close()

			unitOfWork, err := suite.sut.Start(context.Background())
			assert.NoError(t, err)
			collection := suite.createCollection(t, unitOfWork.Context())

			assert.NoError(t, unitOfWork.Commit())
			assert.True(t, suite.collectionExists(t, collection))
		})
	})

	t.Run("Given a committed scope rolled back by its defer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep the writes", func(t *testing.T) {
			t.Parallel()
			suite := WhenCommittingAUnitOfWorkBeforeEach()
			defer suite.close()

			unitOfWork, err := suite.sut.Start(context.Background())
			assert.NoError(t, err)
			collection := suite.createCollection(t, unitOfWork.Context())
			assert.NoError(t, unitOfWork.Commit())

			assert.NoError(t, unitOfWork.Rollback())
			assert.True(t, suite.collectionExists(t, collection))
		})
	})
}
