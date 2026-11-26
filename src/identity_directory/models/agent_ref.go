package models

// AgentRef identifies an agent on its chain; a wallet address never does, anyone can mint with someone else's.
type AgentRef struct {
	CollectionAddress string
	TokenId           string
}
