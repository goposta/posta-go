package posta

import (
	"fmt"
	"net/url"
	"strconv"
)

// FormsService manages web form endpoints: public URLs a website posts to,
// whose submissions land in the workspace as messages.
type FormsService struct{ c *Client }

// CreateFormRequest creates a form endpoint. Only Name is required; the
// anti-spam defaults are sensible and can be tuned afterwards with
// [FormsService.Update].
type CreateFormRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
	// AllowedOrigins lists the site origins permitted to submit. Combined
	// with StrictOrigin it is the main defence against a copied endpoint.
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	StrictOrigin   bool     `json:"strict_origin,omitempty"`
	// RequireNonce demands a short-lived signed nonce with each submission,
	// obtained from [FormsService.Nonce].
	RequireNonce bool `json:"require_nonce,omitempty"`
	// RedirectURL sends browser form posts back to a thank-you page.
	RedirectURL      string `json:"redirect_url,omitempty"`
	AllowAttachments bool   `json:"allow_attachments,omitempty"`
	// NotifyEmails receive an alert for each submission; NotifyMode is
	// "immediate", "hourly", "daily", or "off".
	NotifyEmails  []string `json:"notify_emails,omitempty"`
	NotifyMode    string   `json:"notify_mode,omitempty"`
	ReplyFrom     string   `json:"reply_from,omitempty"`
	ReplyFromName string   `json:"reply_from_name,omitempty"`
}

// UpdateFormRequest changes a form. Nil fields are left unchanged, so a single
// setting can be adjusted without restating the rest. Status is "active",
// "paused", or "archived".
type UpdateFormRequest struct {
	Name                string    `json:"name,omitempty"`
	Slug                string    `json:"slug,omitempty"`
	Description         string    `json:"description,omitempty"`
	Status              string    `json:"status,omitempty"`
	AllowedOrigins      *[]string `json:"allowed_origins,omitempty"`
	StrictOrigin        *bool     `json:"strict_origin,omitempty"`
	HoneypotField       *string   `json:"honeypot_field,omitempty"`
	RequireNonce        *bool     `json:"require_nonce,omitempty"`
	MinFillSeconds      *int      `json:"min_fill_seconds,omitempty"`
	MaxFields           *int      `json:"max_fields,omitempty"`
	MaxBodyBytes        *int64    `json:"max_body_bytes,omitempty"`
	AllowAttachments    *bool     `json:"allow_attachments,omitempty"`
	RedirectURL         *string   `json:"redirect_url,omitempty"`
	ScanEnabled         *bool     `json:"scan_enabled,omitempty"`
	FlagThreshold       *float64  `json:"flag_threshold,omitempty"`
	QuarantineThreshold *float64  `json:"quarantine_threshold,omitempty"`
	RejectThreshold     *float64  `json:"reject_threshold,omitempty"`
	NotifyEnabled       *bool     `json:"notify_enabled,omitempty"`
	NotifyEmails        *[]string `json:"notify_emails,omitempty"`
	NotifyMode          *string   `json:"notify_mode,omitempty"`
	NotifyOnFlagged     *bool     `json:"notify_on_flagged,omitempty"`
	ReplyFrom           *string   `json:"reply_from,omitempty"`
	ReplyFromName       *string   `json:"reply_from_name,omitempty"`
	RetentionDays       *int      `json:"retention_days,omitempty"`
}

// Create adds a form and returns it with its public key.
func (s *FormsService) Create(req *CreateFormRequest) (*Form, error) {
	return post[Form](s.c, wsPath+"/forms", req, nil)
}

// List returns a page of forms.
func (s *FormsService) List(opts *ListOptions) (*PageableResponse[Form], error) {
	return callPage[Form](s.c, wsPath+"/forms", opts.values())
}

// Get returns one form.
func (s *FormsService) Get(id uint) (*Form, error) {
	return get[Form](s.c, fmt.Sprintf("%s/forms/%d", wsPath, id), nil)
}

// Update changes a form.
func (s *FormsService) Update(id uint, req *UpdateFormRequest) (*Form, error) {
	return put[Form](s.c, fmt.Sprintf("%s/forms/%d", wsPath, id), req)
}

// Delete removes a form and the messages collected through it.
func (s *FormsService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/forms/%d", wsPath, id))
}

// RotateKey issues a new public key for a form. Existing embeds stop working
// the moment this returns, so update them together.
func (s *FormsService) RotateKey(id uint) (*Form, error) {
	return post[Form](s.c, fmt.Sprintf("%s/forms/%d/rotate-key", wsPath, id), nil, nil)
}

// FormSnippet is paste-ready embed code for a form.
type FormSnippet struct {
	PublicKey string `json:"public_key"`
	Endpoint  string `json:"endpoint"`
	// HTML is a plain <form> that posts directly to the endpoint.
	HTML string `json:"html"`
	// Fetch is a JavaScript fetch() call posting JSON to the endpoint.
	Fetch string `json:"fetch"`
}

// Snippet returns paste-ready HTML and fetch() code wired to a form.
func (s *FormsService) Snippet(id uint) (*FormSnippet, error) {
	return get[FormSnippet](s.c, fmt.Sprintf("%s/forms/%d/snippet", wsPath, id), nil)
}

// FormNonce is a short-lived, single-use token for a form that requires one.
type FormNonce struct {
	Nonce     string `json:"nonce"`
	IssuedAt  int64  `json:"issued_at,omitempty"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

// Nonce issues a submission nonce for a form whose RequireNonce is set. It is
// a public endpoint keyed by the form's public key, and needs no credential.
func (s *FormsService) Nonce(publicKey string) (*FormNonce, error) {
	return get[FormNonce](s.c, "/f/"+publicKey+"/nonce", nil)
}

// Submit posts a submission to a form's public ingest endpoint, as a website
// would. It needs no credential, and always answers success for a stored or
// silently rejected submission so a spam client learns nothing.
func (s *FormsService) Submit(publicKey string, fields map[string]any) error {
	return s.c.callNoContent("POST", "/f/"+publicKey, fields, nil)
}

// MessagesService reads and triages the submissions collected by web forms.
type MessagesService struct{ c *Client }

// List returns a page of messages.
func (s *MessagesService) List(opts *MessageListOptions) (*PageableResponse[Message], error) {
	return callPage[Message](s.c, wsPath+"/messages", opts.values())
}

// Get returns one message with its fields, attachments, and reply thread.
// Reading a message marks it read.
func (s *MessagesService) Get(uuid string) (*Message, error) {
	return get[Message](s.c, wsPath+"/messages/"+uuid, nil)
}

// Delete removes a message.
func (s *MessagesService) Delete(uuid string) error {
	return s.c.delete(wsPath + "/messages/" + uuid)
}

// MessageStats holds the workspace's message counters.
type MessageStats struct {
	Total  int64 `json:"total"`
	Unread int64 `json:"unread"`
	Spam   int64 `json:"spam"`
	Forms  int64 `json:"forms"`
}

// Stats returns total, unread, and spam message counts plus the number of
// forms in the workspace.
func (s *MessagesService) Stats() (*MessageStats, error) {
	return get[MessageStats](s.c, wsPath+"/messages/stats", nil)
}

// MessageDailyCount is one day's submission volume.
type MessageDailyCount struct {
	Day   string `json:"day"`
	Total int64  `json:"total"`
	Spam  int64  `json:"spam"`
}

// MessageAnalytics is submission volume over time, with the spam share.
type MessageAnalytics struct {
	Total int64               `json:"total"`
	Spam  int64               `json:"spam"`
	Daily []MessageDailyCount `json:"daily"`
}

// Analytics returns submission volume for the last days days (default 30).
func (s *MessagesService) Analytics(days int) (*MessageAnalytics, error) {
	q := url.Values{}
	if days > 0 {
		q.Set("days", strconv.Itoa(days))
	}
	return get[MessageAnalytics](s.c, wsPath+"/messages/analytics", q)
}

// ReplyMessageRequest is an operator reply, sent to the submitter through the
// workspace's normal email pipeline and recorded on the thread.
type ReplyMessageRequest struct {
	Subject string `json:"subject,omitempty"`
	Text    string `json:"text,omitempty"`
	HTML    string `json:"html,omitempty"`
}

// Reply answers a message's sender and records the reply on the thread.
func (s *MessagesService) Reply(uuid string, req *ReplyMessageRequest) (*MessageReply, error) {
	return post[MessageReply](s.c, wsPath+"/messages/"+uuid+"/reply", req, nil)
}

// UpdateState moves a message through triage. State is "new", "open",
// "replied", "closed", or "spam"; read, when set, also marks it read or
// unread.
func (s *MessagesService) UpdateState(uuid, state string, read *bool) (*Message, error) {
	body := struct {
		State string `json:"state"`
		Read  *bool  `json:"read,omitempty"`
	}{State: state, Read: read}
	return put[Message](s.c, wsPath+"/messages/"+uuid+"/state", body)
}

// Assign gives a message to a workspace member, or clears the assignment when
// userID is nil.
func (s *MessagesService) Assign(uuid string, userID *uint) (*Message, error) {
	body := struct {
		UserID *uint `json:"user_id"`
	}{UserID: userID}
	return put[Message](s.c, wsPath+"/messages/"+uuid+"/assign", body)
}

// MarkSpamRequest quarantines a message. Set CreateFilter to also derive a
// reusable filter from it, so later submissions like it are caught on arrival;
// Kind ("keyword", "phrase", "email", "domain", or "ip") and Pattern then
// describe the rule to create.
type MarkSpamRequest struct {
	CreateFilter bool   `json:"create_filter,omitempty"`
	Kind         string `json:"kind,omitempty"`
	Pattern      string `json:"pattern,omitempty"`
}

// MarkSpam quarantines a message and optionally learns a filter from it.
func (s *MessagesService) MarkSpam(uuid string, req *MarkSpamRequest) (*Message, error) {
	return post[Message](s.c, wsPath+"/messages/"+uuid+"/spam", req, nil)
}

// MarkNotSpam clears the spam verdict on a message.
func (s *MessagesService) MarkNotSpam(uuid string) (*Message, error) {
	return post[Message](s.c, wsPath+"/messages/"+uuid+"/not-spam", nil, nil)
}

// DownloadAttachment fetches the file at index idx of a message, returning its
// bytes and Content-Type. Indexes match the order of Message.Attachments.
func (s *MessagesService) DownloadAttachment(uuid string, idx int) ([]byte, string, error) {
	return s.c.callRaw(fmt.Sprintf("%s/messages/%s/attachments/%d", wsPath, uuid, idx), nil)
}

// MessageFiltersService manages the spam rules applied to form submissions as
// they arrive.
type MessageFiltersService struct{ c *Client }

// CreateMessageFilterRequest adds a spam rule.
//
// Kind selects what Pattern matches: "keyword", "phrase", "regex", "email",
// "domain", or "ip". Action is "score", "flag", "quarantine", "reject", or
// "allowlist". FormID limits the rule to one form; leave it nil to apply it
// workspace wide.
type CreateMessageFilterRequest struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
	// Fields names the submission fields to test; empty tests them all.
	Fields        []string `json:"fields,omitempty"`
	Action        string   `json:"action,omitempty"`
	Score         *float64 `json:"score,omitempty"`
	CaseSensitive bool     `json:"case_sensitive,omitempty"`
	FormID        *uint    `json:"form_id,omitempty"`
	Note          string   `json:"note,omitempty"`
}

// UpdateMessageFilterRequest changes a spam rule. Nil fields are left
// unchanged.
type UpdateMessageFilterRequest struct {
	Pattern       string    `json:"pattern,omitempty"`
	Fields        *[]string `json:"fields,omitempty"`
	Action        string    `json:"action,omitempty"`
	Score         *float64  `json:"score,omitempty"`
	CaseSensitive *bool     `json:"case_sensitive,omitempty"`
	Enabled       *bool     `json:"enabled,omitempty"`
	Note          *string   `json:"note,omitempty"`
}

// Create adds a spam filter.
func (s *MessageFiltersService) Create(req *CreateMessageFilterRequest) (*MessageFilter, error) {
	return post[MessageFilter](s.c, wsPath+"/message-filters", req, nil)
}

// List returns a page of spam filters with their hit counts.
func (s *MessageFiltersService) List(opts *ListOptions) (*PageableResponse[MessageFilter], error) {
	return callPage[MessageFilter](s.c, wsPath+"/message-filters", opts.values())
}

// Update changes a spam filter.
func (s *MessageFiltersService) Update(id uint, req *UpdateMessageFilterRequest) (*MessageFilter, error) {
	return put[MessageFilter](s.c, fmt.Sprintf("%s/message-filters/%d", wsPath, id), req)
}

// Delete removes a spam filter.
func (s *MessageFiltersService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/message-filters/%d", wsPath, id))
}

// FilterTestSample is one message a candidate filter would have matched.
type FilterTestSample struct {
	MessageUUID string `json:"message_uuid"`
	Subject     string `json:"subject,omitempty"`
	SenderEmail string `json:"sender_email,omitempty"`
	Excerpt     string `json:"excerpt,omitempty"`
}

// FilterTestResult reports what a candidate filter would have caught.
type FilterTestResult struct {
	Scanned int                `json:"scanned"`
	Matched int                `json:"matched"`
	Samples []FilterTestSample `json:"samples,omitempty"`
}

// TestMessageFilterRequest dry-runs a candidate pattern over recent messages,
// so a rule can be checked before it starts rejecting mail. Kind takes the
// same values as [CreateMessageFilterRequest].
type TestMessageFilterRequest struct {
	Kind          string `json:"kind"`
	Pattern       string `json:"pattern"`
	CaseSensitive bool   `json:"case_sensitive,omitempty"`
	// Limit caps how many recent messages are scanned.
	Limit int `json:"limit,omitempty"`
}

// Test reports how many recent messages a candidate filter would have matched,
// without creating it.
func (s *MessageFiltersService) Test(req *TestMessageFilterRequest) (*FilterTestResult, error) {
	return post[FilterTestResult](s.c, wsPath+"/message-filters/test", req, nil)
}
