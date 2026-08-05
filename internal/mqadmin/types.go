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
