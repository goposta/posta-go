package posta

import (
	"net/http"
	"net/url"
)

// SendOption configures a single send, send-template, or batch call.
type SendOption func(*sendOpts)

// VerifyOption configures a single verify call.
type VerifyOption func(*verifyOpts)

type sendOpts struct{ dryRun bool }

type verifyOpts struct{ fresh bool }

// WithDryRun validates the request without sending it. The typed Send methods
// still decode a SendResponse; use the DryRun variants to receive the
// validation payload in its own shape.
func WithDryRun() SendOption { return func(o *sendOpts) { o.dryRun = true } }

// WithFresh bypasses the verifier cache and re-checks the address.
func WithFresh() VerifyOption { return func(o *verifyOpts) { o.fresh = true } }

func applySendOpts(opts []SendOption) sendOpts {
	var o sendOpts
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

func applyVerifyOpts(opts []VerifyOption) verifyOpts {
	var o verifyOpts
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// EmailsService sends mail and reads the resulting delivery records.
//
// Sending requires an API key with the `send` scope; the list and get calls
// require `read`.
type EmailsService struct{ c *Client }

// Send sends a single email.
//
//	res, err := c.Emails.Send(&posta.SendEmailRequest{
//	    From:    "Acme <hello@example.com>",
//	    To:      []string{"user@example.com"},
//	    Subject: "Hello",
//	    HTML:    "<h1>Hi</h1>",
//	})
func (s *EmailsService) Send(req *SendEmailRequest, opts ...SendOption) (*SendResponse, error) {
	return post[SendResponse](s.c, "/emails/send", req, sendQuery(opts))
}

// SendDryRun validates a send request without delivering it. The payload
// reports what would have happened: which recipients are suppressed, whether
// the sending domain is verified, and how the template renders.
func (s *EmailsService) SendDryRun(req *SendEmailRequest) (*DryRunResponse, error) {
	return post[DryRunResponse](s.c, "/emails/send", req, dryRunValues())
}

// SendTemplate sends an email rendered from a stored template. Identify the
// template by ID (preferred) or by name.
func (s *EmailsService) SendTemplate(req *SendTemplateEmailRequest, opts ...SendOption) (*SendResponse, error) {
	return post[SendResponse](s.c, "/emails/send-template", req, sendQuery(opts))
}

// SendTemplateDryRun validates a template send without delivering it.
func (s *EmailsService) SendTemplateDryRun(req *SendTemplateEmailRequest) (*DryRunResponse, error) {
	return post[DryRunResponse](s.c, "/emails/send-template", req, dryRunValues())
}

// SendBatch sends one template to many recipients, substituting per-recipient
// variables. The response reports each recipient's outcome individually.
func (s *EmailsService) SendBatch(req *BatchRequest, opts ...SendOption) (*BatchResponse, error) {
	return post[BatchResponse](s.c, "/emails/batch", req, sendQuery(opts))
}

// SendBatchDryRun validates a batch send without delivering it.
func (s *EmailsService) SendBatchDryRun(req *BatchRequest) (*DryRunResponse, error) {
	return post[DryRunResponse](s.c, "/emails/batch", req, dryRunValues())
}

// Preview renders a template with variables and returns the subject, HTML, and
// text without sending anything.
func (s *EmailsService) Preview(req *PreviewRequest) (*PreviewResponse, error) {
	return post[PreviewResponse](s.c, "/emails/preview", req, nil)
}

// Verify checks whether an address is worth sending to: syntax, MX records,
// disposable and role-account detection, and the caller's own suppression and
// bounce history. Results are cached; pass [WithFresh] to re-check.
func (s *EmailsService) Verify(req *VerifyEmailRequest, opts ...VerifyOption) (*VerificationResult, error) {
	o := applyVerifyOpts(opts)
	var q url.Values
	if o.fresh {
		q = url.Values{"fresh": []string{"true"}}
	}
	return post[VerificationResult](s.c, "/emails/verify", req, q)
}

// Status returns a lightweight delivery status for one email, suitable for
// polling while it moves through the queue.
func (s *EmailsService) Status(uuid string) (*EmailStatusResponse, error) {
	return get[EmailStatusResponse](s.c, "/emails/"+uuid+"/status", nil)
}

// Retry re-enqueues a failed email. Only emails in the "failed" state can be
// retried, and only up to the SMTP server's retry limit.
func (s *EmailsService) Retry(uuid string) (*SendResponse, error) {
	return post[SendResponse](s.c, "/emails/"+uuid+"/retry", nil, nil)
}

// List returns a page of emails. Requires an API key with the `read` scope.
func (s *EmailsService) List(opts *EmailListOptions) (*PageableResponse[Email], error) {
	return callPage[Email](s.c, "/emails", opts.values())
}

// Get returns the full record of one email by UUID, including its rendered
// bodies. Requires an API key with the `read` scope.
func (s *EmailsService) Get(uuid string) (*Email, error) {
	return get[Email](s.c, "/emails/"+uuid, nil)
}

// ListInWorkspace returns a page of emails through the workspace-scoped
// endpoint, which a session credential can also reach.
func (s *EmailsService) ListInWorkspace(opts *EmailListOptions) (*PageableResponse[Email], error) {
	return callPage[Email](s.c, wsPath+"/emails", opts.values())
}

// GetInWorkspace returns one email through the workspace-scoped endpoint.
func (s *EmailsService) GetInWorkspace(uuid string) (*Email, error) {
	return get[Email](s.c, wsPath+"/emails/"+uuid, nil)
}

func sendQuery(opts []SendOption) url.Values {
	if applySendOpts(opts).dryRun {
		return dryRunValues()
	}
	return nil
}

func dryRunValues() url.Values { return url.Values{"dry_run": []string{"true"}} }

// BouncesService reads recorded bounces and complaints, and records them on
// behalf of an external provider.
type BouncesService struct{ c *Client }

// List returns a page of bounces and complaints. Requires an API key with the
// `read` scope.
func (s *BouncesService) List(opts *ListOptions) (*PageableResponse[Bounce], error) {
	return callPage[Bounce](s.c, "/bounces", opts.values())
}

// ListInWorkspace returns a page of bounces through the workspace-scoped
// endpoint, which a session credential can also reach.
func (s *BouncesService) ListInWorkspace(opts *ListOptions) (*PageableResponse[Bounce], error) {
	return callPage[Bounce](s.c, wsPath+"/bounces", opts.values())
}

// RecordBounceRequest reports a bounce observed elsewhere. Type is "hard" or
// "soft".
type RecordBounceRequest struct {
	// EmailID is the UUID of the email that bounced.
	EmailID string `json:"email_id"`
	// Recipient is the address that rejected or complained about it.
	Recipient string `json:"recipient"`
	Type      string `json:"type"`
	Reason    string `json:"reason,omitempty"`
}

// Record files a bounce or complaint against a recipient, for callers relaying
// notifications from a provider Posta does not poll itself.
func (s *BouncesService) Record(req *RecordBounceRequest) (*Bounce, error) {
	return post[Bounce](s.c, wsPath+"/bounces", req, nil)
}

// SuppressionsService manages the addresses Posta refuses to deliver to.
type SuppressionsService struct{ c *Client }

// List returns a page of suppressed addresses. Set ListID to see the opt-outs
// recorded against a single unsubscribe list.
func (s *SuppressionsService) List(opts *SuppressionListOptions) (*PageableResponse[Suppression], error) {
	return callPage[Suppression](s.c, wsPath+"/suppressions", opts.values())
}

// AddSuppressionRequest suppresses one address. ListID scopes the suppression
// to a single unsubscribe list; without it the address is suppressed workspace
// wide.
type AddSuppressionRequest struct {
	Email  string `json:"email"`
	Reason string `json:"reason,omitempty"`
	ListID *uint  `json:"list_id,omitempty"`
}

// Add suppresses an address.
func (s *SuppressionsService) Add(req *AddSuppressionRequest) (*Suppression, error) {
	return post[Suppression](s.c, wsPath+"/suppressions", req, nil)
}

// Remove lifts a suppression, letting Posta deliver to the address again. Pass
// listID to lift a list-scoped opt-out; pass nil for the workspace-wide entry.
func (s *SuppressionsService) Remove(email string, listID *uint) error {
	body := struct {
		Email  string `json:"email"`
		ListID *uint  `json:"list_id,omitempty"`
	}{Email: email, ListID: listID}
	return s.c.callNoContent(http.MethodDelete, wsPath+"/suppressions", body, nil)
}
