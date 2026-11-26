package common_tests

import (
	"context"

	common_services_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/services_mocks/interfaces_mocks"
	"go.uber.org/mock/gomock"
)

type unitOfWorkScopeKey struct{}

// ScopedContext is the context a unit of work started by ExpectUnitOfWork hands out.
func ScopedContext() context.Context {
	return context.WithValue(context.Background(), unitOfWorkScopeKey{}, true)
}

// InScope matches only a context derived from ScopedContext, proving a call joined the unit of work.
func InScope() gomock.Matcher {
	return gomock.Cond(func(ctx context.Context) bool {
		scoped, _ := ctx.Value(unitOfWorkScopeKey{}).(bool)
		return scoped
	})
}

// OutOfScope matches a context that is not the unit of work's.
func OutOfScope() gomock.Matcher {
	return gomock.Not(InScope())
}

// ExpectUnitOfWork starts one unit of work whose Context is ScopedContext and that is rolled back once.
func ExpectUnitOfWork(
	mockUnitOfWorkFactory *common_services_mocks.MockUnitOfWorkFactoryInterface,
	mockUnitOfWork *common_services_mocks.MockUnitOfWorkInterface,
) {
	mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(mockUnitOfWork, nil)
	mockUnitOfWork.EXPECT().Context().Return(ScopedContext()).AnyTimes()
	mockUnitOfWork.EXPECT().Rollback().Return(nil)
}
