package mqadmin

import "errors"

var (
	// ErrChannelNameRequired indicates a channel name was not supplied.
	ErrChannelNameRequired = errors.New("channel name is required")
	// ErrInvalidChannelType indicates an unsupported channel type was requested.
	ErrInvalidChannelType = errors.New(
		"channel type must be SDR, SVR, RCVR, RQSTR, CLNTCONN, SVRCONN, CLUSSDR, or CLUSRCVR",
	)

	// ErrCHLAUTHChannelRequired indicates a CHLAUTH channel name was not supplied.
	ErrCHLAUTHChannelRequired = errors.New("chlauth channel name is required")
	// ErrInvalidCHLAUTHType indicates an unsupported CHLAUTH rule type was requested.
	ErrInvalidCHLAUTHType = errors.New(
		"chlauth rule type must be ADDRESSMAP, BLOCKUSER, USERMAP, QMGRMAP, or SSLPEERMAP",
	)
	// ErrCHLAUTHIdentityIncomplete indicates required CHLAUTH identity fields are missing.
	ErrCHLAUTHIdentityIncomplete = errors.New("chlauth target identity is incomplete for the rule type")
	// ErrCHLAUTHUserSourceRequired indicates USERSRC was not supplied for define.
	ErrCHLAUTHUserSourceRequired = errors.New("chlauth userSource is required")
	// ErrInvalidCHLAUTHUserSource indicates an unsupported USERSRC value was requested.
	ErrInvalidCHLAUTHUserSource = errors.New("chlauth userSource must be NOACCESS, CHANNEL, or MAP")

	// ErrAuthrecProfileRequired indicates an authority-record profile was not supplied.
	ErrAuthrecProfileRequired = errors.New("authrec profile is required")
	// ErrInvalidAuthrecObjectType indicates an unsupported authrec object type was requested.
	ErrInvalidAuthrecObjectType = errors.New(
		"authrec objectType must be QUEUE, QMGRC, CHANNEL, PROCESS, TOPIC, or NAMELIST",
	)
	// ErrInvalidAuthrecEntityType indicates an unsupported authrec entity type was requested.
	ErrInvalidAuthrecEntityType = errors.New("authrec entityType must be PRINCIPAL or GROUP")
	// ErrAuthrecIdentityIncomplete indicates required authrec identity fields are missing.
	ErrAuthrecIdentityIncomplete = errors.New("authrec target identity is incomplete")
	// ErrAuthrecAuthoritiesRequired indicates no authorities were supplied for define.
	ErrAuthrecAuthoritiesRequired = errors.New("authrec authorities are required")
	// ErrInvalidAuthrecAuthority indicates an unsupported authority grant was requested.
	ErrInvalidAuthrecAuthority = errors.New("authrec authority value is not supported")
)
