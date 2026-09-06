package posta

import (
	"fmt"
	"time"
)

// DomainsService manages sending domains and their DNS verification. A
// workspace that requires verified domains will refuse to send from an
// unverified one.
type DomainsService struct{ c *Client }

// Add registers a domain and returns the DNS records to publish for it.
func (s *DomainsService) Add(domain string) (*DomainWithRecords, error) {
	body := struct {
		Domain string `json:"domain"`
	}{Domain: domain}
	return post[DomainWithRecords](s.c, wsPath+"/domains", body, nil)
}

// List returns a page of domains.
func (s *DomainsService) List(opts *ListOptions) (*PageableResponse[Domain], error) {
	return callPage[Domain](s.c, wsPath+"/domains", opts.values())
}

// Get returns one domain with the DNS records it needs and their current
// verification state.
func (s *DomainsService) Get(id uint) (*DomainWithRecords, error) {
	return get[DomainWithRecords](s.c, fmt.Sprintf("%s/domains/%d", wsPath, id), nil)
}

// Verify re-runs the DNS lookups for a domain and reports the outcome of each
// check. DNS propagates slowly, so this is expected to be called repeatedly
// until FullyVerified is true.
func (s *DomainsService) Verify(id uint) (*DomainVerificationResult, error) {
	return post[DomainVerificationResult](s.c, fmt.Sprintf("%s/domains/%d/verify", wsPath, id), nil, nil)
}

// Delete removes a domain.
func (s *DomainsService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/domains/%d", wsPath, id))
}

// SMTPServersService manages the SMTP relays Posta delivers outbound mail
// through.
type SMTPServersService struct{ c *Client }

// CreateSMTPServerRequest registers an SMTP relay. Encryption is one of
// "none", "tls", or "starttls".
type CreateSMTPServerRequest struct {
	Name       string `json:"name,omitempty"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	Encryption string `json:"encryption,omitempty"`
	// AllowedEmails restricts which From addresses may use this server.
	AllowedEmails []string `json:"allowed_emails,omitempty"`
	MaxRetries    int      `json:"max_retries,omitempty"`
}

// UpdateSMTPServerRequest changes an SMTP relay. Omit Password to keep the
// stored one.
type UpdateSMTPServerRequest struct {
	Name          string   `json:"name,omitempty"`
	Host          string   `json:"host,omitempty"`
	Port          int      `json:"port,omitempty"`
	Username      string   `json:"username,omitempty"`
	Password      string   `json:"password,omitempty"`
	Encryption    string   `json:"encryption,omitempty"`
	Status        string   `json:"status,omitempty"`
	AllowedEmails []string `json:"allowed_emails,omitempty"`
	MaxRetries    *int     `json:"max_retries,omitempty"`
}

// Create registers an SMTP server.
func (s *SMTPServersService) Create(req *CreateSMTPServerRequest) (*SMTPServer, error) {
	return post[SMTPServer](s.c, wsPath+"/smtp-servers", req, nil)
}

// List returns a page of SMTP servers, including any shared server the
// platform offers the workspace.
func (s *SMTPServersService) List(opts *ListOptions) (*PageableResponse[SMTPServer], error) {
	return callPage[SMTPServer](s.c, wsPath+"/smtp-servers", opts.values())
}

// Get returns one SMTP server.
func (s *SMTPServersService) Get(id uint) (*SMTPServer, error) {
	return get[SMTPServer](s.c, fmt.Sprintf("%s/smtp-servers/%d", wsPath, id), nil)
}

// Update changes an SMTP server's configuration.
func (s *SMTPServersService) Update(id uint, req *UpdateSMTPServerRequest) (*SMTPServer, error) {
	return put[SMTPServer](s.c, fmt.Sprintf("%s/smtp-servers/%d", wsPath, id), req)
}

// Delete removes an SMTP server.
func (s *SMTPServersService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/smtp-servers/%d", wsPath, id))
}

// Test opens a connection to the server and authenticates, without sending
// anything. Use it to check credentials before relying on them.
func (s *SMTPServersService) Test(id uint) (*MessageData, error) {
	return post[MessageData](s.c, fmt.Sprintf("%s/smtp-servers/%d/test", wsPath, id), nil, nil)
}

// SMTPCredentialsService manages credentials for Posta's own SMTP relay
// listener, which lets an existing application send through Posta by pointing
// its SMTP client at it instead of calling the HTTP API.
type SMTPCredentialsService struct{ c *Client }

// CreateSMTPCredentialRequest mints a relay credential. AllowedIPs, when set,
// restricts which addresses may authenticate with it.
type CreateSMTPCredentialRequest struct {
	Name       string   `json:"name"`
	AllowedIPs []string `json:"allowed_ips,omitempty"`
}

// SMTPCredentialCreated is the one-time result of creating a relay credential.
// Password is shown only here and cannot be retrieved again. Host and Port
// point at Posta's relay listener, ready to paste into an SMTP client.
type SMTPCredentialCreated struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Host      string    `json:"host,omitempty"`
	Port      int       `json:"port,omitempty"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Create mints a relay credential. The password is returned once; store it now.
func (s *SMTPCredentialsService) Create(req *CreateSMTPCredentialRequest) (*SMTPCredentialCreated, error) {
	return post[SMTPCredentialCreated](s.c, wsPath+"/smtp-credentials", req, nil)
}

// List returns a page of relay credentials.
func (s *SMTPCredentialsService) List(opts *ListOptions) (*PageableResponse[SMTPCredential], error) {
	return callPage[SMTPCredential](s.c, wsPath+"/smtp-credentials", opts.values())
}

// Get returns one relay credential.
func (s *SMTPCredentialsService) Get(id uint) (*SMTPCredential, error) {
	return get[SMTPCredential](s.c, fmt.Sprintf("%s/smtp-credentials/%d", wsPath, id), nil)
}

// Revoke disables a relay credential without deleting its record, so past use
// stays auditable.
func (s *SMTPCredentialsService) Revoke(id uint) (*MessageData, error) {
	return post[MessageData](s.c, fmt.Sprintf("%s/smtp-credentials/%d/revoke", wsPath, id), nil, nil)
}

// Delete removes a relay credential entirely.
func (s *SMTPCredentialsService) Delete(id uint) (*MessageData, error) {
	return call[MessageData](s.c, "DELETE", fmt.Sprintf("%s/smtp-credentials/%d", wsPath, id), nil, nil)
}

// APIKeysService manages the workspace's machine credentials.
type APIKeysService struct{ c *Client }

// API key scopes. A key carrying [ScopeAll] grants every scope. Note that
// [ScopeSend] grants none of the others: a send-only key is confined to the
// public send API and cannot read or modify workspace resources.
const (
	// ScopeSend permits sending, verification, and subscriber-list opt-ins.
	ScopeSend = "send"
	// ScopeRead permits reading emails, bounces, webhook deliveries, and
	// workspace resources.
	ScopeRead = "read"
	// ScopeWrite permits mutating workspace resources.
	ScopeWrite = "write"
	// ScopeWebhooks permits managing webhook endpoints.
	ScopeWebhooks = "webhooks"
	// ScopeAdmin permits tenant administration: API keys, members,
	// invitations, settings, SSO, plan, forms, and message filters.
	ScopeAdmin = "admin"
	// ScopeAll grants every scope.
	ScopeAll = "*"
)

// CreateAPIKeyRequest mints an API key. An empty Scopes defaults to
// [ScopeSend]; ExpiresInDays leaves the key permanent when nil.
type CreateAPIKeyRequest struct {
	Name          string   `json:"name"`
	Scopes        []string `json:"scopes,omitempty"`
	AllowedIPs    []string `json:"allowed_ips,omitempty"`
	ExpiresInDays *int     `json:"expires_in_days,omitempty"`
}

// APIKeyCreated is the one-time result of minting a key. Key holds the secret,
// which is shown only here; Prefix is the fragment that identifies the key
// afterwards.
type APIKeyCreated struct {
	ID      uint     `json:"id"`
	Name    string   `json:"name"`
	Key     string   `json:"key"`
	Prefix  string   `json:"prefix"`
	Scopes  []string `json:"scopes,omitempty"`
	Message string   `json:"message,omitempty"`
}

// Create mints an API key. The secret is returned once; store it now.
func (s *APIKeysService) Create(req *CreateAPIKeyRequest) (*APIKeyCreated, error) {
	return post[APIKeyCreated](s.c, wsPath+"/api-keys", req, nil)
}

// List returns a page of API keys. The secrets are not included.
func (s *APIKeysService) List(opts *ListOptions) (*PageableResponse[APIKey], error) {
	return callPage[APIKey](s.c, wsPath+"/api-keys", opts.values())
}

// Get returns one API key's metadata.
func (s *APIKeysService) Get(id uint) (*APIKey, error) {
	return get[APIKey](s.c, fmt.Sprintf("%s/api-keys/%d", wsPath, id), nil)
}

// Revoke disables a key without deleting its record, so its past use stays
// auditable.
func (s *APIKeysService) Revoke(id uint) (*MessageData, error) {
	return put[MessageData](s.c, fmt.Sprintf("%s/api-keys/%d/revoke", wsPath, id), nil)
}

// Delete removes a key entirely.
func (s *APIKeysService) Delete(id uint) (*MessageData, error) {
	return call[MessageData](s.c, "DELETE", fmt.Sprintf("%s/api-keys/%d", wsPath, id), nil, nil)
}
