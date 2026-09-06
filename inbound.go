package posta

// InboundService reads the email Posta received, whether over its inbound SMTP
// listener or relayed in by an external provider's webhook.
type InboundService struct{ c *Client }

// List returns a page of inbound emails.
func (s *InboundService) List(opts *InboundListOptions) (*PageableResponse[InboundEmail], error) {
	return callPage[InboundEmail](s.c, wsPath+"/inbound-emails", opts.values())
}

// Get returns one inbound email with its parsed bodies.
func (s *InboundService) Get(uuid string) (*InboundEmail, error) {
	return get[InboundEmail](s.c, wsPath+"/inbound-emails/"+uuid, nil)
}

// Delete removes an inbound email and its stored raw message.
func (s *InboundService) Delete(uuid string) error {
	return s.c.delete(wsPath + "/inbound-emails/" + uuid)
}

// Retry re-dispatches the webhook for an inbound email whose forwarding
// failed.
func (s *InboundService) Retry(uuid string) (*InboundEmail, error) {
	return post[InboundEmail](s.c, wsPath+"/inbound-emails/"+uuid+"/retry", nil, nil)
}

// DownloadRaw fetches the original RFC 5322 message (.eml) and its
// Content-Type, for callers that need headers Posta did not parse out.
func (s *InboundService) DownloadRaw(uuid string) ([]byte, string, error) {
	return s.c.callRaw(wsPath+"/inbound-emails/"+uuid+"/raw", nil)
}

// DownloadAttachment fetches the file at index idx of an inbound email,
// returning its bytes and Content-Type.
func (s *InboundService) DownloadAttachment(uuid string, idx int) ([]byte, string, error) {
	return s.c.callRaw(sprintfPath("%s/inbound-emails/%s/attachments/%d", wsPath, uuid, idx), nil)
}
