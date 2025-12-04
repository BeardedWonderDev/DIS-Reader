#!/bin/sh
set -e
id -u disagent >/dev/null 2>&1 || useradd -r -s /sbin/nologin disagent
systemctl daemon-reload || true
systemctl enable dis-agent.service || true
systemctl restart dis-agent.service || true
