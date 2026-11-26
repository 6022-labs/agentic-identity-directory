package models

// CollectionChanges is what a block range touched; the logs only say that it changed, never to what.
type CollectionChanges struct {
	CollectionAddresses []string
	AgentRefs           []AgentRef
}
