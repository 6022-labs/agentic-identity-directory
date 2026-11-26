package models

// ChainWindow is everything a block range changed, ready to be written at once.
type ChainWindow struct {
	ChainId     uint64
	FromBlock   uint64
	ToBlock     uint64
	Collections []*AgentCollection
	Agents      []*Agent
	EnsRecords  []*AgentEnsRecord
}
