package api

import (
	"context"
	"fmt"
)

// RecordType is the type of DNS record, which can be:
//   - A
//   - AAAA
//   - MX
//   - CNAME
//   - ALIAS
//   - TXT
//   - NS
//   - SRV
//   - TLSA
//   - CAA
//   - SSHFP
type RecordType string

type DNSRecord struct {
	// Id is the numeric record ID in the porkbun database
	Id string `json:"id"`
	// Name is the fully-qualified record name (e.g., www.example.com)
	Name string `json:"name"`
	// Type is the DNS record type (e.g., A, AAAA, MX, NS, CNAME, ...)
	Type RecordType `json:"type"`
	// Content is the record value (e.g., 1.2.3.4, srv.example.com)
	Content string `json:"content"`
	// Ttl is the time to live in seconds (e.g., "600")
	Ttl string `json:"ttl"`
	// Priority is used for MX, SRV records. Empty if not applicable (e.g., "10")
	Priority string `json:"prio"`
	// Notes are optional notes attached to the record. Empty if not set.
	Notes string `json:"notes"`
}

type CreateRecordRequest struct {
	// Domain is the domain name (e.g., example.com)
	Domain string `json:"domain"`
	// Name is the subdomain of the record (e.g., www). Leave blank for the root domain.
	Name string `json:"name"`
	// Type is the DNS record type (e.g., A, AAAA, MX, NS, CNAME, ...)
	Type RecordType `json:"type"`
	// Content is the record value (e.g., 1.2.3.4, srv.example.com)
	Content string `json:"content"`
	// Ttl is the time to live in seconds (e.g., "600")
	Ttl *int `json:"ttl,omitempty"`
	// Prio is the record priority used for MX, or SRV records. Null if not applicable.
	Prio *int `json:"prio,omitempty"`
	// Notes are optional notes attached to the record.
	Notes *string `json:"notes,omitempty"`
}

type RetrieveRecordsResponse struct {
	// Status will be set to "SUCCESS"
	Status string `json:"status"`
	// Cloudflare indicates whether the Cloudflare proxy is enabled for this domain.
	Cloudflare string      `json:"cloudflare"`
	Records    []DNSRecord `json:"records"`
}

type CreateRecordResponse struct {
	Status string `json:"status"`
	Id     int    `json:"id"`
}

type DeleteRecordRequest struct {
	Domain    string     `json:"-"`
	Type      RecordType `json:"-"`
	Subdomain string     `json:"-"`
}

type DeleteRecordByIdRequest struct {
	Domain string `json:"-"`
	Id     int    `json:"-"`
}

type UpdateRecordsRequest struct {
	Domain    string     `json:"-"`
	Type      RecordType `json:"-"`
	Subdomain string     `json:"-"`
	Content   string     `json:"content"`
	Notes     *string    `json:"notes,omitempty"`
	Priority  *int       `json:"prio,omitempty"`
	Ttl       *int       `json:"ttl,omitempty"`
}

type UpdateRecordByIdRequest struct {
	Domain   string     `json:"-"`
	Id       int        `json:"-"`
	Type     RecordType `json:"type"`
	Content  string     `json:"content"`
	Notes    *string    `json:"notes,omitempty"`
	Priority *int       `json:"prio,omitempty"`
	Ttl      *int       `json:"ttl,omitempty"`
}

type GetRecordsRequest struct {
	Domain    string     `json:"-"`
	Type      RecordType `json:"-"`
	Subdomain string     `json:"-"`
}

type GetRecordByIdRequest struct {
	Domain string `json:"-"`
	Id     int    `json:"-"`
}

type DNSRecordUpdate struct {
	Content  string  `json:"content"`
	Notes    *string `json:"notes,omitempty"`
	Priority *int    `json:"prio,omitempty"`
	Ttl      *int    `json:"ttl,omitempty"`
}

type RecordsUpdatedResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type DNSService interface {
	// ListRecords returns all DNS records for a given domain
	ListRecords(ctx context.Context, domain string) (resp RetrieveRecordsResponse, err error)

	// GetRecords returns all DNS records for a given domain and type, and optionally a specific subdomain.
	GetRecords(ctx context.Context, req GetRecordsRequest) (resp RetrieveRecordsResponse, err error)

	// GetRecordById returns a specific DNS record by its Porkbun record ID
	GetRecordById(ctx context.Context, req GetRecordByIdRequest) (resp RetrieveRecordsResponse, err error)

	// CreateRecord creates a new DNS record with the specified parameters.
	// See https://porkbun.com/api/json/v3/documentation#tag/dns/POST/dns/create/{domain}
	CreateRecord(ctx context.Context, req CreateRecordRequest) (resp CreateRecordResponse, err error)

	// UpdateRecords updates one or more DNS records for a given domain, type, and subdomain.
	UpdateRecords(ctx context.Context, req UpdateRecordsRequest) (err error)

	// UpdateRecordById updates a specific DNS record by its Porkbun record ID.
	UpdateRecordById(ctx context.Context, req UpdateRecordByIdRequest) (err error)

	// DeleteRecord deletes all DNS records for a given domain, type, and subdomain.
	DeleteRecord(ctx context.Context, req DeleteRecordRequest) (err error)

	// DeleteRecordById deletes a specific DNS record by its Porkbun record ID.
	DeleteRecordById(ctx context.Context, req DeleteRecordByIdRequest) (err error)
}

type dnsService struct {
	client *Client
}

func (s *dnsService) ListRecords(ctx context.Context, domain string) (resp RetrieveRecordsResponse, err error) {
	path := fmt.Sprintf("/dns/retrieve/%s", domain)
	err = s.client.PostJson(ctx, path, EmptyPayload, &resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *dnsService) GetRecordById(ctx context.Context, req GetRecordByIdRequest) (resp RetrieveRecordsResponse, err error) {
	path := fmt.Sprintf("/dns/retrieve/%s/%d", req.Domain, req.Id)
	err = s.client.PostJson(ctx, path, EmptyPayload, &resp)
	if err != nil {
		return resp, err
	}
	if len(resp.Records) == 0 {
		return resp, fmt.Errorf("record with ID %d in domain %s not found", req.Id, req.Domain)
	}
	return resp, nil
}

func (s *dnsService) GetRecords(ctx context.Context, req GetRecordsRequest) (resp RetrieveRecordsResponse, err error) {
	var path string
	if req.Subdomain == "" {
		path = fmt.Sprintf("/dns/retrieveByNameType/%s/%s", req.Domain, req.Type)
	} else {
		path = fmt.Sprintf("/dns/retrieveByNameType/%s/%s/%s", req.Domain, req.Type, req.Subdomain)
	}
	err = s.client.PostJson(ctx, path, EmptyPayload, &resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *dnsService) CreateRecord(ctx context.Context, req CreateRecordRequest) (resp CreateRecordResponse, err error) {
	path := fmt.Sprintf("/dns/create/%s", req.Domain)
	err = s.client.PostJson(ctx, path, req, &resp)
	if err != nil {
		return resp, err
	}
	if resp.Status != "SUCCESS" {
		return resp, fmt.Errorf("failed to create record: status=%s", resp.Status)
	}
	return resp, nil
}

func (s *dnsService) UpdateRecords(ctx context.Context, req UpdateRecordsRequest) (err error) {
	var resp BasicResponse
	path := fmt.Sprintf("/dns/editByNameType/%s/%s/%s", req.Domain, req.Type, req.Subdomain)
	err = s.client.PostJson(ctx, path, req, &resp)
	return err
}

func (s *dnsService) UpdateRecordById(ctx context.Context, req UpdateRecordByIdRequest) (err error) {
	var resp BasicResponse
	path := fmt.Sprintf("/dns/edit/%s/%d", req.Domain, req.Id)
	err = s.client.PostJson(ctx, path, req, &resp)
	return err
}

func (s *dnsService) DeleteRecord(ctx context.Context, req DeleteRecordRequest) (err error) {
	var resp BasicResponse
	var path string
	if req.Subdomain == "" {
		path = fmt.Sprintf("/dns/deleteByNameType/%s/%s", req.Domain, req.Type)
	} else {
		path = fmt.Sprintf("/dns/deleteByNameType/%s/%s/%s", req.Domain, req.Type, req.Subdomain)
	}
	err = s.client.PostJson(ctx, path, EmptyPayload, &resp)
	return err
}

func (s *dnsService) DeleteRecordById(ctx context.Context, req DeleteRecordByIdRequest) (err error) {
	var resp BasicResponse
	path := fmt.Sprintf("/dns/delete/%s/%d", req.Domain, req.Id)
	err = s.client.PostJson(ctx, path, EmptyPayload, &resp)
	return err
}
