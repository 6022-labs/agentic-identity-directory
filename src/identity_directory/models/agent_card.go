package models

// AgentCard is an agent's /.well-known/6022 document; Raw is the body the signatures cover.
type AgentCard struct {
	Protocol        string
	ProtocolVersion string
	Name            string
	EnsDomain       string
	Endpoints       map[string]string
	PaymentMethods  []string
	Signatures      []AgentCardSignature
	Raw             []byte
}

type AgentCardSignature struct {
	Protected string
	Signature string
}
