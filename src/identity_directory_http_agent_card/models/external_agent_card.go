package models

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"

type ExternalAgentCard struct {
	Protocol        string                       `json:"protocol"`
	ProtocolVersion string                       `json:"protocolVersion"`
	Name            string                       `json:"name"`
	EnsDomain       string                       `json:"ensDomain"`
	Endpoints       map[string]string            `json:"endpoints"`
	PaymentMethods  []string                     `json:"paymentMethods"`
	Signatures      []ExternalAgentCardSignature `json:"signatures"`
}

type ExternalAgentCardSignature struct {
	Protected string `json:"protected"`
	Signature string `json:"signature"`
}

func (c *ExternalAgentCard) ToModel(raw []byte) *models.AgentCard {
	signatures := make([]models.AgentCardSignature, 0, len(c.Signatures))
	for _, signature := range c.Signatures {
		signatures = append(signatures, models.AgentCardSignature{
			Protected: signature.Protected,
			Signature: signature.Signature,
		})
	}
	return &models.AgentCard{
		Protocol:        c.Protocol,
		ProtocolVersion: c.ProtocolVersion,
		Name:            c.Name,
		EnsDomain:       c.EnsDomain,
		Endpoints:       c.Endpoints,
		PaymentMethods:  c.PaymentMethods,
		Signatures:      signatures,
		Raw:             raw,
	}
}
