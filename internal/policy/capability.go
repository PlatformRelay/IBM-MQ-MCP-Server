package policy

// Capability names an operation grant from ADR-0003.
type Capability string

// Operation-oriented grants from ADR-0003.
const (
	Inspect     Capability = "inspect"
	Browse      Capability = "browse"
	Consume     Capability = "consume"
	Produce     Capability = "produce"
	Administer  Capability = "administer"
	ExecuteMQSC Capability = "execute_mqsc"
)

var known = map[Capability]struct{}{
	Inspect:     {},
	Browse:      {},
	Consume:     {},
	Produce:     {},
	Administer:  {},
	ExecuteMQSC: {},
}

// AllCapabilities returns the canonical capability vocabulary.
func AllCapabilities() []Capability {
	return []Capability{
		Inspect,
		Browse,
		Consume,
		Produce,
		Administer,
		ExecuteMQSC,
	}
}

// IsKnown reports whether name is a valid capability string.
func IsKnown(name Capability) bool {
	_, ok := known[name]
	return ok
}
