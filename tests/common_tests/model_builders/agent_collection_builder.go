package model_builders

import (
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
)

type AgentCollectionBuilder struct {
	collection *models.AgentCollection
}

func NewAgentCollectionBuilder() *AgentCollectionBuilder {
	return &AgentCollectionBuilder{collection: &models.AgentCollection{
		ChainId:     80002,
		Address:     common_tests.RandomLowerEthAddress(),
		Name:        "agents",
		Symbol:      "AGENTS",
		Description: "agents that scout markets",
	}}
}

func (b *AgentCollectionBuilder) WithChainId(chainId uint64) *AgentCollectionBuilder {
	b.collection.ChainId = chainId
	return b
}

func (b *AgentCollectionBuilder) WithAddress(address string) *AgentCollectionBuilder {
	b.collection.Address = address
	return b
}

func (b *AgentCollectionBuilder) WithName(name string) *AgentCollectionBuilder {
	b.collection.Name = name
	return b
}

func (b *AgentCollectionBuilder) WithDescription(description string) *AgentCollectionBuilder {
	b.collection.Description = description
	return b
}

func (b *AgentCollectionBuilder) Build() *models.AgentCollection {
	return b.collection
}
