package api

import "context"

type DNSRecord struct {
	// Id is the numeric record ID in the porkbun database
	Id string `json:"id"`
	// Name is the fully-qualified record name (e.g., www.example.com)
	Name string `json:"name"`
	// Type is the DNS record type (e.g., A, MX, NS, CNAME, ...)
	Type string `json:"type"`
	// Content is the record value (e.g., 1.2.3.4, srv.example.com)
	Content string `json:"content"`
	// TTL is the time to live in seconds (e.g., "600")
	TTL string `json:"ttl"`
	// Priority is used for MX, SRV records. Empty if not applicable (e.g., "10")
	Priority *string `json:"prio"`
	// Notes are optional notes attached to the record. Empty if not set.
	Notes string `json:"notes"`
}

type listRecordsResponse struct {
	Status  string      `json:"status"`
	Records []DNSRecord `json:"records"`
}

type DNSService interface {
	ListRecords(ctx context.Context, domain string) ([]DNSRecord, error)
}

type dnsService struct {
	client *Client
}

func (s *dnsService) ListRecords(ctx context.Context, domain string) ([]DNSRecord, error) {
	var resp listRecordsResponse
	path := "/dns/retrieve/" + domain
	err := s.client.PostJson(ctx, path, struct{}{}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Records, nil
}
