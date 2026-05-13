package api

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListRecords(t *testing.T) {
	client := NewClient(WithEnvironmentCredentials())
	records, err := client.Dns.ListRecords(context.TODO(), "rauschig.org")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(records, "", "  ")
	t.Logf("Response: %s", string(responseJSON))
}

func TestListRecordsWithoutCredentials(t *testing.T) {
	client := NewClient()
	records, err := client.Dns.ListRecords(context.TODO(), "rauschig.org")

	if records != nil {
		t.Errorf("Expected records to be nil, got %v", records)
	}

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	var apiErr ErrorResponse
	assert := assert.New(t)
	assert.ErrorAs(err, &apiErr, "Expected ErrorResponse error type")
	assert.Equal(400, apiErr.HTTPStatus, "Expected HTTPStatus 400")
	assert.Equal("ERROR", apiErr.Status, "Expected Status 'ERROR'")
	assert.Equal("API_KEY_REQUIRED", apiErr.Code, "Expected Code 'API_KEY_REQUIRED'")
	assert.Equal("All API requests require an API key or API token.", apiErr.Message, "Expected Message 'All API requests require an API key or API token.'")
}

func TestListRecordsInvalidDomain(t *testing.T) {
	client := NewClient(WithEnvironmentCredentials())
	records, err := client.Dns.ListRecords(context.TODO(), "test.asdfg")

	if records != nil {
		t.Errorf("Expected records to be nil, got %v", records)
	}

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	var apiErr ErrorResponse
	assert := assert.New(t)
	assert.ErrorAs(err, &apiErr, "Expected ErrorResponse error type")
	assert.Equal(400, apiErr.HTTPStatus, "Expected HTTPStatus 400")
	assert.Equal("ERROR", apiErr.Status, "Expected Status 'ERROR'")
	assert.Equal("INVALID_DOMAIN", apiErr.Code, "Expected Code 'INVALID_DOMAIN'")
	assert.Equal("Invalid domain.", apiErr.Message, "Expected Message 'Invalid domain.'")
}

func TestDnsService_GetRecordById(t *testing.T) {
	client := NewClient(WithEnvironmentCredentials())
	records, err := client.Dns.GetRecordById(context.TODO(), "rauschig.org", "382004330")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert := assert.New(t)
	assert.Equal(len(records), 1, "Expected 1 record, got %d", len(records))

	record := records[0]
	expected := DNSRecord{
		Id:       "382004330",
		Name:     "rauschig.org",
		Type:     "NS",
		Content:  "fortaleza.porkbun.com",
		Ttl:      "86400",
		Priority: "",
		Notes:    "",
	}
	assert.Equal(expected, record)
}

func TestDnsService_GetRecordByType(t *testing.T) {
	client := NewClient(WithEnvironmentCredentials())
	records, err := client.Dns.GetRecordByType(context.TODO(), "rauschig.org", "MX")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert := assert.New(t)
	assert.Equal(len(records), 1, "Expected 1 record, got %d", len(records))

	record := records[0]
	expected := DNSRecord{
		Id:       "382004521",
		Name:     "rauschig.org",
		Type:     "MX",
		Content:  "srv.divzero.at",
		Ttl:      "600",
		Priority: "10",
		Notes:    "",
	}
	assert.Equal(expected, record)
}

func TestDnsService_GetRecordByNameAndType(t *testing.T) {
	client := NewClient(WithEnvironmentCredentials())
	records, err := client.Dns.GetRecordByNameAndType(context.TODO(), "rauschig.org", "*", "CNAME")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert := assert.New(t)
	assert.Equal(len(records), 1, "Expected 1 record, got %d", len(records))

	record := records[0]
	expected := DNSRecord{
		Id:       "382004466",
		Name:     "*.rauschig.org",
		Type:     "CNAME",
		Content:  "rauschig.org",
		Ttl:      "600",
		Priority: "0",
		Notes:    "",
	}
	assert.Equal(expected, record)
}
