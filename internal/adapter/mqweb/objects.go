package mqweb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

const (
	// objectAPIVersion is the mqweb REST version for channel, listener, and subscription resources.
	objectAPIVersion = "v2"
)

func (c *adminClient) ListChannels(
	ctx context.Context,
	req mqadmin.ListChannelsRequest,
) (collection.Page[mqadmin.ChannelSummary], error) {
	limit := collection.NormalizeLimit(req.Limit)
	start, err := parseCursor(req.Cursor)
	if err != nil {
		return collection.Page[mqadmin.ChannelSummary]{}, err
	}
	query := url.Values{}
	if prefix := strings.TrimSpace(req.Filter.NamePrefix); prefix != "" {
		query.Set("name", prefix+"*")
	}
	if chType := strings.TrimSpace(req.Filter.ChannelType); chType != "" {
		query.Set("type", chType)
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/channel",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
	)
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		return collection.Page[mqadmin.ChannelSummary]{}, err
	}
	if code != http.StatusOK {
		return collection.Page[mqadmin.ChannelSummary]{}, mapFamilyListError("channel", code, body)
	}
	channels, err := parseChannelList(body)
	if err != nil {
		return collection.Page[mqadmin.ChannelSummary]{}, err
	}
	return paginateItems(channels, limit, start)
}

func (c *adminClient) GetChannel(ctx context.Context, name string) (mqadmin.ChannelDetail, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return mqadmin.ChannelDetail{}, errors.New("channel name is required")
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/channel/%s",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(name),
	)
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		return mqadmin.ChannelDetail{}, err
	}
	if code != http.StatusOK {
		return mqadmin.ChannelDetail{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	return parseChannelDetail(body, name)
}

func (c *adminClient) GetChannelStatus(ctx context.Context, name string) (mqadmin.ChannelStatus, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return mqadmin.ChannelStatus{}, errors.New("channel name is required")
	}
	now := time.Now().UTC()
	status := mqadmin.ChannelStatus{
		Name:         name,
		LastChecked:  now,
		Availability: mqadmin.Unavailable,
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/channel/%s?status=*",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(name),
	)
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		status.Error = err.Error()
		return status, err
	}
	if code != http.StatusOK {
		mapErr := mqadmin.ReasonCodeFromHTTPStatus(code)
		status.Error = mapErr.Error()
		return status, mapErr
	}
	parsed, err := parseChannelStatus(body, name)
	if err != nil {
		status.Error = err.Error()
		return status, err
	}
	parsed.LastChecked = now
	return parsed, nil
}

func (c *adminClient) ListListeners(
	ctx context.Context,
	req mqadmin.ListListenersRequest,
) (collection.Page[mqadmin.ListenerSummary], error) {
	limit := collection.NormalizeLimit(req.Limit)
	start, err := parseCursor(req.Cursor)
	if err != nil {
		return collection.Page[mqadmin.ListenerSummary]{}, err
	}
	query := url.Values{}
	if prefix := strings.TrimSpace(req.Filter.NamePrefix); prefix != "" {
		query.Set("name", prefix+"*")
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/listener",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
	)
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		return collection.Page[mqadmin.ListenerSummary]{}, err
	}
	if code != http.StatusOK {
		return collection.Page[mqadmin.ListenerSummary]{}, mapFamilyListError("listener", code, body)
	}
	listeners, err := parseListenerList(body)
	if err != nil {
		return collection.Page[mqadmin.ListenerSummary]{}, err
	}
	return paginateItems(listeners, limit, start)
}

func (c *adminClient) GetListener(ctx context.Context, name string) (mqadmin.ListenerDetail, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return mqadmin.ListenerDetail{}, errors.New("listener name is required")
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/listener/%s",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(name),
	)
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		return mqadmin.ListenerDetail{}, err
	}
	if code != http.StatusOK {
		return mqadmin.ListenerDetail{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	return parseListenerDetail(body, name)
}

func (c *adminClient) GetListenerStatus(ctx context.Context, name string) (mqadmin.ListenerStatus, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return mqadmin.ListenerStatus{}, errors.New("listener name is required")
	}
	now := time.Now().UTC()
	status := mqadmin.ListenerStatus{
		Name:         name,
		LastChecked:  now,
		Availability: mqadmin.Unavailable,
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/listener/%s?status=*",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(name),
	)
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		status.Error = err.Error()
		return status, err
	}
	if code != http.StatusOK {
		mapErr := mqadmin.ReasonCodeFromHTTPStatus(code)
		status.Error = mapErr.Error()
		return status, mapErr
	}
	parsed, err := parseListenerStatus(body, name)
	if err != nil {
		status.Error = err.Error()
		return status, err
	}
	parsed.LastChecked = now
	return parsed, nil
}

func (c *adminClient) ListSubscriptions(
	ctx context.Context,
	req mqadmin.ListSubscriptionsRequest,
) (collection.Page[mqadmin.SubscriptionSummary], error) {
	limit := collection.NormalizeLimit(req.Limit)
	start, err := parseCursor(req.Cursor)
	if err != nil {
		return collection.Page[mqadmin.SubscriptionSummary]{}, err
	}
	query := url.Values{}
	if prefix := strings.TrimSpace(req.Filter.NamePrefix); prefix != "" {
		query.Set("name", prefix+"*")
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/subscription",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
	)
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		return collection.Page[mqadmin.SubscriptionSummary]{}, err
	}
	if code != http.StatusOK {
		return collection.Page[mqadmin.SubscriptionSummary]{}, mapFamilyListError("subscription", code, body)
	}
	subs, err := parseSubscriptionList(body)
	if err != nil {
		return collection.Page[mqadmin.SubscriptionSummary]{}, err
	}
	return paginateItems(subs, limit, start)
}

func (c *adminClient) GetSubscription(ctx context.Context, id string) (mqadmin.SubscriptionDetail, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return mqadmin.SubscriptionDetail{}, errors.New("subscription id is required")
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/subscription/%s",
		objectAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(id),
	)
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		return mqadmin.SubscriptionDetail{}, err
	}
	if code != http.StatusOK {
		return mqadmin.SubscriptionDetail{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	return parseSubscriptionDetail(body, id)
}

type channelListResponse struct {
	Channel []channelJSON `json:"channel"`
}

type channelJSON struct {
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	General *channelGeneral `json:"general,omitempty"`
	Status  *channelStatusJ `json:"status,omitempty"`
}

type channelGeneral struct {
	Description       string `json:"description"`
	ConnectionName    string `json:"connectionName"`
	TransmissionQueue string `json:"transmissionQueue"`
}

type channelStatusJ struct {
	State string `json:"state"`
}

type listenerListResponse struct {
	Listener []listenerJSON `json:"listener"`
}

type listenerJSON struct {
	Name    string           `json:"name"`
	General *listenerGeneral `json:"general,omitempty"`
	Status  *listenerStatusJ `json:"status,omitempty"`
}

type listenerGeneral struct {
	Description string `json:"description"`
	Port        int    `json:"port"`
	Transport   string `json:"transportType"`
}

type listenerStatusJ struct {
	State string `json:"state"`
}

type subscriptionListResponse struct {
	Subscription []subscriptionJSON `json:"subscription"`
}

type subscriptionJSON struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	TopicString string               `json:"topicString"`
	Type        string               `json:"type"`
	General     *subscriptionGeneral `json:"general,omitempty"`
}

type subscriptionGeneral struct {
	Description string `json:"description"`
	Destination string `json:"destination"`
}

func parseChannelList(body []byte) ([]mqadmin.ChannelSummary, error) {
	var payload channelListResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode channel list: %w", err)
	}
	out := make([]mqadmin.ChannelSummary, 0, len(payload.Channel))
	for _, ch := range payload.Channel {
		name := strings.TrimSpace(ch.Name)
		if name == "" {
			continue
		}
		out = append(out, mqadmin.ChannelSummary{Name: name, Type: ch.Type})
	}
	return out, nil
}

func parseChannelDetail(body []byte, name string) (mqadmin.ChannelDetail, error) {
	ch, err := firstChannel(body)
	if err != nil {
		return mqadmin.ChannelDetail{}, err
	}
	detail := mqadmin.ChannelDetail{
		Name: strings.TrimSpace(ch.Name),
		Type: ch.Type,
	}
	if detail.Name == "" {
		detail.Name = name
	}
	if ch.General != nil {
		detail.Description = ch.General.Description
		detail.ConnectionName = ch.General.ConnectionName
		detail.TransmissionQueue = ch.General.TransmissionQueue
		detail.MKuratorTag = parseMKuratorTag(ch.General.Description)
	}
	return detail, nil
}

func parseChannelStatus(body []byte, requested string) (mqadmin.ChannelStatus, error) {
	ch, err := firstChannel(body)
	if err != nil {
		return mqadmin.ChannelStatus{}, err
	}
	status := mqadmin.ChannelStatus{
		Name:         strings.TrimSpace(ch.Name),
		Type:         ch.Type,
		Availability: mqadmin.Unavailable,
	}
	if status.Name == "" {
		status.Name = requested
	}
	if status.Name != "" && !strings.EqualFold(status.Name, requested) {
		status.Availability = mqadmin.Stale
		status.Error = "observed channel name differs from request"
		return status, nil
	}
	if ch.Status == nil || strings.TrimSpace(ch.Status.State) == "" {
		status.Error = "runtime status not returned by mqweb"
		return status, nil
	}
	status.State = ch.Status.State
	status.StatusText = ch.Status.State
	status.Availability = mqadmin.Available
	return status, nil
}

func parseListenerList(body []byte) ([]mqadmin.ListenerSummary, error) {
	var payload listenerListResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode listener list: %w", err)
	}
	out := make([]mqadmin.ListenerSummary, 0, len(payload.Listener))
	for _, ln := range payload.Listener {
		name := strings.TrimSpace(ln.Name)
		if name == "" {
			continue
		}
		out = append(out, mqadmin.ListenerSummary{Name: name})
	}
	return out, nil
}

func parseListenerDetail(body []byte, name string) (mqadmin.ListenerDetail, error) {
	ln, err := firstListener(body)
	if err != nil {
		return mqadmin.ListenerDetail{}, err
	}
	detail := mqadmin.ListenerDetail{Name: strings.TrimSpace(ln.Name)}
	if detail.Name == "" {
		detail.Name = name
	}
	if ln.General != nil {
		detail.Description = ln.General.Description
		detail.Port = ln.General.Port
		detail.Transport = ln.General.Transport
	}
	return detail, nil
}

func parseListenerStatus(body []byte, requested string) (mqadmin.ListenerStatus, error) {
	ln, err := firstListener(body)
	if err != nil {
		return mqadmin.ListenerStatus{}, err
	}
	status := mqadmin.ListenerStatus{
		Name:         strings.TrimSpace(ln.Name),
		Availability: mqadmin.Unavailable,
	}
	if status.Name == "" {
		status.Name = requested
	}
	if status.Name != "" && !strings.EqualFold(status.Name, requested) {
		status.Availability = mqadmin.Stale
		status.Error = "observed listener name differs from request"
		return status, nil
	}
	if ln.Status == nil || strings.TrimSpace(ln.Status.State) == "" {
		status.Error = "runtime status not returned by mqweb"
		return status, nil
	}
	status.State = ln.Status.State
	status.StatusText = ln.Status.State
	status.Availability = mqadmin.Available
	return status, nil
}

func parseSubscriptionList(body []byte) ([]mqadmin.SubscriptionSummary, error) {
	var payload subscriptionListResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode subscription list: %w", err)
	}
	out := make([]mqadmin.SubscriptionSummary, 0, len(payload.Subscription))
	for _, sub := range payload.Subscription {
		name := strings.TrimSpace(sub.Name)
		if name == "" {
			continue
		}
		out = append(out, mqadmin.SubscriptionSummary{
			ID:          strings.TrimSpace(sub.ID),
			Name:        name,
			TopicString: sub.TopicString,
			Type:        sub.Type,
		})
	}
	return out, nil
}

func parseSubscriptionDetail(body []byte, id string) (mqadmin.SubscriptionDetail, error) {
	sub, err := firstSubscription(body)
	if err != nil {
		return mqadmin.SubscriptionDetail{}, err
	}
	detail := mqadmin.SubscriptionDetail{
		ID:          strings.TrimSpace(sub.ID),
		Name:        strings.TrimSpace(sub.Name),
		TopicString: sub.TopicString,
		Type:        sub.Type,
	}
	if detail.ID == "" {
		detail.ID = id
	}
	if detail.Name == "" {
		detail.Name = id
	}
	if sub.General != nil {
		detail.Description = sub.General.Description
		detail.Destination = sub.General.Destination
	}
	return detail, nil
}

func firstChannel(body []byte) (channelJSON, error) {
	var payload channelListResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return channelJSON{}, fmt.Errorf("decode channel: %w", err)
	}
	if len(payload.Channel) == 0 {
		return channelJSON{}, mqadmin.MapReasonCode(2085)
	}
	return payload.Channel[0], nil
}

func firstListener(body []byte) (listenerJSON, error) {
	var payload listenerListResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return listenerJSON{}, fmt.Errorf("decode listener: %w", err)
	}
	if len(payload.Listener) == 0 {
		return listenerJSON{}, mqadmin.MapReasonCode(2085)
	}
	return payload.Listener[0], nil
}

func firstSubscription(body []byte) (subscriptionJSON, error) {
	var payload subscriptionListResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return subscriptionJSON{}, fmt.Errorf("decode subscription: %w", err)
	}
	if len(payload.Subscription) == 0 {
		return subscriptionJSON{}, mqadmin.MapReasonCode(2085)
	}
	return payload.Subscription[0], nil
}

// paginateItems applies the shared collection envelope client-side. mqweb v2 list
// endpoints return the full filtered set in one response — server-side cursors are
// not used until IBM MQ documents them for every object family.
func paginateItems[T any](all []T, limit, start int) (collection.Page[T], error) {
	if start < 0 || start > len(all) {
		return collection.Page[T]{}, fmt.Errorf("invalid cursor offset %d", start)
	}
	end := start + limit
	truncated := end < len(all)
	if end > len(all) {
		end = len(all)
	}
	page := collection.Page[T]{
		Items:     all[start:end],
		Limit:     limit,
		Truncated: truncated,
	}
	if start > 0 {
		page.Cursor = strconv.Itoa(start)
	}
	if truncated {
		page.NextCursor = strconv.Itoa(end)
		page.TruncationReason = collection.TruncationLimitReached
	}
	return page, nil
}

func mapFamilyListError(family string, code int, body []byte) error {
	if code == http.StatusNotImplemented || code == http.StatusMethodNotAllowed {
		return mqadmin.UnsupportedFamily(family)
	}
	if code == http.StatusNotFound && isUnknownRESTResource(body) {
		return mqadmin.UnsupportedFamily(family)
	}
	return mqadmin.ReasonCodeFromHTTPStatus(code)
}

func isUnknownRESTResource(body []byte) bool {
	if len(body) == 0 {
		return true
	}
	lower := strings.ToLower(string(body))
	return strings.Contains(lower, "unknown url") ||
		strings.Contains(lower, "not supported") ||
		strings.Contains(lower, "not available")
}
