package services

import "fmt"

// EnsNameBuilder names agents the way AgentEnsRegistry provisions them under 6022.eth.
type EnsNameBuilder struct{}

func NewEnsNameBuilder() *EnsNameBuilder {
	return &EnsNameBuilder{}
}

func (b *EnsNameBuilder) Build(chainId uint64, collectionName, agentName string) string {
	return fmt.Sprintf("%s.%s.%d.6022.eth", agentName, collectionName, chainId)
}
