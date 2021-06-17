#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

# Lookup the current referee assignment
python "${SCRIPT_DIR}/update_referee_caddy_passwords.py" 

# Update the caddy config file with the new list of passwords
"${SCRIPT_DIR}"/../caddy/generate_caddyfile.py

# Reload the caddy webserver
"${SCRIPT_DIR}"/../caddy/update_caddy_config.sh