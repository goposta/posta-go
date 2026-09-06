package posta

import (
	"fmt"
)

// TemplatesService manages templates and the versions and localizations
// beneath them.
//
// A template is a named container. Its content lives on immutable versions,
// and each version carries one localization per language. Sending resolves the
// template's active version unless a caller names another.
type TemplatesService struct{ c *Client }

// CreateTemplateRequest creates a template. Only Name is required; content is
// added afterwards as a version with localizations.
type CreateTemplateRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	DefaultLanguage string `json:"default_language,omitempty"`
	// SampleData is JSON used to render previews of this template.
	SampleData string `json:"sample_data,omitempty"`
}

// UpdateTemplateRequest changes a template's metadata. Nil fields are left
// unchanged.
type UpdateTemplateRequest struct {
	Name            string  `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	DefaultLanguage string  `json:"default_language,omitempty"`
	SampleData      *string `json:"sample_data,omitempty"`
}

// Create adds a template.
func (s *TemplatesService) Create(req *CreateTemplateRequest) (*Template, error) {
	return post[Template](s.c, wsPath+"/templates", req, nil)
}

// List returns a page of templates in the summary shape, without version
// bodies.
func (s *TemplatesService) List(opts *SearchListOptions) (*PageableResponse[TemplateListItem], error) {
	return callPage[TemplateListItem](s.c, wsPath+"/templates", opts.values())
}

// Get returns one template with its active version.
func (s *TemplatesService) Get(id uint) (*Template, error) {
	return get[Template](s.c, fmt.Sprintf("%s/templates/%d", wsPath, id), nil)
}

// Update changes a template's metadata.
func (s *TemplatesService) Update(id uint, req *UpdateTemplateRequest) (*Template, error) {
	return put[Template](s.c, fmt.Sprintf("%s/templates/%d", wsPath, id), req)
}

// Delete removes a template and every version under it.
func (s *TemplatesService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/templates/%d", wsPath, id))
}

// PreviewTemplateRequest renders template source directly, without saving it.
// It is the preview used while editing, before a version exists.
type PreviewTemplateRequest struct {
	SubjectTemplate string         `json:"subject_template"`
	HTMLTemplate    string         `json:"html_template,omitempty"`
	TextTemplate    string         `json:"text_template,omitempty"`
	StylesheetID    *uint          `json:"stylesheet_id,omitempty"`
	TemplateData    map[string]any `json:"template_data,omitempty"`
}

// PreviewResult is rendered template output.
type PreviewResult struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
}

// Preview renders unsaved template source with the given variables.
func (s *TemplatesService) Preview(req *PreviewTemplateRequest) (*PreviewResult, error) {
	return post[PreviewResult](s.c, wsPath+"/templates/preview", req, nil)
}

// SendTestRequest sends a template to a handful of addresses so an editor can
// see the real thing in a real inbox.
type SendTestRequest struct {
	To           []string       `json:"to"`
	From         string         `json:"from,omitempty"`
	Language     string         `json:"language,omitempty"`
	TemplateData map[string]any `json:"template_data,omitempty"`
}

// SendTest sends a test rendering of a template.
func (s *TemplatesService) SendTest(id uint, req *SendTestRequest) (*SendResponse, error) {
	return post[SendResponse](s.c, fmt.Sprintf("%s/templates/%d/send-test", wsPath, id), req, nil)
}

// TemplateExportVersion is one version inside an exported template.
type TemplateExportVersion struct {
	Version       int                    `json:"version"`
	SampleData    string                 `json:"sample_data,omitempty"`
	Active        bool                   `json:"active,omitempty"`
	Localizations []TemplateLocalization `json:"localizations,omitempty"`
}

// TemplateExport is a template and all its versions in a portable shape. It is
// what Export returns and what Import accepts, so a template can be moved
// between workspaces or kept in version control.
type TemplateExport struct {
	Name            string                  `json:"name"`
	Description     string                  `json:"description,omitempty"`
	DefaultLanguage string                  `json:"default_language,omitempty"`
	SampleData      string                  `json:"sample_data,omitempty"`
	Versions        []TemplateExportVersion `json:"versions,omitempty"`
	PostaVersion    string                  `json:"posta_version,omitempty"`
	ExportedAt      string                  `json:"exported_at,omitempty"`
}

// Export returns a template and all its versions in portable form.
func (s *TemplatesService) Export(id uint) (*TemplateExport, error) {
	return get[TemplateExport](s.c, fmt.Sprintf("%s/templates/%d/export", wsPath, id), nil)
}

// Import recreates a template from an [TemplateExport] payload.
func (s *TemplatesService) Import(req *TemplateExport) (*Template, error) {
	return post[Template](s.c, wsPath+"/templates/import", req, nil)
}

// ImportHTMLRequest creates a template from a raw HTML document, letting Posta
// derive the text alternative.
type ImportHTMLRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	SubjectTemplate string `json:"subject_template,omitempty"`
	HTML            string `json:"html"`
	Language        string `json:"language,omitempty"`
}

// ImportHTML creates a template from a raw HTML document.
func (s *TemplatesService) ImportHTML(req *ImportHTMLRequest) (*Template, error) {
	return post[Template](s.c, wsPath+"/templates/import-html", req, nil)
}

// CreateVersionRequest opens a new version of a template, copying the current
// active version's localizations as a starting point.
type CreateVersionRequest struct {
	SampleData   string `json:"sample_data,omitempty"`
	StylesheetID *uint  `json:"stylesheet_id,omitempty"`
}

// UpdateVersionRequest changes the stylesheet attached to a version.
type UpdateVersionRequest struct {
	StylesheetID *uint `json:"stylesheet_id,omitempty"`
}

// ListVersions returns every version of a template, newest first.
func (s *TemplatesService) ListVersions(templateID uint) (*[]TemplateVersion, error) {
	return get[[]TemplateVersion](s.c, fmt.Sprintf("%s/templates/%d/versions", wsPath, templateID), nil)
}

// CreateVersion opens a new draft version of a template.
func (s *TemplatesService) CreateVersion(templateID uint, req *CreateVersionRequest) (*TemplateVersion, error) {
	return post[TemplateVersion](s.c, fmt.Sprintf("%s/templates/%d/versions", wsPath, templateID), req, nil)
}

// UpdateVersion changes a version's stylesheet.
func (s *TemplatesService) UpdateVersion(templateID, versionID uint, req *UpdateVersionRequest) (*TemplateVersion, error) {
	return put[TemplateVersion](s.c, fmt.Sprintf("%s/templates/%d/versions/%d", wsPath, templateID, versionID), req)
}

// DeleteVersion removes a version. The active version cannot be deleted.
func (s *TemplatesService) DeleteVersion(templateID, versionID uint) error {
	return s.c.delete(fmt.Sprintf("%s/templates/%d/versions/%d", wsPath, templateID, versionID))
}

// ActivateVersion makes a version the one that sends.
func (s *TemplatesService) ActivateVersion(templateID, versionID uint) (*Template, error) {
	return post[Template](s.c, fmt.Sprintf("%s/templates/%d/activate/%d", wsPath, templateID, versionID), nil, nil)
}

// CreateLocalizationRequest adds one language's content to a version.
type CreateLocalizationRequest struct {
	Language        string `json:"language"`
	SubjectTemplate string `json:"subject_template"`
	HTMLTemplate    string `json:"html_template,omitempty"`
	TextTemplate    string `json:"text_template,omitempty"`
	// BuilderJSON stores the visual editor's document for this localization.
	BuilderJSON string `json:"builder_json,omitempty"`
}

// UpdateLocalizationRequest changes one language's content. Nil fields are
// left unchanged.
type UpdateLocalizationRequest struct {
	SubjectTemplate *string `json:"subject_template,omitempty"`
	HTMLTemplate    *string `json:"html_template,omitempty"`
	TextTemplate    *string `json:"text_template,omitempty"`
	BuilderJSON     *string `json:"builder_json,omitempty"`
}

// ListLocalizations returns every language defined on a version.
func (s *TemplatesService) ListLocalizations(templateID, versionID uint) (*[]TemplateLocalization, error) {
	return get[[]TemplateLocalization](s.c, fmt.Sprintf("%s/templates/%d/versions/%d/localizations", wsPath, templateID, versionID), nil)
}

// CreateLocalization adds a language to a version.
func (s *TemplatesService) CreateLocalization(templateID, versionID uint, req *CreateLocalizationRequest) (*TemplateLocalization, error) {
	return post[TemplateLocalization](s.c, fmt.Sprintf("%s/templates/%d/versions/%d/localizations", wsPath, templateID, versionID), req, nil)
}

// UpdateLocalization changes a language's content. Localizations are addressed
// by their own ID, not by template and version.
func (s *TemplatesService) UpdateLocalization(localizationID uint, req *UpdateLocalizationRequest) (*TemplateLocalization, error) {
	return put[TemplateLocalization](s.c, fmt.Sprintf("%s/localizations/%d", wsPath, localizationID), req)
}

// DeleteLocalization removes a language from its version.
func (s *TemplatesService) DeleteLocalization(localizationID uint) error {
	return s.c.delete(fmt.Sprintf("%s/localizations/%d", wsPath, localizationID))
}

// PreviewLocalization renders a saved version in one language.
func (s *TemplatesService) PreviewLocalization(templateID, versionID uint, language string, data map[string]any) (*PreviewResult, error) {
	body := struct {
		Language     string         `json:"language"`
		TemplateData map[string]any `json:"template_data,omitempty"`
	}{Language: language, TemplateData: data}
	return post[PreviewResult](s.c, fmt.Sprintf("%s/templates/%d/versions/%d/preview", wsPath, templateID, versionID), body, nil)
}

// LanguagesService manages the languages a workspace's templates can be
// localized into.
type LanguagesService struct{ c *Client }

// CreateLanguageRequest adds a language. Code is a BCP 47 tag such as "en" or
// "pt-BR".
type CreateLanguageRequest struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default,omitempty"`
}

// UpdateLanguageRequest renames a language or makes it the default.
type UpdateLanguageRequest struct {
	Code      string `json:"code,omitempty"`
	Name      string `json:"name,omitempty"`
	IsDefault *bool  `json:"is_default,omitempty"`
}

// Create adds a language.
func (s *LanguagesService) Create(req *CreateLanguageRequest) (*Language, error) {
	return post[Language](s.c, wsPath+"/languages", req, nil)
}

// List returns a page of languages.
func (s *LanguagesService) List(opts *ListOptions) (*PageableResponse[Language], error) {
	return callPage[Language](s.c, wsPath+"/languages", opts.values())
}

// Update changes a language.
func (s *LanguagesService) Update(id uint, req *UpdateLanguageRequest) (*Language, error) {
	return put[Language](s.c, fmt.Sprintf("%s/languages/%d", wsPath, id), req)
}

// Delete removes a language. Localizations already written in it are not
// removed.
func (s *LanguagesService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/languages/%d", wsPath, id))
}

// StylesheetsService manages reusable CSS that template versions share, so a
// house style can be changed in one place.
type StylesheetsService struct{ c *Client }

// StylesheetRequest creates or replaces a stylesheet.
type StylesheetRequest struct {
	Name string `json:"name"`
	CSS  string `json:"css"`
}

// Create adds a stylesheet.
func (s *StylesheetsService) Create(req *StylesheetRequest) (*Stylesheet, error) {
	return post[Stylesheet](s.c, wsPath+"/stylesheets", req, nil)
}

// List returns a page of stylesheets.
func (s *StylesheetsService) List(opts *ListOptions) (*PageableResponse[Stylesheet], error) {
	return callPage[Stylesheet](s.c, wsPath+"/stylesheets", opts.values())
}

// Update replaces a stylesheet's name and CSS.
func (s *StylesheetsService) Update(id uint, req *StylesheetRequest) (*Stylesheet, error) {
	return put[Stylesheet](s.c, fmt.Sprintf("%s/stylesheets/%d", wsPath, id), req)
}

// Delete removes a stylesheet. Versions referencing it fall back to no
// stylesheet.
func (s *StylesheetsService) Delete(id uint) error {
	return s.c.delete(fmt.Sprintf("%s/stylesheets/%d", wsPath, id))
}
