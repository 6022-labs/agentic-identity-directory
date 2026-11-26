package model_builders

import (
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
)

type AgentBuilder struct {
	agent *models.Agent
}

func NewAgentBuilder() *AgentBuilder {
	collectionAddress := common_tests.RandomLowerEthAddress()
	evmAddress := common_tests.RandomLowerEthAddress()
	return &AgentBuilder{agent: &models.Agent{
		ChainId:           80002,
		CollectionAddress: collectionAddress,
		TokenId:           "1",
		Name:              "scout",
		EnsName:           "scout.agents.80002.6022.eth",
		Creator:           common_tests.RandomLowerEthAddress(),
		Owner:             common_tests.RandomLowerEthAddress(),
		Images:            map[string]string{"default": "bafybeibsulc7sgpck3w7ur7eqa6fwaazcsswn3dg3rfl65zka3knajx43i"},
		Attributes:        map[string]string{"role": "expert", "mission": "scouting"},
		Addresses: []models.AgentAddress{{
			ChainId: 80002, CollectionAddress: collectionAddress, TokenId: "1", Type: "evm", Value: evmAddress,
		}},
	}}
}

func (b *AgentBuilder) WithChainId(chainId uint64) *AgentBuilder {
	b.agent.ChainId = chainId
	for index := range b.agent.Addresses {
		b.agent.Addresses[index].ChainId = chainId
	}
	return b
}

func (b *AgentBuilder) WithCollectionAddress(collectionAddress string) *AgentBuilder {
	b.agent.CollectionAddress = collectionAddress
	for index := range b.agent.Addresses {
		b.agent.Addresses[index].CollectionAddress = collectionAddress
	}
	return b
}

func (b *AgentBuilder) WithTokenId(tokenId string) *AgentBuilder {
	b.agent.TokenId = tokenId
	for index := range b.agent.Addresses {
		b.agent.Addresses[index].TokenId = tokenId
	}
	return b
}

func (b *AgentBuilder) WithName(name string) *AgentBuilder {
	b.agent.Name = name
	return b
}

func (b *AgentBuilder) WithEnsName(ensName string) *AgentBuilder {
	b.agent.EnsName = ensName
	return b
}

func (b *AgentBuilder) WithOwner(owner string) *AgentBuilder {
	b.agent.Owner = owner
	return b
}

func (b *AgentBuilder) WithAttributes(attributes map[string]string) *AgentBuilder {
	b.agent.Attributes = attributes
	return b
}

func (b *AgentBuilder) WithAddresses(addresses ...models.AgentAddress) *AgentBuilder {
	b.agent.Addresses = addresses
	return b
}

func (b *AgentBuilder) WithEnsRecord(key, value string) *AgentBuilder {
	b.agent.EnsRecords = append(b.agent.EnsRecords, models.AgentEnsRecord{
		ChainId: b.agent.ChainId, CollectionAddress: b.agent.CollectionAddress, TokenId: b.agent.TokenId,
		Key: key, Value: value,
	})
	return b
}

func (b *AgentBuilder) WithCollection(collection *models.AgentCollection) *AgentBuilder {
	b.agent.Collection = collection
	return b
}

func (b *AgentBuilder) Build() *models.Agent {
	return b.agent
}
