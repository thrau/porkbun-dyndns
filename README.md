Porkbun DynDNS: DNS utilities for Porkbun
=========================================

Client utilities for the [Porkbun](https://porkbun.com/) DNS API.

It provides
* `porkbun-dns` a command-line interface for basic DNS record management via the Porkbun API.
* `porkbun-ddnsd` a Dynamic DNS daemon that automatically keeps your DNS records synchronized with your current IP address.

## Installation

* Download the latest release from the [releases page](https://github.com/porkbun/porkbun-dyndns/releases).
* Unpack the archive and place the `porkbun-dns` and/or `porkbun-ddnsd` binary into your path.
* Make sure API access is enabled for your domain. 
* Create an API key, you can do this in the [account API settings](https://porkbun.com/account/api).

## Use cases

### Cron script to update DNS record with the current IP address

If you prefer a manual approach to Dynamic DNS over using the daemon, you can use a cron job to update your DNS records with the current IP address.
First, unless the record already exists, create a record in Porkbun for your domain with your current IP address:

```sh
porkbun-dns create-record --name dyndns.example.com --type A --content $(porkbun-dns myip)
```

Now, you can use a cron job to update the record with the current IP address every 15 minutes.
Add a note to make sure the record is updated correctly (see update-records help for more details).
```sh
crontab -e
```

Add:
```
*/15 * * * * porkbun-dns update-records --name dyndns.example.com --type A --content $(porkbun-dns myip) --notes "last updated at $(date)"
```

### Systemd service to keep DNS record up to date

You can add the DynDNS daemon as a system service to run in the background.
See the [DynDNS Daemon](#dyndns-daemon) section below.

## Usage

### `porkbun-dns` CLI

Porkbun DynDNS provides a CLI for basic DNS record management via the Porkbun API.
You can use it to build your own custom scripts or cron jobs to update your DNS records as needed.

Set your API key and secret key in your environment:

```sh
export PORKBUN_API_KEY=pk1_...
export PORKBUN_SECRET_KEY=sk1_...
```

You may see the following error if you don't have the API keys set correctly:
```
Error: porkbun api error (400): status=ERROR code=API_KEY_REQUIRED message=All API requests require an API key or API token.
```

Show your IP address:

```sh
porkbun-dns myip
```

List all domain records via the API:

```sh
porkbun-dns list-records --domain example.com
```
Might return:
```json
[
  {
    "id": "123456789",
    "name": "srv.example.com",
    "type": "A",
    "content": "127.0.0.1",
    "ttl": "600",
    "prio": "0",
    "notes": ""
  },
  ...
}
```

Show specific A records

```sh
porkbun-dns get-records --name example.com --type A
porkbun-dns get-records --name www.example.com --type A
```

Get a specific record by Porkbun record ID

```sh
porkbun-dns get-record --domain example.com --id 123456789
```

Create a new DNS record (prints the new record ID on success):

```sh
porkbun-dns create-record --name srv.example.com --type A --content $(porkbun-dns myip)
porkbun-dns create-record --name _dmarc.example.com --type TXT --content "v=DMARC1; p=none" --ttl 3600
```

Update records by name and type (replaces *all* records of that type!).
Note that update-records will only succeed if the record would change, and will otherwise return an `EDIT_ERROR_WE_WERE_UNABLE_TO_EDIT_THE_DNS_RECORD` error.
You can work around this by using either the `update-record` method (which checks this explicitly, but requires the record ID), or update the `--notes` field of the record to include the current time.  

```sh
porkbun-dns update-records --name www.example.com --type CNAME --content srv.example.com --ttl 3600
porkbun-dns update-records --name www.example.com --type A --content $(porkbun-dns myip) --notes "updated at $(date)"
```

Update a specific record by its Porkbun record ID by merging the new values with the existing ones.
Note that this method behaves differently from the underlying API method, which would replace the existing record entirely.

```sh
# suppose this is the ID of a TXT record for _acme-challenge.example.com
porkbun-dns update-record --domain example.com --id 123456789 --content "my-new-dns-challenge"
# you can also completely change the record type and content of a record
porkbun-dns update-record --domain example.com --id 123456789 --type A --content "127.0.0.1" --ttl 600
```

Delete a specific record by its Porkbun record ID:

```sh
porkbun-dns delete-record --domain example.com --id 123456789
```

Delete all records matching a name and type:

```sh
porkbun-dns delete-records --name srv.example.com --type A
porkbun-dns delete-records --name _acme-challenge.example.com --type TXT
```

### DynDNS Daemon

The daemon runs in the background and automatically updates DNS records based on configuration.

Run the daemon with a specific configuration file:

```sh
porkbun-ddnds --config /etc/porkbun-ddnsd/config.toml
```

You should see something like
```
[2026-05-16T02:00:21+02:00] loading daemon config from /etc/porkbun-ddnsd/config.toml
[2026-05-16T02:00:24+02:00] updated DNS record: mydomain.example.org A 12.34.56.78
```

Here's an example configuration file. Make sure the config file has limited permissions to keep secrets safe.

```toml
# name of the DNS record to update
name = "mymachine.example.com"
# interval between updates (in Go time.Duration format)
interval = "5m"

# porkbun API credentials
api_key = "pk1_..."
secret_key = "sk1_..."
```

#### Systemd Service

You can run the DynDNS daemon as a systemd service to ensure it runs automatically at boot and restarts if it fails.

Create a systemd service file at `/etc/systemd/system/porkbun-ddnsd.service`:

```ini
[Unit]
Description=Porkbun Dynamic DNS Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/porkbun-ddnsd --config /etc/porkbun-ddnsd/config.toml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Create your configuration file at `/etc/porkbun-ddnsd/config.toml` and set restrictive permissions:

```sh
sudo mkdir -p /etc/porkbun-ddnsd
sudo nano /etc/porkbun-ddnsd/config.toml
sudo chown root:porkbun-ddns /etc/porkbun-ddnsd/config.toml
sudo chmod 600 /etc/porkbun-ddnsd/config.toml
```

Enable and start the service:

```sh
sudo systemctl daemon-reload
sudo systemctl enable porkbun-ddnsd
sudo systemctl start porkbun-ddnsd
```

Check the service status:

```sh
sudo systemctl status porkbun-ddnsd
```

View logs:

```sh
sudo journalctl -u porkbun-ddnsd -f
```
