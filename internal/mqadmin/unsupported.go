package mqadmin

import (
	"errors"
	"fmt"
)

// UnsupportedError means the connected mqweb deployment does not expose an object family.
type UnsupportedError struct {
	Family  string `json:"family"`
	Message string `json:"message"`
	Action  string `json:"action"`
}

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("unsupported %s: %s — %s", e.Family, e.Message, e.Action)
}

// UnsupportedFamily reports that mqweb does not expose REST resources for family.
func UnsupportedFamily(family string) error {
	return &UnsupportedError{
		Family:  family,
		Message: fmt.Sprintf("mqweb deployment does not expose %s REST resources", family),
		Action:  "connect to a full IBM MQ installation with the administrative REST API enabled",
	}
}

// AsUnsupportedError extracts an UnsupportedError from err when present.
func AsUnsupportedError(err error) (*UnsupportedError, bool) {
	var ue *UnsupportedError
	if errors.As(err, &ue) {
		return ue, true
	}
	return nil, false
}
