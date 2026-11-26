package enums

// AddressType is the kind of address an agent declares on-chain.
type AddressType string

// AddressTypeEvm is the only type that controls the agent and signs its card.
const AddressTypeEvm AddressType = "evm"

func (t AddressType) String() string {
	return string(t)
}
