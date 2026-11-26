package models

import (
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/helpers"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models/enums"
)

type Agent struct {
	ChainId           uint64            `gorm:"primaryKey;autoIncrement:false"`
	CollectionAddress string            `gorm:"primaryKey"`
	TokenId           string            `gorm:"primaryKey;type:numeric(78,0)"`
	Name              string            `gorm:"not null"`
	EnsName           string            `gorm:"not null;index"`
	Creator           string            `gorm:"not null"`
	Owner             string            `gorm:"not null;index"`
	CloneOf           *string           `gorm:"type:numeric(78,0)"`
	Images            map[string]string `gorm:"serializer:json;type:jsonb;not null"`
	Attributes        map[string]string `gorm:"serializer:json;type:jsonb;not null"`
	CreatedAt         time.Time
	UpdatedAt         time.Time

	Collection *AgentCollection `gorm:"foreignKey:ChainId,CollectionAddress;references:ChainId,Address;constraint:-"`
	Addresses  []AgentAddress   `gorm:"foreignKey:ChainId,CollectionAddress,TokenId;references:ChainId,CollectionAddress,TokenId;constraint:-"`
	EnsRecords []AgentEnsRecord `gorm:"foreignKey:ChainId,CollectionAddress,TokenId;references:ChainId,CollectionAddress,TokenId;constraint:-"`
}

func (a *Agent) Role() string {
	return a.Attributes[enums.AgentAttributeKeyRole.String()]
}

// Url is the agent's published endpoint, empty until it publishes its ENS records.
func (a *Agent) Url() string {
	for _, record := range a.EnsRecords {
		if record.Key == enums.EnsRecordKeyUrl.String() {
			return record.Value
		}
	}
	return ""
}

func (a *Agent) EvmAddresses() []string {
	addresses := make([]string, 0, len(a.Addresses))
	for _, address := range a.Addresses {
		if address.Type == enums.AddressTypeEvm.String() {
			addresses = append(addresses, address.Value)
		}
	}
	return addresses
}

// Sanitize cleans every on-chain string; their values are written by anyone who mints.
func (a *Agent) Sanitize() {
	a.Name = helpers.SanitizeText(a.Name)
	a.Images = helpers.SanitizeTexts(a.Images)
	a.Attributes = helpers.SanitizeTexts(a.Attributes)
	for index := range a.Addresses {
		a.Addresses[index].Sanitize()
	}
}
