package mqweb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

const messagingAPIVersion = "v3"

// NewMessagingClient builds an mqweb REST messaging client for profile.
func NewMessagingClient(profile catalog.Profile, resolver *secrets.Resolver) (messaging.Client, error) {
	base, err := newBaseClient(profile, resolver)
	if err != nil {
		return nil, err
	}
	return &messagingClient{base: base, queueManager: profile.QueueManager}, nil
}

type messagingClient struct {
	base         baseClient
	queueManager string
}

func (c *messagingClient) ProfileName() string { return c.base.name }

func (c *messagingClient) Ping(ctx context.Context) error {
	return c.base.ping(ctx)
}

func (c *messagingClient) Close() error { return c.base.close() }

func (c *messagingClient) BrowseMessages(
	ctx context.Context,
	queueName string,
	req messaging.BrowseRequest,
) (collection.Page[messaging.MessageRecord], error) {
	count := messaging.NormalizeBrowseCount(req.Count)
	page := collection.Page[messaging.MessageRecord]{
		Limit: count,
		Items: []messaging.MessageRecord{},
	}
	listPath := c.messageListPath(queueName, count, req.WaitIntervalMs)
	body, code, err := c.base.get(ctx, listPath)
	if err != nil {
		return page, err
	}
	switch code {
	case http.StatusNoContent:
		return page, nil
	case http.StatusOK:
	default:
		return page, mapHTTPError(code, body)
	}
	records, err := decodeMessageList(body)
	if err != nil {
		return page, err
	}
	if len(records) > count {
		records = records[:count]
		page.Truncated = true
		page.TruncationReason = collection.TruncationLimitReached
	}
	if req.IncludePayload {
		maxBytes := messaging.NormalizeMaxPayloadBytes(req.MaxPayloadBytes)
		for i := range records {
			payload, enc, truncated, payloadErr := c.fetchPayload(ctx, queueName, records[i].MessageID, maxBytes)
			if payloadErr != nil {
				return page, payloadErr
			}
			if enc != messaging.EncodingOmitted {
				records[i].Payload = payload
				records[i].Encoding = enc
				records[i].PayloadTruncated = truncated
			}
		}
	} else {
		for i := range records {
			records[i].Encoding = messaging.EncodingOmitted
		}
	}
	page.Items = records
	return page, nil
}

func (c *messagingClient) messageListPath(queueName string, count, waitMs int) string {
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/messaging/qmgr/%s/queue/%s/messagelist",
		messagingAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(queueName),
	)
	query := url.Values{}
	query.Set("numberOfMessages", strconv.Itoa(count))
	if waitMs > 0 {
		query.Set("waitInterval", strconv.Itoa(waitMs))
	}
	return path + "?" + query.Encode()
}

func (c *messagingClient) fetchPayload(
	ctx context.Context,
	queueName, messageID string,
	maxBytes int,
) (text string, encoding messaging.PayloadEncoding, truncated bool, err error) {
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/messaging/qmgr/%s/queue/%s/message",
		messagingAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(queueName),
	)
	query := url.Values{}
	if messageID != "" {
		query.Set("messageId", messageID)
	}
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	body, code, err := c.base.request(ctx, http.MethodGet, path, "text/plain")
	if err != nil {
		return "", messaging.EncodingMalformed, false, err
	}
	switch code {
	case http.StatusNoContent:
		return "", messaging.EncodingOmitted, false, nil
	case http.StatusOK:
		text, encoding, truncated = messaging.FormatPayload(body, maxBytes)
		return text, encoding, truncated, nil
	default:
		return "", messaging.EncodingMalformed, false, mapHTTPError(code, body)
	}
}

func decodeMessageList(body []byte) ([]messaging.MessageRecord, error) {
	var raw []messageListEntry
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode message list: %w", err)
	}
	out := make([]messaging.MessageRecord, 0, len(raw))
	for _, entry := range raw {
		out = append(out, entry.toRecord())
	}
	return out, nil
}

type messageListEntry struct {
	MessageID     string `json:"messageId"`
	CorrelationID string `json:"correlationId"`
	Format        string `json:"format"`
	PutDate       string `json:"putDate"`
	PutTime       string `json:"putTime"`
	MessageLength int    `json:"messageLength"`
	Priority      int    `json:"priority"`
	Persistence   string `json:"persistence"`
}

func (e messageListEntry) toRecord() messaging.MessageRecord {
	return messaging.MessageRecord{
		MessageID:     strings.TrimSpace(e.MessageID),
		CorrelationID: strings.TrimSpace(e.CorrelationID),
		Format:        strings.TrimSpace(e.Format),
		PutDate:       strings.TrimSpace(e.PutDate),
		PutTime:       strings.TrimSpace(e.PutTime),
		MessageLength: e.MessageLength,
		Priority:      e.Priority,
		Persistence:   strings.TrimSpace(e.Persistence),
	}
}

func mapHTTPError(code int, body []byte) error {
	if reason := parseReasonCode(body); reason != 0 {
		return mqadmin.MapReasonCode(reason)
	}
	return fmt.Errorf("mqweb messaging request failed with status %d", code)
}

func (b *baseClient) request(ctx context.Context, method, path, accept string) ([]byte, int, error) {
	if b.closed {
		return nil, 0, errors.New("mqweb client closed")
	}
	req, err := http.NewRequestWithContext(ctx, method, b.endpoint+path, nil)
	if err != nil {
		return nil, 0, err
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if b.authType == catalog.AuthBasic && b.username != "" {
		req.SetBasicAuth(b.username, b.password)
	}
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if reason := parseReasonCode(body); reason != 0 && resp.StatusCode >= 400 {
		return body, resp.StatusCode, mqadmin.MapReasonCode(reason)
	}
	return body, resp.StatusCode, nil
}
