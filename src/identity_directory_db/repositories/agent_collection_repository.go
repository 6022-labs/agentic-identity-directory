package repositories

import (
	"context"
	"errors"

	domain_repositories "github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AgentCollectionRepository struct {
	client *identity_directory_db.DatabaseClient
}

func NewAgentCollectionRepository(client *identity_directory_db.DatabaseClient) *AgentCollectionRepository {
	return &AgentCollectionRepository{client: client}
}

func (r *AgentCollectionRepository) Upsert(ctx context.Context, agentCollection *models.AgentCollection) error {
	return r.client.GetDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chain_id"}, {Name: "address"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "symbol", "description", "updated_at"}),
	}).Create(agentCollection).Error
}

func (r *AgentCollectionRepository) Get(
	ctx context.Context,
	chainId uint64,
	address string,
) (*models.AgentCollection, error) {
	var collection models.AgentCollection
	err := r.client.GetDB(ctx).First(&collection, "chain_id = ? AND address = ?", chainId, address).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

func (r *AgentCollectionRepository) ListAddresses(ctx context.Context, chainId uint64) ([]string, error) {
	var addresses []string
	err := r.client.GetDB(ctx).Model(&models.AgentCollection{}).
		Where("chain_id = ?", chainId).
		Order("address").
		Pluck("address", &addresses).Error
	return addresses, err
}

func (r *AgentCollectionRepository) List(
	ctx context.Context,
	params *domain_repositories.AgentCollectionListParams,
) ([]*models.AgentCollection, int64, error) {
	query := r.client.GetDB(ctx).Model(&models.AgentCollection{})
	if params.ChainId != 0 {
		query = query.Where("chain_id = ?", params.ChainId)
	}
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var collections []*models.AgentCollection
	err := query.Order("chain_id, name, address").
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&collections).Error
	return collections, total, err
}
