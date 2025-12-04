#!/bin/sh
set -e
systemctl stop dis-agent.service 2>/dev/null || true
systemctl disable dis-agent.service 2>/dev/null || true
