# Posta Go Client

Official Go client for the [Posta](https://github.com/goposta/posta) email
platform, built on [okapi/client](https://github.com/jkaninda/okapi).

It covers the whole Posta API: transactional and templated sending, batch
sends, address verification, templates with versions and localizations,
campaigns, subscribers and lists, suppressions and bounces, domains, SMTP
servers and relay credentials, webhooks, web forms and the messages they
collect, inbound email, workspace administration, and the platform admin
surface.

## Installation

```bash
go get github.com/goposta/posta-go
```

**Requires:** Go 1.25+

## Quick start

```go
package main

import (
    "fmt"
    "log"

    posta "github.com/goposta/posta-go"
)

func main() {
    client := posta.New("https://posta.example.com", "psk_your_api_key")

    resp, err := client.Emails.Send(&posta.SendEmailRequest{
        From:    "Acme <hello@example.com>",
        To:      []string{"user@example.com"},
        Subject: "Hello from Posta",
        HTML:    "<h1>Hello!</h1>",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("sent: id=%s status=%s\n", resp.ID, resp.Status)
}
```

## Credentials

Most machine-facing endpoints take an API key:

```go
client := posta.New("https://posta.example.com", "psk_...")
```

Account-level endpoints (`/users/me/*`) and the platform admin surface accept
only a user session token — an API key is never a valid credential there:

```go
auth, err := posta.New("https://posta.example.com", "").
    Auth.Login("admin@example.com", "password", "")
if err != nil {
    log.Fatal(err)
}
admin := posta.NewWithToken("https://posta.example.com", auth.Token)
```

### Workspaces

Workspace-scoped endpoints resolve the active workspace from the
`X-Posta-Workspace-Id` header. A workspace-bound API key already carries its
workspace; an account-wide key or a user session must name one:

```go
client := posta.New(baseURL, apiKey, posta.WithWorkspace(42))
```

### API key scopes

A key reaches only what its scopes allow. `posta.ScopeSend` covers the public
send API; `ScopeRead` and `ScopeWrite` cover reading and mutating workspace
resources; `ScopeWebhooks` covers webhook management; `ScopeAdmin` covers
tenant administration (keys, members, settings); `ScopeAll` grants everything.

A 403 from an endpoint you expect to work usually means a missing scope —
`posta.IsForbidden(err)` distinguishes it.

## Client options

```go
client := posta.New(baseURL, apiKey,
    posta.WithTimeout(15*time.Second),
    posta.WithWorkspace(42),
    posta.WithUserAgent("my-app/1.0"),
    posta.WithHeader("X-Request-Source", "batch-job"),
    posta.WithRetry(client.RetryPolicy{MaxAttempts: 3, BaseDelay: 200 * time.Millisecond}),
    posta.WithHTTPClient(myHTTPClient),
)
```

Pass a context per call chain with `WithContext`:

```go
emails, err := client.WithContext(ctx).Emails.List(nil)
```

## Services

| Field | Covers |
|---|---|
| `Emails` | send, send-template, batch, preview, verify, status, retry, list, get |
| `Bounces` | list, record |
| `Suppressions` | list, add, remove |
| `Webhooks` | list, create, delete, deliveries, signature verification |
| `Templates` | CRUD, versions, localizations, preview, send-test, import/export |
| `Languages`, `Stylesheets` | CRUD |
| `Domains` | add, list, get, verify, delete |
| `SMTPServers`, `SMTPCredentials` | CRUD, test, revoke |
| `Subscribers` | CRUD, JSON and CSV bulk import |
| `SubscriberLists` | CRUD, members, segments, subscribe/unsubscribe/resubscribe |
| `UnsubscribeLists`, `Contacts` | CRUD / read |
| `Campaigns` | CRUD, send, pause, resume, cancel, duplicate, messages, analytics |
| `Analytics` | email, dashboard, provider breakdown, dashboard stats |
| `Forms` | CRUD, rotate key, embed snippet, nonce, public submit |
| `Messages`, `MessageFilters` | list, triage, reply, attachments; filter CRUD and dry-run |
| `Inbound` | list, get, retry, raw `.eml`, attachments |
| `APIKeys` | create, list, get, revoke, delete |
| `Workspaces` | CRUD, members, invitations, settings, SSO, audit log, export/import, GDPR |
| `Users` | profile, password, 2FA, sessions, settings, notifications (session credential) |
| `Auth` | login, register, password reset, email verification, SSO discovery |
| `Admin` | users, plans, shared servers, domains, settings, announcements, events, metrics |
| `System` | info, health, readiness |

## Examples

### Templated and batch sends

```go
_, err := client.Emails.SendTemplate(&posta.SendTemplateEmailRequest{
    Template:     "welcome",
    To:           []string{"user@example.com"},
    TemplateData: map[string]any{"name": "Ada"},
})

batch, err := client.Emails.SendBatch(&posta.BatchRequest{
    Template: "welcome",
    Recipients: []posta.BatchRecipient{
        {Email: "a@example.com", TemplateData: map[string]any{"name": "Ada"}},
        {Email: "b@example.com", TemplateData: map[string]any{"name": "Grace"}},
    },
})
fmt.Printf("%d sent, %d failed\n", batch.Sent, batch.Failed)
```

Validate without sending:

```go
report, err := client.Emails.SendDryRun(req)
```

### One-click unsubscribe

Reference a Posta-managed unsubscribe list and Posta mints the signed
one-click URL, recording opt-outs against that list alone:

```go
listID := uint(7)
_, err := client.Emails.Send(&posta.SendEmailRequest{
    From:        "news@example.com",
    To:          []string{"user@example.com"},
    Subject:     "This week",
    HTML:        "<p>…</p>",
    Unsubscribe: &posta.Unsubscribe{ListID: &listID},
})
```

### Templates, versions, localizations

```go
tpl, _ := client.Templates.Create(&posta.CreateTemplateRequest{
    Name: "welcome", DefaultLanguage: "en",
})
ver, _ := client.Templates.CreateVersion(tpl.ID, &posta.CreateVersionRequest{})
client.Templates.CreateLocalization(tpl.ID, ver.ID, &posta.CreateLocalizationRequest{
    Language:        "en",
    SubjectTemplate: "Welcome, {{.name}}",
    HTMLTemplate:    "<h1>Welcome, {{.name}}</h1>",
})
client.Templates.ActivateVersion(tpl.ID, ver.ID)
```

### Campaigns

```go
camp, _ := client.Campaigns.Create(&posta.CreateCampaignRequest{
    Name: "Launch", Subject: "We're live", FromEmail: "news@example.com",
    ListID: listID, TemplateID: tpl.ID,
})
client.Campaigns.Send(camp.ID)

stats, _ := client.Campaigns.Analytics(camp.ID)
fmt.Printf("open rate %.1f%%\n", stats.Analytics.OpenRate)
```

### Paging

`ListOptions` is zero-based; a zero `Size` lets the server apply its default.

```go
page, err := client.Emails.List(&posta.EmailListOptions{
    ListOptions: posta.ListOptions{Page: 0, Size: 50},
    Query:       "user@example.com",
    Sort:        "-created_at",
})
fmt.Println(page.Pageable.TotalElements)
```

### Verifying webhooks

Posta signs each delivery with HMAC-SHA256 over the raw body, in the
`X-Posta-Signature` header as `sha256=<hex>`. Verify against the exact bytes
received — re-serializing the JSON changes them:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    if !posta.VerifySignature(body, r.Header.Get(posta.SignatureHeader), secret) {
        http.Error(w, "bad signature", http.StatusUnauthorized)
        return
    }

    var event posta.WebhookEvent
    json.Unmarshal(body, &event)
    switch event.Event {
    case posta.EventEmailSent:
        // …
    case posta.EventEmailFailed:
        // …
    }
    w.WriteHeader(http.StatusOK)
}
```

Event names are constants: `EventEmailSent`, `EventEmailFailed`,
`EventEmailInbound`, `EventEmailUnsubscribed`, `EventEmailComplained`,
`EventCampaignStarted`, `EventCampaignCompleted`, `EventMessageReceived`,
`EventMessageSpam`. Each has a typed payload — `WebhookEvent`,
`CampaignWebhookEvent`, `ComplaintWebhookEvent`, `UnsubscribeWebhookEvent`,
`InboundWebhookEvent`, `MessageWebhookEvent`.

### Web forms

```go
form, _ := client.Forms.Create(&posta.CreateFormRequest{
    Name:           "Contact",
    AllowedOrigins: []string{"https://example.com"},
    StrictOrigin:   true,
    NotifyEmails:   []string{"team@example.com"},
})
snippet, _ := client.Forms.Snippet(form.ID)
fmt.Println(snippet.HTML)

msgs, _ := client.Messages.List(&posta.MessageListOptions{State: "new"})
```

## Errors

Non-2xx responses come back as `*posta.APIError`, carrying the status and the
decoded error envelope:

```go
if _, err := client.Emails.Send(req); err != nil {
    var apiErr *posta.APIError
    if errors.As(err, &apiErr) {
        log.Printf("posta %d: %s", apiErr.StatusCode, apiErr.Info.Message)
    }
}
```

Helpers cover the common cases: `posta.IsNotFound`, `posta.IsUnauthorized`,
`posta.IsForbidden`, `posta.IsRateLimited`.

## License

Apache-2.0
