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

	mu sync.Mutex

	QMStatusCalls   int
	ListQueuesCalls int
	GetQueueCalls   int
	PingCalls       int

	ListChannelsCalls     int
	GetChannelCalls       int
	GetChannelStatusCalls int

	ListListenersCalls     int
	GetListenerCalls       int
	GetListenerStatusCalls int

	ListSubscriptionsCalls int
	GetSubscriptionCalls   int

	DefineQueueCalls int
	AlterQueueCalls  int
	DeleteQueueCalls int

	QMStatus       mqadmin.QueueManagerStatus
	ListQueuesPage collection.Page[mqadmin.QueueSummary]
	Queue          mqadmin.QueueDetail

	ListChannelsPage      collection.Page[mqadmin.ChannelSummary]
	Channel               mqadmin.ChannelDetail
	ChannelStatus         mqadmin.ChannelStatus
	ListListenersPage     collection.Page[mqadmin.ListenerSummary]
	Listener              mqadmin.ListenerDetail
	ListenerStatus        mqadmin.ListenerStatus
	ListSubscriptionsPage collection.Page[mqadmin.SubscriptionSummary]
	Subscription          mqadmin.SubscriptionDetail

	QMStatusErr          error
	ListQueuesErr        error
	GetQueueErr          error
	PingErr              error
	ListChannelsErr      error
	GetChannelErr        error
	GetChannelStatusErr  error
	ListListenersErr     error
	GetListenerErr       error
	GetListenerStatusErr error
	ListSubscriptionsErr error
	GetSubscriptionErr   error
	Closed               bool

	DefineQueueResult mqadmin.QueueMutationResult
	AlterQueueResult  mqadmin.QueueMutationResult
	DeleteQueueResult mqadmin.QueueMutationResult
	DefineQueueErr    error
	AlterQueueErr     error
	DeleteQueueErr    error
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

// ListChannels records the call and returns configured stub data.
func (c *Client) ListChannels(
	_ context.Context,
	_ mqadmin.ListChannelsRequest,
) (collection.Page[mqadmin.ChannelSummary], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ListChannelsCalls++
	if c.ListChannelsErr != nil {
		return collection.Page[mqadmin.ChannelSummary]{}, c.ListChannelsErr
	}
	return c.ListChannelsPage, nil
}

// GetChannel records the call and returns configured stub data.
func (c *Client) GetChannel(_ context.Context, name string) (mqadmin.ChannelDetail, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GetChannelCalls++
	if c.GetChannelErr != nil {
		return mqadmin.ChannelDetail{}, c.GetChannelErr
	}
	ch := c.Channel
	if ch.Name == "" {
		ch.Name = name
	}
	return ch, nil
}

// GetChannelStatus records the call and returns configured stub data.
func (c *Client) GetChannelStatus(_ context.Context, name string) (mqadmin.ChannelStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GetChannelStatusCalls++
	if c.GetChannelStatusErr != nil {
		return mqadmin.ChannelStatus{}, c.GetChannelStatusErr
	}
	st := c.ChannelStatus
	if st.Name == "" {
		st.Name = name
	}
	return st, nil
}

// ListListeners records the call and returns configured stub data.
func (c *Client) ListListeners(
	_ context.Context,
	_ mqadmin.ListListenersRequest,
) (collection.Page[mqadmin.ListenerSummary], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ListListenersCalls++
	if c.ListListenersErr != nil {
		return collection.Page[mqadmin.ListenerSummary]{}, c.ListListenersErr
	}
	return c.ListListenersPage, nil
}

// GetListener records the call and returns configured stub data.
func (c *Client) GetListener(_ context.Context, name string) (mqadmin.ListenerDetail, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GetListenerCalls++
	if c.GetListenerErr != nil {
		return mqadmin.ListenerDetail{}, c.GetListenerErr
	}
	ln := c.Listener
	if ln.Name == "" {
		ln.Name = name
	}
	return ln, nil
}

// GetListenerStatus records the call and returns configured stub data.
func (c *Client) GetListenerStatus(_ context.Context, name string) (mqadmin.ListenerStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GetListenerStatusCalls++
	if c.GetListenerStatusErr != nil {
		return mqadmin.ListenerStatus{}, c.GetListenerStatusErr
	}
	st := c.ListenerStatus
	if st.Name == "" {
		st.Name = name
	}
	return st, nil
}

// ListSubscriptions records the call and returns configured stub data.
func (c *Client) ListSubscriptions(
	_ context.Context,
	_ mqadmin.ListSubscriptionsRequest,
) (collection.Page[mqadmin.SubscriptionSummary], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ListSubscriptionsCalls++
	if c.ListSubscriptionsErr != nil {
		return collection.Page[mqadmin.SubscriptionSummary]{}, c.ListSubscriptionsErr
	}
	return c.ListSubscriptionsPage, nil
}

// GetSubscription records the call and returns configured stub data.
func (c *Client) GetSubscription(_ context.Context, id string) (mqadmin.SubscriptionDetail, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.GetSubscriptionCalls++
	if c.GetSubscriptionErr != nil {
		return mqadmin.SubscriptionDetail{}, c.GetSubscriptionErr
	}
	sub := c.Subscription
	if sub.ID == "" {
		sub.ID = id
	}
	if sub.Name == "" {
		sub.Name = id
	}
	return sub, nil
}

// DefineQueue records the call and returns configured stub data.
func (c *Client) DefineQueue(
	_ context.Context,
	name string,
	req mqadmin.DefineQueueRequest,
) (mqadmin.QueueMutationResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.DefineQueueCalls++
	if c.DefineQueueErr != nil {
		return mqadmin.QueueMutationResult{}, c.DefineQueueErr
	}
	result := c.DefineQueueResult
	if result.QueueName == "" {
		result.QueueName = name
	}
	if result.Operation == "" {
		result.Operation = mqadmin.MutationDefine
	}
	if result.After == nil {
		result.After = &mqadmin.QueueSnapshot{
			Name: name,
			Type: string(req.QueueType),
		}
		if req.MaxDepth != nil {
			result.After.MaxDepth = *req.MaxDepth
		}
	}
	return result, nil
}

// AlterQueue records the call and returns configured stub data.
func (c *Client) AlterQueue(
	_ context.Context,
	name string,
	req mqadmin.AlterQueueRequest,
) (mqadmin.QueueMutationResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AlterQueueCalls++
	if c.AlterQueueErr != nil {
		return mqadmin.QueueMutationResult{}, c.AlterQueueErr
	}
	result := c.AlterQueueResult
	if result.QueueName == "" {
		result.QueueName = name
	}
	if result.Operation == "" {
		result.Operation = mqadmin.MutationAlter
	}
	if result.After == nil {
		after := mqadmin.QueueSnapshot{Name: name}
		if req.MaxDepth != nil {
			after.MaxDepth = *req.MaxDepth
		}
		if req.Description != nil {
			after.Description = *req.Description
		}
		result.After = &after
	}
	return result, nil
}

// DeleteQueue records the call and returns configured stub data.
func (c *Client) DeleteQueue(_ context.Context, name string) (mqadmin.QueueMutationResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.DeleteQueueCalls++
	if c.DeleteQueueErr != nil {
		return mqadmin.QueueMutationResult{}, c.DeleteQueueErr
	}
	result := c.DeleteQueueResult
	if result.QueueName == "" {
		result.QueueName = name
	}
	if result.Operation == "" {
		result.Operation = mqadmin.MutationDelete
	}
	return result, nil
}

// Calls returns invocation counts for policy-deny assertions.
func (c *Client) Calls() (qmStatus, listQueues, getQueue, ping int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.QMStatusCalls, c.ListQueuesCalls, c.GetQueueCalls, c.PingCalls
}

// TotalCalls returns the sum of all recorded adapter invocations.
func (c *Client) TotalCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.QMStatusCalls + c.ListQueuesCalls + c.GetQueueCalls + c.PingCalls +
		c.ListChannelsCalls + c.GetChannelCalls + c.GetChannelStatusCalls +
		c.ListListenersCalls + c.GetListenerCalls + c.GetListenerStatusCalls +
		c.ListSubscriptionsCalls + c.GetSubscriptionCalls +
		c.DefineQueueCalls + c.AlterQueueCalls + c.DeleteQueueCalls
}
