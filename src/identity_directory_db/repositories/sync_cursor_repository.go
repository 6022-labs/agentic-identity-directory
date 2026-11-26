package repositories

import (
	"context"
	"errors"

	common_errors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SyncCursorRepository struct {
	client *identity_directory_db.DatabaseClient
}

func NewSyncCursorRepository(client *identity_directory_db.DatabaseClient) *SyncCursorRepository {
	return &SyncCursorRepository{client: client}
}

func (r *SyncCursorRepository) Get(ctx context.Context, chainId uint64) (*uint64, error) {
	var cursor models.SyncCursor
	err := r.client.GetDB(ctx).First(&cursor, "chain_id = ?", chainId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cursor.Block, nil
}

func (r *SyncCursorRepository) Advance(ctx context.Context, chainId uint64, expected *uint64, block uint64) error {
	db := r.client.GetDB(ctx)
	var result *gorm.DB
	if expected == nil {
		result = db.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&models.SyncCursor{ChainId: chainId, Block: block})
	} else {
		result = db.Model(&models.SyncCursor{}).
			Where("chain_id = ? AND block = ?", chainId, *expected).
			Updates(map[string]any{"block": block, "updated_at": gorm.Expr("now()")})
	}
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return common_errors.ErrSyncCursorMoved
	}
	return nil
}
