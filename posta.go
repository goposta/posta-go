// Package posta provides a Go client for the Posta public email API.
//
// It supports sending emails, template emails, batch emails,
// and checking email delivery status using API key authentication.
package posta

import (
	"fmt"

	"github.com/jkaninda/kitoko"
)

// Client is the Posta API client.
type Client struct {
	http *kitoko.Client
}

// APIError is returned when the Posta API responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Info       *ErrorInfo
}

// New creates a new Posta client authenticated with an API key (Bearer token).
//
//	client := posta.New("https://posta.example.com", "your-api-key")
func New(baseURL, apiKey string) *Client {
	c := kitoko.NewClient(baseURL + "/api/v1")
	c.Headers["Authorization"] = "Bearer " + apiKey
	c.Headers["Content-Type"] = "application/json"
	return &Client{http: c}
}

// SendEmail sends a single email.
func (c *Client) SendEmail(req *SendEmailRequest) (*SendResponse, error) {
	return post[SendResponse](c, "/emails/send", req)
}

// SendTemplateEmail sends an email using a template.
func (c *Client) SendTemplateEmail(req *SendTemplateEmailRequest) (*SendResponse, error) {
	return post[SendResponse](c, "/emails/send-template", req)
}

// SendBatch sends a batch of emails using a template.
func (c *Client) SendBatch(req *BatchRequest) (*BatchResponse, error) {
	return post[BatchResponse](c, "/emails/batch", req)
}

// GetEmailStatus returns the delivery status of an email by UUID.
func (c *Client) GetEmailStatus(emailID string) (*EmailStatusResponse, error) {
	return get[EmailStatusResponse](c, "/emails/"+emailID+"/status")
}

// RetryEmail re-enqueues a failed email for another delivery attempt.
// Only emails with status "failed" can be retried, subject to the retry limit.
func (c *Client) RetryEmail(emailID string) (*SendResponse, error) {
	return post[SendResponse](c, "/emails/"+emailID+"/retry", nil)
}

// SubscribeToList adds an email to a named list via API key. The list is
// created on first use. Any prior list-scoped opt-out for the same
// (list, email) is cleared. Idempotent.
func (c *Client) SubscribeToList(req *ListSubscribeRequest) (*ListSubscribeResponse, error) {
	return post[ListSubscribeResponse](c, "/subscriber-lists/subscribe", req)
}

// UnsubscribeFromList opts an email out of a specific list. Idempotent; does
// not change the subscriber's global status. Returns the subscriber and the
// action taken.
func (c *Client) UnsubscribeFromList(listID uint, req *ListUnsubscribeRequest) (*ListSubscribeResponse, error) {
	return post[ListSubscribeResponse](c, fmt.Sprintf("/subscriber-lists/%d/unsubscribe", listID), req)
}

// ResubscribeToList reverses a list-scoped opt-out and (for static lists)
// re-adds the subscriber. Idempotent.
func (c *Client) ResubscribeToList(listID uint, email string) (*ListSubscribeResponse, error) {
	return post[ListSubscribeResponse](c, fmt.Sprintf("/subscriber-lists/%d/resubscribe", listID), map[string]string{"email": email})
}


func (e *APIError) Error() string {
	if e.Info != nil && e.Info.Message != "" {
		return fmt.Sprintf("posta: %d %s", e.StatusCode, e.Info.Message)
	}
	return fmt.Sprintf("posta: unexpected status %d", e.StatusCode)
}

func checkStatus(r *kitoko.Response) error {
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}
	apiErr := &APIError{StatusCode: r.StatusCode}
	var errResp ErrorResponse
	if err := r.JSON(&errResp); err == nil && errResp.Error != nil {
		apiErr.Info = errResp.Error
	}
	return apiErr
}

func get[T any](c *Client, path string) (*T, error) {
	r, err := c.http.GET(path).Execute()
	if err != nil {
		return nil, err
	}
	if err := checkStatus(r); err != nil {
		return nil, err
	}
	var resp Response[T]
	if err := r.JSON(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp.Data, nil
}

func post[T any](c *Client, path string, body any) (*T, error) {
	rb := c.http.POST(path)
	if body != nil {
		rb = rb.JSONBody(body)
	}
	r, err := rb.Execute()
	if err != nil {
		return nil, err
	}
	if err := checkStatus(r); err != nil {
		return nil, err
	}
	var resp Response[T]
	if err := r.JSON(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp.Data, nil
}
