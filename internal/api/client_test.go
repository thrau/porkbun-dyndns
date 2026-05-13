package api

import (
	"context"
	"encoding/json"
	"testing"
)

func TestPostJson_Ping(t *testing.T) {
	client := NewClient()

	var response map[string]interface{}
	err := client.PostJson(context.Background(), "/ping", nil, &response)
	if err != nil {
		t.Logf("Error: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")
	t.Logf("Response: %s", string(responseJSON))

	if status, ok := response["status"].(string); !ok || status != "SUCCESS" {
		t.Errorf("Expected status to be 'SUCCESS', got: %v", response["status"])
	}

}
