// Package mqweb implements mqadmin.Client against IBM MQ mqweb REST APIs (ADR-0002).
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
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	mqtls "github.com/platformrelay/ibm-mq-mcp-server/internal/config/tls"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

const (
	adminAPIVersion = "v3"
	// queueListAPIVersion uses v1 list/get queue resources; v3 requires fixed internal MQSC (see package doc).
	queueListAPIVersion = "v1"
)

// NewAdminClient builds an mqweb REST admin client for profile.
func NewAdminClient(profile catalog.Profile, resolver *secrets.Resolver) (mqadmin.Client, error) {
	base, err := newBaseClient(profile, resolver)
	if err != nil {
		return nil, err
	}
	return &adminClient{base: base, queueManager: profile.QueueManager}, nil
}

type adminClient struct {
	base         baseClient
	queueManager string
}

func (c *adminClient) ProfileName() string { return c.base.name }

func (c *adminClient) Ping(ctx context.Context) error {
	return c.base.ping(ctx)
}

func (c *adminClient) Close() error { return c.base.close() }

func (c *adminClient) QueueManagerStatus(
	ctx context.Context,
	configuredName string,
) (mqadmin.QueueManagerStatus, error) {
	now := time.Now().UTC()
	status := mqadmin.QueueManagerStatus{
		Profile:     c.base.name,
		LastChecked: now,
		Identity: mqadmin.Identity{
			Configured: configuredName,
		},
	}
	path := fmt.Sprintf("/ibmmq/rest/%s/admin/qmgr/%s", adminAPIVersion, url.PathEscape(c.queueManager))
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		status.Availability = mqadmin.Unavailable
		status.Error = err.Error()
		return status, err
	}
	if code != http.StatusOK {
		mapErr := mqadmin.ReasonCodeFromHTTPStatus(code)
		status.Availability = mqadmin.Unavailable
		status.Error = mapErr.Error()
		return status, mapErr
	}
	var payload qmgrResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		status.Availability = mqadmin.Unavailable
		status.Error = err.Error()
		return status, fmt.Errorf("decode queue manager status: %w", err)
	}
	observed := strings.TrimSpace(payload.Name)
	if observed == "" {
		observed = strings.TrimSpace(payload.QMgrName)
	}
	status.Identity.Observed = observed
	status.Running = payload.Running || strings.EqualFold(payload.State, "running")
	status.StatusText = payload.State
	if observed != "" && !strings.EqualFold(observed, configuredName) {
		status.Availability = mqadmin.Stale
	} else {
		status.Availability = mqadmin.Available
	}
	return status, nil
}

func (c *adminClient) ListQueues(
	ctx context.Context,
	req mqadmin.ListQueuesRequest,
) (collection.Page[mqadmin.QueueSummary], error) {
	limit := collection.NormalizeLimit(req.Limit)
	start, err := parseCursor(req.Cursor)
	if err != nil {
		return collection.Page[mqadmin.QueueSummary]{}, err
	}
	query := url.Values{}
	if prefix := strings.TrimSpace(req.Filter.NamePrefix); prefix != "" {
		query.Set("name", prefix+"*")
	}
	if qtype := strings.TrimSpace(req.Filter.QueueType); qtype != "" {
		query.Set("type", qtype)
	}
	path := fmt.Sprintf("/ibmmq/rest/%s/admin/qmgr/%s/queue", queueListAPIVersion, url.PathEscape(c.queueManager))
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		return collection.Page[mqadmin.QueueSummary]{}, err
	}
	if code != http.StatusOK {
		return collection.Page[mqadmin.QueueSummary]{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	queues, err := parseQueueList(body)
	if err != nil {
		return collection.Page[mqadmin.QueueSummary]{}, err
	}
	page, err := paginateQueues(queues, limit, start)
	if err != nil {
		return collection.Page[mqadmin.QueueSummary]{}, err
	}
	return page, nil
}

func (c *adminClient) GetQueue(ctx context.Context, name string) (mqadmin.QueueDetail, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return mqadmin.QueueDetail{}, errors.New("queue name is required")
	}
	path := fmt.Sprintf(
		"/ibmmq/rest/%s/admin/qmgr/%s/queue/%s?status=*",
		queueListAPIVersion,
		url.PathEscape(c.queueManager),
		url.PathEscape(name),
	)
	body, code, err := c.base.get(ctx, path)
	if err != nil {
		return mqadmin.QueueDetail{}, err
	}
	if code != http.StatusOK {
		return mqadmin.QueueDetail{}, mqadmin.ReasonCodeFromHTTPStatus(code)
	}
	return parseQueueDetail(body, name)
}

type qmgrResponse struct {
	Name     string `json:"name"`
	QMgrName string `json:"qmgrName"`
	State    string `json:"state"`
	Running  bool   `json:"running"`
}

type queueListResponse struct {
	Queue []queueJSON `json:"queue"`
}

type queueJSON struct {
	Name    string           `json:"name"`
	Type    string           `json:"type"`
	General *queueGeneralJ   `json:"general,omitempty"`
	Storage *storageJSON     `json:"storage,omitempty"`
	Status  *queueStatusJSON `json:"status,omitempty"`
}

type queueGeneralJ struct {
	Description string `json:"description"`
}

type storageJSON struct {
	MaximumDepth int `json:"maximumDepth"`
}

type queueStatusJSON struct {
	CurrentDepth    int `json:"currentDepth"`
	OpenInputCount  int `json:"openInputCount"`
	OpenOutputCount int `json:"openOutputCount"`
}

func parseQueueList(body []byte) ([]mqadmin.QueueSummary, error) {
	var payload queueListResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode queue list: %w", err)
	}
	out := make([]mqadmin.QueueSummary, 0, len(payload.Queue))
	for _, q := range payload.Queue {
		name := strings.TrimSpace(q.Name)
		if name == "" {
			continue
		}
		out = append(out, mqadmin.QueueSummary{Name: name, Type: q.Type})
	}
	return out, nil
}

func parseQueueDetail(body []byte, name string) (mqadmin.QueueDetail, error) {
	var queues queueListResponse
	if err := json.Unmarshal(body, &queues); err != nil {
		return mqadmin.QueueDetail{}, fmt.Errorf("decode queue detail: %w", err)
	}
	if len(queues.Queue) == 0 {
		return mqadmin.QueueDetail{}, mqadmin.MapReasonCode(2085)
	}
	q := queues.Queue[0]
	detail := mqadmin.QueueDetail{
		Name: strings.TrimSpace(q.Name),
		Type: q.Type,
	}
	if detail.Name == "" {
		detail.Name = name
	}
	if q.Storage != nil {
		detail.MaxDepth = q.Storage.MaximumDepth
	}
	if q.General != nil {
		detail.Description = q.General.Description
		detail.MKuratorTag = parseMKuratorTag(q.General.Description)
	}
	if q.Status != nil {
		detail.CurrentDepth = q.Status.CurrentDepth
		detail.OpenInputCount = q.Status.OpenInputCount
		detail.OpenOutputCount = q.Status.OpenOutputCount
	}
	return detail, nil
}

func paginateQueues(all []mqadmin.QueueSummary, limit, start int) (collection.Page[mqadmin.QueueSummary], error) {
	if start < 0 || start > len(all) {
		return collection.Page[mqadmin.QueueSummary]{}, fmt.Errorf("invalid cursor offset %d", start)
	}
	end := start + limit
	truncated := end < len(all)
	if end > len(all) {
		end = len(all)
	}
	page := collection.Page[mqadmin.QueueSummary]{
		Items:     all[start:end],
		Limit:     limit,
		Truncated: truncated,
	}
	if reqCursor := start; reqCursor > 0 {
		page.Cursor = strconv.Itoa(start)
	}
	if truncated {
		page.NextCursor = strconv.Itoa(end)
		page.TruncationReason = collection.TruncationLimitReached
	}
	return page, nil
}

func parseCursor(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(raw)
	if err != nil || offset < 0 {
		return 0, fmt.Errorf("invalid cursor %q", raw)
	}
	return offset, nil
}

type baseClient struct {
	name       string
	endpoint   string
	httpClient *http.Client
	authType   catalog.AuthType
	username   string
	password   string
	closed     bool
}

func (b *baseClient) ping(ctx context.Context) error {
	if b.closed {
		return errors.New("mqweb client closed")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.endpoint+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

func (b *baseClient) close() error {
	b.closed = true
	return nil
}

func (b *baseClient) get(ctx context.Context, path string) ([]byte, int, error) {
	if b.closed {
		return nil, 0, errors.New("mqweb client closed")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.endpoint+path, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
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

func parseReasonCode(body []byte) int {
	var payload struct {
		Reason         int `json:"reason"`
		CompletionCode int `json:"completionCode"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0
	}
	if payload.Reason != 0 {
		return payload.Reason
	}
	return 0
}

func newBaseClient(profile catalog.Profile, resolver *secrets.Resolver) (baseClient, error) {
	creds, err := resolveAuth(profile.Authentication, resolver)
	if err != nil {
		return baseClient{}, err
	}
	tlsCfg, err := mqtls.BuildConfig(profile.TLS, resolver)
	if err != nil {
		return baseClient{}, err
	}
	if profile.Authentication.Type == catalog.AuthMTLS {
		auth := profile.Authentication
		certErr := mqtls.ApplyClientCertificate(
			tlsCfg,
			auth.CertificateRef,
			auth.PrivateKeyRef,
			auth.PassphraseRef,
			resolver,
		)
		if certErr != nil {
			return baseClient{}, certErr
		}
	}
	timeout, err := profileTimeout(profile.Timeout)
	if err != nil {
		return baseClient{}, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsCfg
	return baseClient{
		name:     profile.Name,
		endpoint: strings.TrimRight(profile.Endpoint, "/"),
		authType: profile.Authentication.Type,
		username: creds.username,
		password: creds.password,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}, nil
}

func profileTimeout(raw string) (time.Duration, error) {
	if raw == "" {
		return 30 * time.Second, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("timeout: %w", err)
	}
	return d, nil
}

type resolvedCredentials struct {
	username string
	password string
}

func resolveAuth(auth catalog.Authentication, resolver *secrets.Resolver) (resolvedCredentials, error) {
	switch auth.Type {
	case catalog.AuthBasic:
		ref, err := secrets.Parse(auth.SecretRef)
		if err != nil {
			return resolvedCredentials{}, err
		}
		secret, err := resolver.Resolve(ref)
		if err != nil {
			return resolvedCredentials{}, fmt.Errorf("resolve basic credentials: %w", err)
		}
		user, pass, err := parseBasicSecret(secret)
		if err != nil {
			return resolvedCredentials{}, err
		}
		return resolvedCredentials{username: user, password: pass}, nil
	case catalog.AuthMTLS:
		for _, raw := range []string{auth.CertificateRef, auth.PrivateKeyRef, auth.PassphraseRef} {
			if raw == "" {
				continue
			}
			ref, err := secrets.Parse(raw)
			if err != nil {
				return resolvedCredentials{}, err
			}
			if _, err := resolver.Resolve(ref); err != nil {
				return resolvedCredentials{}, fmt.Errorf("resolve mtls material: %w", err)
			}
		}
	default:
		return resolvedCredentials{}, fmt.Errorf("unsupported authentication type %q", auth.Type)
	}
	return resolvedCredentials{}, nil
}

func parseBasicSecret(value string) (username, password string, err error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", errors.New("basic credentials secret is empty")
	}
	user, pass, ok := strings.Cut(value, ":")
	if !ok || strings.TrimSpace(user) == "" {
		return "", "", errors.New("basic credentials must be username:password")
	}
	if strings.TrimSpace(pass) == "" {
		return "", "", errors.New("basic credentials password must not be empty")
	}
	return user, pass, nil
}
