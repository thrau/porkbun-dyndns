package api

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
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
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected ErrorResponse, got %T: %v", err, err)
	}

	if apiErr.HTTPStatus != 400 {
		t.Errorf("Expected HTTPStatus 400, got %d", apiErr.HTTPStatus)
	}

	if apiErr.Status != "ERROR" {
		t.Errorf("Expected Status 'ERROR', got '%s'", apiErr.Status)
	}

	if apiErr.Code != "API_KEY_REQUIRED" {
		t.Errorf("Expected Code 'API_KEY_REQUIRED', got '%s'", apiErr.Code)
	}

	if apiErr.Message != "All API requests require an API key or API token." {
		t.Errorf("Expected Message 'All API requests require an API key or API token.', got '%s'", apiErr.Message)
	}
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
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected ErrorResponse, got %T: %v", err, err)
	}

	if apiErr.HTTPStatus != 400 {
		t.Errorf("Expected HTTPStatus 400, got %d", apiErr.HTTPStatus)
	}

	if apiErr.Status != "ERROR" {
		t.Errorf("Expected Status 'ERROR', got '%s'", apiErr.Status)
	}

	if apiErr.Code != "INVALID_DOMAIN" {
		t.Errorf("Expected Code 'INVALID_DOMAIN', got '%s'", apiErr.Code)
	}

	if apiErr.Message != "Invalid domain." {
		t.Errorf("Expected Message 'Invalid domain.', got '%s'", apiErr.Message)
	}

}
