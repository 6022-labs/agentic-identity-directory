package requests_test

import (
	"math"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenComputingAPaginateOffset(t *testing.T) {
	t.Parallel()

	t.Run("Given the first page", func(t *testing.T) {
		t.Parallel()

		t.Run("Should skip nothing", func(t *testing.T) {
			t.Parallel()

			request := &requests.PaginateRequest{Page: new(uint(1)), ElementsPerPage: new(uint(10))}

			assert.Equal(t, 0, request.Offset())
		})
	})

	t.Run("Given the third page of 20", func(t *testing.T) {
		t.Parallel()

		t.Run("Should skip the first 40 elements", func(t *testing.T) {
			t.Parallel()

			request := &requests.PaginateRequest{Page: new(uint(3)), ElementsPerPage: new(uint(20))}

			assert.Equal(t, 40, request.Offset())
		})
	})

	t.Run("Given the last page Sanitize allows at the largest size", func(t *testing.T) {
		t.Parallel()

		t.Run("Should stay positive", func(t *testing.T) {
			t.Parallel()

			request := &requests.PaginateRequest{Page: new(uint(math.MaxInt32)), ElementsPerPage: new(uint(100))}

			assert.Positive(t, request.Offset())
		})
	})
}
