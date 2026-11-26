package services

import (
	"context"

	identity_directory_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory/services/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db"
)

type UnitOfWorkFactory struct {
	client *identity_directory_db.DatabaseClient
}

func NewUnitOfWorkFactory(client *identity_directory_db.DatabaseClient) *UnitOfWorkFactory {
	return &UnitOfWorkFactory{client: client}
}

func (f *UnitOfWorkFactory) Start(ctx context.Context) (identity_directory_interfaces.UnitOfWorkInterface, error) {
	scopedCtx, transaction, err := f.client.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return &unitOfWork{ctx: scopedCtx, transaction: transaction}, nil
}

type unitOfWork struct {
	ctx         context.Context
	transaction *identity_directory_db.Transaction
}

func (u *unitOfWork) Context() context.Context { return u.ctx }
func (u *unitOfWork) Commit() error            { return u.transaction.Commit() }
func (u *unitOfWork) Rollback() error          { return u.transaction.Rollback() }
