package posta

import "time"

// Template is a named, versioned email template. The rendered content lives on
// its versions; the template itself holds the metadata and points at the
// version that sends.
type Template struct {
	ID              uint             `json:"id"`
	UserID          uint             `json:"user_id"`
	WorkspaceID     *uint            `json:"workspace_id,omitempty"`
	Name            string           `json:"name"`
	Description     string           `json:"description,omitempty"`
	DefaultLanguage string           `json:"default_language,omitempty"`
	SampleData      string           `json:"sample_data,omitempty"`
	ActiveVersionID *uint            `json:"active_version_id,omitempty"`
	ActiveVersion   *TemplateVersion `json:"active_version,omitempty"`
	LastEditedByID  *uint            `json:"last_edited_by_id,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       *time.Time       `json:"updated_at,omitempty"`
}

// TemplateListItem is the shape returned when listing templates: the metadata
// and the active version, without the other versions' bodies.
type TemplateListItem struct {
	ID              uint             `json:"id"`
	UserID          uint             `json:"user_id"`
	WorkspaceID     *uint            `json:"workspace_id,omitempty"`
	Name            string           `json:"name"`
	Description     string           `json:"description,omitempty"`
	DefaultLanguage string           `json:"default_language,omitempty"`
	SampleData      string           `json:"sample_data,omitempty"`
	ActiveVersionID *uint            `json:"active_version_id,omitempty"`
	ActiveVersion   *TemplateVersion `json:"active_version,omitempty"`
	// Languages lists the language codes the active version is localized
	// into.
	Languages      []string   `json:"languages,omitempty"`
	LastEditedByID *uint      `json:"last_edited_by_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

// TemplateVersion is one immutable revision of a template. Sending uses the
// template's active version unless a specific one is named.
type TemplateVersion struct {
	ID            uint                   `json:"id"`
	TemplateID    uint                   `json:"template_id"`
	Version       int                    `json:"version"`
	SampleData    string                 `json:"sample_data,omitempty"`
	StylesheetID  *uint                  `json:"stylesheet_id,omitempty"`
	Stylesheet    *Stylesheet            `json:"stylesheet,omitempty"`
	Localizations []TemplateLocalization `json:"localizations,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// TemplateLocalization holds the subject and body for one language of a
// template version.
type TemplateLocalization struct {
	ID              uint       `json:"id"`
	VersionID       uint       `json:"version_id"`
	Language        string     `json:"language"`
	SubjectTemplate string     `json:"subject_template"`
	HTMLTemplate    string     `json:"html_template,omitempty"`
	TextTemplate    string     `json:"text_template,omitempty"`
	BuilderJSON     string     `json:"builder_json,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

// Language is a language code available to template localizations.
type Language struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	WorkspaceID *uint     `json:"workspace_id,omitempty"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
}

// Stylesheet is reusable CSS that template versions can share.
type Stylesheet struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	WorkspaceID *uint      `json:"workspace_id,omitempty"`
	Name        string     `json:"name"`
	CSS         string     `json:"css"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// Domain is a sending domain and the state of its DNS verification. Posta will
// only send from a domain whose ownership has been verified when the workspace
// requires it.
type Domain struct {
	ID                uint      `json:"id"`
	UserID            uint      `json:"user_id"`
	WorkspaceID       *uint     `json:"workspace_id,omitempty"`
	Domain            string    `json:"domain"`
	VerificationToken string    `json:"verification_token,omitempty"`
	OwnershipVerified bool      `json:"ownership_verified"`
	SPFVerified       bool      `json:"spf_verified"`
	DKIMVerified      bool      `json:"dkim_verified"`
	DMARCVerified     bool      `json:"dmarc_verified"`
	CreatedAt         time.Time `json:"created_at"`
}

// DNSRecord is one record a domain needs published.
type DNSRecord struct {
	Type  string `json:"type"`
	Host  string `json:"host"`
	Value string `json:"value"`
}

// DNSRecords are the four records a sending domain needs: a TXT proving
// ownership, and the SPF, DKIM, and DMARC records that make mail from it
// deliverable.
type DNSRecords struct {
	Verification DNSRecord `json:"verification"`
	SPF          DNSRecord `json:"spf"`
	DKIM         DNSRecord `json:"dkim"`
	DMARC        DNSRecord `json:"dmarc"`
}

// DomainWithRecords is a domain together with the DNS records to publish for
// it, as returned when adding or fetching one.
type DomainWithRecords struct {
	Domain
	DNSRecords DNSRecords `json:"dns_records"`
}

// DomainVerification is the outcome of one round of DNS checks. The Record
// fields carry what was actually found, which is what makes a failed check
// diagnosable.
type DomainVerification struct {
	OwnershipVerified bool   `json:"ownership_verified"`
	SPFVerified       bool   `json:"spf_verified"`
	DKIMVerified      bool   `json:"dkim_verified"`
	DMARCVerified     bool   `json:"dmarc_verified"`
	SPFRecord         string `json:"spf_record,omitempty"`
	DKIMRecord        string `json:"dkim_record,omitempty"`
	DMARCRecord       string `json:"dmarc_record,omitempty"`
}

// DomainVerificationResult is what a verification run returns: the updated
// domain, whether every check now passes, and the detail of each check.
type DomainVerificationResult struct {
	Domain        Domain             `json:"domain"`
	FullyVerified bool               `json:"fully_verified"`
	Verification  DomainVerification `json:"verification"`
}

// SMTPServer is an SMTP relay Posta delivers through. Passwords are never
// returned by the API.
type SMTPServer struct {
	ID              uint       `json:"id"`
	UserID          uint       `json:"user_id"`
	WorkspaceID     *uint      `json:"workspace_id,omitempty"`
	Name            string     `json:"name"`
	Host            string     `json:"host"`
	Port            int        `json:"port"`
	Username        string     `json:"username,omitempty"`
	Encryption      string     `json:"encryption,omitempty"`
	AllowedEmails   []string   `json:"allowed_emails,omitempty"`
	MaxRetries      int        `json:"max_retries"`
	Status          string     `json:"status,omitempty"`
	IsSystem        bool       `json:"is_system"`
	ValidatedAt     *time.Time `json:"validated_at,omitempty"`
	ValidationError string     `json:"validation_error,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// SMTPCredential authenticates a client against Posta's own SMTP relay
// listener. The password is returned once, at creation.
type SMTPCredential struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	WorkspaceID uint       `json:"workspace_id"`
	Name        string     `json:"name"`
	Username    string     `json:"username"`
	AllowedIPs  []string   `json:"allowed_ips,omitempty"`
	Revoked     bool       `json:"revoked"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Subscriber is a mailing-list recipient. Status is one of "subscribed",
// "unsubscribed", "bounced", "complained".
type Subscriber struct {
	ID             uint           `json:"id"`
	UserID         uint           `json:"user_id"`
	WorkspaceID    *uint          `json:"workspace_id,omitempty"`
	Email          string         `json:"email"`
	Name           string         `json:"name,omitempty"`
	Status         string         `json:"status"`
	Language       string         `json:"language,omitempty"`
	Timezone       string         `json:"timezone,omitempty"`
	CustomFields   map[string]any `json:"custom_fields,omitempty"`
	SubscribedAt   *time.Time     `json:"subscribed_at,omitempty"`
	UnsubscribedAt *time.Time     `json:"unsubscribed_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      *time.Time     `json:"updated_at,omitempty"`
}

// FilterRule is one clause of a segment list's membership query.
type FilterRule struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value,omitempty"`
}

// SubscriberList groups subscribers. Type is "static" (explicit membership) or
// "segment" (membership derived from FilterRules).
type SubscriberList struct {
	ID          uint         `json:"id"`
	UserID      uint         `json:"user_id"`
	WorkspaceID *uint        `json:"workspace_id,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Type        string       `json:"type"`
	FilterRules []FilterRule `json:"filter_rules,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   *time.Time   `json:"updated_at,omitempty"`
}

// SubscriberListWithCount is a list together with its current member count.
type SubscriberListWithCount struct {
	SubscriberList
	SubscriberCount int64 `json:"subscriber_count"`
}

// UnsubscribeList is a named opt-out list referenced by List-Unsubscribe
// headers. A recipient who unsubscribes is suppressed on this list alone.
type UnsubscribeList struct {
	ID          uint       `json:"id"`
	UUID        string     `json:"uuid"`
	UserID      uint       `json:"user_id"`
	WorkspaceID *uint      `json:"workspace_id,omitempty"`
	Name        string     `json:"name"`
	PublicName  string     `json:"public_name,omitempty"`
	Description string     `json:"description,omitempty"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// Suppression blocks delivery to an address. Kind is one of "bounce",
// "complaint", "unsubscribe", "manual".
type Suppression struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	WorkspaceID *uint     `json:"workspace_id,omitempty"`
	Email       string    `json:"email"`
	Kind        string    `json:"kind"`
	Reason      string    `json:"reason,omitempty"`
	ListID      *uint     `json:"list_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Contact is the derived record of an address Posta has mailed, with its
// delivery counters.
type Contact struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	WorkspaceID *uint      `json:"workspace_id,omitempty"`
	Email       string     `json:"email"`
	Name        string     `json:"name,omitempty"`
	SentCount   int64      `json:"sent_count"`
	FailCount   int64      `json:"fail_count"`
	LastSentAt  *time.Time `json:"last_sent_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ABTestVariant is one arm of a campaign A/B test.
type ABTestVariant struct {
	Name            string `json:"name"`
	Subject         string `json:"subject,omitempty"`
	TemplateID      *uint  `json:"template_id,omitempty"`
	SplitPercentage int    `json:"split_percentage,omitempty"`
}

// Campaign is a bulk send to a subscriber list. Status is one of "draft",
// "scheduled", "sending", "paused", "completed", "cancelled", "failed".
type Campaign struct {
	ID                uint            `json:"id"`
	UserID            uint            `json:"user_id"`
	WorkspaceID       *uint           `json:"workspace_id,omitempty"`
	Name              string          `json:"name"`
	Subject           string          `json:"subject,omitempty"`
	FromEmail         string          `json:"from_email,omitempty"`
	FromName          string          `json:"from_name,omitempty"`
	ListID            uint            `json:"list_id"`
	TemplateID        uint            `json:"template_id"`
	TemplateVersionID *uint           `json:"template_version_id,omitempty"`
	TemplateData      map[string]any  `json:"template_data,omitempty"`
	Language          string          `json:"language,omitempty"`
	Status            string          `json:"status"`
	SendRate          int             `json:"send_rate,omitempty"`
	SendAtLocalTime   bool            `json:"send_at_local_time,omitempty"`
	ABTestEnabled     bool            `json:"ab_test_enabled,omitempty"`
	ABTestVariants    []ABTestVariant `json:"ab_test_variants,omitempty"`
	ABTestWinner      string          `json:"ab_test_winner,omitempty"`
	ScheduledAt       *time.Time      `json:"scheduled_at,omitempty"`
	StartedAt         *time.Time      `json:"started_at,omitempty"`
	CompletedAt       *time.Time      `json:"completed_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         *time.Time      `json:"updated_at,omitempty"`
}

// CampaignWithStats is a campaign together with its delivery counters.
type CampaignWithStats struct {
	Campaign
	TotalRecipients  int64 `json:"total_recipients"`
	SentCount        int64 `json:"sent_count"`
	FailedCount      int64 `json:"failed_count"`
	OpenedCount      int64 `json:"opened_count"`
	ClickedCount     int64 `json:"clicked_count"`
	BouncedCount     int64 `json:"bounced_count"`
	UnsubscribeCount int64 `json:"unsubscribed_count"`
}

// CampaignMessage is one campaign send to one subscriber, with the engagement
// timestamps recorded for it.
type CampaignMessage struct {
	ID             uint       `json:"id"`
	CampaignID     uint       `json:"campaign_id"`
	SubscriberID   uint       `json:"subscriber_id"`
	EmailID        *uint      `json:"email_id,omitempty"`
	Status         string     `json:"status"`
	Variant        string     `json:"variant,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	OpenedAt       *time.Time `json:"opened_at,omitempty"`
	ClickedAt      *time.Time `json:"clicked_at,omitempty"`
	BouncedAt      *time.Time `json:"bounced_at,omitempty"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// APIKey is a machine credential. The secret itself is returned only once, by
// [APIKeysService.Create]; afterwards only KeyPrefix identifies it.
type APIKey struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	WorkspaceID *uint      `json:"workspace_id,omitempty"`
	Name        string     `json:"name"`
	KeyPrefix   string     `json:"key_prefix"`
	Scopes      []string   `json:"scopes,omitempty"`
	AllowedIPs  []string   `json:"allowed_ips,omitempty"`
	Revoked     bool       `json:"revoked"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Form is a web form endpoint that accepts public submissions and turns them
// into messages.
type Form struct {
	ID                  uint       `json:"id"`
	UUID                string     `json:"uuid"`
	WorkspaceID         *uint      `json:"workspace_id,omitempty"`
	Name                string     `json:"name"`
	Slug                string     `json:"slug,omitempty"`
	Description         string     `json:"description,omitempty"`
	PublicKey           string     `json:"public_key"`
	Status              string     `json:"status"`
	AllowedOrigins      []string   `json:"allowed_origins,omitempty"`
	StrictOrigin        bool       `json:"strict_origin"`
	HoneypotField       string     `json:"honeypot_field,omitempty"`
	RequireNonce        bool       `json:"require_nonce"`
	MinFillSeconds      int        `json:"min_fill_seconds"`
	MaxFields           int        `json:"max_fields"`
	MaxBodyBytes        int64      `json:"max_body_bytes"`
	AllowAttachments    bool       `json:"allow_attachments"`
	RedirectURL         string     `json:"redirect_url,omitempty"`
	ScanEnabled         bool       `json:"scan_enabled"`
	FlagThreshold       float64    `json:"flag_threshold"`
	QuarantineThreshold float64    `json:"quarantine_threshold"`
	RejectThreshold     float64    `json:"reject_threshold"`
	NotifyEnabled       bool       `json:"notify_enabled"`
	NotifyEmails        []string   `json:"notify_emails,omitempty"`
	NotifyMode          string     `json:"notify_mode,omitempty"`
	NotifyOnFlagged     bool       `json:"notify_on_flagged"`
	ReplyFrom           string     `json:"reply_from,omitempty"`
	ReplyFromName       string     `json:"reply_from_name,omitempty"`
	RetentionDays       int        `json:"retention_days"`
	MessageCount        int64      `json:"message_count"`
	SpamCount           int64      `json:"spam_count"`
	LastMessageAt       *time.Time `json:"last_message_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}

// MessageField is one submitted form field, preserved in submission order.
type MessageField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MessageAttachment describes a file submitted with a form. Content is present
// only for small inline attachments; otherwise fetch it with
// [MessagesService.DownloadAttachment].
type MessageAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type,omitempty"`
	Size        int64  `json:"size"`
	StorageKey  string `json:"storage_key,omitempty"`
	Content     string `json:"content,omitempty"`
}

// MessageReply is an operator reply sent on a message thread, or an inbound
// message that continued it.
type MessageReply struct {
	ID             uint      `json:"id"`
	UUID           string    `json:"uuid"`
	MessageID      uint      `json:"message_id"`
	WorkspaceID    *uint     `json:"workspace_id,omitempty"`
	AuthorID       uint      `json:"author_id"`
	Kind           string    `json:"kind"`
	Subject        string    `json:"subject,omitempty"`
	FromAddr       string    `json:"from_addr,omitempty"`
	ToAddr         string    `json:"to_addr,omitempty"`
	HTMLBody       string    `json:"html_body,omitempty"`
	TextBody       string    `json:"text_body,omitempty"`
	EmailUUID      string    `json:"email_uuid,omitempty"`
	InboundEmailID *uint     `json:"inbound_email_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// Message is a web form submission. State tracks triage ("new", "open",
// "replied", "closed", "spam"); Status records the spam verdict ("received",
// "flagged", "spam", "rejected").
type Message struct {
	ID           uint                `json:"id"`
	UUID         string              `json:"uuid"`
	WorkspaceID  *uint               `json:"workspace_id,omitempty"`
	FormID       uint                `json:"form_id"`
	Form         *Form               `json:"form,omitempty"`
	Subject      string              `json:"subject,omitempty"`
	Body         string              `json:"body,omitempty"`
	Fields       []MessageField      `json:"fields,omitempty"`
	Attachments  []MessageAttachment `json:"attachments,omitempty"`
	SenderName   string              `json:"sender_name,omitempty"`
	SenderEmail  string              `json:"sender_email,omitempty"`
	SenderPhone  string              `json:"sender_phone,omitempty"`
	State        string              `json:"state"`
	Status       string              `json:"status"`
	SpamScore    float64             `json:"spam_score"`
	ScanReasons  []string            `json:"scan_reasons,omitempty"`
	AssignedToID *uint               `json:"assigned_to_id,omitempty"`
	ClientIP     string              `json:"client_ip,omitempty"`
	Origin       string              `json:"origin,omitempty"`
	Referer      string              `json:"referer,omitempty"`
	UserAgent    string              `json:"user_agent,omitempty"`
	Replies      []MessageReply      `json:"replies,omitempty"`
	ReplyCount   int                 `json:"reply_count"`
	ReadAt       *time.Time          `json:"read_at,omitempty"`
	RepliedAt    *time.Time          `json:"replied_at,omitempty"`
	NotifiedAt   *time.Time          `json:"notified_at,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    *time.Time          `json:"updated_at,omitempty"`
}

// MessageFilter is a spam rule applied to incoming form submissions. Kind
// selects what the pattern matches ("keyword", "phrase", "regex", "email",
// "domain", "ip"); Action is "score", "flag", "quarantine", "reject", or
// "allowlist".
type MessageFilter struct {
	ID            uint       `json:"id"`
	WorkspaceID   *uint      `json:"workspace_id,omitempty"`
	FormID        *uint      `json:"form_id,omitempty"`
	Kind          string     `json:"kind"`
	Pattern       string     `json:"pattern"`
	Fields        []string   `json:"fields,omitempty"`
	Action        string     `json:"action"`
	Score         float64    `json:"score"`
	CaseSensitive bool       `json:"case_sensitive"`
	Enabled       bool       `json:"enabled"`
	Note          string     `json:"note,omitempty"`
	HitCount      int64      `json:"hit_count"`
	LastHitAt     *time.Time `json:"last_hit_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

// InboundEmail is a message Posta received, either over its inbound SMTP
// listener or by webhook relay from an external provider.
type InboundEmail struct {
	ID              uint       `json:"id"`
	UUID            string     `json:"uuid"`
	UserID          uint       `json:"user_id"`
	WorkspaceID     *uint      `json:"workspace_id,omitempty"`
	DomainID        uint       `json:"domain_id"`
	MessageID       string     `json:"message_id,omitempty"`
	Sender          string     `json:"sender"`
	Recipients      []string   `json:"recipients"`
	Subject         string     `json:"subject,omitempty"`
	HTMLBody        string     `json:"html_body,omitempty"`
	TextBody        string     `json:"text_body,omitempty"`
	HeadersJSON     string     `json:"headers_json,omitempty"`
	AttachmentsJSON string     `json:"attachments_json,omitempty"`
	RawStorageKey   string     `json:"raw_storage_key,omitempty"`
	Size            int64      `json:"size"`
	Source          string     `json:"source,omitempty"`
	SpamScore       *float64   `json:"spam_score,omitempty"`
	Status          string     `json:"status"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	RetryCount      int        `json:"retry_count"`
	ForwardedAt     *time.Time `json:"forwarded_at,omitempty"`
	ReceivedAt      time.Time  `json:"received_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Workspace is a tenant: the boundary that owns templates, domains, keys, and
// every other resource in this API. Role is the caller's own role in it.
type Workspace struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
	OwnerID     uint   `json:"owner_id,omitempty"`
	Role        string `json:"role,omitempty"`
	IsPersonal  bool   `json:"is_personal,omitempty"`
	// System marks the workspace Posta itself sends from, such as
	// verification and password-reset mail.
	System    bool      `json:"system,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// WorkspaceMember is a user's membership of a workspace. Role is one of
// "owner", "admin", "editor", "viewer".
type WorkspaceMember struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email,omitempty"`
	Name      string    `json:"name,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Invitation is a pending offer of workspace membership.
type Invitation struct {
	ID          uint `json:"id"`
	WorkspaceID uint `json:"workspace_id"`
	// Workspace is the workspace's name, for showing an invitee what they
	// are being asked to join.
	Workspace string    `json:"workspace,omitempty"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Plan is the quota and feature set applied to a workspace or user.
type Plan struct {
	ID                    uint      `json:"id"`
	Name                  string    `json:"name"`
	Description           string    `json:"description,omitempty"`
	IsDefault             bool      `json:"is_default"`
	IsActive              bool      `json:"is_active"`
	DailyRateLimit        int       `json:"daily_rate_limit"`
	HourlyRateLimit       int       `json:"hourly_rate_limit"`
	MaxBatchSize          int       `json:"max_batch_size"`
	MaxAttachmentSizeMB   int       `json:"max_attachment_size_mb"`
	MaxAPIKeys            int       `json:"max_api_keys"`
	MaxDomains            int       `json:"max_domains"`
	MaxSMTPServers        int       `json:"max_smtp_servers"`
	MaxWorkspaces         int       `json:"max_workspaces"`
	EmailLogRetentionDays int       `json:"email_log_retention_days"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// User is an account on the platform.
type User struct {
	ID                    uint       `json:"id"`
	Email                 string     `json:"email"`
	Name                  string     `json:"name,omitempty"`
	Role                  string     `json:"role"`
	Active                bool       `json:"active"`
	AvatarURL             string     `json:"avatar_url,omitempty"`
	AuthMethod            string     `json:"auth_method,omitempty"`
	TwoFactorEnabled      bool       `json:"two_factor_enabled"`
	RequireVerifiedDomain bool       `json:"require_verified_domain"`
	EmailVerifiedAt       *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt           *time.Time `json:"last_login_at,omitempty"`
	ScheduledDeletionAt   *time.Time `json:"scheduled_deletion_at,omitempty"`
	DefaultWorkspaceID    *uint      `json:"default_workspace_id,omitempty"`
	PlanID                *uint      `json:"plan_id,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
}

// Notification is a dashboard message addressed to the signed-in user.
type Notification struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	WorkspaceID *uint  `json:"workspace_id,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Category    string `json:"category,omitempty"`
	Severity    string `json:"severity,omitempty"`
	Title       string `json:"title"`
	Body        string `json:"body,omitempty"`
	Link        string `json:"link,omitempty"`
	ActionText  string `json:"action_text,omitempty"`
	// DedupKey collapses repeats of the same underlying condition into one
	// notification rather than a stream of them.
	DedupKey    string     `json:"dedup_key,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	DismissedAt *time.Time `json:"dismissed_at,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Event is an audit or system event recorded by the platform. Metadata is a
// JSON document whose shape depends on Type.
type Event struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"`
	Category    string    `json:"category"`
	Message     string    `json:"message,omitempty"`
	ActorID     *uint     `json:"actor_id,omitempty"`
	ActorName   string    `json:"actor_name,omitempty"`
	WorkspaceID *uint     `json:"workspace_id,omitempty"`
	ClientIP    string    `json:"client_ip,omitempty"`
	Metadata    string    `json:"metadata,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Announcement is a platform-wide notice broadcast by an administrator.
// Recipients counts the users it reached.
type Announcement struct {
	ID         uint       `json:"id"`
	Title      string     `json:"title"`
	Message    string     `json:"message,omitempty"`
	Severity   string     `json:"severity,omitempty"`
	Link       string     `json:"link,omitempty"`
	CreatedBy  uint       `json:"created_by,omitempty"`
	AuthorName string     `json:"author_name,omitempty"`
	Recipients int        `json:"recipients"`
	SentAt     *time.Time `json:"sent_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Server is a shared SMTP server administered at the platform level and
// offered to workspaces that have none of their own.
type Server struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Host            string     `json:"host"`
	Port            int        `json:"port"`
	Username        string     `json:"username,omitempty"`
	Encryption      string     `json:"encryption,omitempty"`
	SecurityMode    string     `json:"security_mode,omitempty"`
	AllowedDomains  []string   `json:"allowed_domains,omitempty"`
	MaxRetries      int        `json:"max_retries"`
	Status          string     `json:"status,omitempty"`
	SentCount       int64      `json:"sent_count"`
	FailedCount     int64      `json:"failed_count"`
	ValidatedAt     *time.Time `json:"validated_at,omitempty"`
	ValidationError string     `json:"validation_error,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Setting is one platform or workspace configuration entry.
type Setting struct {
	ID    uint   `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MessageData is the envelope returned by endpoints whose only result is a
// confirmation string.
type MessageData struct {
	Message string `json:"message"`
}
