package posta

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// wsPath prefixes every workspace-scoped endpoint. The active workspace comes
// from the credential or the X-Posta-Workspace-Id header, never the path.
const wsPath = "/workspaces/current"

// ListOptions is the pagination shared by every list endpoint. Page is
// zero-based; a zero Size lets the server apply its default (20).
type ListOptions struct {
	Page int
	Size int
}

func (o *ListOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	if o.Page > 0 {
		v.Set("page", strconv.Itoa(o.Page))
	}
	if o.Size > 0 {
		v.Set("size", strconv.Itoa(o.Size))
	}
	return v
}

func setStr(v url.Values, key, value string) {
	if value != "" {
		v.Set(key, value)
	}
}

func setUint(v url.Values, key string, value *uint) {
	if value != nil {
		v.Set(key, strconv.FormatUint(uint64(*value), 10))
	}
}

func setBool(v url.Values, key string, value *bool) {
	if value != nil {
		v.Set(key, strconv.FormatBool(*value))
	}
}

func setTime(v url.Values, key string, value *time.Time) {
	if value != nil {
		v.Set(key, value.UTC().Format(time.RFC3339))
	}
}

// EmailListOptions filters a list of emails. Query is a free-text search over
// recipient and subject; Sort names a column with an optional "-" prefix for
// descending order.
type EmailListOptions struct {
	ListOptions
	Query string
	Sort  string
}

func (o *EmailListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "q", o.Query)
	setStr(v, "sort", o.Sort)
	return v
}

// SearchListOptions filters lists that accept a single free-text `search`
// parameter, such as templates and contacts.
type SearchListOptions struct {
	ListOptions
	Search string
}

func (o *SearchListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "search", o.Search)
	return v
}

// QueryListOptions filters lists that accept a free-text `q` parameter and an
// optional sort, such as unsubscribe lists and subscriber lists.
type QueryListOptions struct {
	ListOptions
	Query string
	Sort  string
}

func (o *QueryListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "q", o.Query)
	setStr(v, "sort", o.Sort)
	return v
}

// SubscriberListOptions filters a list of subscribers.
type SubscriberListOptions struct {
	ListOptions
	Search string
	Status string
}

func (o *SubscriberListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "search", o.Search)
	setStr(v, "status", o.Status)
	return v
}

// CampaignListOptions filters a list of campaigns by status.
type CampaignListOptions struct {
	ListOptions
	Status string
}

func (o *CampaignListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "status", o.Status)
	return v
}

// SuppressionListOptions filters suppressions to a single unsubscribe list.
type SuppressionListOptions struct {
	ListOptions
	ListID string
}

func (o *SuppressionListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "list_id", o.ListID)
	return v
}

// MessageListOptions filters web form submissions.
type MessageListOptions struct {
	ListOptions
	// FormID restricts the result to one form.
	FormID *uint
	// Status filters on the spam verdict: "received", "flagged", "spam",
	// "rejected".
	Status string
	// State filters on the triage state: "new", "open", "closed".
	State string
	// Unread, when set, keeps only read or only unread messages.
	Unread *bool
	// Query is a free-text search over sender, subject, and body.
	Query string
	// After and Before bound the result by submission time.
	After  *time.Time
	Before *time.Time
}

func (o *MessageListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setUint(v, "form_id", o.FormID)
	setStr(v, "status", o.Status)
	setStr(v, "state", o.State)
	setBool(v, "unread", o.Unread)
	setStr(v, "q", o.Query)
	setTime(v, "after", o.After)
	setTime(v, "before", o.Before)
	return v
}

// InboundListOptions filters received inbound email.
type InboundListOptions struct {
	ListOptions
	// Status filters on processing state, such as "received" or "failed".
	Status string
	// Source filters on how the message arrived: "smtp" or "webhook".
	Source string
	// Sender filters on the envelope sender.
	Sender string
	// Query is a free-text search over subject and body.
	Query string
}

func (o *InboundListOptions) values() url.Values {
	if o == nil {
		return url.Values{}
	}
	v := o.ListOptions.values()
	setStr(v, "status", o.Status)
	setStr(v, "source", o.Source)
	setStr(v, "sender", o.Sender)
	setStr(v, "q", o.Query)
	return v
}

// AnalyticsOptions bounds an analytics query by date. From and To are
// inclusive and default to the server's own window when omitted.
type AnalyticsOptions struct {
	From *time.Time
	To   *time.Time
	// Status narrows delivery analytics to one email status.
	Status string
}

func (o *AnalyticsOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	setTime(v, "from", o.From)
	setTime(v, "to", o.To)
	setStr(v, "status", o.Status)
	return v
}

// sprintfPath formats a request path. It exists so path building reads the
// same in every service without each importing fmt for a single call.
func sprintfPath(format string, args ...any) string { return fmt.Sprintf(format, args...) }
