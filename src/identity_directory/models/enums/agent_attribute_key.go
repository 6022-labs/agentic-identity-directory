package enums

type AgentAttributeKey string

// AgentAttributeKeyRole is mandatory at mint.
const AgentAttributeKeyRole AgentAttributeKey = "role"

func (k AgentAttributeKey) String() string {
	return string(k)
}
