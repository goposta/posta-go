package posta

// Methods kept on Client for source compatibility with earlier releases of
// this package, which exposed the send surface directly rather than through
// [Client.Emails] and its siblings. New code should call the services, which
// cover the whole API rather than this subset.

// SendEmail sends a single email.
//
// Deprecated: use [EmailsService.Send] via c.Emails.
func (c *Client) SendEmail(req *SendEmailRequest, opts ...SendOption) (*SendResponse, error) {
	return c.Emails.Send(req, opts...)
}

// SendEmailDryRun validates a send request without sending.
//
// Deprecated: use [EmailsService.SendDryRun] via c.Emails.
func (c *Client) SendEmailDryRun(req *SendEmailRequest) (*DryRunResponse, error) {
	return c.Emails.SendDryRun(req)
}

// SendTemplateEmail sends an email using a template.
//
// Deprecated: use [EmailsService.SendTemplate] via c.Emails.
func (c *Client) SendTemplateEmail(req *SendTemplateEmailRequest, opts ...SendOption) (*SendResponse, error) {
	return c.Emails.SendTemplate(req, opts...)
}

// SendTemplateEmailDryRun validates a template send without sending.
//
// Deprecated: use [EmailsService.SendTemplateDryRun] via c.Emails.
func (c *Client) SendTemplateEmailDryRun(req *SendTemplateEmailRequest) (*DryRunResponse, error) {
	return c.Emails.SendTemplateDryRun(req)
}

// SendBatch sends a batch of emails using a template.
//
// Deprecated: use [EmailsService.SendBatch] via c.Emails.
func (c *Client) SendBatch(req *BatchRequest, opts ...SendOption) (*BatchResponse, error) {
	return c.Emails.SendBatch(req, opts...)
}

// SendBatchDryRun validates a batch request without sending.
//
// Deprecated: use [EmailsService.SendBatchDryRun] via c.Emails.
func (c *Client) SendBatchDryRun(req *BatchRequest) (*DryRunResponse, error) {
	return c.Emails.SendBatchDryRun(req)
}

// PreviewTemplate renders a template without sending.
//
// Deprecated: use [EmailsService.Preview] via c.Emails.
func (c *Client) PreviewTemplate(req *PreviewRequest) (*PreviewResponse, error) {
	return c.Emails.Preview(req)
}

// VerifyEmail checks whether an email address is valid and deliverable.
//
// Deprecated: use [EmailsService.Verify] via c.Emails.
func (c *Client) VerifyEmail(req *VerifyEmailRequest, opts ...VerifyOption) (*VerificationResult, error) {
	return c.Emails.Verify(req, opts...)
}

// ListEmails returns a paginated list of emails.
//
// Deprecated: use [EmailsService.List] via c.Emails, which takes an
// [EmailListOptions] and can also filter and sort.
func (c *Client) ListEmails(page, size int) (*PageableResponse[Email], error) {
	return c.Emails.List(&EmailListOptions{ListOptions: ListOptions{Page: page, Size: size}})
}

// GetEmail returns a single email by its UUID.
//
// Deprecated: use [EmailsService.Get] via c.Emails.
func (c *Client) GetEmail(id string) (*Email, error) { return c.Emails.Get(id) }

// GetEmailStatus returns the delivery status of an email by UUID.
//
// Deprecated: use [EmailsService.Status] via c.Emails.
func (c *Client) GetEmailStatus(emailID string) (*EmailStatusResponse, error) {
	return c.Emails.Status(emailID)
}

// RetryEmail re-enqueues a failed email for another delivery attempt.
//
// Deprecated: use [EmailsService.Retry] via c.Emails.
func (c *Client) RetryEmail(emailID string) (*SendResponse, error) {
	return c.Emails.Retry(emailID)
}

// ListBounces returns a paginated list of bounces and complaints.
//
// Deprecated: use [BouncesService.List] via c.Bounces.
func (c *Client) ListBounces(page, size int) (*PageableResponse[Bounce], error) {
	return c.Bounces.List(&ListOptions{Page: page, Size: size})
}

// ListWebhookDeliveries returns a paginated list of webhook delivery records.
//
// Deprecated: use [WebhooksService.ListDeliveries] via c.Webhooks.
func (c *Client) ListWebhookDeliveries(page, size int) (*PageableResponse[WebhookDelivery], error) {
	return c.Webhooks.ListDeliveries(&ListOptions{Page: page, Size: size})
}

// ListWebhooks returns a paginated list of registered webhooks.
//
// Deprecated: use [WebhooksService.List] via c.Webhooks.
func (c *Client) ListWebhooks(page, size int) (*PageableResponse[Webhook], error) {
	return c.Webhooks.List(&ListOptions{Page: page, Size: size})
}

// CreateWebhook registers a new webhook.
//
// Deprecated: use [WebhooksService.Create] via c.Webhooks.
func (c *Client) CreateWebhook(req *CreateWebhookRequest) (*Webhook, error) {
	return c.Webhooks.Create(req)
}

// DeleteWebhook removes a webhook by its numeric ID.
//
// Deprecated: use [WebhooksService.Delete] via c.Webhooks.
func (c *Client) DeleteWebhook(id int) error { return c.Webhooks.Delete(uint(id)) }

// SubscribeToList adds an email to a named list.
//
// Deprecated: use [SubscriberListsService.Subscribe] via c.SubscriberLists.
func (c *Client) SubscribeToList(req *ListSubscribeRequest) (*ListSubscribeResponse, error) {
	return c.SubscriberLists.Subscribe(req)
}

// UnsubscribeFromList opts an email out of a specific list.
//
// Deprecated: use [SubscriberListsService.Unsubscribe] via c.SubscriberLists.
func (c *Client) UnsubscribeFromList(listID uint, req *ListUnsubscribeRequest) (*ListSubscribeResponse, error) {
	return c.SubscriberLists.Unsubscribe(listID, req)
}

// ResubscribeToList reverses a list-scoped opt-out.
//
// Deprecated: use [SubscriberListsService.Resubscribe] via c.SubscriberLists.
func (c *Client) ResubscribeToList(listID uint, email string) (*ListSubscribeResponse, error) {
	return c.SubscriberLists.Resubscribe(listID, email)
}
