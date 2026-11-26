package enums

type EnsRecordKey string

// EnsRecordKeyUrl is the agent's endpoint, mandatory once published and never cleared.
const EnsRecordKeyUrl EnsRecordKey = "url"

func (k EnsRecordKey) String() string {
	return string(k)
}
