package models

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/helpers"

type AgentEnsRecord struct {
	ChainId           uint64 `gorm:"primaryKey;autoIncrement:false"`
	CollectionAddress string `gorm:"primaryKey"`
	TokenId           string `gorm:"primaryKey;type:numeric(78,0)"`
	Key               string `gorm:"primaryKey"`
	Value             string `gorm:"not null"`
}

func (r *AgentEnsRecord) Sanitize() {
	r.Key = helpers.SanitizeText(r.Key)
	r.Value = helpers.SanitizeText(r.Value)
}
