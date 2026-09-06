package posta

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// UsersService manages the signed-in account: profile, password, two-factor
// authentication, sessions, settings, and notifications.
//
// These endpoints accept only a user session token — an API key is never a
// valid credential here — so build the client with [NewWithToken].
type UsersService struct{ c *Client }

// UserProfile is the signed-in account.
type UserProfile struct {
	ID                        uint       `json:"id"`
	Email                     string     `json:"email"`
	Name                      string     `json:"name,omitempty"`
	Role                      string     `json:"role"`
	TwoFactorEnabled          bool       `json:"two_factor_enabled"`
	EmailVerifiedAt           *time.Time `json:"email_verified_at,omitempty"`
	EmailVerificationRequired bool       `json:"email_verification_required"`
	RequireVerifiedDomain     bool       `json:"require_verified_domain"`
	DefaultWorkspaceID        *uint      `json:"default_workspace_id,omitempty"`
	PersonalWorkspaceID       *uint      `json:"personal_workspace_id,omitempty"`
	// ScheduledDeletionAt is set once deletion has been requested, and
	// cleared by [UsersService.CancelDeletion].
	ScheduledDeletionAt *time.Time `json:"scheduled_deletion_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

// Me returns the signed-in account's profile.
func (s *UsersService) Me() (*UserProfile, error) {
	return get[UserProfile](s.c, "/users/me", nil)
}

// UpdateProfile changes the account's display name, and whether its sends
// require a verified domain.
func (s *UsersService) UpdateProfile(name string, requireVerifiedDomain *bool) (*UserProfile, error) {
	body := struct {
		Name                  string `json:"name"`
		RequireVerifiedDomain *bool  `json:"require_verified_domain,omitempty"`
	}{Name: name, RequireVerifiedDomain: requireVerifiedDomain}
	return put[UserProfile](s.c, "/users/me", body)
}

// ChangePassword sets a new password, confirming the current one.
func (s *UsersService) ChangePassword(currentPassword, newPassword string) error {
	body := struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}{CurrentPassword: currentPassword, NewPassword: newPassword}
	return s.c.callNoContent("PUT", "/users/me/password", body, nil)
}

// ResendVerificationEmail sends the address confirmation email again.
func (s *UsersService) ResendVerificationEmail() error {
	return s.c.callNoContent("POST", "/users/me/verify-email/resend", nil, nil)
}

// Plan returns the quota and feature set applied to the account.
func (s *UsersService) Plan() (*Plan, error) {
	return get[Plan](s.c, "/users/me/plan", nil)
}

// Enable2FAResponse carries the TOTP secret to enrol an authenticator app.
// URL is an otpauth:// URI, usually rendered as a QR code.
type Enable2FAResponse struct {
	Secret string `json:"secret"`
	URL    string `json:"url"`
}

// Setup2FA begins two-factor enrolment and returns the TOTP secret. Two-factor
// is not active until [UsersService.Verify2FA] confirms a code from it.
func (s *UsersService) Setup2FA() (*Enable2FAResponse, error) {
	return post[Enable2FAResponse](s.c, "/users/me/2fa/setup", nil, nil)
}

// Verify2FA confirms a code from the authenticator and switches two-factor on.
func (s *UsersService) Verify2FA(code string) error {
	body := struct {
		Code string `json:"code"`
	}{Code: code}
	return s.c.callNoContent("POST", "/users/me/2fa/verify", body, nil)
}

// Disable2FA switches two-factor off, confirming a current code.
func (s *UsersService) Disable2FA(code string) error {
	body := struct {
		Code string `json:"code"`
	}{Code: code}
	return s.c.callNoContent("POST", "/users/me/2fa/disable", body, nil)
}

// RequestDeletion schedules the account for deletion after a grace period.
// [UsersService.CancelDeletion] reverses it while the grace period lasts.
func (s *UsersService) RequestDeletion() error {
	return s.c.callNoContent("POST", "/users/me/delete", nil, nil)
}

// CancelDeletion calls off a scheduled account deletion.
func (s *UsersService) CancelDeletion() error {
	return s.c.callNoContent("POST", "/users/me/cancel-deletion", nil, nil)
}

// Session is one signed-in browser or client.
type Session struct {
	ID        uint   `json:"id"`
	Label     string `json:"label,omitempty"`
	Device    string `json:"device,omitempty"`
	Browser   string `json:"browser,omitempty"`
	OS        string `json:"os,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	// Current marks the session making this request.
	Current   bool   `json:"current"`
	CreatedAt string `json:"created_at,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// ListSessions returns the account's active sessions.
func (s *UsersService) ListSessions() (*[]Session, error) {
	return get[[]Session](s.c, "/users/me/sessions", nil)
}

// RevokeSession signs one session out.
func (s *UsersService) RevokeSession(id uint) error {
	return s.c.delete(fmt.Sprintf("/users/me/sessions/%d", id))
}

// RevokeOtherSessions signs out every session but this one — what to call
// after a password change on a possibly compromised account.
func (s *UsersService) RevokeOtherSessions() error {
	return s.c.callNoContent("POST", "/users/me/sessions/revoke-others", nil, nil)
}

// Logout signs out the session making the request.
func (s *UsersService) Logout() error {
	return s.c.callNoContent("POST", "/users/me/sessions/logout", nil, nil)
}

// SetDefaultWorkspace chooses the workspace a request lands in when it names
// none.
func (s *UsersService) SetDefaultWorkspace(workspaceID uint) error {
	body := struct {
		WorkspaceID uint `json:"workspace_id"`
	}{WorkspaceID: workspaceID}
	return s.c.callNoContent("PUT", "/users/me/default-workspace", body, nil)
}

// UserAuditLogOptions filters the account's audit log.
type UserAuditLogOptions struct {
	ListOptions
	Category string
	Search   string
}

func (o *UserAuditLogOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "category", o.Category)
	setStr(v, "search", o.Search)
	return v
}

// AuditLog returns a page of the account's own audit events.
func (s *UsersService) AuditLog(opts *UserAuditLogOptions) (*PageableResponse[Event], error) {
	return callPage[Event](s.c, "/users/me/audit-log", opts.values())
}

// UserSettings holds the account's personal defaults and notification
// preferences.
type UserSettings struct {
	UserID                  uint      `json:"user_id"`
	DefaultSenderEmail      string    `json:"default_sender_email,omitempty"`
	DefaultSenderName       string    `json:"default_sender_name,omitempty"`
	DefaultLanguage         string    `json:"default_language,omitempty"`
	DefaultTemplateID       *uint     `json:"default_template_id,omitempty"`
	NotificationEmail       string    `json:"notification_email,omitempty"`
	EmailNotifications      bool      `json:"email_notifications"`
	DailyReport             bool      `json:"daily_report"`
	NotifyBounceAlerts      bool      `json:"notify_bounce_alerts"`
	NotifyAPIKeyExpiry      bool      `json:"notify_api_key_expiry"`
	NotifyNewMessage        bool      `json:"notify_new_message"`
	NotifyWorkspaceActivity bool      `json:"notify_workspace_activity"`
	BounceAutoSuppress      bool      `json:"bounce_auto_suppress"`
	WebhookRetryCount       int       `json:"webhook_retry_count"`
	APIKeyExpiryDays        int       `json:"api_key_expiry_days"`
	Timezone                string    `json:"timezone,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// UpdateUserSettingsRequest changes account settings. Nil fields are left
// unchanged.
type UpdateUserSettingsRequest struct {
	DefaultSenderEmail      *string `json:"default_sender_email,omitempty"`
	DefaultSenderName       *string `json:"default_sender_name,omitempty"`
	DefaultLanguage         *string `json:"default_language,omitempty"`
	DefaultTemplateID       *uint   `json:"default_template_id,omitempty"`
	NotificationEmail       *string `json:"notification_email,omitempty"`
	EmailNotifications      *bool   `json:"email_notifications,omitempty"`
	DailyReport             *bool   `json:"daily_report,omitempty"`
	NotifyBounceAlerts      *bool   `json:"notify_bounce_alerts,omitempty"`
	NotifyAPIKeyExpiry      *bool   `json:"notify_api_key_expiry,omitempty"`
	NotifyNewMessage        *bool   `json:"notify_new_message,omitempty"`
	NotifyWorkspaceActivity *bool   `json:"notify_workspace_activity,omitempty"`
	BounceAutoSuppress      *bool   `json:"bounce_auto_suppress,omitempty"`
	WebhookRetryCount       *int    `json:"webhook_retry_count,omitempty"`
	APIKeyExpiryDays        *int    `json:"api_key_expiry_days,omitempty"`
	Timezone                *string `json:"timezone,omitempty"`
}

// Settings returns the account's settings.
func (s *UsersService) Settings() (*UserSettings, error) {
	return get[UserSettings](s.c, "/users/me/settings", nil)
}

// UpdateSettings changes the account's settings.
func (s *UsersService) UpdateSettings(req *UpdateUserSettingsRequest) (*UserSettings, error) {
	return put[UserSettings](s.c, "/users/me/settings", req)
}

// NotificationOptions filters the account's notifications.
type NotificationOptions struct {
	// Unread keeps only unread notifications.
	Unread *bool
	// Open keeps only notifications that have not been dismissed.
	Open     *bool
	Category string
	// Before pages backwards from a notification ID.
	Before uint
	Limit  int
	// Scoped restricts the result to the active workspace.
	Scoped *bool
}

func (o *NotificationOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	setBool(v, "unread", o.Unread)
	setBool(v, "open", o.Open)
	setStr(v, "category", o.Category)
	if o.Before > 0 {
		v.Set("before", strconv.FormatUint(uint64(o.Before), 10))
	}
	if o.Limit > 0 {
		v.Set("limit", strconv.Itoa(o.Limit))
	}
	setBool(v, "scoped", o.Scoped)
	return v
}

// ListNotifications returns the account's notifications.
func (s *UsersService) ListNotifications(opts *NotificationOptions) (*[]Notification, error) {
	return get[[]Notification](s.c, "/users/me/notifications", opts.values())
}

// ListBannerNotifications returns the notifications meant for the dashboard
// banner: platform announcements and anything needing attention now.
func (s *UsersService) ListBannerNotifications() (*[]Notification, error) {
	return get[[]Notification](s.c, "/users/me/notifications/banner", nil)
}

// NotificationCounts is the unread and open notification totals, for a badge.
type NotificationCounts struct {
	Unread int64 `json:"unread"`
	Open   int64 `json:"open"`
}

// NotificationCounts returns the unread and open notification totals.
func (s *UsersService) NotificationCounts() (*NotificationCounts, error) {
	return get[NotificationCounts](s.c, "/users/me/notifications/counts", nil)
}

// MarkNotificationsRead marks the given notifications read.
func (s *UsersService) MarkNotificationsRead(ids []uint) error {
	return s.c.callNoContent("POST", "/users/me/notifications/read", idsBody(ids), nil)
}

// MarkAllNotificationsRead marks every notification read.
func (s *UsersService) MarkAllNotificationsRead() error {
	return s.c.callNoContent("POST", "/users/me/notifications/read-all", nil, nil)
}

// DismissNotifications removes the given notifications from the list.
func (s *UsersService) DismissNotifications(ids []uint) error {
	return s.c.callNoContent("POST", "/users/me/notifications/dismiss", idsBody(ids), nil)
}

// DismissAllNotifications removes every notification from the list.
func (s *UsersService) DismissAllNotifications() error {
	return s.c.callNoContent("POST", "/users/me/notifications/dismiss-all", nil, nil)
}

func idsBody(ids []uint) any {
	return struct {
		IDs []uint `json:"ids"`
	}{IDs: ids}
}

// LinkedOAuthAccount is an external identity linked to the account.
type LinkedOAuthAccount struct {
	ID           uint   `json:"id"`
	ProviderID   uint   `json:"provider_id"`
	ProviderName string `json:"provider_name"`
	ProviderType string `json:"provider_type,omitempty"`
	Email        string `json:"email,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
}

// ListLinkedOAuthAccounts returns the external identities linked to the
// account.
func (s *UsersService) ListLinkedOAuthAccounts() (*[]LinkedOAuthAccount, error) {
	return get[[]LinkedOAuthAccount](s.c, "/users/me/oauth", nil)
}

// UnlinkOAuthAccount detaches an external identity from the account.
func (s *UsersService) UnlinkOAuthAccount(providerID uint) error {
	return s.c.delete(fmt.Sprintf("/users/me/oauth/%d", providerID))
}
