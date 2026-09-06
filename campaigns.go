package posta

import (
	"fmt"
	"time"
)

// CampaignsService manages bulk sends to a subscriber list, and the lifecycle
// that carries one from draft through sending to completion.
type CampaignsService struct{ c *Client }

// CreateCampaignRequest schedules a bulk send. Set ScheduledAt to send later;
// leave it nil and the campaign stays a draft until
// [CampaignsService.Send] is called.
type CreateCampaignRequest struct {
	Name      string `json:"name"`
	Subject   string `json:"subject"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name,omitempty"`
	// ListID names the subscriber list to send to.
	ListID uint `json:"list_id"`
	// TemplateID names the template to render.
	TemplateID uint `json:"template_id"`
	// TemplateVersionID pins a specific version; nil uses the active one.
	TemplateVersionID *uint          `json:"template_version_id,omitempty"`
	TemplateData      map[string]any `json:"template_data,omitempty"`
	Language          string         `json:"language,omitempty"`
	ScheduledAt       *time.Time     `json:"scheduled_at,omitempty"`
	// SendRate caps deliveries per hour, to stay within a provider's limits.
	SendRate int `json:"send_rate,omitempty"`
	// SendAtLocalTime staggers delivery so each subscriber receives the
	// campaign at the scheduled hour in their own timezone.
	SendAtLocalTime bool            `json:"send_at_local_time,omitempty"`
	ABTestEnabled   bool            `json:"ab_test_enabled,omitempty"`
	ABTestVariants  []ABTestVariant `json:"ab_test_variants,omitempty"`
}

// UpdateCampaignRequest changes a draft or scheduled campaign. Nil fields are
// left unchanged; a campaign that has started cannot be edited.
type UpdateCampaignRequest struct {
	Name              string          `json:"name,omitempty"`
	Subject           string          `json:"subject,omitempty"`
	FromEmail         string          `json:"from_email,omitempty"`
	FromName          string          `json:"from_name,omitempty"`
	ListID            *uint           `json:"list_id,omitempty"`
	TemplateID        *uint           `json:"template_id,omitempty"`
	TemplateVersionID *uint           `json:"template_version_id,omitempty"`
	TemplateData      map[string]any  `json:"template_data,omitempty"`
	Language          string          `json:"language,omitempty"`
	ScheduledAt       *time.Time      `json:"scheduled_at,omitempty"`
	SendRate          *int            `json:"send_rate,omitempty"`
	SendAtLocalTime   *bool           `json:"send_at_local_time,omitempty"`
	ABTestEnabled     *bool           `json:"ab_test_enabled,omitempty"`
	ABTestVariants    []ABTestVariant `json:"ab_test_variants,omitempty"`
}

// Create adds a campaign.
func (s *CampaignsService) Create(req *CreateCampaignRequest) (*Campaign, error) {
	return post[Campaign](s.c, wsPath+"/campaigns", req, nil)
}

// List returns a page of campaigns with their delivery counters.
func (s *CampaignsService) List(opts *CampaignListOptions) (*PageableResponse[CampaignWithStats], error) {
	return callPage[CampaignWithStats](s.c, wsPath+"/campaigns", opts.values())
}

// Get returns one campaign with its delivery counters.
func (s *CampaignsService) Get(id uint) (*CampaignWithStats, error) {
	return get[CampaignWithStats](s.c, fmt.Sprintf("%s/campaigns/%d", wsPath, id), nil)
}

// Update changes a campaign that has not started sending.
func (s *CampaignsService) Update(id uint, req *UpdateCampaignRequest) (*Campaign, error) {
	return put[Campaign](s.c, fmt.Sprintf("%s/campaigns/%d", wsPath, id), req)
}

// Delete removes a campaign.
func (s *CampaignsService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/campaigns/%d", wsPath, id))
}

// Send starts a campaign immediately, ignoring any schedule on it.
func (s *CampaignsService) Send(id uint) (*Campaign, error) {
	return post[Campaign](s.c, fmt.Sprintf("%s/campaigns/%d/send", wsPath, id), nil, nil)
}

// Pause halts a sending campaign. Recipients already sent to are not resent.
func (s *CampaignsService) Pause(id uint) (*Campaign, error) {
	return post[Campaign](s.c, fmt.Sprintf("%s/campaigns/%d/pause", wsPath, id), nil, nil)
}

// Resume continues a paused campaign from where it stopped.
func (s *CampaignsService) Resume(id uint) (*Campaign, error) {
	return post[Campaign](s.c, fmt.Sprintf("%s/campaigns/%d/resume", wsPath, id), nil, nil)
}

// Cancel stops a campaign for good. It cannot be resumed afterwards.
func (s *CampaignsService) Cancel(id uint) (*Campaign, error) {
	return post[Campaign](s.c, fmt.Sprintf("%s/campaigns/%d/cancel", wsPath, id), nil, nil)
}

// Duplicate copies a campaign into a fresh draft, so a recurring send can be
// repeated without rebuilding it.
func (s *CampaignsService) Duplicate(id uint) (*Campaign, error) {
	return post[Campaign](s.c, fmt.Sprintf("%s/campaigns/%d/duplicate", wsPath, id), nil, nil)
}

// ListMessages returns a page of per-subscriber sends for a campaign, with the
// engagement timestamps recorded against each.
func (s *CampaignsService) ListMessages(id uint, opts *ListOptions) (*PageableResponse[CampaignMessage], error) {
	return callPage[CampaignMessage](s.c, fmt.Sprintf("%s/campaigns/%d/messages", wsPath, id), opts.values())
}

// CampaignAnalytics holds the headline rates and counts for a campaign.
type CampaignAnalytics struct {
	TotalMessages   int64   `json:"total_messages"`
	SentMessages    int64   `json:"sent_messages"`
	FailedMessages  int64   `json:"failed_messages"`
	OpenedMessages  int64   `json:"opened_messages"`
	ClickedMessages int64   `json:"clicked_messages"`
	BouncedMessages int64   `json:"bounced_messages"`
	Unsubscribed    int64   `json:"unsubscribed"`
	DeliveryRate    float64 `json:"delivery_rate"`
	OpenRate        float64 `json:"open_rate"`
	ClickRate       float64 `json:"click_rate"`
	BounceRate      float64 `json:"bounce_rate"`
	UnsubscribeRate float64 `json:"unsubscribe_rate"`
}

// TimeSeriesPoint is one bucket of a time series.
type TimeSeriesPoint struct {
	Time  string `json:"time"`
	Count int64  `json:"count"`
}

// CampaignLink is a tracked link in a campaign and how often it was clicked.
type CampaignLink struct {
	ID          uint      `json:"id"`
	CampaignID  uint      `json:"campaign_id"`
	Hash        string    `json:"hash"`
	OriginalURL string    `json:"original_url"`
	ClickCount  int64     `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// CampaignAnalyticsResponse is the full analytics view of a campaign: headline
// rates, open and click series over time, per-link click counts, and — for an
// A/B test — the same figures per variant.
type CampaignAnalyticsResponse struct {
	Analytics        *CampaignAnalytics           `json:"analytics,omitempty"`
	OpenSeries       []TimeSeriesPoint            `json:"open_series,omitempty"`
	ClickSeries      []TimeSeriesPoint            `json:"click_series,omitempty"`
	Links            []CampaignLink               `json:"links,omitempty"`
	VariantAnalytics map[string]CampaignAnalytics `json:"variant_analytics,omitempty"`
}

// Analytics returns the engagement analytics for a campaign.
func (s *CampaignsService) Analytics(id uint) (*CampaignAnalyticsResponse, error) {
	return get[CampaignAnalyticsResponse](s.c, fmt.Sprintf("%s/campaigns/%d/analytics", wsPath, id), nil)
}
