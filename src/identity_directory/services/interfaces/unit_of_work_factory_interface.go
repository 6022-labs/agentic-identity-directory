package interfaces

import "context"

type UnitOfWorkFactoryInterface interface {
	Start(ctx context.Context) (UnitOfWorkInterface, error)
}
