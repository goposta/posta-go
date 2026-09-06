package posta

import "net/url"

// AuthService covers the public endpoints: signing in, registering, and
// recovering a password. They need no credential, so a client built for them
// can be created with an empty key.
type AuthService struct{ c *Client }

// AuthUser is the account summary returned alongside a session token.
type AuthUser struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	Role  string `json:"role"`
}

// AuthResponse is a successful sign-in: the session token and who it belongs
// to. Pass Token to [NewWithToken] to reach account-level endpoints.
type AuthResponse struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

// Login exchanges an email and password for a session token. Supply
// twoFactorCode when the account has 2FA enabled; pass an empty string
// otherwise.
func (s *AuthService) Login(email, password, twoFactorCode string) (*AuthResponse, error) {
	body := struct {
		Email         string `json:"email"`
		Password      string `json:"password"`
		TwoFactorCode string `json:"two_factor_code,omitempty"`
	}{Email: email, Password: password, TwoFactorCode: twoFactorCode}
	return post[AuthResponse](s.c, "/auth/login", body, nil)
}

// Register creates an account, when the deployment allows self-registration.
// Check first with [AuthService.RegistrationStatus].
func (s *AuthService) Register(name, email, password string) (*AuthResponse, error) {
	body := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{Name: name, Email: email, Password: password}
	return post[AuthResponse](s.c, "/auth/register", body, nil)
}

// RegistrationStatus reports whether self-registration is enabled.
func (s *AuthService) RegistrationStatus() (map[string]any, error) {
	res, err := get[map[string]any](s.c, "/auth/registration-status", nil)
	if err != nil {
		return nil, err
	}
	return *res, nil
}

// ForgotPassword emails a reset link. It always succeeds, whether or not the
// address has an account, so it cannot be used to enumerate users.
func (s *AuthService) ForgotPassword(email string) error {
	body := struct {
		Email string `json:"email"`
	}{Email: email}
	return s.c.callNoContent("POST", "/auth/forgot-password", body, nil)
}

// ResetPassword redeems a reset token and sets a new password.
func (s *AuthService) ResetPassword(token, newPassword string) error {
	body := struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}{Token: token, NewPassword: newPassword}
	return s.c.callNoContent("POST", "/auth/reset-password", body, nil)
}

// VerifyEmail redeems the token from a verification email.
func (s *AuthService) VerifyEmail(token string) error {
	return s.c.callNoContent("GET", "/auth/verify-email", nil, url.Values{"token": []string{token}})
}

// SSOProvider identifies the single sign-on provider an address should use.
type SSOProvider struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Type string `json:"type"`
}

// DiscoverSSO reports which SSO provider, if any, an email domain is bound to,
// so a login page can send the user straight to it.
func (s *AuthService) DiscoverSSO(email string) (*SSOProvider, error) {
	body := struct {
		Email string `json:"email"`
	}{Email: email}
	return post[SSOProvider](s.c, "/auth/oauth/discover", body, nil)
}

// ListOAuthProviders returns the SSO providers offered on the sign-in page.
func (s *AuthService) ListOAuthProviders() (map[string]any, error) {
	res, err := get[map[string]any](s.c, "/auth/oauth/providers", nil)
	if err != nil {
		return nil, err
	}
	return *res, nil
}

// AuthorizeURL returns the URL that begins an OAuth sign-in with provider.
// Redirect the browser here; Posta handles the callback itself.
func (c *Client) AuthorizeURL(provider string) string {
	return c.http.BaseURL() + "/auth/oauth/" + provider + "/authorize"
}

// SystemService reads build and health information.
type SystemService struct{ c *Client }

// AppInfo identifies the running build.
type AppInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	CommitID    string `json:"commit_id"`
	OpenAPIDocs bool   `json:"openapi_docs"`
}

// Info returns the running build's name, version, and commit. It is
// authenticated: the exact build is what an attacker needs to match a
// deployment against known CVEs.
func (s *SystemService) Info() (*AppInfo, error) {
	return get[AppInfo](s.c, "/info", nil)
}

// HealthStatus is a liveness or readiness answer. Database and Redis are
// filled in only by the readiness probe.
type HealthStatus struct {
	Status   string `json:"status"`
	Database string `json:"database,omitempty"`
	Redis    string `json:"redis,omitempty"`
}

// Healthz reports process liveness. Public: it needs no credential.
func (s *SystemService) Healthz() (*HealthStatus, error) {
	return getAbsolute[HealthStatus](s.c, "/healthz")
}

// Readyz reports whether dependencies (database, Redis) are reachable, which
// is what a load balancer should gate traffic on. Public.
func (s *SystemService) Readyz() (*HealthStatus, error) {
	return getAbsolute[HealthStatus](s.c, "/readyz")
}
