// Package fake provides test doubles for messaging.Client.
package fake

import (
	"context"
	"sync"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
)

// Client implements messaging.Client for tests and records invocation counts.
type Client struct {
	Name string

	mu sync.Mutex

	BrowseCalls  int
	ConsumeCalls int
	PutCalls     int
	PingCalls    int
	Closed       bool

	BrowsePage collection.Page[messaging.MessageRecord]
	BrowseErr  error
	PutResult  messaging.PutResult
	PutErr     error
	PingErr    error
}

// New returns a fake messaging client for the given profile name.
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

// BrowseMessages records the call and returns configured results.
func (c *Client) BrowseMessages(
	_ context.Context,
	_ string,
	_ messaging.BrowseRequest,
) (collection.Page[messaging.MessageRecord], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.BrowseCalls++
	return c.BrowsePage, c.BrowseErr
}

// PutMessage records the call and returns configured results.
func (c *Client) PutMessage(
	_ context.Context,
	_ string,
	_ messaging.PutRequest,
) (messaging.PutResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.PutCalls++
	return c.PutResult, c.PutErr
}

// RecordConsume simulates a destructive consume call for spy assertions.
func (c *Client) RecordConsume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ConsumeCalls++
}

// Close marks the fake client closed.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Closed = true
	return nil
}

// TotalCalls returns browse + consume + put invocations.
func (c *Client) TotalCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.BrowseCalls + c.ConsumeCalls + c.PutCalls
}

// BrowseOnlyCalls returns browse invocations.
func (c *Client) BrowseOnlyCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.BrowseCalls
}

// ConsumeOnlyCalls returns consume invocations.
func (c *Client) ConsumeOnlyCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ConsumeCalls
}
