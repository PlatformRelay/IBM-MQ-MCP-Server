// Package fake provides test doubles for mqadmin.Client.
package fake

import (
	"context"
	"sync"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

// Client implements mqadmin.Client for tests and records invocation counts.
type Client struct {
	Name string

	mu              sync.Mutex
	QMStatusCalls   int
	ListQueuesCalls int
	GetQueueCalls   int
	PingCalls       int

	QMStatus       mqadmin.QueueManagerStatus
	ListQueuesPage collection.Page[mqadmin.QueueSummary]
	Queue          mqadmin.QueueDetail

	QMStatusErr   error
	ListQueuesErr error
	GetQueueErr   error
	PingErr       error
	Closed        bool
}

// New returns a fake admin client for the given profile name.
func New(name string) *Client {
	return &Client{Name: name}
}

// ProfileName returns the configured profile name.
func (c *Client) ProfileName() string { return c.Name }

// Ping records the call and returns PingErr when set.
func (c *Client) Ping(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.PingCalls++
	return c.PingErr
}

// Close marks the fake client closed.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Closed = true
	return nil
}

// QueueManagerStatus records the call and returns configured stub data.
func (c *Client) QueueManagerStatus(_ context.Context, _ string) (mqadmin.QueueManagerStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.QMStatusCalls++
	if c.QMStatusErr != nil {
		return mqadmin.QueueManagerStatus{}, c.QMStatusErr
	}
	status := c.QMStatus
	if status.Profile == "" {
		status.Profile = c.Name
	}
	return status, nil
}

// ListQueues records the call and returns configured stub data.
func (c *Client) ListQueues(
	_ context.Context,
	_ mqadmin.ListQueuesRequest,
) (collection.Page[mqadmin.QueueSummary], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ListQueuesCalls++
	if c.ListQueuesErr != nil {
		return collection.Page[mqadmin.QueueSummary]{}, c.ListQueuesErr
	}
	return c.ListQueuesPage, nil
}

// GetQueue records the call and returns configured stub data.
func (c *Client) GetQueue(_ context.Context, name string) (mqadmin.QueueDetail, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GetQueueCalls++
	if c.GetQueueErr != nil {
		return mqadmin.QueueDetail{}, c.GetQueueErr
	}
	q := c.Queue
	if q.Name == "" {
		q.Name = name
	}
	return q, nil
}

// Calls returns invocation counts for policy-deny assertions.
func (c *Client) Calls() (qmStatus, listQueues, getQueue, ping int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.QMStatusCalls, c.ListQueuesCalls, c.GetQueueCalls, c.PingCalls
}
