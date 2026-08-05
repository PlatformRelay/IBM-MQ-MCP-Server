package mqadmin

import "errors"

var (
	// ErrQueueNameRequired indicates a queue name was not supplied.
	ErrQueueNameRequired = errors.New("queue name is required")
	// ErrInvalidQueueType indicates an unsupported queue type was requested.
	ErrInvalidQueueType = errors.New("queue type must be LOCAL, ALIAS, REMOTE, or MODEL")
	// ErrAlterNoChanges indicates an alter request carried no supported attributes.
	ErrAlterNoChanges = errors.New("alter request must include at least one supported attribute")
)
