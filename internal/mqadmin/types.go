package mqadmin

import "time"

// Availability describes whether live MQ data was observed for a profile.
type Availability string

const (
	// Available means live MQ data was observed successfully.
	Available Availability = "available"
	// Stale means observed identity differs from configured catalog values.
	Stale Availability = "stale"
	// Unavailable means mqweb or the queue manager could not be reached.
	Unavailable Availability = "unavailable"
)

// Identity distinguishes configured catalog values from observed MQ responses.
type Identity struct {
	Configured string `json:"configured"`
	Observed   string `json:"observed,omitempty"`
}

// QueueManagerStatus reports configured vs observed queue manager identity and reachability.
type QueueManagerStatus struct {
	Profile      string       `json:"profile"`
	Identity     Identity     `json:"identity"`
	Availability Availability `json:"availability"`
	Running      bool         `json:"running,omitempty"`
	StatusText   string       `json:"status,omitempty"`
	LastChecked  time.Time    `json:"lastChecked"`
	Error        string       `json:"error,omitempty"`
}

// QueueSummary is a lightweight queue listing entry.
type QueueSummary struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// QueueDetail includes definition and live status for one queue.
type QueueDetail struct {
	Name            string `json:"name"`
	Type            string `json:"type,omitempty"`
	MaxDepth        int    `json:"maxDepth,omitempty"`
	CurrentDepth    int    `json:"currentDepth"`
	OpenInputCount  int    `json:"openInputCount,omitempty"`
	OpenOutputCount int    `json:"openOutputCount,omitempty"`
	InhibitGet      string `json:"inhibitGet,omitempty"`
	InhibitPut      string `json:"inhibitPut,omitempty"`
	Description     string `json:"description,omitempty"`
	MKuratorTag     string `json:"-"`
}

// ListQueuesFilter narrows queue listing without accepting arbitrary MQSC.
type ListQueuesFilter struct {
	NamePrefix string
	QueueType  string
}

// ListQueuesRequest carries pagination and filters for queue listing.
type ListQueuesRequest struct {
	Filter ListQueuesFilter
	Limit  int
	Cursor string
}

// ChannelSummary is a lightweight channel listing entry.
type ChannelSummary struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// ChannelDetail holds channel definition attributes from mqweb.
type ChannelDetail struct {
	Name              string `json:"name"`
	Type              string `json:"type,omitempty"`
	Description       string `json:"description,omitempty"`
	ConnectionName    string `json:"connectionName,omitempty"`
	TransmissionQueue string `json:"transmissionQueue,omitempty"`
	MKuratorTag       string `json:"-"`
}

// ChannelStatus reports runtime channel state separately from definition.
type ChannelStatus struct {
	Name         string       `json:"name"`
	Type         string       `json:"type,omitempty"`
	State        string       `json:"state,omitempty"`
	Availability Availability `json:"availability"`
	StatusText   string       `json:"status,omitempty"`
	LastChecked  time.Time    `json:"lastChecked"`
	Error        string       `json:"error,omitempty"`
}

// ListChannelsFilter narrows channel listing without accepting arbitrary MQSC.
type ListChannelsFilter struct {
	NamePrefix  string
	ChannelType string
}

// ListChannelsRequest carries pagination and filters for channel listing.
type ListChannelsRequest struct {
	Filter ListChannelsFilter
	Limit  int
	Cursor string
}

// ListenerSummary is a lightweight listener listing entry.
type ListenerSummary struct {
	Name string `json:"name"`
}

// ListenerDetail holds listener definition attributes from mqweb.
type ListenerDetail struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Port        int    `json:"port,omitempty"`
	Transport   string `json:"transport,omitempty"`
}

// ListenerStatus reports runtime listener state separately from definition.
type ListenerStatus struct {
	Name         string       `json:"name"`
	State        string       `json:"state,omitempty"`
	Availability Availability `json:"availability"`
	StatusText   string       `json:"status,omitempty"`
	LastChecked  time.Time    `json:"lastChecked"`
	Error        string       `json:"error,omitempty"`
}

// ListListenersFilter narrows listener listing.
type ListListenersFilter struct {
	NamePrefix string
}

// ListListenersRequest carries pagination and filters for listener listing.
type ListListenersRequest struct {
	Filter ListListenersFilter
	Limit  int
	Cursor string
}

// SubscriptionSummary is a lightweight subscription listing entry.
type SubscriptionSummary struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	TopicString string `json:"topicString,omitempty"`
	Type        string `json:"type,omitempty"`
}

// SubscriptionDetail holds subscription definition attributes from mqweb.
type SubscriptionDetail struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	TopicString string `json:"topicString,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Destination string `json:"destination,omitempty"`
}

// ListSubscriptionsFilter narrows subscription listing.
type ListSubscriptionsFilter struct {
	NamePrefix string
}

// ListSubscriptionsRequest carries pagination and filters for subscription listing.
type ListSubscriptionsRequest struct {
	Filter ListSubscriptionsFilter
	Limit  int
	Cursor string
}
