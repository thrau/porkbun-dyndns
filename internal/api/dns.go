package api

import (
	"context"
	"fmt"
)

type DNSRecord struct {
	// Id is the numeric record ID in the porkbun database
	Id string `json:"id"`
	// Name is the fully-qualified record name (e.g., www.example.com)
	Name string `json:"name"`
	// Type is the DNS record type (e.g., A, MX, NS, CNAME, ...)
	Type string `json:"type"`
	// Content is the record value (e.g., 1.2.3.4, srv.example.com)
	Content string `json:"content"`
	// Ttl is the time to live in seconds (e.g., "600")
	Ttl string `json:"ttl"`
	// Priority is used for MX, SRV records. Empty if not applicable (e.g., "10")
	Priority string `json:"prio"`
	// Notes are optional notes attached to the record. Empty if not set.
	Notes string `json:"notes"`
}

type retrieveRecordsResponse struct {
	Status  string      `json:"status"`
	Records []DNSRecord `json:"records"`
}

type DNSService interface {
	ListRecords(ctx context.Context, domain string) ([]DNSRecord, error)
	GetRecordById(ctx context.Context, domain string, id string) ([]DNSRecord, error)
	GetRecordByType(ctx context.Context, domain string, recordType string) ([]DNSRecord, error)
	GetRecordByNameAndType(ctx context.Context, domain string, subdomain string, recordType string) ([]DNSRecord, error)
}

type dnsService struct {
	client *Client
}

func (s *dnsService) ListRecords(ctx context.Context, domain string) ([]DNSRecord, error) {
	var resp retrieveRecordsResponse
	path := fmt.Sprintf("/dns/retrieve/%s", domain)
	err := s.client.PostJson(ctx, path, EmptyPayload, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Records, nil
}

func (s *dnsService) GetRecordById(ctx context.Context, domain string, id string) ([]DNSRecord, error) {
	var resp retrieveRecordsResponse
	path := fmt.Sprintf("/dns/retrieve/%s/%s", domain, id)
	err := s.client.PostJson(ctx, path, EmptyPayload, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Records) == 0 {
		return nil, fmt.Errorf("record with ID %s in domain %s not found", id, domain)
	}
	return resp.Records, nil
}

func (s *dnsService) GetRecordByType(ctx context.Context, domain string, recordType string) ([]DNSRecord, error) {
	var resp retrieveRecordsResponse
	path := fmt.Sprintf("/dns/retrieveByNameType/%s/%s", domain, recordType)
	err := s.client.PostJson(ctx, path, EmptyPayload, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Records) == 0 {
		return nil, fmt.Errorf("%s record for domain %s not found", recordType, domain)
	}
	return resp.Records, nil
}

func (s *dnsService) GetRecordByNameAndType(ctx context.Context, domain string, subdomain string, recordType string) ([]DNSRecord, error) {
	var resp retrieveRecordsResponse
	path := fmt.Sprintf("/dns/retrieveByNameType/%s/%s/%s", domain, recordType, subdomain)
	err := s.client.PostJson(ctx, path, EmptyPayload, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Records) == 0 {
		return nil, fmt.Errorf("%s record %s in domain %s not found", recordType, subdomain, domain)
	}
	return resp.Records, nil
}
