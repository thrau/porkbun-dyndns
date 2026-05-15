package main

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/thrau/porkbun-dyndns/internal/api"
	"github.com/thrau/porkbun-dyndns/internal/util"
)

// UpdateRequestBuilder is a helper used in the update-request command of the CLI, which builds an
// `UpdateRecordByIdRequest` from CLI flags and merges default values from an existing record.
type UpdateRequestBuilder struct {
	Domain    string
	Id        int
	Type      *string
	Content   *string
	Subdomain *string
	Notes     *string
	Priority  *int
	Ttl       *int
}

func NewUpdateRequestBuilder() *UpdateRequestBuilder {
	return &UpdateRequestBuilder{}
}

// SetValuesFromCommandFlags is meant as an initial population step. It sets the values of the builder from the given
// command flags. Expects the flags passed to `porkbun-dns update-record --domain <domain> --id <id> [flags]`.
func (b *UpdateRequestBuilder) SetValuesFromCommandFlags(cmd *cobra.Command) error {
	domain, err := cmd.Flags().GetString("domain")
	if err != nil {
		return err
	}
	id, err := cmd.Flags().GetInt("id")
	if err != nil {
		return err
	}
	recordType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	content, err := cmd.Flags().GetString("content")
	if err != nil {
		return err
	}
	notes, err := cmd.Flags().GetString("notes")
	if err != nil {
		return err
	}
	prio, err := cmd.Flags().GetInt("priority")
	if err != nil {
		return err
	}
	ttl, err := cmd.Flags().GetInt("ttl")
	if err != nil {
		return err
	}
	subdomain, err := cmd.Flags().GetString("subdomain")
	if err != nil {
		return err
	}

	b.Domain = domain
	b.Id = id

	if domain == "" {
		return errors.New("domain cannot be empty")
	}

	if cmd.Flags().Changed("content") {
		b.Content = api.String(content)
	}
	if cmd.Flags().Changed("type") {
		b.Type = api.String(recordType)
	}
	if cmd.Flags().Changed("notes") {
		b.Notes = api.String(notes)
	}
	if cmd.Flags().Changed("priority") {
		b.Priority = api.Int(prio)
	}
	if cmd.Flags().Changed("ttl") {
		b.Ttl = api.Int(ttl)
	}
	if cmd.Flags().Changed("subdomain") {
		b.Subdomain = api.String(subdomain)
	}
	return nil
}

// SetDefaultValuesFromRecord sets nil values of the builder to the values of the given record.
func (b *UpdateRequestBuilder) SetDefaultValuesFromRecord(record api.DNSRecord) error {
	if b.Type == nil {
		b.Type = api.String(string(record.Type))
	}
	if b.Subdomain == nil {
		if _, subdomain := util.SplitDomain(record.Name); subdomain != "" {
			b.Subdomain = api.String(subdomain)
		}
	}
	if b.Priority == nil {
		b.Priority = toInt(record.Priority)
	}
	if b.Ttl == nil {
		b.Ttl = toInt(record.Ttl)
	}
	if b.Notes == nil {
		b.Notes = api.String(record.Notes)
	}

	return nil
}

func CheckWouldUpdate(record api.DNSRecord, updateRequest api.UpdateRecordByIdRequest) bool {
	if record.Type != updateRequest.Type {
		return true
	}
	if record.Content != updateRequest.Content {
		return true
	}

	_, recordSubdomain := util.SplitDomain(record.Name)
	updateSubdomain := ""
	if updateRequest.Name != nil {
		updateSubdomain = *updateRequest.Name
	}
	if recordSubdomain != updateSubdomain {
		return true
	}

	if *toInt(record.Priority) != *updateRequest.Priority {
		// empty priority and nil priority are treated as equal
		// TODO: not sure this is correct
		if !(record.Priority == "" && updateRequest.Priority == nil) {
			return true
		}
	}
	if *toInt(record.Ttl) != *updateRequest.Ttl {
		return true
	}
	if *api.String(record.Notes) != *updateRequest.Notes {
		return true
	}

	return false
}

// Build creates an UpdateRecordByIdRequest from the builder's current state.
func (b *UpdateRequestBuilder) Build() (req api.UpdateRecordByIdRequest, err error) {
	if b.Content == nil {
		return req, errors.New("content cannot be nil")
	}
	if b.Type == nil {
		return req, errors.New("type cannot be nil")
	}

	req.Domain = b.Domain
	req.Id = b.Id
	req.Type = api.RecordType(*b.Type)
	req.Content = *b.Content
	req.Name = b.Subdomain
	req.Priority = b.Priority
	req.Ttl = b.Ttl
	req.Notes = b.Notes

	return req, err
}
