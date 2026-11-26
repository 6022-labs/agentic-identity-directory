package interfaces

import "context"

// UnitOfWorkInterface is one atomic scope, owned by a single goroutine.
type UnitOfWorkInterface interface {
	// Context carries the scope; repository calls made with it join the transaction.
	Context() context.Context
	Commit() error
	// Rollback does nothing once committed.
	Rollback() error
}
