#!/usr/bin/env bash
set -euo pipefail

NODE_BIN="${NODE20_BIN:-/Users/zhaoguiyang/.nvm/versions/node/v20.12.0/bin/node}"

if [ ! -x "$NODE_BIN" ]; then
  NODE_BIN="$(command -v node)"
fi

exec "$NODE_BIN" "$@"
