package interfaces

type EnsNameBuilderInterface interface {
	Build(chainId uint64, collectionName, agentName string) string
}
