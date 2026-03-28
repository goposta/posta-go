package posta

import "time"

// Response is the standard Posta API response envelope.
type Response[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data,omitempty"`
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

// SendEmailRequest is the request body for sending a single email.
type SendEmailRequest struct {
	From                string            `json:"from"`
	To                  []string          `json:"to"`
	Subject             string            `json:"subject"`
	HTML                string            `json:"html,omitempty"`
	Text                string            `json:"text,omitempty"`
	Attachments         []Attachment      `json:"attachments,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	ListUnsubscribeURL  string            `json:"list_unsubscribe_url,omitempty"`
	ListUnsubscribePost bool              `json:"list_unsubscribe_post,omitempty"`
	SendAt              *time.Time        `json:"send_at,omitempty"`
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
