// Command example exercises a representative slice of the Posta Go client:
// sending, templates, subscribers, campaigns, and webhook verification.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	posta "github.com/goposta/posta-go"
)

func main() {
	client := posta.New("https://posta.example.com", "psk_your_api_key",
		posta.WithWorkspace(1),
	)

	// A plain transactional send.
	resp, err := client.Emails.Send(&posta.SendEmailRequest{
		From:    "Acme <hello@example.com>",
		To:      []string{"user@example.com"},
		Subject: "Hello from Posta",
		HTML:    "<h1>Hello!</h1><p>This is a test email.</p>",
		Text:    "Hello! This is a test email.",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sent: id=%s status=%s\n", resp.ID, resp.Status)

	// Poll its delivery status.
	status, err := client.Emails.Status(resp.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("status: %s (retries: %d)\n", status.Status, status.RetryCount)

	// Send from a stored template.
	if _, err := client.Emails.SendTemplate(&posta.SendTemplateEmailRequest{
		Template:     "welcome",
		To:           []string{"user@example.com"},
		From:         "noreply@example.com",
		TemplateData: map[string]any{"name": "Alice"},
	}); err != nil {
		log.Fatal(err)
	}

	// Batch send with per-recipient variables.
	batch, err := client.Emails.SendBatch(&posta.BatchRequest{
		Template: "welcome",
		From:     "noreply@example.com",
		Recipients: []posta.BatchRecipient{
			{Email: "a@example.com", TemplateData: map[string]any{"name": "Ada"}},
			{Email: "b@example.com", TemplateData: map[string]any{"name": "Grace"}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("batch: %d sent, %d failed\n", batch.Sent, batch.Failed)

	// Check an address before adding it to a list.
	verdict, err := client.Emails.Verify(&posta.VerifyEmailRequest{Email: "user@example.com"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("verify: %s (score %d)\n", verdict.Status, verdict.Score)

	// Page through recent emails.
	page, err := client.Emails.List(&posta.EmailListOptions{
		ListOptions: posta.ListOptions{Size: 10},
		Sort:        "-created_at",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("emails: %d of %d\n", len(page.Data), page.Pageable.TotalElements)

	// Register a webhook. The secret is returned only here.
	hook, err := client.Webhooks.Create(&posta.CreateWebhookRequest{
		URL:    "https://example.com/hooks/posta",
		Events: []string{posta.EventEmailSent, posta.EventEmailFailed},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("webhook %d registered; store secret %q\n", hook.ID, hook.Secret)

	// Errors carry the API's status and message.
	if _, err := client.Emails.Get("does-not-exist"); err != nil {
		if posta.IsNotFound(err) {
			fmt.Println("no such email, as expected")
		}
	}
}

// webhookHandler shows how to authenticate an incoming Posta webhook. Verify
// the signature against the exact bytes received: decoding and re-encoding the
// JSON changes them, and the HMAC will not match.
func webhookHandler(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "unreadable body", http.StatusBadRequest)
			return
		}
		if !posta.VerifySignature(body, r.Header.Get(posta.SignatureHeader), secret) {
			http.Error(w, "bad signature", http.StatusUnauthorized)
			return
		}

		var event posta.WebhookEvent
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}
		switch event.Event {
		case posta.EventEmailSent:
			log.Printf("delivered: %s", event.EmailID)
		case posta.EventEmailFailed:
			log.Printf("failed: %s", event.EmailID)
		}
		w.WriteHeader(http.StatusOK)
	}
}
