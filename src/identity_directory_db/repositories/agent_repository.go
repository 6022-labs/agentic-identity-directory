package repositories

import (
	"context"
	"errors"
	"strings"

	domain_repositories "github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models/enums"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AgentRepository struct {
	client *identity_directory_db.DatabaseClient
}

func NewAgentRepository(client *identity_directory_db.DatabaseClient) *AgentRepository {
	return &AgentRepository{client: client}
}

func (r *AgentRepository) Upsert(ctx context.Context, agent *models.Agent) error {
	db := r.client.GetDB(ctx)
	err := db.Omit(clause.Associations).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "chain_id"}, {Name: "collection_address"}, {Name: "token_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "ens_name", "creator", "owner", "clone_of", "images", "attributes", "updated_at",
		}),
	}).Create(agent).Error
	if err != nil {
		return err
	}

	err = db.Where("chain_id = ? AND collection_address = ? AND token_id = ?",
		agent.ChainId, agent.CollectionAddress, agent.TokenId).
		Delete(&models.AgentAddress{}).Error
	if err != nil {
		return err
	}
	// Sanitizing can turn two on-chain addresses into one.
	addresses := lo.UniqBy(agent.Addresses, func(address models.AgentAddress) string {
		return address.Type + "\x00" + address.Value
	})
	if len(addresses) == 0 {
		return nil
	}
	return db.Create(&addresses).Error
}

func (r *AgentRepository) Get(ctx context.Context, chainId uint64, agentRef models.AgentRef) (*models.Agent, error) {
	return r.first(r.withRelations(ctx).Where(
		"agents.chain_id = ? AND agents.collection_address = ? AND agents.token_id = ?",
		chainId, agentRef.CollectionAddress, agentRef.TokenId,
	))
}

func (r *AgentRepository) GetByEnsName(ctx context.Context, ensName string) (*models.Agent, error) {
	return r.first(r.withRelations(ctx).Where("agents.ens_name = ?", ensName))
}

func (r *AgentRepository) Search(
	ctx context.Context,
	params *domain_repositories.AgentSearchParams,
) ([]*models.Agent, int64, error) {
	query := r.client.GetDB(ctx).Model(&models.Agent{}).
		Joins("LEFT JOIN agent_collections ON agent_collections.chain_id = agents.chain_id" +
			" AND agent_collections.address = agents.collection_address").
		Session(&gorm.Session{})

	if params.Query != "" {
		pattern := "%" + r.escapeLike(params.Query) + "%"
		query = query.Where(
			"agents.name ILIKE ? OR agent_collections.name ILIKE ? OR agent_collections.description ILIKE ?"+
				" OR agents.attributes::text ILIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}
	if params.Role != "" {
		query = query.Where("agents.attributes->>'role' = ?", params.Role)
	}
	if params.ChainId != 0 {
		query = query.Where("agents.chain_id = ?", params.ChainId)
	}
	if params.CollectionAddress != "" {
		query = query.Where("agents.collection_address = ?", params.CollectionAddress)
	}
	if params.Owner != "" {
		query = query.Where("agents.owner = ?", params.Owner)
	}
	if params.Address != "" {
		query = query.Where("EXISTS (SELECT 1 FROM agent_addresses WHERE agent_addresses.chain_id = agents.chain_id"+
			" AND agent_addresses.collection_address = agents.collection_address"+
			" AND agent_addresses.token_id = agents.token_id AND agent_addresses.value = ?)", params.Address)
	}
	if params.Published != nil {
		published := "EXISTS (SELECT 1 FROM agent_ens_records WHERE agent_ens_records.chain_id = agents.chain_id" +
			" AND agent_ens_records.collection_address = agents.collection_address" +
			" AND agent_ens_records.token_id = agents.token_id AND agent_ens_records.key = ?)"
		if !*params.Published {
			published = "NOT " + published
		}
		query = query.Where(published, enums.EnsRecordKeyUrl.String())
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var agents []*models.Agent
	err := query.Select("agents.*").
		Preload("Collection").Preload("Addresses").Preload("EnsRecords").
		Order("agents.chain_id, agents.collection_address, agents.token_id").
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&agents).Error
	return agents, total, err
}

func (r *AgentRepository) withRelations(ctx context.Context) *gorm.DB {
	return r.client.GetDB(ctx).Preload("Collection").Preload("Addresses").Preload("EnsRecords")
}

func (r *AgentRepository) first(query *gorm.DB) (*models.Agent, error) {
	var agent models.Agent
	err := query.First(&agent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) escapeLike(text string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(text)
}
