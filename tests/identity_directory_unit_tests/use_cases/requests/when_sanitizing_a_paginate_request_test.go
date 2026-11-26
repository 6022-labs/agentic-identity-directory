package requests_test

import (
	"math"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenSanitizingAPaginateRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given no page and no size", func(t *testing.T) {
		t.Parallel()

		t.Run("Should default to the first page of 10", func(t *testing.T) {
			t.Parallel()

			request := &requests.PaginateRequest{}

			request.Sanitize()

			assert.Equal(t, uint(1), *request.Page)
			assert.Equal(t, uint(10), *request.ElementsPerPage)
		})
	})

	t.Run("Given a size above 100", func(t *testing.T) {
		t.Parallel()

		t.Run("Should clamp it to 100", func(t *testing.T) {
			t.Parallel()

			elementsPerPage := uint(500)
			request := &requests.PaginateRequest{ElementsPerPage: &elementsPerPage}

			request.Sanitize()

			assert.Equal(t, uint(100), *request.ElementsPerPage)
		})
	})

	t.Run("Given a page beyond what an offset can hold", func(t *testing.T) {
		t.Parallel()

		t.Run("Should clamp it so the offset stays positive", func(t *testing.T) {
			t.Parallel()

			page := uint(math.MaxUint64)
			elementsPerPage := uint(100)
			request := &requests.PaginateRequest{Page: &page, ElementsPerPage: &elementsPerPage}

			request.Sanitize()

			assert.Equal(t, uint(math.MaxInt32), *request.Page)
		})
	})
}
