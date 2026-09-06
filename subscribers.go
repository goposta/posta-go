package posta

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
)

// SubscribersService manages the people a workspace sends campaigns to.
type SubscribersService struct{ c *Client }

// CreateSubscriberRequest adds a subscriber. Status defaults to "subscribed".
type CreateSubscriberRequest struct {
	Email        string         `json:"email"`
	Name         string         `json:"name,omitempty"`
	Status       string         `json:"status,omitempty"`
	Language     string         `json:"language,omitempty"`
	Timezone     string         `json:"timezone,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
}

// UpdateSubscriberRequest changes a subscriber. The email address itself
// cannot be changed; delete and re-add instead.
type UpdateSubscriberRequest struct {
	Name         string         `json:"name,omitempty"`
	Status       string         `json:"status,omitempty"`
	Language     string         `json:"language,omitempty"`
	Timezone     string         `json:"timezone,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
}

// Create adds a subscriber.
func (s *SubscribersService) Create(req *CreateSubscriberRequest) (*Subscriber, error) {
	return post[Subscriber](s.c, wsPath+"/subscribers", req, nil)
}

// List returns a page of subscribers.
func (s *SubscribersService) List(opts *SubscriberListOptions) (*PageableResponse[Subscriber], error) {
	return callPage[Subscriber](s.c, wsPath+"/subscribers", opts.values())
}

// Get returns one subscriber.
func (s *SubscribersService) Get(id uint) (*Subscriber, error) {
	return get[Subscriber](s.c, fmt.Sprintf("%s/subscribers/%d", wsPath, id), nil)
}

// Update changes a subscriber.
func (s *SubscribersService) Update(id uint, req *UpdateSubscriberRequest) (*Subscriber, error) {
	return put[Subscriber](s.c, fmt.Sprintf("%s/subscribers/%d", wsPath, id), req)
}

// Delete removes a subscriber and its list memberships.
func (s *SubscribersService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/subscribers/%d", wsPath, id))
}

// BulkImportResult reports what a bulk import did. Errors names the rows that
// were rejected and why.
type BulkImportResult struct {
	Total    int      `json:"total"`
	Imported int      `json:"imported"`
	Updated  int      `json:"updated"`
	Skipped  int      `json:"skipped"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}

// ImportJSON adds or updates many subscribers at once. Existing addresses are
// updated rather than duplicated.
func (s *SubscribersService) ImportJSON(subscribers []CreateSubscriberRequest) (*BulkImportResult, error) {
	body := struct {
		Subscribers []CreateSubscriberRequest `json:"subscribers"`
	}{Subscribers: subscribers}
	return post[BulkImportResult](s.c, wsPath+"/subscribers/import/json", body, nil)
}

// ImportCSV adds or updates many subscribers from a CSV document, uploaded as
// multipart/form-data under the "file" field.
//
// columnMapping maps zero-based column indexes to subscriber fields and
// defaults to {0: "email", 1: "name"} when nil. A "custom_fields." prefix
// routes a column into the subscriber's custom fields, as in
// {0: "email", 1: "name", 2: "custom_fields.company"}. The header row is
// always skipped.
func (s *SubscribersService) ImportCSV(filename string, csvData []byte, columnMapping map[int]string) (*BulkImportResult, error) {
	rb := s.c.request(http.MethodPost, wsPath+"/subscribers/import/csv", nil).
		Multipart(func(w *multipart.Writer) error {
			part, err := w.CreateFormFile("file", filename)
			if err != nil {
				return err
			}
			if _, err := part.Write(csvData); err != nil {
				return err
			}
			if len(columnMapping) == 0 {
				return nil
			}
			// The server reads the mapping as JSON with string keys.
			raw := make(map[string]string, len(columnMapping))
			for idx, field := range columnMapping {
				raw[strconv.Itoa(idx)] = field
			}
			encoded, err := json.Marshal(raw)
			if err != nil {
				return err
			}
			return w.WriteField("column_mapping", string(encoded))
		})
	resp, err := s.c.do(rb)
	if err != nil {
		return nil, err
	}
	var envelope Response[BulkImportResult]
	if err := resp.JSON(&envelope); err != nil {
		return nil, fmt.Errorf("posta: decode response: %w", err)
	}
	return &envelope.Data, nil
}

// SubscriberListsService manages lists and who belongs to them.
//
// A "static" list has explicit members. A "segment" list derives its members
// from FilterRules evaluated at send time.
type SubscriberListsService struct{ c *Client }

// CreateSubscriberListRequest creates a list. Type is "static" (the default)
// or "segment", in which case FilterRules defines membership.
type CreateSubscriberListRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Type        string       `json:"type,omitempty"`
	FilterRules []FilterRule `json:"filter_rules,omitempty"`
}

// UpdateSubscriberListRequest changes a list. The type cannot be changed after
// creation.
type UpdateSubscriberListRequest struct {
	Name        string       `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	FilterRules []FilterRule `json:"filter_rules,omitempty"`
}

// Create adds a list.
func (s *SubscriberListsService) Create(req *CreateSubscriberListRequest) (*SubscriberList, error) {
	return post[SubscriberList](s.c, wsPath+"/subscriber-lists", req, nil)
}

// List returns a page of lists with their member counts.
func (s *SubscriberListsService) List(opts *QueryListOptions) (*PageableResponse[SubscriberListWithCount], error) {
	return callPage[SubscriberListWithCount](s.c, wsPath+"/subscriber-lists", opts.values())
}

// Get returns one list with its member count.
func (s *SubscriberListsService) Get(id uint) (*SubscriberListWithCount, error) {
	return get[SubscriberListWithCount](s.c, fmt.Sprintf("%s/subscriber-lists/%d", wsPath, id), nil)
}

// Update changes a list.
func (s *SubscriberListsService) Update(id uint, req *UpdateSubscriberListRequest) (*SubscriberList, error) {
	return put[SubscriberList](s.c, fmt.Sprintf("%s/subscriber-lists/%d", wsPath, id), req)
}

// Delete removes a list. The subscribers themselves are not deleted.
func (s *SubscriberListsService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/subscriber-lists/%d", wsPath, id))
}

// ListMembers returns a page of the subscribers on a list. For a segment list
// this evaluates the filter rules.
func (s *SubscriberListsService) ListMembers(id uint, opts *ListOptions) (*PageableResponse[Subscriber], error) {
	return callPage[Subscriber](s.c, fmt.Sprintf("%s/subscriber-lists/%d/members", wsPath, id), opts.values())
}

// AddMember puts an existing subscriber on a static list.
func (s *SubscriberListsService) AddMember(listID, subscriberID uint) error {
	body := struct {
		SubscriberID uint `json:"subscriber_id"`
	}{SubscriberID: subscriberID}
	return s.c.callNoContent(http.MethodPost, fmt.Sprintf("%s/subscriber-lists/%d/members", wsPath, listID), body, nil)
}

// RemoveMember takes a subscriber off a static list. This is not an opt-out:
// use [SubscriberListsService.Unsubscribe] to record one.
func (s *SubscriberListsService) RemoveMember(listID, subscriberID uint) error {
	body := struct {
		SubscriberID uint `json:"subscriber_id"`
	}{SubscriberID: subscriberID}
	return s.c.callNoContent(http.MethodDelete, fmt.Sprintf("%s/subscriber-lists/%d/members", wsPath, listID), body, nil)
}

// SegmentPreview reports how many subscribers a set of filter rules would
// select, so a segment can be checked before it is saved.
type SegmentPreview struct {
	Count int64 `json:"count"`
}

// PreviewSegment counts the subscribers a candidate segment would match.
func (s *SubscriberListsService) PreviewSegment(rules []FilterRule) (*SegmentPreview, error) {
	body := struct {
		FilterRules []FilterRule `json:"filter_rules"`
	}{FilterRules: rules}
	return post[SegmentPreview](s.c, wsPath+"/subscriber-lists/preview-segment", body, nil)
}

// Subscribe adds an address to a list by name, creating the list on first use
// and clearing any prior opt-out for it. Idempotent, and reachable with a
// `send`-scoped API key, so a signup form can call it directly.
func (s *SubscriberListsService) Subscribe(req *ListSubscribeRequest) (*ListSubscribeResponse, error) {
	return post[ListSubscribeResponse](s.c, "/subscriber-lists/subscribe", req, nil)
}

// Unsubscribe opts an address out of one list. The subscriber's global status
// is untouched. Idempotent.
func (s *SubscriberListsService) Unsubscribe(listID uint, req *ListUnsubscribeRequest) (*ListSubscribeResponse, error) {
	return post[ListSubscribeResponse](s.c, fmt.Sprintf("/subscriber-lists/%d/unsubscribe", listID), req, nil)
}

// Resubscribe reverses a list-scoped opt-out and, for a static list, puts the
// subscriber back on it. Idempotent.
func (s *SubscriberListsService) Resubscribe(listID uint, email string) (*ListSubscribeResponse, error) {
	body := struct {
		Email string `json:"email"`
	}{Email: email}
	return post[ListSubscribeResponse](s.c, fmt.Sprintf("/subscriber-lists/%d/resubscribe", listID), body, nil)
}

// UnsubscribeInWorkspace opts an address out of a list through the
// workspace-scoped endpoint, which a session credential can also reach.
func (s *SubscriberListsService) UnsubscribeInWorkspace(listID uint, req *ListUnsubscribeRequest) (*ListSubscribeResponse, error) {
	return post[ListSubscribeResponse](s.c, fmt.Sprintf("%s/subscriber-lists/%d/unsubscribe", wsPath, listID), req, nil)
}

// ResubscribeInWorkspace reverses an opt-out through the workspace-scoped
// endpoint.
func (s *SubscriberListsService) ResubscribeInWorkspace(listID uint, email string) (*ListSubscribeResponse, error) {
	body := struct {
		Email string `json:"email"`
	}{Email: email}
	return post[ListSubscribeResponse](s.c, fmt.Sprintf("%s/subscriber-lists/%d/resubscribe", wsPath, listID), body, nil)
}

// UnsubscribeListsService manages the named opt-out lists that
// List-Unsubscribe headers point at. Referencing one from a send lets Posta
// mint the signed one-click URL and record the opt-out against that list alone.
type UnsubscribeListsService struct{ c *Client }

// UnsubscribeListRequest creates or changes an unsubscribe list. PublicName is
// what a recipient sees on the opt-out page.
type UnsubscribeListRequest struct {
	Name        string `json:"name"`
	PublicName  string `json:"public_name,omitempty"`
	Description string `json:"description,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

// Create adds an unsubscribe list.
func (s *UnsubscribeListsService) Create(req *UnsubscribeListRequest) (*UnsubscribeList, error) {
	return post[UnsubscribeList](s.c, wsPath+"/unsubscribe-lists", req, nil)
}

// List returns a page of unsubscribe lists.
func (s *UnsubscribeListsService) List(opts *QueryListOptions) (*PageableResponse[UnsubscribeList], error) {
	return callPage[UnsubscribeList](s.c, wsPath+"/unsubscribe-lists", opts.values())
}

// Get returns one unsubscribe list.
func (s *UnsubscribeListsService) Get(id uint) (*UnsubscribeList, error) {
	return get[UnsubscribeList](s.c, fmt.Sprintf("%s/unsubscribe-lists/%d", wsPath, id), nil)
}

// Update changes an unsubscribe list.
func (s *UnsubscribeListsService) Update(id uint, req *UnsubscribeListRequest) (*UnsubscribeList, error) {
	return put[UnsubscribeList](s.c, fmt.Sprintf("%s/unsubscribe-lists/%d", wsPath, id), req)
}

// Delete removes an unsubscribe list.
func (s *UnsubscribeListsService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/unsubscribe-lists/%d", wsPath, id))
}

// ContactsService reads the derived record of every address the workspace has
// mailed, with delivery counters. Contacts are created by sending; they are
// not managed directly.
type ContactsService struct{ c *Client }

// List returns a page of contacts.
func (s *ContactsService) List(opts *SearchListOptions) (*PageableResponse[Contact], error) {
	return callPage[Contact](s.c, wsPath+"/contacts", opts.values())
}

// Get returns one contact.
func (s *ContactsService) Get(id uint) (*Contact, error) {
	return get[Contact](s.c, fmt.Sprintf("%s/contacts/%d", wsPath, id), nil)
}
