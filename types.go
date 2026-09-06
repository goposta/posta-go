package posta

import "time"

// Response is the standard Posta API response envelope.
type Response[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data,omitempty"`
}

// Pageable holds pagination metadata returned alongside paginated lists.
type Pageable struct {
	CurrentPage   int   `json:"current_page"`
	Size          int   `json:"size"`
	TotalPages    int   `json:"total_pages"`
	TotalElements int64 `json:"total_elements"`
	Empty         bool  `json:"empty"`
}

// PageableResponse is the standard Posta API response envelope for paginated
// lists. Data holds the page of items and Pageable holds the page metadata.
type PageableResponse[T any] struct {
	Success  bool     `json:"success"`
	Data     []T      `json:"data"`
	Pageable Pageable `json:"pageable"`
}

// ErrorResponse is the error envelope returned by Posta.
type ErrorResponse struct {
	Success bool       `json:"success"`
	Error   *ErrorInfo `json:"error"`
}

// ErrorInfo holds error details.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// Unsubscribe configures the List-Unsubscribe (RFC 2369) and
// List-Unsubscribe-Post (RFC 8058) headers. ListID and URL are mutually
// exclusive.
type Unsubscribe struct {
	// ListID references an existing Posta-managed UnsubscribeList by id.
	// Posta mints the signed one-click URL and a click suppresses the
	// recipient on this list only. Mutually exclusive with URL.
	ListID *uint `json:"list_id,omitempty"`
	// URL is the caller-managed unsubscribe endpoint. Posta only emits the
	// List-Unsubscribe header; you own the endpoint. Also the RFC 8058
	// POST target when OneClick is true (which requires https). Mutually
	// exclusive with ListID.
	URL string `json:"url,omitempty"`
	// Mailto is an optional mailto: URI emitted alongside URL in the
	// List-Unsubscribe header (RFC 2369). A bare address is accepted; Posta
	// prepends "mailto:" if missing.
	Mailto string `json:"mailto,omitempty"`
	// OneClick emits "List-Unsubscribe-Post: List-Unsubscribe=One-Click"
	// (RFC 8058). Applies to the https URL target only; the caller-managed
	// path requires an https URL. Implied true on the Posta-managed
	// (ListID) path.
	OneClick bool `json:"one_click,omitempty"`
}

// SendEmailRequest is the request body for sending a single email.
type SendEmailRequest struct {
	From        string            `json:"from"`
	To          []string          `json:"to"`
	Subject     string            `json:"subject"`
	HTML        string            `json:"html,omitempty"`
	Text        string            `json:"text,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	// Unsubscribe configures the List-Unsubscribe / List-Unsubscribe-Post
	// headers. Supersedes the deprecated ListUnsubscribeURL /
	// ListUnsubscribePost fields.
	Unsubscribe *Unsubscribe `json:"unsubscribe,omitempty"`
	// Deprecated: use Unsubscribe.URL.
	ListUnsubscribeURL string `json:"list_unsubscribe_url,omitempty"`
	// Deprecated: use Unsubscribe.OneClick.
	ListUnsubscribePost bool       `json:"list_unsubscribe_post,omitempty"`
	SendAt              *time.Time `json:"send_at,omitempty"`
}

// SendTemplateEmailRequest is the request body for sending a template email.
// Provide either TemplateID or Template (name). TemplateID is preferred (primary key lookup);
// Template is a fallback when the ID is not known.
type SendTemplateEmailRequest struct {
	// TemplateID identifies the template by its numeric ID (preferred).
	TemplateID *uint `json:"template_id,omitempty"`
	// Template identifies the template by name (fallback when TemplateID is not provided).
	Template     string         `json:"template,omitempty"`
	Language     string         `json:"language,omitempty"`
	From         string         `json:"from,omitempty"`
	To           []string       `json:"to"`
	TemplateData map[string]any `json:"template_data,omitempty"`
	Attachments  []Attachment   `json:"attachments,omitempty"`
	// Unsubscribe configures the List-Unsubscribe / List-Unsubscribe-Post headers.
	Unsubscribe *Unsubscribe `json:"unsubscribe,omitempty"`
}

// BatchRequest is the request body for sending batch emails.
// Provide either TemplateID or Template (name). TemplateID is preferred (primary key lookup);
// Template is a fallback when the ID is not known.
type BatchRequest struct {
	// TemplateID identifies the template by its numeric ID (preferred).
	TemplateID *uint `json:"template_id,omitempty"`
	// Template identifies the template by name (fallback when TemplateID is not provided).
	Template   string           `json:"template,omitempty"`
	Language   string           `json:"language,omitempty"`
	From       string           `json:"from,omitempty"`
	Recipients []BatchRecipient `json:"recipients"`
	// Unsubscribe applies to every recipient in the batch
	// (per-recipient unsubscribe is intentionally not supported).
	Unsubscribe *Unsubscribe `json:"unsubscribe,omitempty"`
}

// BatchRecipient represents a single recipient in a batch send.
type BatchRecipient struct {
	Email        string         `json:"email"`
	Language     string         `json:"language,omitempty"`
	TemplateData map[string]any `json:"template_data,omitempty"`
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

// SendResponse is the response after sending an email.
type SendResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ListUnsubscribeRequest is the body for list-scoped unsubscribe/resubscribe.
type ListUnsubscribeRequest struct {
	Email  string `json:"email"`
	Reason string `json:"reason,omitempty"`
}

// ListSubscribeRequest is the body for the explicit list subscribe endpoint.
// The list is identified by name (created on first use); any prior
// list-scoped opt-out for this (list, email) is cleared.
type ListSubscribeRequest struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	List  string `json:"list"`
}

// ListSubscribeResponse is the result of a subscribe / unsubscribe /
// resubscribe call. Action is one of "subscribed", "unsubscribed",
// "resubscribed".
type ListSubscribeResponse struct {
	ListID            uint   `json:"list_id"`
	SubscriberID      uint   `json:"subscriber_id"`
	Email             string `json:"email"`
	Action            string `json:"action"`
	ListCreated       bool   `json:"list_created,omitempty"`
	SubscriberCreated bool   `json:"subscriber_created,omitempty"`
	MemberAdded       bool   `json:"member_added,omitempty"`
}

// BatchResponse is the response after a batch send.
type BatchResponse struct {
	Total   int           `json:"total"`
	Sent    int           `json:"sent"`
	Failed  int           `json:"failed"`
	Skipped int           `json:"skipped"`
	Results []BatchResult `json:"results"`
}

// BatchResult represents the result for a single recipient in a batch.
type BatchResult struct {
	Email  string `json:"email"`
	ID     string `json:"id,omitempty"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// EmailStatusResponse is a lightweight view of an email's delivery status.
type EmailStatusResponse struct {
	ID           string  `json:"id"`
	Status       string  `json:"status"`
	ErrorMessage string  `json:"error_message,omitempty"`
	RetryCount   int     `json:"retry_count"`
	CreatedAt    string  `json:"created_at"`
	SentAt       *string `json:"sent_at,omitempty"`
}

// PreviewRequest is the request body for rendering a template without sending.
// Provide either TemplateID or Template (name).
type PreviewRequest struct {
	TemplateID   *uint          `json:"template_id,omitempty"`
	Template     string         `json:"template,omitempty"`
	Language     string         `json:"language,omitempty"`
	TemplateData map[string]any `json:"template_data,omitempty"`
}

// PreviewResponse is the rendered template returned by /emails/preview.
type PreviewResponse struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
}

// VerifyEmailRequest is the body for POST /emails/verify.
type VerifyEmailRequest struct {
	Email string `json:"email"`
}

// VerificationChecks records which individual checks passed.
type VerificationChecks struct {
	Syntax      bool `json:"syntax"`
	MX          bool `json:"mx"`
	Disposable  bool `json:"disposable"`
	RoleAccount bool `json:"role_account"`
	// SMTP is always "skipped"; no SMTP probe is performed.
	SMTP string `json:"smtp"`
}

// VerificationResult is the outcome returned by POST /emails/verify.
// Status is one of "valid", "invalid", "risky", "disposable", "unknown".
type VerificationResult struct {
	Email             string             `json:"email"`
	Status            string             `json:"status"`
	Score             int                `json:"score"`
	Checks            VerificationChecks `json:"checks"`
	Reason            string             `json:"reason,omitempty"`
	MailboxVerified   bool               `json:"mailbox_verified"`
	Suppressed        bool               `json:"suppressed"`
	PreviouslyBounced bool               `json:"previously_bounced"`
	Cached            bool               `json:"cached"`
	CheckedAt         time.Time          `json:"checked_at"`
}

// DryRunResponse is the validation payload returned when ?dry_run=true is
// passed to send / send-template / batch. The exact shape is decided
// server-side; treat as an opaque verification payload.
type DryRunResponse map[string]any

// Email is the full record of an email handled by Posta, as returned by
// ListEmails and GetEmail. Status is one of "pending", "queued",
// "processing", "sent", "failed", "suppressed", "scheduled".
type Email struct {
	ID           uint       `json:"id"`
	UUID         string     `json:"uuid"`
	UserID       uint       `json:"user_id"`
	WorkspaceID  *uint      `json:"workspace_id,omitempty"`
	APIKeyID     *uint      `json:"api_key_id"`
	Sender       string     `json:"sender"`
	Recipients   []string   `json:"recipients"`
	Subject      string     `json:"subject"`
	TemplateName string     `json:"template_name,omitempty"`
	HTMLBody     string     `json:"html_body"`
	TextBody     string     `json:"text_body"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message"`
	RetryCount   int        `json:"retry_count"`
	CreatedAt    time.Time  `json:"created_at"`
	SentAt       *time.Time `json:"sent_at"`
	ScheduledAt  *time.Time `json:"scheduled_at"`
	Provider     string     `json:"provider,omitempty"`
	SMTPHostname string     `json:"smtp_hostname,omitempty"`
}

// Bounce records a bounce or complaint against a recipient. Type is one of
// "hard", "soft", "complaint".
type Bounce struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	WorkspaceID *uint     `json:"workspace_id,omitempty"`
	EmailID     uint      `json:"email_id"`
	Recipient   string    `json:"recipient"`
	Type        string    `json:"type"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

// Webhook is a registered webhook endpoint that Posta notifies on events.
type Webhook struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	WorkspaceID *uint     `json:"workspace_id,omitempty"`
	URL         string    `json:"url"`
	Events      []string  `json:"events"`
	Filters     []string  `json:"filters"`
	Secret      string    `json:"secret,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// WebhookDelivery records the result of delivering a webhook notification.
// Status is one of "success", "failed".
type WebhookDelivery struct {
	ID             uint      `json:"id"`
	WebhookID      uint      `json:"webhook_id"`
	UserID         uint      `json:"user_id"`
	WorkspaceID    *uint     `json:"workspace_id,omitempty"`
	Event          string    `json:"event"`
	Status         string    `json:"status"`
	HTTPStatusCode int       `json:"http_status_code"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	Attempt        int       `json:"attempt"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateWebhookRequest is the body for POST /webhooks. URL and at least one
// entry in Events are required. Valid event values are: email.sent,
// email.failed, email.inbound, email.unsubscribed, email.complained,
// campaign.started, campaign.completed.
type CreateWebhookRequest struct {
	URL     string   `json:"url"`
	Events  []string `json:"events"`
	Filters []string `json:"filters,omitempty"`
}
