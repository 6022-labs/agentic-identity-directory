package repositories

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db"
	"gorm.io/gorm/clause"
)

type AgentEnsRecordRepository struct {
	client *identity_directory_db.DatabaseClient
}

func NewAgentEnsRecordRepository(client *identity_directory_db.DatabaseClient) *AgentEnsRecordRepository {
	return &AgentEnsRecordRepository{client: client}
}

func (r *AgentEnsRecordRepository) Set(ctx context.Context, agentEnsRecord *models.AgentEnsRecord) error {
	db := r.client.GetDB(ctx)
	if agentEnsRecord.Value == "" {
		return db.Where("chain_id = ? AND collection_address = ? AND token_id = ? AND key = ?",
			agentEnsRecord.ChainId, agentEnsRecord.CollectionAddress, agentEnsRecord.TokenId, agentEnsRecord.Key).
			Delete(&models.AgentEnsRecord{}).Error
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "chain_id"}, {Name: "collection_address"}, {Name: "token_id"}, {Name: "key"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(agentEnsRecord).Error
}
