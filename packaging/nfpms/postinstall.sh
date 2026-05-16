#!/bin/sh
set -e

chown root:root /etc/porkbun-ddnsd/config.toml
chmod 600 /etc/porkbun-ddnsd/config.toml

systemctl daemon-reload || true
