package posta

// AnalyticsService reads delivery and engagement analytics for the workspace.
type AnalyticsService struct{ c *Client }

// DailyCount is one day's total in a volume series.
type DailyCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// StatusCount is the number of emails in one delivery status.
type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// AnalyticsResponse is the email volume view: a daily series plus the current
// breakdown by delivery status.
type AnalyticsResponse struct {
	DailyCounts     []DailyCount  `json:"daily_counts"`
	StatusBreakdown []StatusCount `json:"status_breakdown"`
}

// Emails returns email volume and status analytics over the requested window.
func (s *AnalyticsService) Emails(opts *AnalyticsOptions) (*AnalyticsResponse, error) {
	return get[AnalyticsResponse](s.c, wsPath+"/analytics", opts.values())
}

// DeliveryRatePoint is one day's delivery outcome.
type DeliveryRatePoint struct {
	Date         string  `json:"date"`
	Total        int64   `json:"total"`
	Sent         int64   `json:"sent"`
	Failed       int64   `json:"failed"`
	DeliveryRate float64 `json:"delivery_rate"`
}

// BounceRatePoint is one day's bounces, split by kind.
type BounceRatePoint struct {
	Date      string `json:"date"`
	Total     int64  `json:"total"`
	Hard      int64  `json:"hard"`
	Soft      int64  `json:"soft"`
	Complaint int64  `json:"complaint"`
}

// DashboardAnalyticsResponse is the trend view behind the dashboard charts.
type DashboardAnalyticsResponse struct {
	DeliveryRateTrends []DeliveryRatePoint `json:"delivery_rate_trends"`
	BounceRateTrends   []BounceRatePoint   `json:"bounce_rate_trends"`
	LatencyPercentiles map[string]float64  `json:"latency_percentiles,omitempty"`
}

// Dashboard returns delivery and bounce trends plus send-latency percentiles.
func (s *AnalyticsService) Dashboard(opts *AnalyticsOptions) (*DashboardAnalyticsResponse, error) {
	return get[DashboardAnalyticsResponse](s.c, wsPath+"/analytics/dashboard", opts.values())
}

// ProviderStats is one receiving provider's deliverability.
type ProviderStats struct {
	Provider     string  `json:"provider"`
	Total        int64   `json:"total"`
	Sent         int64   `json:"sent"`
	Failed       int64   `json:"failed"`
	Suppressed   int64   `json:"suppressed"`
	DeliveryRate float64 `json:"delivery_rate"`
}

// ProviderBreakdownResponse groups deliverability by recipient provider, which
// is how a reputation problem at one mailbox provider shows up.
type ProviderBreakdownResponse struct {
	Providers []ProviderStats `json:"providers"`
}

// Providers returns deliverability broken down by recipient mailbox provider.
func (s *AnalyticsService) Providers(opts *AnalyticsOptions) (*ProviderBreakdownResponse, error) {
	return get[ProviderBreakdownResponse](s.c, wsPath+"/analytics/providers", opts.values())
}

// DailyVolume is one day of send volume.
type DailyVolume struct {
	Date   string `json:"date"`
	Sent   int64  `json:"sent"`
	Failed int64  `json:"failed"`
}

// WorkspaceFeatures reports which optional subsystems this deployment has
// enabled, so a dashboard can hide what is not configured.
type WorkspaceFeatures struct {
	Inbound  bool `json:"inbound"`
	Messages bool `json:"messages"`
	Relay    bool `json:"relay"`
}

// DashboardStats is the workspace's headline counters, as shown on the
// dashboard landing page.
type DashboardStats struct {
	TotalEmails       int64             `json:"total_emails"`
	SentEmails        int64             `json:"sent_emails"`
	FailedEmails      int64             `json:"failed_emails"`
	QueuedEmails      int64             `json:"queued_emails"`
	ProcessingEmails  int64             `json:"processing_emails"`
	SuppressedEmails  int64             `json:"suppressed_emails"`
	FailureRate       float64           `json:"failure_rate"`
	BounceRate        float64           `json:"bounce_rate"`
	TotalBounces      int64             `json:"total_bounces"`
	TotalSuppressions int64             `json:"total_suppressions"`
	TotalTemplates    int64             `json:"total_templates"`
	TotalDomains      int64             `json:"total_domains"`
	UnverifiedDomains int64             `json:"unverified_domains"`
	TotalSMTPServers  int64             `json:"total_smtp_servers"`
	TotalWebhooks     int64             `json:"total_webhooks"`
	TotalAPIKeys      int64             `json:"total_api_keys"`
	ActiveAPIKeys     int64             `json:"active_api_keys"`
	ExpiringAPIKeys   int64             `json:"expiring_api_keys"`
	TotalContacts     int64             `json:"total_contacts"`
	TotalSubscribers  int64             `json:"total_subscribers"`
	TotalCampaigns    int64             `json:"total_campaigns"`
	TotalInbound      int64             `json:"total_inbound"`
	ForwardedInbound  int64             `json:"forwarded_inbound"`
	FailedInbound     int64             `json:"failed_inbound"`
	TotalForms        int64             `json:"total_forms"`
	TotalMessages     int64             `json:"total_messages"`
	UnreadMessages    int64             `json:"unread_messages"`
	SpamMessages      int64             `json:"spam_messages"`
	DailyVolume       []DailyVolume     `json:"daily_volume,omitempty"`
	WebhookDeliveries map[string]int64  `json:"webhook_deliveries,omitempty"`
	Features          WorkspaceFeatures `json:"features"`
}

// DashboardStats returns the workspace's headline counters.
func (s *AnalyticsService) DashboardStats() (*DashboardStats, error) {
	return get[DashboardStats](s.c, wsPath+"/dashboard/stats", nil)
}
