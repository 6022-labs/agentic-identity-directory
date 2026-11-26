package models

import (
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/helpers"
)

type AgentCollection struct {
	ChainId     uint64 `gorm:"primaryKey;autoIncrement:false"`
	Address     string `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Symbol      string `gorm:"not null"`
	Description string `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (c *AgentCollection) Sanitize() {
	c.Name = helpers.SanitizeText(c.Name)
	c.Symbol = helpers.SanitizeText(c.Symbol)
	c.Description = helpers.SanitizeText(c.Description)
}
