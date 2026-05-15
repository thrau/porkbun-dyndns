package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thrau/porkbun-dyndns/internal/util"
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
	response, err := client.Dns.ListRecords(context.TODO(), "rauschig.org")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	var apiErr ApiError
	assert := assert.New(t)
	assert.Empty(response.Records, "Expected records to be empty")
	assert.ErrorAs(err, &apiErr, "Expected ErrorResponse error type")
	assert.Equal(400, apiErr.HTTPStatus, "Expected HTTPStatus 400")
	assert.Equal("ERROR", apiErr.Status, "Expected Status 'ERROR'")
	assert.Equal("API_KEY_REQUIRED", apiErr.Code, "Expected Code 'API_KEY_REQUIRED'")
	assert.Equal("All API requests require an API key or API token.", apiErr.Message, "Expected Message 'All API requests require an API key or API token.'")
}

func TestListRecordsInvalidDomain(t *testing.T) {
	client := NewClient(WithEnvironmentCredentials())
	response, err := client.Dns.ListRecords(context.TODO(), "test.asdfg")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	var apiErr ApiError
	assert := assert.New(t)
	assert.Empty(response.Records, "Expected records to be empty")
	assert.ErrorAs(err, &apiErr, "Expected ErrorResponse error type")
	assert.Equal(400, apiErr.HTTPStatus, "Expected HTTPStatus 400")
	assert.Equal("ERROR", apiErr.Status, "Expected Status 'ERROR'")
	assert.Equal("INVALID_DOMAIN", apiErr.Code, "Expected Code 'INVALID_DOMAIN'")
	assert.Equal("Invalid domain.", apiErr.Message, "Expected Message 'Invalid domain.'")
}

func TestDnsService_GetRecordById(t *testing.T) {
	client := NewClient(WithEnvironmentCredentials())
	resp, err := client.Dns.GetRecordById(context.TODO(), GetRecordByIdRequest{
		Domain: "rauschig.org",
		Id:     382004330,
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert := assert.New(t)
	assert.Equal(1, len(resp.Records), "Expected 1 record, got %d", len(resp.Records))

	record := resp.Records[0]
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
	resp, err := client.Dns.GetRecords(context.TODO(), GetRecordsRequest{
		Domain: "rauschig.org",
		Type:   "MX",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert := assert.New(t)
	assert.Equal(1, len(resp.Records), "Expected 1 record, got %d", len(resp.Records))

	record := resp.Records[0]
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
	resp, err := client.Dns.GetRecords(context.TODO(), GetRecordsRequest{
		Domain:    "rauschig.org",
		Subdomain: "*",
		Type:      "CNAME",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert := assert.New(t)
	assert.Equal(1, len(resp.Records), "Expected 1 record, got %d", len(resp.Records))

	record := resp.Records[0]
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

func TestDnsService_CRUDRoundtrip(t *testing.T) {
	client := NewClient(WithEnvironmentCredentials())

	domain := "rauschig.org"
	subdomain := fmt.Sprintf("test-%s", util.RandomShortId())

	// create the record
	response, err := client.Dns.CreateRecord(context.TODO(), CreateRecordRequest{
		Domain:  domain,
		Name:    subdomain,
		Type:    "A",
		Content: "127.0.0.1",
		Ttl:     Int(600),
	})

	// make sure the record even if an error is raised during the test
	defer func() {
		client.Dns.DeleteRecordById(context.TODO(), DeleteRecordByIdRequest{
			Domain: domain,
			Id:     response.Id,
		})
	}()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	recordId := response.Id

	// assert that it's there
	recordsResponse, err := client.Dns.GetRecordById(context.TODO(), GetRecordByIdRequest{domain, recordId})
	record := recordsResponse.Records[0]
	assert.Equal(t, fmt.Sprintf("%s.%s", subdomain, domain), record.Name)
	assert.Equal(t, RecordType("A"), record.Type)
	assert.Equal(t, strconv.Itoa(recordId), record.Id)
	assert.Equal(t, "127.0.0.1", record.Content)

	// sleep for 5 seconds (wait for porkbun provisioning)
	time.Sleep(5 * time.Second)

	// update the record by id
	err = client.Dns.UpdateRecords(context.TODO(), UpdateRecordsRequest{
		Domain:    domain,
		Type:      "A",
		Subdomain: subdomain,
		Content:   "127.0.0.2",
		Ttl:       Int(1200),
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// check that it was updated (use a different Get method)
	recordsResponse, err = client.Dns.GetRecords(context.TODO(), GetRecordsRequest{domain, "A", subdomain})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	record = recordsResponse.Records[0]
	assert.Equal(t, fmt.Sprintf("%s.%s", subdomain, domain), record.Name)
	assert.Equal(t, RecordType("A"), record.Type)
	assert.Equal(t, strconv.Itoa(recordId), record.Id)
	assert.Equal(t, "127.0.0.2", record.Content)

	// delete the record by id
	err = client.Dns.DeleteRecordById(context.TODO(), DeleteRecordByIdRequest{
		Domain: domain,
		Id:     response.Id,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// assert that it's gone
	recordsResponse, err = client.Dns.GetRecordById(context.TODO(), GetRecordByIdRequest{domain, recordId})
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestDnsService_UpdateRecordById(t *testing.T) {
	// documents the somewhat peculiar behavior of `/edit/{domain}/{id}`.
	client := NewClient(WithEnvironmentCredentials())

	domain := "rauschig.org"
	subdomain := fmt.Sprintf("test-%s", util.RandomShortId())

	// create the record
	response, err := client.Dns.CreateRecord(context.TODO(), CreateRecordRequest{
		Domain:  domain,
		Name:    subdomain,
		Type:    "TXT",
		Content: "my-original-value",
		Ttl:     Int(1200),
		Notes:   String("created by porkbun-dyndns"),
		Prio:    Int(10),
	})
	// make sure the record even if an error is raised during the test
	defer func() {
		client.Dns.DeleteRecordById(context.TODO(), DeleteRecordByIdRequest{
			Domain: domain,
			Id:     response.Id,
		})
	}()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// assert the record was created as expected
	resp, err := client.Dns.GetRecordById(context.TODO(), GetRecordByIdRequest{domain, response.Id})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	record := resp.Records[0]
	assert.Equal(t, RecordType("TXT"), record.Type)
	assert.Equal(t, "my-original-value", record.Content)
	assert.Equal(t, "created by porkbun-dyndns", record.Notes)
	assert.Equal(t, subdomain+"."+domain, record.Name)
	assert.Equal(t, "1200", record.Ttl)
	assert.Equal(t, "10", record.Priority)

	// update the record by id by not passing a TTL or the original name (subdomain)
	err = client.Dns.UpdateRecordById(context.TODO(), UpdateRecordByIdRequest{
		Domain:  domain,
		Id:      response.Id,
		Type:    "TXT",
		Content: "my-updated-value",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resp, err = client.Dns.GetRecordById(context.TODO(), GetRecordByIdRequest{domain, response.Id})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	record = resp.Records[0]
	assert.Equal(t, RecordType("TXT"), record.Type)
	assert.Equal(t, "my-updated-value", record.Content)
	assert.Equal(t, "created by porkbun-dyndns", record.Notes)
	// note how *some* values we originally created were overwritten with a default because they were omitted in the
	// update request, including the record name! (subdomain removed)
	assert.Equal(t, domain, record.Name)
	assert.Equal(t, "600", record.Ttl)
	assert.Equal(t, "0", record.Priority)
}

func TestDnsService_CreateRecord_WithInvalidType(t *testing.T) {
	client := NewClient(WithEnvironmentCredentials())

	_, err := client.Dns.CreateRecord(context.TODO(), CreateRecordRequest{
		Domain:  "rauschig.org",
		Name:    "test",
		Type:    "INVALID",
		Content: "127.0.0.1",
	})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	assert := assert.New(t)
	assert.Equal(ApiError{
		HTTPStatus: 400,
		Status:     "ERROR",
		Code:       "INVALID_TYPE",
		Message:    "Invalid type.",
	}, err)
}

func TestDnsService_CreateRecord_WithEmptyContent(t *testing.T) {
	client := NewClient(WithEnvironmentCredentials())

	_, err := client.Dns.CreateRecord(context.TODO(), CreateRecordRequest{
		Domain:  "rauschig.org",
		Name:    "test",
		Type:    "A",
		Content: "",
	})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	assert := assert.New(t)
	assert.Equal(ApiError{
		HTTPStatus: 400,
		Status:     "ERROR",
		Code:       "CREATE_ERROR_YOUR_DNS_RECORD_MUST_HAVE_AN_ANSWER",
		Message:    "Create error: Your DNS record must have an answer.",
	}, err)
}
