package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/thrau/porkbun-dyndns/internal/api"
)

type Daemon struct {
	dns       api.DNSService
	util      api.UtilService
	config    *Config
	logger    *slog.Logger
	closeChan chan struct{}
}

type Config struct {
	Domain    string
	Subdomain string
	Interval  time.Duration
}

func (c *Config) String() string {
	return fmt.Sprintf("DaemonConfig{Domain: %s, Subdomain: %s, Interval: %s}", c.Domain, c.Subdomain, c.Interval)
}

func (c *Config) GetRecordFqdn() string {
	if c.Subdomain == "" {
		return c.Domain
	}
	return c.Subdomain + "." + c.Domain
}

func NewDaemon(dns api.DNSService, util api.UtilService, config *Config, logger *slog.Logger) *Daemon {
	return &Daemon{
		dns:       dns,
		util:      util,
		config:    config,
		logger:    logger,
		closeChan: make(chan struct{}),
	}
}

type SyncResult struct {
	ActiveIpAddress   string
	RecordedIpAddress string
	Changed           bool
}

func (d *Daemon) syncRecord(ctx context.Context) (result SyncResult, err error) {
	ip, err := d.util.Ip(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get IP: %w", err)
		return
	}
	result.ActiveIpAddress = ip.YourIp

	d.logger.Debug("IP address retrieved", "ip", ip.YourIp)
	recordType := api.RecordType("A")

	resp, err := d.dns.GetRecords(ctx, api.GetRecordsRequest{
		Domain:    d.config.Domain,
		Subdomain: d.config.Subdomain,
		Type:      recordType,
	})

	if err != nil {
		err = fmt.Errorf("failed to get DNS record %s.%s: %w", d.config.Subdomain, d.config.Domain, err)
		return
	}

	// case: new record is needed
	if len(resp.Records) == 0 {
		d.logger.Info(
			"creating a new DNS record",
			"fqdn", d.config.GetRecordFqdn(),
			"type", recordType,
			"content", ip.YourIp,
		)

		_, err = d.dns.CreateRecord(ctx, api.CreateRecordRequest{
			Domain:  d.config.Domain,
			Name:    d.config.Subdomain,
			Type:    recordType,
			Content: ip.YourIp,
			Notes:   api.String("updated by porkbun-ddnsd at " + time.Now().Format(time.RFC3339)),
		})
		if err != nil {
			err = fmt.Errorf("failed to create DNS record %s: %w", d.config.GetRecordFqdn(), err)
			return
		}
		return
	}

	// case: more than one record exist that would be overwritten
	if len(resp.Records) > 1 {
		// TODO: handle multiple records. this could be useful to support load balancing, but will require different
		//  configuration approach
		err = fmt.Errorf(
			"expected 1 DNS record, got %d; skipping update since we would overwrite all existing records",
			len(resp.Records),
		)
		return
	}
	record := resp.Records[0]
	result.RecordedIpAddress = record.Content
	d.logger.Debug("current DNS record returned", "fqdn", d.config.GetRecordFqdn(), "type", recordType, "content", ip.YourIp)

	if record.Content == ip.YourIp {
		// skip, since record already exists with the same IP
		d.logger.Debug("DNS record already up to date")
		return
	}

	request := api.UpdateRecordsRequest{
		Domain:    d.config.Domain,
		Subdomain: d.config.Subdomain,
		Type:      recordType,
		Content:   ip.YourIp,
		Ttl:       toOptionalInt(record.Ttl),
		Priority:  toOptionalInt(record.Priority),
		Notes:     api.String("updated by porkbun-ddnsd at " + time.Now().Format(time.RFC3339)),
	}

	err = d.dns.UpdateRecords(ctx, request)
	if err != nil {
		err = fmt.Errorf("failed to update DNS record %s: %w", d.config.GetRecordFqdn(), err)
		return
	}
	result.Changed = true
	d.logger.Info("updated DNS record", "fqdn", d.config.GetRecordFqdn(), "type", recordType, "content", ip.YourIp)
	return
}

func (d *Daemon) Run(ctx context.Context) {
	// run once immediately (otherwise we would wait for the first tick)
	d.logger.Info("running initial sync")
	result, err := d.syncRecord(ctx)
	if err != nil {
		d.logger.Error("failed to sync DNS record", "error", err)
	} else if result.Changed {
		d.logger.Info("initial sync completed, IP address was updated", "previous", result.RecordedIpAddress, "new", result.ActiveIpAddress)
	} else {
		d.logger.Info("initial sync completed, IP address is already up to date", "ip", result.ActiveIpAddress)
	}

	d.logger.Info(fmt.Sprintf("starting daemon loop, next sync at %s", time.Now().Add(d.config.Interval)))
	// now, start the loop
	ticker := time.NewTicker(d.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := d.syncRecord(ctx)
			if err != nil {
				d.logger.Error("failed to sync DNS record", "error", err)
			}
			break
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
