#!/usr/bin/env bash
# Stops every devnet validator process started by start.sh.
set -uo pipefail

OUTPUT_DIR="${OUTPUT_DIR:-$HOME/.lcv-devnet}"
PID_DIR="$OUTPUT_DIR/pids"

if [ ! -d "$PID_DIR" ]; then
  echo "No pidfiles under $PID_DIR — nothing to stop."
  exit 0
fi

for pidfile in "$PID_DIR"/*.pid; do
  [ -e "$pidfile" ] || continue
  pid="$(cat "$pidfile")"
  name="$(basename "$pidfile" .pid)"
  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid"
    echo "stopped $name (pid $pid)"
  else
    echo "$name (pid $pid) was not running"
  fi
  rm -f "$pidfile"
done
