package identity_directory_db

import (
	"context"

	"gorm.io/gorm"
)

// transactionKey carries the open Transaction so repositories join it without knowing it exists.
type transactionKey struct{}

// Transaction is one open database transaction, owned by a single goroutine.
type Transaction struct {
	tx     *gorm.DB
	closed bool
}

func (t *Transaction) Commit() error {
	if t.closed {
		return nil
	}
	t.closed = true
	return t.tx.Commit().Error
}

func (t *Transaction) Rollback() error {
	if t.closed {
		return nil
	}
	t.closed = true
	return t.tx.Rollback().Error
}

// BeginTransaction opens a transaction and returns a context whose repository calls join it.
func (c *DatabaseClient) BeginTransaction(ctx context.Context) (context.Context, *Transaction, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	transaction := &Transaction{tx: tx}
	return context.WithValue(ctx, transactionKey{}, transaction), transaction, nil
}
