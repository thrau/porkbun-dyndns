package daemon

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thrau/porkbun-dyndns/internal/api"
)

// Mocks

type mockUtilService struct {
	ipResp api.IpResponse
	ipErr  error
}

func (m *mockUtilService) Ip(_ context.Context) (api.IpResponse, error) {
	return m.ipResp, m.ipErr
}

type mockDNSService struct {
	getResp      api.RetrieveRecordsResponse
	getErr       error
	createResp   api.CreateRecordResponse
	createErr    error
	updateErr    error
	createCalled bool
	updateCalled bool
	createReq    api.CreateRecordRequest
	updateReq    api.UpdateRecordsRequest
	getCallCount atomic.Int64
}

func (m *mockDNSService) ListRecords(ctx context.Context, domain string) (api.RetrieveRecordsResponse, error) {
	panic("not implemented")
}

func (m *mockDNSService) GetRecords(ctx context.Context, req api.GetRecordsRequest) (api.RetrieveRecordsResponse, error) {
	m.getCallCount.Add(1)
	return m.getResp, m.getErr
}

func (m *mockDNSService) GetRecordById(ctx context.Context, req api.GetRecordByIdRequest) (api.RetrieveRecordsResponse, error) {
	panic("not implemented")
}

func (m *mockDNSService) CreateRecord(ctx context.Context, req api.CreateRecordRequest) (api.CreateRecordResponse, error) {
	m.createCalled = true
	m.createReq = req
	return m.createResp, m.createErr
}

func (m *mockDNSService) UpdateRecords(ctx context.Context, req api.UpdateRecordsRequest) error {
	m.updateCalled = true
	m.updateReq = req
	return m.updateErr
}

func (m *mockDNSService) UpdateRecordById(ctx context.Context, req api.UpdateRecordByIdRequest) error {
	panic("not implemented")
}

func (m *mockDNSService) DeleteRecord(ctx context.Context, req api.DeleteRecordRequest) error {
	panic("not implemented")
}

func (m *mockDNSService) DeleteRecordById(ctx context.Context, req api.DeleteRecordByIdRequest) error {
	panic("not implemented")
}

func newConfig() *Config {
	return &Config{
		Domain:    "example.com",
		Subdomain: "www",
		Interval:  time.Minute,
	}
}

func TestSyncRecord_IpError(t *testing.T) {
	ctx := t.Context()
	util := &mockUtilService{ipErr: errors.New("network down")}
	dns := &mockDNSService{}
	d := NewDaemon(dns, util, newConfig())

	err := d.syncRecord(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get IP")
}

func TestSyncRecord_GetRecordsError(t *testing.T) {
	ctx := t.Context()
	util := &mockUtilService{ipResp: api.IpResponse{YourIp: "1.2.3.4"}}
	dns := &mockDNSService{getErr: errors.New("dns lookup failed")}
	d := NewDaemon(dns, util, newConfig())

	err := d.syncRecord(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get DNS record")
}

func TestSyncRecord_NoRecords_CreateSuccess(t *testing.T) {
	ctx := t.Context()
	ip := "1.2.3.4"
	util := &mockUtilService{ipResp: api.IpResponse{YourIp: ip}}
	dns := &mockDNSService{
		getResp: api.RetrieveRecordsResponse{Status: "SUCCESS", Records: []api.DNSRecord{}},
	}
	config := newConfig()
	d := NewDaemon(dns, util, config)

	err := d.syncRecord(ctx)

	assert.NoError(t, err)
	assert.True(t, dns.createCalled)
	assert.Equal(t, config.Domain, dns.createReq.Domain)
	assert.Equal(t, config.Subdomain, dns.createReq.Name)
	assert.Equal(t, ip, dns.createReq.Content)
}

func TestSyncRecord_NoRecords_CreateError(t *testing.T) {
	ctx := t.Context()
	util := &mockUtilService{ipResp: api.IpResponse{YourIp: "1.2.3.4"}}
	dns := &mockDNSService{
		getResp:   api.RetrieveRecordsResponse{Status: "SUCCESS", Records: []api.DNSRecord{}},
		createErr: errors.New("creation failed"),
	}
	d := NewDaemon(dns, util, newConfig())

	err := d.syncRecord(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create DNS record")
}

func TestSyncRecord_MultipleRecords(t *testing.T) {
	ctx := t.Context()
	util := &mockUtilService{ipResp: api.IpResponse{YourIp: "1.2.3.4"}}
	dns := &mockDNSService{
		getResp: api.RetrieveRecordsResponse{
			Status: "SUCCESS",
			Records: []api.DNSRecord{
				{Content: "1.1.1.1"},
				{Content: "2.2.2.2"},
			},
		},
	}
	d := NewDaemon(dns, util, newConfig())

	err := d.syncRecord(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 1 DNS record, got 2")
}

func TestSyncRecord_RecordUpToDate(t *testing.T) {
	ctx := t.Context()
	ip := "1.2.3.4"
	util := &mockUtilService{ipResp: api.IpResponse{YourIp: ip}}
	dns := &mockDNSService{
		getResp: api.RetrieveRecordsResponse{
			Status: "SUCCESS",
			Records: []api.DNSRecord{
				{Content: ip},
			},
		},
	}
	d := NewDaemon(dns, util, newConfig())

	err := d.syncRecord(ctx)

	assert.NoError(t, err)
	assert.False(t, dns.updateCalled)
}

func TestSyncRecord_RecordOutdated_UpdateSuccess(t *testing.T) {
	ctx := t.Context()
	newIp := "1.2.3.4"
	oldIp := "1.1.1.1"
	util := &mockUtilService{ipResp: api.IpResponse{YourIp: newIp}}
	dns := &mockDNSService{
		getResp: api.RetrieveRecordsResponse{
			Status: "SUCCESS",
			Records: []api.DNSRecord{
				{Content: oldIp, Ttl: "300", Priority: "10"},
			},
		},
	}
	config := newConfig()
	d := NewDaemon(dns, util, config)

	err := d.syncRecord(ctx)

	assert.NoError(t, err)
	assert.True(t, dns.updateCalled)
	assert.Equal(t, config.Domain, dns.updateReq.Domain)
	assert.Equal(t, config.Subdomain, dns.updateReq.Subdomain)
	assert.Equal(t, newIp, dns.updateReq.Content)
	assert.Equal(t, 300, *dns.updateReq.Ttl)
	assert.Equal(t, 10, *dns.updateReq.Priority)
}

func TestSyncRecord_RecordOutdated_UpdateError(t *testing.T) {
	ctx := t.Context()
	newIp := "1.2.3.4"
	oldIp := "1.1.1.1"
	util := &mockUtilService{ipResp: api.IpResponse{YourIp: newIp}}
	dns := &mockDNSService{
		getResp: api.RetrieveRecordsResponse{
			Status: "SUCCESS",
			Records: []api.DNSRecord{
				{Content: oldIp},
			},
		},
		updateErr: errors.New("update failed"),
	}
	d := NewDaemon(dns, util, newConfig())

	err := d.syncRecord(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update DNS record")
}
func TestConfig_GetRecordFqdn(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		subdomain string
		expected  string
	}{
		{"domain only", "example.com", "", "example.com"},
		{"subdomain set", "example.com", "www", "www.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Domain: tt.domain, Subdomain: tt.subdomain}
			assert.Equal(t, tt.expected, c.GetRecordFqdn())
		})
	}
}

func TestToOptionalInt(t *testing.T) {
	tests := []struct {
		input    string
		expected *int
	}{
		{"", nil},
		{"abc", nil},
		{"300", intPtr(300)},
		{"0", intPtr(0)},
		{"-1", intPtr(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, toOptionalInt(tt.input))
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func TestDaemon_Close_StopsRun(t *testing.T) {
	ctx := t.Context()
	util := &mockUtilService{ipResp: api.IpResponse{YourIp: "1.2.3.4"}}
	dns := &mockDNSService{
		getResp: api.RetrieveRecordsResponse{Status: "SUCCESS", Records: []api.DNSRecord{{Content: "1.2.3.4"}}},
	}
	config := &Config{Interval: time.Hour} // long interval to avoid noise
	d := NewDaemon(dns, util, config)

	done := make(chan struct{})
	go func() {
		d.Run(ctx)
		close(done)
	}()

	d.Close()

	select {
	case <-done:
		// success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Daemon.Run did not stop after Close()")
	}
}

func TestDaemon_Run_EventLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	util := &mockUtilService{ipResp: api.IpResponse{YourIp: "1.2.3.4"}}
	dns := &mockDNSService{
		getResp: api.RetrieveRecordsResponse{Status: "SUCCESS", Records: []api.DNSRecord{{Content: "1.2.3.4"}}},
	}
	interval := 50 * time.Millisecond
	config := &Config{Interval: interval}
	d := NewDaemon(dns, util, config)

	go d.Run(ctx)

	waitDuration := 275 * time.Millisecond
	time.Sleep(waitDuration)
	cancel()

	expected := int64(waitDuration / interval)
	actual := dns.getCallCount.Load()
	delta := int64(1)
	assert.InDelta(t, expected, actual, float64(delta))
}

func TestDaemon_ContextCancel_StopsRun(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	util := &mockUtilService{ipResp: api.IpResponse{YourIp: "1.2.3.4"}}
	dns := &mockDNSService{
		getResp: api.RetrieveRecordsResponse{Status: "SUCCESS", Records: []api.DNSRecord{{Content: "1.2.3.4"}}},
	}
	config := &Config{Interval: time.Hour}
	d := NewDaemon(dns, util, config)

	done := make(chan struct{})
	go func() {
		d.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Daemon.Run did not stop after context cancellation")
	}
}
