package posta

import (
	"fmt"
	"time"
)

// WorkspacesService manages workspaces and everything that governs one:
// membership, invitations, settings, SSO, the audit log, and the data export
// and GDPR erasure tools.
//
// Endpoints under "current" act on the workspace the credential resolves to —
// the one a workspace-bound API key names, or the one selected with
// [WithWorkspace].
type WorkspacesService struct{ c *Client }

// CreateWorkspaceRequest creates a workspace. SeedDefaults, when true, fills
// the new workspace with a starter set of languages and templates.
type CreateWorkspaceRequest struct {
	Name            string `json:"name"`
	Slug            string `json:"slug,omitempty"`
	Description     string `json:"description,omitempty"`
	DefaultLanguage string `json:"default_language,omitempty"`
	SeedDefaults    *bool  `json:"seed_defaults,omitempty"`
}

// UpdateWorkspaceRequest renames or re-describes a workspace.
type UpdateWorkspaceRequest struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	DefaultLanguage string `json:"default_language,omitempty"`
}

// Create adds a workspace, owned by the caller.
func (s *WorkspacesService) Create(req *CreateWorkspaceRequest) (*Workspace, error) {
	return post[Workspace](s.c, "/workspaces", req, nil)
}

// List returns every workspace the caller belongs to, with their own role in
// each.
func (s *WorkspacesService) List() (*[]Workspace, error) {
	return get[[]Workspace](s.c, "/workspaces", nil)
}

// Get returns the active workspace.
func (s *WorkspacesService) Get() (*Workspace, error) {
	return get[Workspace](s.c, wsPath, nil)
}

// Update changes the active workspace.
func (s *WorkspacesService) Update(req *UpdateWorkspaceRequest) (*Workspace, error) {
	return put[Workspace](s.c, wsPath, req)
}

// Delete removes the active workspace and everything in it. This cannot be
// undone; export first with [WorkspacesService.ExportData].
func (s *WorkspacesService) Delete() error {
	return s.c.delete(wsPath)
}

// ListMembers returns the workspace's members and their roles.
func (s *WorkspacesService) ListMembers() (*[]WorkspaceMember, error) {
	return get[[]WorkspaceMember](s.c, wsPath+"/members", nil)
}

// UpdateMemberRole changes a member's role. Role is one of "owner", "admin",
// "editor", "viewer".
func (s *WorkspacesService) UpdateMemberRole(memberID uint, role string) (*WorkspaceMember, error) {
	body := struct {
		Role string `json:"role"`
	}{Role: role}
	return put[WorkspaceMember](s.c, fmt.Sprintf("%s/members/%d", wsPath, memberID), body)
}

// RemoveMember removes a member from the workspace.
func (s *WorkspacesService) RemoveMember(memberID uint) error {
	return s.c.delete(fmt.Sprintf("%s/members/%d", wsPath, memberID))
}

// Invite offers workspace membership to an email address, at the given role.
func (s *WorkspacesService) Invite(email, role string) (*Invitation, error) {
	body := struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}{Email: email, Role: role}
	return post[Invitation](s.c, wsPath+"/invitations", body, nil)
}

// ListInvitations returns the workspace's pending invitations.
func (s *WorkspacesService) ListInvitations() (*[]Invitation, error) {
	return get[[]Invitation](s.c, wsPath+"/invitations", nil)
}

// CancelInvitation withdraws a pending invitation.
func (s *WorkspacesService) CancelInvitation(invitationID uint) error {
	return s.c.delete(fmt.Sprintf("%s/invitations/%d", wsPath, invitationID))
}

// ListMyInvitations returns the invitations addressed to the signed-in user,
// across every workspace.
func (s *WorkspacesService) ListMyInvitations() (*[]Invitation, error) {
	return get[[]Invitation](s.c, "/invitations", nil)
}

// AcceptInvitation joins a workspace using the token from an invitation email.
func (s *WorkspacesService) AcceptInvitation(token string) error {
	body := struct {
		Token string `json:"token"`
	}{Token: token}
	return s.c.callNoContent("POST", "/invitations/accept", body, nil)
}

// DeclineInvitation refuses an invitation using its token.
func (s *WorkspacesService) DeclineInvitation(token string) error {
	body := struct {
		Token string `json:"token"`
	}{Token: token}
	return s.c.callNoContent("POST", "/invitations/decline", body, nil)
}

// AcceptInvitationByID joins a workspace using an invitation listed by
// [WorkspacesService.ListMyInvitations], which needs no token.
func (s *WorkspacesService) AcceptInvitationByID(id uint) error {
	return s.c.callNoContent("POST", fmt.Sprintf("/invitations/%d/accept", id), nil, nil)
}

// DeclineInvitationByID refuses an invitation by its ID.
func (s *WorkspacesService) DeclineInvitationByID(id uint) error {
	return s.c.callNoContent("POST", fmt.Sprintf("/invitations/%d/decline", id), nil, nil)
}

// WorkspaceSettings holds the workspace's sending defaults and policy.
type WorkspaceSettings struct {
	WorkspaceID uint `json:"workspace_id"`
	// DefaultSenderEmail and DefaultSenderName fill in a send's From when it
	// omits one.
	DefaultSenderEmail string `json:"default_sender_email,omitempty"`
	DefaultSenderName  string `json:"default_sender_name,omitempty"`
	// RequireVerifiedDomain refuses sends from a domain that has not passed
	// DNS verification.
	RequireVerifiedDomain bool `json:"require_verified_domain"`
	// BounceAutoSuppress adds a hard-bounced address to the suppression list
	// automatically.
	BounceAutoSuppress bool      `json:"bounce_auto_suppress"`
	WebhookRetryCount  int       `json:"webhook_retry_count"`
	APIKeyExpiryDays   int       `json:"api_key_expiry_days"`
	Timezone           string    `json:"timezone,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// UpdateWorkspaceSettingsRequest changes workspace settings. Nil fields are
// left unchanged.
type UpdateWorkspaceSettingsRequest struct {
	DefaultSenderEmail    *string `json:"default_sender_email,omitempty"`
	DefaultSenderName     *string `json:"default_sender_name,omitempty"`
	RequireVerifiedDomain *bool   `json:"require_verified_domain,omitempty"`
	BounceAutoSuppress    *bool   `json:"bounce_auto_suppress,omitempty"`
	WebhookRetryCount     *int    `json:"webhook_retry_count,omitempty"`
	APIKeyExpiryDays      *int    `json:"api_key_expiry_days,omitempty"`
	Timezone              *string `json:"timezone,omitempty"`
}

// Settings returns the workspace's settings.
func (s *WorkspacesService) Settings() (*WorkspaceSettings, error) {
	return get[WorkspaceSettings](s.c, wsPath+"/settings", nil)
}

// UpdateSettings changes the workspace's settings.
func (s *WorkspacesService) UpdateSettings(req *UpdateWorkspaceSettingsRequest) (*WorkspaceSettings, error) {
	return put[WorkspaceSettings](s.c, wsPath+"/settings", req)
}

// Plan returns the quota and feature set applied to the workspace.
func (s *WorkspacesService) Plan() (*Plan, error) {
	return get[Plan](s.c, wsPath+"/plan", nil)
}

// ListAuditLog returns a page of the workspace's audit events.
func (s *WorkspacesService) ListAuditLog(opts *ListOptions) (*PageableResponse[Event], error) {
	return callPage[Event](s.c, wsPath+"/audit-log", opts.values())
}

// GetAuditEvent returns one audit event with its full metadata.
func (s *WorkspacesService) GetAuditEvent(id uint) (*Event, error) {
	return get[Event](s.c, fmt.Sprintf("%s/audit-log/%d", wsPath, id), nil)
}

// WorkspaceDataExport is a portable snapshot of a workspace's configuration
// and content, as produced by [WorkspacesService.ExportData] and consumed by
// [WorkspacesService.ImportData].
type WorkspaceDataExport struct {
	Templates        []TemplateExport  `json:"templates,omitempty"`
	Languages        []Language        `json:"languages,omitempty"`
	Stylesheets      []Stylesheet      `json:"stylesheets,omitempty"`
	Domains          []Domain          `json:"domains,omitempty"`
	SMTPServers      []SMTPServer      `json:"smtp_servers,omitempty"`
	Webhooks         []Webhook         `json:"webhooks,omitempty"`
	Subscribers      []Subscriber      `json:"subscribers,omitempty"`
	SubscriberLists  []SubscriberList  `json:"subscriber_lists,omitempty"`
	ContactLists     []SubscriberList  `json:"contact_lists,omitempty"`
	Contacts         []Contact         `json:"contacts,omitempty"`
	Suppressions     []Suppression     `json:"suppressions,omitempty"`
	UnsubscribeLists []UnsubscribeList `json:"unsubscribe_lists,omitempty"`
	Campaigns        []Campaign        `json:"campaigns,omitempty"`
	Forms            []Form            `json:"forms,omitempty"`
	MessageFilters   []MessageFilter   `json:"message_filters,omitempty"`
	PostaVersion     string            `json:"posta_version,omitempty"`
	ExportedAt       string            `json:"exported_at,omitempty"`
}

// ExportData returns a portable snapshot of the workspace, for backup or for
// moving it to another Posta deployment.
func (s *WorkspacesService) ExportData() (*WorkspaceDataExport, error) {
	return get[WorkspaceDataExport](s.c, wsPath+"/data/export", nil)
}

// ImportData restores a snapshot into the active workspace.
func (s *WorkspacesService) ImportData(data *WorkspaceDataExport) (*MessageData, error) {
	return post[MessageData](s.c, wsPath+"/data/import", data, nil)
}

// GDPRDeleteResult reports how many records an erasure removed.
type GDPRDeleteResult struct {
	Deleted int64  `json:"deleted"`
	Message string `json:"message,omitempty"`
}

// DeleteContactData erases a data subject's contact, subscriber, and
// suppression records. Passing an empty email erases every contact in the
// workspace, so pass the address you mean.
func (s *WorkspacesService) DeleteContactData(email string) (*GDPRDeleteResult, error) {
	body := struct {
		Email string `json:"email,omitempty"`
	}{Email: email}
	return post[GDPRDeleteResult](s.c, wsPath+"/gdpr/delete-contacts", body, nil)
}

// DeleteEmailLogs erases stored email records older than olderThanDays, for
// meeting a retention policy.
func (s *WorkspacesService) DeleteEmailLogs(olderThanDays int) (*GDPRDeleteResult, error) {
	body := struct {
		OlderThanDays int `json:"older_than_days"`
	}{OlderThanDays: olderThanDays}
	return post[GDPRDeleteResult](s.c, wsPath+"/gdpr/delete-email-logs", body, nil)
}

// WorkspaceSSO binds a workspace to an OAuth provider for single sign-on.
type WorkspaceSSO struct {
	ProviderID   uint   `json:"provider_id"`
	ProviderName string `json:"provider_name,omitempty"`
	// AllowedDomains restricts SSO to these email domains, comma separated.
	AllowedDomains string `json:"allowed_domains,omitempty"`
	// AutoProvision creates a workspace membership on first SSO login.
	AutoProvision bool `json:"auto_provision"`
	// EnforceSSO refuses password logins for members of this workspace.
	EnforceSSO bool `json:"enforce_sso"`
}

// SSO returns the workspace's single sign-on configuration.
func (s *WorkspacesService) SSO() (*WorkspaceSSO, error) {
	return get[WorkspaceSSO](s.c, wsPath+"/sso", nil)
}

// SetSSO configures single sign-on for the workspace.
func (s *WorkspacesService) SetSSO(req *WorkspaceSSO) (*WorkspaceSSO, error) {
	return put[WorkspaceSSO](s.c, wsPath+"/sso", req)
}

// DeleteSSO removes the workspace's single sign-on configuration.
func (s *WorkspacesService) DeleteSSO() error {
	return s.c.delete(wsPath + "/sso")
}
