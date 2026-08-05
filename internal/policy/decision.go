package policy

// Decision records one capability authorization outcome for audit hooks.
type Decision struct {
	Profile   string
	Required  Capability
	Granted   bool
	Operation string
}
