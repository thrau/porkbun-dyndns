package daemon

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/thrau/porkbun-dyndns/internal/api"
)

type Daemon struct {
	dns       api.DNSService
	util      api.UtilService
	config    *Config
	closeChan chan struct{}
}

type Config struct {
	Domain    string
	Subdomain string
	Interval  time.Duration
}

func (c *Config) GetRecordFqdn() string {
	if c.Subdomain == "" {
		return c.Domain
	}
	return c.Subdomain + "." + c.Domain
}

func NewDaemon(dns api.DNSService, util api.UtilService, config *Config) *Daemon {
	return &Daemon{
		dns:       dns,
		util:      util,
		config:    config,
		closeChan: make(chan struct{}),
	}
}

func (d *Daemon) syncRecord(ctx context.Context) error {
	ip, err := d.util.Ip(ctx)
	if err != nil {
		return fmt.Errorf("failed to get IP: %w", err)
	}

	resp, err := d.dns.GetRecords(ctx, api.GetRecordsRequest{
		Domain:    d.config.Domain,
		Subdomain: d.config.Subdomain,
		Type:      "A",
	})

	if err != nil {
		return fmt.Errorf("failed to get DNS record %s.%s: %w", d.config.Subdomain, d.config.Domain, err)
	}

	// case: new record is needed
	if len(resp.Records) == 0 {
		fmt.Printf(
			"[%s] creating a new DNS record: %s A %s\n", time.Now().Format(time.RFC3339),
			d.config.GetRecordFqdn(),
			ip.YourIp,
		)

		_, err := d.dns.CreateRecord(ctx, api.CreateRecordRequest{
			Domain:  d.config.Domain,
			Name:    d.config.Subdomain,
			Type:    "A",
			Content: ip.YourIp,
			Notes:   api.String("updated by porkbun-ddnsd at " + time.Now().Format(time.RFC3339)),
		})
		if err != nil {
			return fmt.Errorf("failed to create DNS record %s: %w", d.config.GetRecordFqdn(), err)
		}
		return nil
	}

	// case: more than one record exist that would be overwritten
	if len(resp.Records) > 1 {
		// TODO: handle multiple records. this could be useful to support load balancing, but will require different
		//  configuration approach
		return fmt.Errorf(
			"expected 1 DNS record, got %d; skipping update since we would overwrite all existing records",
			len(resp.Records),
		)
	}
	record := resp.Records[0]

	if record.Content == ip.YourIp {
		// skip, since record already exists with the same IP
		return nil
	}

	request := api.UpdateRecordsRequest{
		Domain:    d.config.Domain,
		Subdomain: d.config.Subdomain,
		Type:      "A",
		Content:   ip.YourIp,
		Ttl:       toOptionalInt(record.Ttl),
		Priority:  toOptionalInt(record.Priority),
		Notes:     api.String("updated by porkbun-ddnsd at " + time.Now().Format(time.RFC3339)),
	}

	err = d.dns.UpdateRecords(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to update DNS record %s: %w", d.config.GetRecordFqdn(), err)
	}
	fmt.Printf("[%s] updated DNS record: %s A %s\n", time.Now().Format(time.RFC3339), d.config.GetRecordFqdn(), ip.YourIp)
	return nil
}

func (d *Daemon) Run(ctx context.Context) {
	// run once immediately (otherwise we would wait for the first tick)
	err := d.syncRecord(ctx)
	if err != nil {
		fmt.Printf("[%s] error: failed to sync DNS record: %v\n", time.Now().Format(time.RFC3339), err)
	}

	// now, start the loop
	ticker := time.NewTicker(d.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := d.syncRecord(ctx)
			if err != nil {
				fmt.Printf("[%s] error: failed to sync DNS record: %v\n", time.Now().Format(time.RFC3339), err)
			}
		case <-d.closeChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (d *Daemon) Close() {
	close(d.closeChan)
}

func toOptionalInt(s string) *int {
	if s == "" {
		return nil
	}
	if i, err := strconv.Atoi(s); err == nil {
		return &i
	}
	return nil
}
