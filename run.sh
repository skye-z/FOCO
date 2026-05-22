#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../.env"

if [ -f "$ENV_FILE" ]; then
  set -a
  source "$ENV_FILE"
  set +a
fi

show_help() {
  echo "Usage: $0 --t <target>"
  echo ""
  echo "Targets:"
  echo "  1  Start admin frontend    (port 3001)"
  echo "  2  Start learner frontend  (port 3000)"
  echo "  3  Start backend API       (port 8080)"
  echo ""
  echo "Examples:"
  echo "  $0 --t 1        # start admin only"
  echo "  $0 --t 1 --t 3  # start admin + backend"
  echo "  $0 --t 1 --t 2 --t 3  # start all"
}

TARGETS=()

while [ $# -gt 0 ]; do
  case "$1" in
    --t)
      if [ -z "${2:-}" ]; then
        echo "Error: --t requires a value (1, 2, or 3)"
        show_help
        exit 1
      fi
      TARGETS+=("$2")
      shift 2
      ;;
    -h|--help)
      show_help
      exit 0
      ;;
    *)
      echo "Error: unknown option $1"
      show_help
      exit 1
      ;;
  esac
done

if [ ${#TARGETS[@]} -eq 0 ]; then
  show_help
  exit 1
fi

PIDS=()

cleanup() {
  echo ""
  echo "Stopping all processes..."
  for pid in "${PIDS[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  wait 2>/dev/null
  echo "All processes stopped."
}

trap cleanup EXIT INT TERM

ensure_node_deps() {
  local dir="$1"
  if [ ! -d "$dir/node_modules" ]; then
    echo "Installing dependencies for $dir..."
    (cd "$dir" && npm install --no-fund --no-audit)
  fi
}

for t in "${TARGETS[@]}"; do
  case "$t" in
    1)
      echo "==> Starting Admin Frontend (port 3001)..."
      ADMIN_DIR="$SCRIPT_DIR/frontend/admin"
      ensure_node_deps "$ADMIN_DIR"
      (cd "$ADMIN_DIR" && npx next dev -p 3001) &
      PIDS+=($!)
      ;;
    2)
      echo "==> Starting Learner Frontend (port 3000)..."
      LEARNER_DIR="$SCRIPT_DIR/frontend/learner"
      ensure_node_deps "$LEARNER_DIR"
      (cd "$LEARNER_DIR" && npx next dev -p 3000) &
      PIDS+=($!)
      ;;
    3)
      echo "==> Starting Backend API (port 8080)..."
      BACKEND_DIR="$SCRIPT_DIR/backend"
      echo "Building backend..."
      (cd "$BACKEND_DIR" && go build -o cmd/api/main ./cmd/api/)
      (cd "$BACKEND_DIR" && ./cmd/api/main) &
      PIDS+=($!)
      ;;
    *)
      echo "Error: unknown target '$t' (must be 1, 2, or 3)"
      exit 1
      ;;
  esac
done

echo ""
echo "All requested services are starting. Press Ctrl+C to stop all."
echo ""

wait
