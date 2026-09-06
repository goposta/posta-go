// Package posta provides a Go client for the Posta email platform.
//
// The client covers the whole Posta API: transactional and templated sending,
// batch sends, address verification, templates and their versions and
// localizations, campaigns, subscribers and lists, suppressions and bounces,
// domains, SMTP servers and relay credentials, webhooks, web forms and the
// messages they collect, inbound email, workspace administration, and the
// platform admin surface.
//
// # Credentials
//
// Most machine-facing endpoints authenticate with an API key:
//
//	c := posta.New("https://posta.example.com", "psk_...")
//
// Account-level and platform-admin endpoints (users/me, /admin/*) accept only a
// user session token, which [NewWithToken] supplies:
//
//	c := posta.NewWithToken("https://posta.example.com", jwt)
//
// # Workspaces
//
// Workspace-scoped endpoints resolve the active workspace from the
// X-Posta-Workspace-Id header. A workspace-bound API key carries its workspace
// already and needs nothing further; an account-wide key or a user session must
// name one with [WithWorkspace].
package posta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jkaninda/okapi/client"
)

// Version is the client library version, reported in the User-Agent header.
const Version = "0.2.0"

// WorkspaceHeader names the header that selects the active workspace for
// workspace-scoped endpoints.
const WorkspaceHeader = "X-Posta-Workspace-Id"

// Client is the Posta API client. Its exported fields group the API by
// resource; the client itself is safe for concurrent use.
type Client struct {
	http *client.Client
	ctx  context.Context
	// rootURL is the deployment root, without the /api/v1 prefix. The health
	// probes live there rather than under the versioned API.
	rootURL string

	// Emails sends mail and reads the resulting delivery records.
	Emails *EmailsService
	// Bounces reads recorded bounces and complaints.
	Bounces *BouncesService
	// Suppressions manages the workspace suppression list.
	Suppressions *SuppressionsService
	// Webhooks registers endpoints and reads delivery attempts.
	Webhooks *WebhooksService
	// Templates manages templates, versions, and localizations.
	Templates *TemplatesService
	// Languages manages the workspace's template languages.
	Languages *LanguagesService
	// Stylesheets manages reusable CSS for templates.
	Stylesheets *StylesheetsService
	// Domains manages sending domains and their DNS verification.
	Domains *DomainsService
	// SMTPServers manages the SMTP servers Posta delivers through.
	SMTPServers *SMTPServersService
	// SMTPCredentials manages credentials for the SMTP relay listener.
	SMTPCredentials *SMTPCredentialsService
	// Subscribers manages subscriber records and bulk imports.
	Subscribers *SubscribersService
	// SubscriberLists manages lists, their members, and opt-outs.
	SubscriberLists *SubscriberListsService
	// UnsubscribeLists manages the lists behind List-Unsubscribe headers.
	UnsubscribeLists *UnsubscribeListsService
	// Contacts reads the derived contact view of everyone mailed.
	Contacts *ContactsService
	// Campaigns manages bulk campaigns and their lifecycle.
	Campaigns *CampaignsService
	// Analytics reads delivery and engagement analytics.
	Analytics *AnalyticsService
	// Forms manages web form endpoints and their embed snippets.
	Forms *FormsService
	// Messages reads and triages web form submissions.
	Messages *MessagesService
	// MessageFilters manages the spam filters applied to submissions.
	MessageFilters *MessageFiltersService
	// Inbound reads inbound email received by Posta.
	Inbound *InboundService
	// APIKeys manages the workspace's API keys.
	APIKeys *APIKeysService
	// Workspaces manages workspaces, members, invitations, and settings.
	Workspaces *WorkspacesService
	// Users manages the signed-in account (session credential only).
	Users *UsersService
	// Auth covers login, registration, and password recovery.
	Auth *AuthService
	// Admin covers platform administration (admin session only).
	Admin *AdminService
	// System reads health and build information.
	System *SystemService
}

// Option configures a [Client] at construction time.
type Option func(*clientConfig)

type clientConfig struct {
	opts    []client.Option
	headers map[string]string
}

// WithHTTPClient supplies the underlying *http.Client, for callers that need
// custom transports, proxies, or TLS settings.
func WithHTTPClient(h *http.Client) Option {
	return func(c *clientConfig) { c.opts = append(c.opts, client.WithHTTPClient(h)) }
}

// WithTimeout sets the per-request timeout. The default is 30 seconds.
func WithTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.opts = append(c.opts, client.WithTimeout(d)) }
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *clientConfig) { c.opts = append(c.opts, client.WithUserAgent(ua)) }
}

// WithHeader sets an additional header on every request.
func WithHeader(key, value string) Option {
	return func(c *clientConfig) { c.headers[key] = value }
}

// WithWorkspace selects the active workspace for workspace-scoped endpoints.
// A workspace-bound API key already names its workspace and does not need this;
// an account-wide key or a user session does.
func WithWorkspace(id uint) Option {
	return func(c *clientConfig) { c.headers[WorkspaceHeader] = strconv.FormatUint(uint64(id), 10) }
}

// WithRetry retries failed requests according to policy. By default no retry is
// attempted, so a caller opts in explicitly:
//
//	posta.WithRetry(client.RetryPolicy{MaxAttempts: 3, BaseDelay: 200 * time.Millisecond})
func WithRetry(policy client.RetryPolicy) Option {
	return func(c *clientConfig) { c.opts = append(c.opts, client.WithRetry(policy)) }
}

// New creates a client authenticated with an API key.
//
//	c := posta.New("https://posta.example.com", "psk_...")
func New(baseURL, apiKey string, opts ...Option) *Client {
	return newClient(baseURL, apiKey, opts)
}

// NewWithToken creates a client authenticated with a user session token (JWT),
// as returned by [AuthService.Login]. Account-level endpoints under
// /users/me and the platform admin surface accept only this credential.
func NewWithToken(baseURL, token string, opts ...Option) *Client {
	return newClient(baseURL, token, opts)
}

func newClient(baseURL, credential string, opts []Option) *Client {
	cfg := &clientConfig{headers: map[string]string{}}
	for _, fn := range opts {
		fn(cfg)
	}

	// Content-Type is set per request rather than here: a global header would
	// override the multipart boundary type on file uploads.
	base := []client.Option{
		client.WithBearerToken(credential),
		client.WithUserAgent("posta-go/" + Version),
	}
	base = append(base, cfg.opts...)
	if len(cfg.headers) > 0 {
		base = append(base, client.WithHeaders(cfg.headers))
	}

	c := &Client{
		http:    client.New(baseURL+"/api/v1", base...),
		rootURL: strings.TrimSuffix(baseURL, "/"),
	}
	c.initServices()
	return c
}

// WithContext returns a copy of the client whose requests carry ctx. The
// returned client shares the underlying HTTP connection pool.
//
//	emails, err := c.WithContext(ctx).Emails.List(nil)
func (c *Client) WithContext(ctx context.Context) *Client {
	clone := &Client{http: c.http, ctx: ctx, rootURL: c.rootURL}
	clone.initServices()
	return clone
}

func (c *Client) initServices() {
	c.Emails = &EmailsService{c: c}
	c.Bounces = &BouncesService{c: c}
	c.Suppressions = &SuppressionsService{c: c}
	c.Webhooks = &WebhooksService{c: c}
	c.Templates = &TemplatesService{c: c}
	c.Languages = &LanguagesService{c: c}
	c.Stylesheets = &StylesheetsService{c: c}
	c.Domains = &DomainsService{c: c}
	c.SMTPServers = &SMTPServersService{c: c}
	c.SMTPCredentials = &SMTPCredentialsService{c: c}
	c.Subscribers = &SubscribersService{c: c}
	c.SubscriberLists = &SubscriberListsService{c: c}
	c.UnsubscribeLists = &UnsubscribeListsService{c: c}
	c.Contacts = &ContactsService{c: c}
	c.Campaigns = &CampaignsService{c: c}
	c.Analytics = &AnalyticsService{c: c}
	c.Forms = &FormsService{c: c}
	c.Messages = &MessagesService{c: c}
	c.MessageFilters = &MessageFiltersService{c: c}
	c.Inbound = &InboundService{c: c}
	c.APIKeys = &APIKeysService{c: c}
	c.Workspaces = &WorkspacesService{c: c}
	c.Users = &UsersService{c: c}
	c.Auth = &AuthService{c: c}
	c.Admin = &AdminService{c: c}
	c.System = &SystemService{c: c}
}

// APIError is returned when the Posta API responds with a non-2xx status.
type APIError struct {
	// StatusCode is the HTTP status returned by the API.
	StatusCode int
	// Info holds the decoded error envelope, when the response carried one.
	Info *ErrorInfo
}

func (e *APIError) Error() string {
	if e.Info != nil && e.Info.Message != "" {
		return fmt.Sprintf("posta: %d %s", e.StatusCode, e.Info.Message)
	}
	if e.Info != nil && e.Info.Error != "" {
		return fmt.Sprintf("posta: %d %s", e.StatusCode, e.Info.Error)
	}
	return fmt.Sprintf("posta: unexpected status %d", e.StatusCode)
}

// IsNotFound reports whether err is an [APIError] carrying HTTP 404. It is the
// usual way to tell "no such record" apart from a transport failure.
func IsNotFound(err error) bool { return statusIs(err, http.StatusNotFound) }

// IsUnauthorized reports whether err is an [APIError] carrying HTTP 401,
// meaning the credential was missing, malformed, or revoked.
func IsUnauthorized(err error) bool { return statusIs(err, http.StatusUnauthorized) }

// IsForbidden reports whether err is an [APIError] carrying HTTP 403. For an
// API key this usually means the key lacks the scope the endpoint requires.
func IsForbidden(err error) bool { return statusIs(err, http.StatusForbidden) }

// IsRateLimited reports whether err is an [APIError] carrying HTTP 429.
func IsRateLimited(err error) bool { return statusIs(err, http.StatusTooManyRequests) }

func statusIs(err error, code int) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == code
}

// request builds a request with the client's context and optional query.
func (c *Client) request(method, path string, query url.Values) *client.RequestBuilder {
	rb := c.http.Request(method, path)
	if c.ctx != nil {
		rb = rb.WithContext(c.ctx)
	}
	for k, vs := range query {
		for _, v := range vs {
			rb = rb.QueryParam(k, v)
		}
	}
	return rb
}

func (c *Client) do(rb *client.RequestBuilder) (*client.Response, error) {
	resp, err := rb.Do()
	if err != nil {
		return nil, err
	}
	if resp.IsSuccess() {
		return resp, nil
	}
	apiErr := &APIError{StatusCode: resp.StatusCode}
	var envelope ErrorResponse
	if json.Unmarshal(resp.Body, &envelope) == nil && envelope.Error != nil {
		apiErr.Info = envelope.Error
	}
	return nil, apiErr
}

// call issues a request and decodes the `data` field of the response envelope.
func call[T any](c *Client, method, path string, body any, query url.Values) (*T, error) {
	rb := c.request(method, path, query)
	if body != nil {
		rb = rb.JSONBody(body)
	}
	resp, err := c.do(rb)
	if err != nil {
		return nil, err
	}
	var envelope Response[T]
	if err := json.Unmarshal(resp.Body, &envelope); err != nil {
		return nil, fmt.Errorf("posta: decode response: %w", err)
	}
	return &envelope.Data, nil
}

// callPage issues a request and decodes a paginated envelope.
func callPage[T any](c *Client, path string, query url.Values) (*PageableResponse[T], error) {
	resp, err := c.do(c.request(http.MethodGet, path, query))
	if err != nil {
		return nil, err
	}
	var envelope PageableResponse[T]
	if err := json.Unmarshal(resp.Body, &envelope); err != nil {
		return nil, fmt.Errorf("posta: decode response: %w", err)
	}
	return &envelope, nil
}

// callNoContent issues a request and discards the body, for endpoints that
// answer 204 or whose envelope carries nothing worth returning.
func (c *Client) callNoContent(method, path string, body any, query url.Values) error {
	rb := c.request(method, path, query)
	if body != nil {
		rb = rb.JSONBody(body)
	}
	_, err := c.do(rb)
	return err
}

// callRaw issues a request and returns the raw body together with its
// Content-Type, for endpoints that answer with a file rather than JSON.
func (c *Client) callRaw(path string, query url.Values) ([]byte, string, error) {
	resp, err := c.do(c.request(http.MethodGet, path, query))
	if err != nil {
		return nil, "", err
	}
	return resp.Body, resp.Header.Get("Content-Type"), nil
}

func get[T any](c *Client, path string, query url.Values) (*T, error) {
	return call[T](c, http.MethodGet, path, nil, query)
}

func post[T any](c *Client, path string, body any, query url.Values) (*T, error) {
	return call[T](c, http.MethodPost, path, body, query)
}

func put[T any](c *Client, path string, body any) (*T, error) {
	return call[T](c, http.MethodPut, path, body, nil)
}

// getAbsolute fetches a URL outside the versioned API and decodes the body
// as-is: the health probes answer with a bare object, not the {success, data}
// envelope every /api/v1 endpoint uses.
func getAbsolute[T any](c *Client, path string) (*T, error) {
	resp, err := c.do(c.request(http.MethodGet, c.rootURL+path, nil))
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal(resp.Body, &out); err != nil {
		return nil, fmt.Errorf("posta: decode response: %w", err)
	}
	return &out, nil
}

func (c *Client) delete(path string) error {
	return c.callNoContent(http.MethodDelete, path, nil, nil)
}
