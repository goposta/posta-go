package posta

import (
	"fmt"
	"net/url"
	"time"
)

// AdminService covers platform administration: users, plans, shared SMTP
// servers, domains across every workspace, platform settings, announcements,
// the event log, and the update check.
//
// These endpoints accept only an administrator's session token — an API key is
// never a valid credential here, whatever scopes it carries — so build the
// client with [NewWithToken].
type AdminService struct{ c *Client }

// AdminUserListOptions filters the platform user list.
type AdminUserListOptions struct {
	ListOptions
	Search string
}

func (o *AdminUserListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "search", o.Search)
	return v
}

// CreateUserRequest creates an account directly, bypassing self-registration.
// Role is "admin" or "user".
type CreateUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// UpdateUserRequest changes an account's role or standing. Nil fields are left
// unchanged.
type UpdateUserRequest struct {
	Role          string `json:"role,omitempty"`
	Active        *bool  `json:"active,omitempty"`
	EmailVerified *bool  `json:"email_verified,omitempty"`
}

// CreateUser adds an account.
func (s *AdminService) CreateUser(req *CreateUserRequest) (*User, error) {
	return post[User](s.c, "/admin/users", req, nil)
}

// ListUsers returns a page of platform accounts.
func (s *AdminService) ListUsers(opts *AdminUserListOptions) (*PageableResponse[User], error) {
	return callPage[User](s.c, "/admin/users", opts.values())
}

// UpdateUser changes an account's role or standing.
func (s *AdminService) UpdateUser(id uint, req *UpdateUserRequest) (*User, error) {
	return put[User](s.c, fmt.Sprintf("/admin/users/%d", id), req)
}

// DeleteUser schedules an account for deletion after the usual grace period.
func (s *AdminService) DeleteUser(id uint) error {
	return s.c.delete(fmt.Sprintf("/admin/users/%d", id))
}

// ForceDeleteUser removes an account and its data immediately, skipping the
// grace period. This cannot be undone.
func (s *AdminService) ForceDeleteUser(id uint) error {
	return s.c.delete(fmt.Sprintf("/admin/users/%d/force", id))
}

// CancelUserDeletion calls off a scheduled account deletion.
func (s *AdminService) CancelUserDeletion(id uint) error {
	return s.c.callNoContent("POST", fmt.Sprintf("/admin/users/%d/cancel-deletion", id), nil, nil)
}

// DisableUser2FA turns off an account's two-factor authentication, for
// recovering a user who has lost their authenticator.
func (s *AdminService) DisableUser2FA(id uint) error {
	return s.c.delete(fmt.Sprintf("/admin/users/%d/2fa", id))
}

// RevokeUserSessions signs an account out everywhere.
func (s *AdminService) RevokeUserSessions(id uint) error {
	return s.c.callNoContent("POST", fmt.Sprintf("/admin/users/%d/revoke-sessions", id), nil, nil)
}

// UserMetrics is one account's usage across the platform.
type UserMetrics struct {
	User              *User            `json:"user,omitempty"`
	TotalEmails       int64            `json:"total_emails"`
	SentEmails        int64            `json:"sent_emails"`
	FailedEmails      int64            `json:"failed_emails"`
	SuppressedEmails  int64            `json:"suppressed_emails"`
	FailureRate       float64          `json:"failure_rate"`
	TotalBounces      int64            `json:"total_bounces"`
	TotalSuppressions int64            `json:"total_suppressions"`
	TotalContacts     int64            `json:"total_contacts"`
	TotalDomains      int64            `json:"total_domains"`
	TotalSMTPServers  int64            `json:"total_smtp_servers"`
	TotalAPIKeys      int64            `json:"total_api_keys"`
	ActiveAPIKeys     int64            `json:"active_api_keys"`
	TotalInbound      int64            `json:"total_inbound"`
	ForwardedInbound  int64            `json:"forwarded_inbound"`
	FailedInbound     int64            `json:"failed_inbound"`
	WebhookDeliveries map[string]int64 `json:"webhook_deliveries,omitempty"`
}

// UserMetrics returns one account's usage figures.
func (s *AdminService) UserMetrics(id uint) (*UserMetrics, error) {
	return get[UserMetrics](s.c, fmt.Sprintf("/admin/users/%d/metrics", id), nil)
}

// AdminWorkspace is a workspace as seen from platform administration, with the
// plan assigned to it.
type AdminWorkspace struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug,omitempty"`
	OwnerID   uint      `json:"owner_id"`
	PlanID    *uint     `json:"plan_id,omitempty"`
	PlanName  string    `json:"plan_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListUserWorkspaces returns the workspaces an account belongs to.
func (s *AdminService) ListUserWorkspaces(id uint) (*[]AdminWorkspace, error) {
	return get[[]AdminWorkspace](s.c, fmt.Sprintf("/admin/users/%d/workspaces", id), nil)
}

// UserPlan returns the plan assigned to an account.
func (s *AdminService) UserPlan(id uint) (*Plan, error) {
	return get[Plan](s.c, fmt.Sprintf("/admin/users/%d/plan", id), nil)
}

// AssignUserPlan puts an account on a plan.
func (s *AdminService) AssignUserPlan(userID, planID uint) (*Plan, error) {
	return post[Plan](s.c, fmt.Sprintf("/admin/users/%d/plan", userID), planIDBody(planID), nil)
}

// WorkspacePlan returns the plan assigned to a workspace.
func (s *AdminService) WorkspacePlan(workspaceID uint) (*Plan, error) {
	return get[Plan](s.c, fmt.Sprintf("/admin/workspaces/%d/plan", workspaceID), nil)
}

// AssignWorkspacePlan puts a workspace on a plan.
func (s *AdminService) AssignWorkspacePlan(workspaceID, planID uint) (*Plan, error) {
	return post[Plan](s.c, fmt.Sprintf("/admin/workspaces/%d/plan", workspaceID), planIDBody(planID), nil)
}

func planIDBody(planID uint) any {
	return struct {
		PlanID uint `json:"plan_id"`
	}{PlanID: planID}
}

// CreatePlanRequest defines a plan's quotas.
type CreatePlanRequest struct {
	Name                  string `json:"name"`
	Description           string `json:"description,omitempty"`
	IsDefault             bool   `json:"is_default,omitempty"`
	DailyRateLimit        int    `json:"daily_rate_limit,omitempty"`
	HourlyRateLimit       int    `json:"hourly_rate_limit,omitempty"`
	MaxBatchSize          int    `json:"max_batch_size,omitempty"`
	MaxAttachmentSizeMB   int    `json:"max_attachment_size_mb,omitempty"`
	MaxAPIKeys            int    `json:"max_api_keys,omitempty"`
	MaxDomains            int    `json:"max_domains,omitempty"`
	MaxSMTPServers        int    `json:"max_smtp_servers,omitempty"`
	MaxWorkspaces         int    `json:"max_workspaces,omitempty"`
	EmailLogRetentionDays int    `json:"email_log_retention_days,omitempty"`
}

// UpdatePlanRequest changes a plan's quotas. Nil fields are left unchanged.
type UpdatePlanRequest struct {
	Name                  *string `json:"name,omitempty"`
	Description           *string `json:"description,omitempty"`
	IsDefault             *bool   `json:"is_default,omitempty"`
	IsActive              *bool   `json:"is_active,omitempty"`
	DailyRateLimit        *int    `json:"daily_rate_limit,omitempty"`
	HourlyRateLimit       *int    `json:"hourly_rate_limit,omitempty"`
	MaxBatchSize          *int    `json:"max_batch_size,omitempty"`
	MaxAttachmentSizeMB   *int    `json:"max_attachment_size_mb,omitempty"`
	MaxAPIKeys            *int    `json:"max_api_keys,omitempty"`
	MaxDomains            *int    `json:"max_domains,omitempty"`
	MaxSMTPServers        *int    `json:"max_smtp_servers,omitempty"`
	MaxWorkspaces         *int    `json:"max_workspaces,omitempty"`
	EmailLogRetentionDays *int    `json:"email_log_retention_days,omitempty"`
}

// CreatePlan adds a plan.
func (s *AdminService) CreatePlan(req *CreatePlanRequest) (*Plan, error) {
	return post[Plan](s.c, "/admin/plans", req, nil)
}

// ListPlans returns a page of plans.
func (s *AdminService) ListPlans(opts *SearchListOptions) (*PageableResponse[Plan], error) {
	return callPage[Plan](s.c, "/admin/plans", opts.values())
}

// GetPlan returns one plan.
func (s *AdminService) GetPlan(id uint) (*Plan, error) {
	return get[Plan](s.c, fmt.Sprintf("/admin/plans/%d", id), nil)
}

// UpdatePlan changes a plan.
func (s *AdminService) UpdatePlan(id uint, req *UpdatePlanRequest) (*Plan, error) {
	return put[Plan](s.c, fmt.Sprintf("/admin/plans/%d", id), req)
}

// DeletePlan removes a plan. Accounts on it fall back to the default plan.
func (s *AdminService) DeletePlan(id uint) error {
	return s.c.delete(fmt.Sprintf("/admin/plans/%d", id))
}

// SetDefaultPlan makes a plan the one new accounts receive.
func (s *AdminService) SetDefaultPlan(id uint) (*Plan, error) {
	return call[Plan](s.c, "PATCH", fmt.Sprintf("/admin/plans/%d/default", id), nil, nil)
}

// CreateServerRequest registers a shared SMTP server, offered to workspaces
// that have configured none of their own.
type CreateServerRequest struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	Encryption string `json:"encryption,omitempty"`
	// SecurityMode governs how strictly TLS is enforced on the connection.
	SecurityMode string `json:"security_mode,omitempty"`
	// AllowedDomains restricts which sender domains may use this server.
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	MaxRetries     int      `json:"max_retries,omitempty"`
}

// UpdateServerRequest changes a shared SMTP server. Omit Password to keep the
// stored one.
type UpdateServerRequest struct {
	Name           string   `json:"name,omitempty"`
	Host           string   `json:"host,omitempty"`
	Port           int      `json:"port,omitempty"`
	Username       string   `json:"username,omitempty"`
	Password       string   `json:"password,omitempty"`
	Encryption     string   `json:"encryption,omitempty"`
	SecurityMode   string   `json:"security_mode,omitempty"`
	Status         string   `json:"status,omitempty"`
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	MaxRetries     *int     `json:"max_retries,omitempty"`
}

// CreateServer registers a shared SMTP server.
func (s *AdminService) CreateServer(req *CreateServerRequest) (*Server, error) {
	return post[Server](s.c, "/admin/servers", req, nil)
}

// ListServers returns a page of shared SMTP servers.
func (s *AdminService) ListServers(opts *SearchListOptions) (*PageableResponse[Server], error) {
	return callPage[Server](s.c, "/admin/servers", opts.values())
}

// GetServer returns one shared SMTP server.
func (s *AdminService) GetServer(id uint) (*Server, error) {
	return get[Server](s.c, fmt.Sprintf("/admin/servers/%d", id), nil)
}

// UpdateServer changes a shared SMTP server.
func (s *AdminService) UpdateServer(id uint, req *UpdateServerRequest) (*Server, error) {
	return put[Server](s.c, fmt.Sprintf("/admin/servers/%d", id), req)
}

// DeleteServer removes a shared SMTP server.
func (s *AdminService) DeleteServer(id uint) error {
	return s.c.delete(fmt.Sprintf("/admin/servers/%d", id))
}

// EnableServer puts a shared SMTP server back into rotation.
func (s *AdminService) EnableServer(id uint) (*Server, error) {
	return post[Server](s.c, fmt.Sprintf("/admin/servers/%d/enable", id), nil, nil)
}

// DisableServer takes a shared SMTP server out of rotation without deleting
// it.
func (s *AdminService) DisableServer(id uint) (*Server, error) {
	return post[Server](s.c, fmt.Sprintf("/admin/servers/%d/disable", id), nil, nil)
}

// TestServer opens a connection to a shared SMTP server and authenticates,
// without sending anything.
func (s *AdminService) TestServer(id uint) (*MessageData, error) {
	return post[MessageData](s.c, fmt.Sprintf("/admin/servers/%d/test", id), nil, nil)
}

// AdminDomain is a domain as seen across every workspace, with its owner.
type AdminDomain struct {
	ID                uint      `json:"id"`
	Domain            string    `json:"domain"`
	OwnerID           uint      `json:"owner_id"`
	OwnerEmail        string    `json:"owner_email,omitempty"`
	WorkspaceID       *uint     `json:"workspace_id,omitempty"`
	WorkspaceName     string    `json:"workspace_name,omitempty"`
	OwnershipVerified bool      `json:"ownership_verified"`
	SPFVerified       bool      `json:"spf_verified"`
	DKIMVerified      bool      `json:"dkim_verified"`
	DMARCVerified     bool      `json:"dmarc_verified"`
	FullyVerified     bool      `json:"fully_verified"`
	CreatedAt         time.Time `json:"created_at"`
}

// AdminDomainListOptions filters the platform-wide domain list.
type AdminDomainListOptions struct {
	ListOptions
	Search string
	// Status filters on verification state, such as "verified" or "pending".
	Status string
	// Workspace restricts the result to one workspace.
	Workspace *uint
}

func (o *AdminDomainListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "search", o.Search)
	setStr(v, "status", o.Status)
	setUint(v, "workspace", o.Workspace)
	return v
}

// ListDomains returns a page of domains across every workspace.
func (s *AdminService) ListDomains(opts *AdminDomainListOptions) (*PageableResponse[AdminDomain], error) {
	return callPage[AdminDomain](s.c, "/admin/domains", opts.values())
}

// GetDomain returns one domain with the DNS records it needs.
func (s *AdminService) GetDomain(id uint) (*DomainWithRecords, error) {
	return get[DomainWithRecords](s.c, fmt.Sprintf("/admin/domains/%d", id), nil)
}

// VerifyDomain re-runs DNS verification for a domain in any workspace.
func (s *AdminService) VerifyDomain(id uint) error {
	return s.c.callNoContent("POST", fmt.Sprintf("/admin/domains/%d/verify", id), nil, nil)
}

// SetDomainVerification marks a domain's ownership verified, or withdraws
// that, without a DNS lookup — an override for a domain that cannot publish
// the record. Reason is recorded in the audit log.
func (s *AdminService) SetDomainVerification(id uint, verified bool, reason string) (*Domain, error) {
	body := struct {
		OwnershipVerified bool   `json:"ownership_verified"`
		Reason            string `json:"reason,omitempty"`
	}{OwnershipVerified: verified, Reason: reason}
	return put[Domain](s.c, fmt.Sprintf("/admin/domains/%d/verification", id), body)
}

// PlatformMetrics is the deployment's overall usage and runtime health.
type PlatformMetrics struct {
	TotalUsers            int64            `json:"total_users"`
	TotalWorkspaces       int64            `json:"total_workspaces"`
	UsersWithoutWorkspace int64            `json:"users_without_workspace"`
	TwoFactorUsers        int64            `json:"two_factor_users"`
	TwoFactorAdoptionRate float64          `json:"two_factor_adoption_rate"`
	ActiveSessions        int64            `json:"active_sessions"`
	FailedLoginsLast24h   int64            `json:"failed_logins_last_24h"`
	TotalEmails           int64            `json:"total_emails"`
	SentEmails            int64            `json:"sent_emails"`
	FailedEmails          int64            `json:"failed_emails"`
	QueuedEmails          int64            `json:"queued_emails"`
	ProcessingEmails      int64            `json:"processing_emails"`
	SuppressedEmails      int64            `json:"suppressed_emails"`
	FailureRate           float64          `json:"failure_rate"`
	TotalBounces          int64            `json:"total_bounces"`
	TotalSuppressions     int64            `json:"total_suppressions"`
	TotalDomains          int64            `json:"total_domains"`
	TotalAPIKeys          int64            `json:"total_api_keys"`
	ActiveAPIKeys         int64            `json:"active_api_keys"`
	SharedSMTPServers     int64            `json:"shared_smtp_servers"`
	TotalInbound          int64            `json:"total_inbound"`
	ReceivedInbound       int64            `json:"received_inbound"`
	ForwardedInbound      int64            `json:"forwarded_inbound"`
	FailedInbound         int64            `json:"failed_inbound"`
	RejectedInbound       int64            `json:"rejected_inbound"`
	ActiveWorkers         int              `json:"active_workers"`
	CurrentGoroutines     int              `json:"current_goroutines"`
	CurrentMemoryUsage    int64            `json:"current_memory_usage"`
	ServerUptimeSeconds   float64          `json:"server_uptime_seconds"`
	WebhookDeliveries     map[string]int64 `json:"webhook_deliveries,omitempty"`
}

// Metrics returns the deployment's overall usage and runtime health.
func (s *AdminService) Metrics() (*PlatformMetrics, error) {
	return get[PlatformMetrics](s.c, "/admin/metrics", nil)
}

// Analytics returns platform-wide email volume and status analytics.
func (s *AdminService) Analytics(opts *AnalyticsOptions) (*AnalyticsResponse, error) {
	return get[AnalyticsResponse](s.c, "/admin/analytics", opts.values())
}

// DashboardAnalytics returns platform-wide delivery and bounce trends.
func (s *AdminService) DashboardAnalytics(opts *AnalyticsOptions) (*DashboardAnalyticsResponse, error) {
	return get[DashboardAnalyticsResponse](s.c, "/admin/analytics/dashboard", opts.values())
}

// ProviderAnalytics returns platform-wide deliverability by recipient
// provider.
func (s *AdminService) ProviderAnalytics(opts *AnalyticsOptions) (*ProviderBreakdownResponse, error) {
	return get[ProviderBreakdownResponse](s.c, "/admin/analytics/providers", opts.values())
}

// EventListOptions filters the platform event log.
type EventListOptions struct {
	ListOptions
	Category string
	Search   string
}

func (o *EventListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "category", o.Category)
	setStr(v, "search", o.Search)
	return v
}

// ListEvents returns a page of platform events.
func (s *AdminService) ListEvents(opts *EventListOptions) (*PageableResponse[Event], error) {
	return callPage[Event](s.c, "/admin/events", opts.values())
}

// GetEvent returns one platform event with its full metadata.
func (s *AdminService) GetEvent(id uint) (*Event, error) {
	return get[Event](s.c, fmt.Sprintf("/admin/events/%d", id), nil)
}

// Settings returns the platform's configuration entries.
func (s *AdminService) Settings() (*[]Setting, error) {
	return get[[]Setting](s.c, "/admin/settings", nil)
}

// UpdateSettings changes platform configuration entries. Only the keys
// supplied are touched.
func (s *AdminService) UpdateSettings(settings []Setting) (*[]Setting, error) {
	body := struct {
		Settings []Setting `json:"settings"`
	}{Settings: settings}
	return put[[]Setting](s.c, "/admin/settings", body)
}

// CreateAnnouncementRequest broadcasts a notice to every user. Severity is
// "info", "warning", or "critical".
type CreateAnnouncementRequest struct {
	Title    string `json:"title"`
	Message  string `json:"message,omitempty"`
	Severity string `json:"severity,omitempty"`
	Link     string `json:"link,omitempty"`
}

// CreateAnnouncement broadcasts a notice to every user.
func (s *AdminService) CreateAnnouncement(req *CreateAnnouncementRequest) (*Announcement, error) {
	return post[Announcement](s.c, "/admin/announcements", req, nil)
}

// ListAnnouncements returns a page of announcements.
func (s *AdminService) ListAnnouncements(opts *ListOptions) (*PageableResponse[Announcement], error) {
	return callPage[Announcement](s.c, "/admin/announcements", opts.values())
}

// DeleteAnnouncement retracts an announcement, removing it from every user's
// notifications.
func (s *AdminService) DeleteAnnouncement(id uint) error {
	return s.c.delete(fmt.Sprintf("/admin/announcements/%d", id))
}

// UpdateInfo reports whether a newer Posta release is available.
type UpdateInfo struct {
	Enabled         bool       `json:"enabled"`
	CurrentVersion  string     `json:"current_version"`
	LatestVersion   string     `json:"latest_version,omitempty"`
	UpdateAvailable bool       `json:"update_available"`
	ReleaseURL      string     `json:"release_url,omitempty"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
	CheckedAt       *time.Time `json:"checked_at,omitempty"`
	LastError       string     `json:"last_error,omitempty"`
}

// UpdateStatus reports whether a newer Posta release is available.
func (s *AdminService) UpdateStatus() (*UpdateInfo, error) {
	return get[UpdateInfo](s.c, "/admin/update", nil)
}

// DismissUpdate hides the update notice for one version, until a later one
// appears.
func (s *AdminService) DismissUpdate(version string) (*UpdateInfo, error) {
	body := struct {
		Version string `json:"version"`
	}{Version: version}
	return post[UpdateInfo](s.c, "/admin/update/dismiss", body, nil)
}

// OAuthProvider is an SSO provider configured for the platform.
type OAuthProvider struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	// Type is the protocol family, such as "oidc", "google", or "github".
	Type   string `json:"type"`
	Issuer string `json:"issuer,omitempty"`
	Scopes string `json:"scopes,omitempty"`
	// AllowedDomains restricts sign-in to these email domains, comma
	// separated.
	AllowedDomains string `json:"allowed_domains,omitempty"`
	// AutoRegister creates an account on first sign-in through this provider.
	AutoRegister bool `json:"auto_register"`
	Enabled      bool `json:"enabled"`
	// Hidden keeps the provider off the sign-in page, reachable only by a
	// direct authorize URL.
	Hidden    bool   `json:"hidden"`
	CreatedAt string `json:"created_at,omitempty"`
}

// CreateOAuthProviderRequest configures an SSO provider. For a standards
// compliant OIDC provider, Issuer alone is enough — Posta discovers the
// endpoints. Set AuthURL, TokenURL, and UserinfoURL only for one that does not
// publish a discovery document.
type CreateOAuthProviderRequest struct {
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Type           string `json:"type"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	Issuer         string `json:"issuer,omitempty"`
	AuthURL        string `json:"auth_url,omitempty"`
	TokenURL       string `json:"token_url,omitempty"`
	UserinfoURL    string `json:"userinfo_url,omitempty"`
	Scopes         string `json:"scopes,omitempty"`
	AllowedDomains string `json:"allowed_domains,omitempty"`
	AutoRegister   *bool  `json:"auto_register,omitempty"`
	Hidden         *bool  `json:"hidden,omitempty"`
}

// UpdateOAuthProviderRequest changes an SSO provider. Omit ClientSecret to
// keep the stored one.
type UpdateOAuthProviderRequest struct {
	Name           string `json:"name,omitempty"`
	ClientID       string `json:"client_id,omitempty"`
	ClientSecret   string `json:"client_secret,omitempty"`
	Issuer         string `json:"issuer,omitempty"`
	AuthURL        string `json:"auth_url,omitempty"`
	TokenURL       string `json:"token_url,omitempty"`
	UserinfoURL    string `json:"userinfo_url,omitempty"`
	Scopes         string `json:"scopes,omitempty"`
	AllowedDomains string `json:"allowed_domains,omitempty"`
	AutoRegister   *bool  `json:"auto_register,omitempty"`
	Enabled        *bool  `json:"enabled,omitempty"`
	Hidden         *bool  `json:"hidden,omitempty"`
}

// ListOAuthProviders returns every configured SSO provider, including hidden
// and disabled ones.
func (s *AdminService) ListOAuthProviders() (*[]OAuthProvider, error) {
	return get[[]OAuthProvider](s.c, "/admin/oauth/providers", nil)
}

// CreateOAuthProvider configures an SSO provider.
func (s *AdminService) CreateOAuthProvider(req *CreateOAuthProviderRequest) (*OAuthProvider, error) {
	return post[OAuthProvider](s.c, "/admin/oauth/providers", req, nil)
}

// UpdateOAuthProvider changes an SSO provider.
func (s *AdminService) UpdateOAuthProvider(id uint, req *UpdateOAuthProviderRequest) (*OAuthProvider, error) {
	return put[OAuthProvider](s.c, fmt.Sprintf("/admin/oauth/providers/%d", id), req)
}

// DeleteOAuthProvider removes an SSO provider. Accounts linked to it fall back
// to password sign-in.
func (s *AdminService) DeleteOAuthProvider(id uint) error {
	return s.c.delete(fmt.Sprintf("/admin/oauth/providers/%d", id))
}
