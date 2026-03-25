package main

import (
	"fmt"
	"log"

	posta "github.com/goposta/posta-go"
)

func main() {
	client := posta.New("https://posta.example.com", "your-api-key")

	// Send a simple email
	resp, err := client.SendEmail(&posta.SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Hello from Posta",
		HTML:    "<h1>Hello!</h1><p>This is a test email.</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Email sent: id=%s status=%s\n", resp.ID, resp.Status)

	// Check email status
	status, err := client.GetEmailStatus(resp.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Email status: %s\n", status.Status)

	// Send using a template
	templateResp, err := client.SendTemplateEmail(&posta.SendTemplateEmailRequest{
		Template: "welcome",
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
		TemplateData: map[string]any{
			"name": "Alice",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Template email sent: id=%s\n", templateResp.ID)

	// Batch send
	batchResp, err := client.SendBatch(&posta.BatchRequest{
		Template: "newsletter",
		From:     "news@example.com",
		Recipients: []posta.BatchRecipient{
			{Email: "user1@example.com", TemplateData: map[string]any{"name": "Bob"}},
			{Email: "user2@example.com", TemplateData: map[string]any{"name": "Carol"}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Batch sent: total=%d sent=%d failed=%d\n", batchResp.Total, batchResp.Sent, batchResp.Failed)

	// Error handling
	_, err = client.GetEmailStatus("invalid-uuid")
	if err != nil {
		if apiErr, ok := err.(*posta.APIError); ok {
			fmt.Printf("API error: status=%d message=%s\n", apiErr.StatusCode, apiErr.Info.Message)
		}
	}
}
