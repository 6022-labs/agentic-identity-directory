package models

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/helpers"

type AgentAddress struct {
	ChainId           uint64 `gorm:"primaryKey;autoIncrement:false"`
	CollectionAddress string `gorm:"primaryKey"`
	TokenId           string `gorm:"primaryKey;type:numeric(78,0)"`
	Type              string `gorm:"primaryKey;index:idx_agent_addresses_type_value,priority:1"`
	Value             string `gorm:"primaryKey;index:idx_agent_addresses_type_value,priority:2"`
}

func (a *AgentAddress) Sanitize() {
	a.Type = helpers.SanitizeText(a.Type)
	a.Value = helpers.SanitizeText(a.Value)
}
