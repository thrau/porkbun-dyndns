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

Show specific A records

```sh
porkbun-dns get-records --name example.com --type A
porkbun-dns get-records --name www.example.com --type A
```

Get a specific record by Porkbun record ID

```sh
porkbun-dns get-records --domain example.com --id 123456789
```

Update records by name and type (replaces *all* records of that type!).
Note that update-records will only succeed if the value is different from the current value.

```sh
porkbun-dns update-records --name www.example.com --type A --content 192.168.1.1
porkbun-dns update-records --name www.example.com --type A --content $(porkbun-dns myip)
porkbun-dns update-records --name www.example.com --type CNAME --content "srv.example.com" --ttl 3600 --notes "set by $(whoami) at $(date)"
```

### DynDNS Daemon

The daemon runs in the background and automatically updates DNS records based on configuration.

Run the daemon with a configuration file:

```sh
porkbun-ddnds --config /etc/porkbun-dyndns/daemon.conf
```

#### Systemd Service

TODO
