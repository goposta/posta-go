package posta

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Webhook event names. Register a subset of these on a webhook to choose what
// Posta notifies you about.
const (
	// EventEmailSent fires when a message is accepted by the destination MTA.
	EventEmailSent = "email.sent"
	// EventEmailFailed fires when a message permanently fails after retries.
	EventEmailFailed = "email.failed"
	// EventEmailInbound fires when an inbound email is received and parsed.
	EventEmailInbound = "email.inbound"
	// EventEmailUnsubscribed fires when a recipient opts out via one-click
	// unsubscribe.
	EventEmailUnsubscribed = "email.unsubscribed"
	// EventEmailComplained fires when a recipient marks a message as spam.
	EventEmailComplained = "email.complained"
	// EventCampaignStarted fires when a campaign begins sending.
	EventCampaignStarted = "campaign.started"
	// EventCampaignCompleted fires when a campaign finishes sending.
	EventCampaignCompleted = "campaign.completed"
	// EventMessageReceived fires when a web form submission passes scanning.
	EventMessageReceived = "message.received"
	// EventMessageSpam fires when a submission is quarantined or rejected.
	EventMessageSpam = "message.spam"
)

// SignatureHeader names the header carrying a webhook's HMAC signature.
const SignatureHeader = "X-Posta-Signature"

// WebhooksService registers endpoints Posta notifies, and reads the delivery
// attempts it made.
type WebhooksService struct{ c *Client }

// List returns a page of webhooks. Requires an API key with the `webhooks`
// scope.
func (s *WebhooksService) List(opts *ListOptions) (*PageableResponse[Webhook], error) {
	return callPage[Webhook](s.c, "/webhooks", opts.values())
}

// Create registers a webhook endpoint. The response carries the signing
// secret, which is shown only here — store it to verify deliveries.
//
//	wh, err := c.Webhooks.Create(&posta.CreateWebhookRequest{
//	    URL:    "https://example.com/hooks/posta",
//	    Events: []string{posta.EventEmailSent, posta.EventEmailFailed},
//	})
func (s *WebhooksService) Create(req *CreateWebhookRequest) (*Webhook, error) {
	return post[Webhook](s.c, "/webhooks", req, nil)
}

// Delete removes a webhook.
func (s *WebhooksService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("/webhooks/%d", id))
}

// ListDeliveries returns a page of delivery attempts, with the HTTP status and
// error of each. Requires an API key with the `read` scope.
func (s *WebhooksService) ListDeliveries(opts *ListOptions) (*PageableResponse[WebhookDelivery], error) {
	return callPage[WebhookDelivery](s.c, "/webhook-deliveries", opts.values())
}

// ListInWorkspace returns a page of webhooks through the workspace-scoped
// endpoint, which a session credential can also reach.
func (s *WebhooksService) ListInWorkspace(opts *ListOptions) (*PageableResponse[Webhook], error) {
	return callPage[Webhook](s.c, wsPath+"/webhooks", opts.values())
}

// CreateInWorkspace registers a webhook through the workspace-scoped endpoint.
func (s *WebhooksService) CreateInWorkspace(req *CreateWebhookRequest) (*Webhook, error) {
	return post[Webhook](s.c, wsPath+"/webhooks", req, nil)
}

// DeleteInWorkspace removes a webhook through the workspace-scoped endpoint.
func (s *WebhooksService) DeleteInWorkspace(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/webhooks/%d", wsPath, id))
}

// ListDeliveriesInWorkspace returns delivery attempts through the
// workspace-scoped endpoint.
func (s *WebhooksService) ListDeliveriesInWorkspace(opts *ListOptions) (*PageableResponse[WebhookDelivery], error) {
	return callPage[WebhookDelivery](s.c, wsPath+"/webhook-deliveries", opts.values())
}

// VerifySignature reports whether signature authenticates body under secret.
//
// Posta signs each delivery with HMAC-SHA256 over the raw request body and
// sends it as "sha256=<hex>" in the X-Posta-Signature header. Pass the header
// value verbatim, along with the exact bytes received — re-serializing the JSON
// changes them and the check will fail.
//
//	if !posta.VerifySignature(body, r.Header.Get(posta.SignatureHeader), secret) {
//	    http.Error(w, "bad signature", http.StatusUnauthorized)
//	    return
//	}
func VerifySignature(body []byte, signature, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}
	provided := strings.TrimPrefix(signature, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	// Compare as bytes in constant time so a wrong signature leaks nothing
	// about how much of it was right.
	return hmac.Equal([]byte(provided), []byte(expected))
}

// WebhookEvent is the payload of the email.sent and email.failed events.
type WebhookEvent struct {
	Event     string `json:"event"`
	EmailID   string `json:"email_id"`
	Timestamp string `json:"timestamp"`
}

// CampaignWebhookEvent is the payload of the campaign.started and
// campaign.completed events.
type CampaignWebhookEvent struct {
	Event      string `json:"event"`
	CampaignID uint   `json:"campaign_id"`
	Name       string `json:"name"`
	Timestamp  string `json:"timestamp"`
}

// ComplaintWebhookEvent is the payload of the email.complained event.
type ComplaintWebhookEvent struct {
	Event     string `json:"event"`
	Email     string `json:"email"`
	EmailUUID string `json:"email_uuid"`
	Timestamp string `json:"timestamp"`
}

// UnsubscribeWebhookEvent is the payload of the email.unsubscribed event.
// ListID names the unsubscribe list the recipient opted out of, if any.
type UnsubscribeWebhookEvent struct {
	Event     string `json:"event"`
	Email     string `json:"email"`
	EmailUUID string `json:"email_uuid"`
	ListID    *uint  `json:"list_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// InboundAttachment describes a file on an inbound message.
type InboundAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type,omitempty"`
	Size        int64  `json:"size"`
}

// InboundWebhookEvent is the payload of the email.inbound event: a received
// message, already parsed.
type InboundWebhookEvent struct {
	Event       string              `json:"event"`
	InboundID   string              `json:"inbound_id"`
	MessageID   string              `json:"message_id,omitempty"`
	From        string              `json:"from"`
	To          []string            `json:"to"`
	Subject     string              `json:"subject,omitempty"`
	HTMLBody    string              `json:"html_body,omitempty"`
	TextBody    string              `json:"text_body,omitempty"`
	Headers     map[string]any      `json:"headers,omitempty"`
	Attachments []InboundAttachment `json:"attachments,omitempty"`
	Size        int64               `json:"size"`
	Source      string              `json:"source,omitempty"`
	ReceivedAt  string              `json:"received_at"`
	Timestamp   string              `json:"timestamp"`
}

// MessageWebhookEvent is the payload of the message.received and message.spam
// events: a web form submission and the verdict scanning gave it.
type MessageWebhookEvent struct {
	Event       string         `json:"event"`
	MessageID   string         `json:"message_id"`
	FormID      string         `json:"form_id"`
	FormName    string         `json:"form_name,omitempty"`
	Subject     string         `json:"subject,omitempty"`
	Body        string         `json:"body,omitempty"`
	Fields      []MessageField `json:"fields,omitempty"`
	SenderName  string         `json:"sender_name,omitempty"`
	SenderEmail string         `json:"sender_email,omitempty"`
	SenderPhone string         `json:"sender_phone,omitempty"`
	Status      string         `json:"status"`
	SpamScore   float64        `json:"spam_score"`
	ScanReasons []string       `json:"scan_reasons,omitempty"`
	ClientIP    string         `json:"client_ip,omitempty"`
	ReceivedAt  string         `json:"received_at"`
	Timestamp   string         `json:"timestamp"`
}
