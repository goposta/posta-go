# Posta Go Client

Official Go client for the [Posta](https://github.com/goposta/posta) email delivery platform, built with [kitoko](https://github.com/jkaninda/kitoko).

## Installation

```bash
go get github.com/goposta/posta-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    posta "github.com/goposta/posta-go"
)

func main() {
    client := posta.New("https://posta.example.com", "your-api-key")

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
}
```

## API Reference

### Send Email

```go
client.SendEmail(&posta.SendEmailRequest{
    From:    "sender@example.com",
    To:      []string{"recipient@example.com"},
    Subject: "Hello",
    HTML:    "<h1>Hello!</h1>",
    Text:    "Hello!",                          // optional plain text
    Attachments: []posta.Attachment{{           // optional
        Filename:    "doc.pdf",
        Content:     "<base64-encoded>",
        ContentType: "application/pdf",
    }},
    Headers:             map[string]string{"X-Custom": "value"},  // optional
    ListUnsubscribeURL:  "https://example.com/unsub",             // optional
    SendAt:              &scheduledTime,                           // optional
})
```

### Send Template Email

```go
client.SendTemplateEmail(&posta.SendTemplateEmailRequest{
    Template:     "welcome",
    To:           []string{"user@example.com"},
    From:         "noreply@example.com",
    Language:     "en",                          // optional
    TemplateData: map[string]any{"name": "Alice"},
})
```

### Batch Send

```go
client.SendBatch(&posta.BatchRequest{
    Template: "newsletter",
    From:     "news@example.com",
    Recipients: []posta.BatchRecipient{
        {Email: "user1@example.com", TemplateData: map[string]any{"name": "Bob"}},
        {Email: "user2@example.com", Language: "fr", TemplateData: map[string]any{"name": "Carol"}},
    },
})
```

### Check Delivery Status

```go
status, _ := client.GetEmailStatus("email-uuid")
fmt.Printf("Status: %s\n", status.Status)
```

## Error Handling

All methods return typed errors. API errors can be inspected:

```go
_, err := client.GetEmailStatus("invalid-uuid")
if err != nil {
    var apiErr *posta.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("Status: %d\n", apiErr.StatusCode)
        if apiErr.Info != nil {
            fmt.Printf("Message: %s\n", apiErr.Info.Message)
        }
    }
}
```

## Contributing

Contributions are welcome! Please open an issue to discuss proposed changes before submitting a pull request.
---
## License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.

---

## Copyright

Copyright © 2026 Jonas Kaninda
