package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thrau/porkbun-dyndns/internal/api"
)

func TestCheckWouldUpdate_NoChange(t *testing.T) {
	record := api.DNSRecord{
		Id:       "1",
		Name:     "www.example.com",
		Type:     "A",
		Content:  "1.2.3.4",
		Priority: "0",
		Ttl:      "600",
		Notes:    "my note",
	}
	req := api.UpdateRecordByIdRequest{
		Domain:   "example.com",
		Id:       1,
		Type:     "A",
		Content:  "1.2.3.4",
		Name:     api.String("www"),
		Priority: api.Int(0),
		Ttl:      api.Int(600),
		Notes:    api.String("my note"),
	}
	assert.False(t, CheckWouldUpdate(record, req))
}

func TestCheckWouldUpdate_ContentChanged(t *testing.T) {
	record := api.DNSRecord{
		Id:       "1",
		Name:     "www.example.com",
		Type:     "A",
		Content:  "1.2.3.4",
		Priority: "0",
		Ttl:      "600",
		Notes:    "my note",
	}
	req := api.UpdateRecordByIdRequest{
		Domain:   "example.com",
		Id:       1,
		Type:     "A",
		Content:  "1.2.3.5",
		Name:     api.String("www"),
		Priority: api.Int(0),
		Ttl:      api.Int(600),
		Notes:    api.String("my note"),
	}
	assert.True(t, CheckWouldUpdate(record, req))
}

func TestCheckWouldUpdate_TtlChanged(t *testing.T) {
	record := api.DNSRecord{
		Id:       "1",
		Name:     "www.example.com",
		Type:     "A",
		Content:  "1.2.3.4",
		Priority: "0",
		Ttl:      "600",
		Notes:    "my note",
	}
	req := api.UpdateRecordByIdRequest{
		Domain:   "example.com",
		Id:       1,
		Type:     "A",
		Content:  "1.2.3.4",
		Name:     api.String("www"),
		Priority: api.Int(0),
		Ttl:      api.Int(800),
		Notes:    api.String("my note"),
	}
	assert.True(t, CheckWouldUpdate(record, req))
}
