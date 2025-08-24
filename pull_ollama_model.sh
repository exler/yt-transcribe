#!/usr/bin/env bash
set -euo pipefail

# Pulls an Ollama model via REST. Defaults: base URL http://localhost:11434, model phi3:mini.
usage() {
  cat <<'EOF'
Usage: pull_ollama_model.sh [-u BASE_URL] [-m MODEL]

Options:
  -u, --base-url   Ollama base URL (default: http://localhost:11434)
  -m, --model      Model name:tag to pull (default: phi3:mini)
  -h, --help       Show this help

Environment:
  OLLAMA_BASE_URL  Base URL for Ollama (overrides default)
  OLLAMA_MODEL     Model name:tag to pull (overrides default)

Examples:
  ./pull_ollama_model.sh
  ./pull_ollama_model.sh -m phi3
  OLLAMA_BASE_URL=http://ollama:11434 ./pull_ollama_model.sh -m phi3:mini
EOF
}

# Defaults with env overrides
BASE_URL="${OLLAMA_BASE_URL:-${BASE_URL:-http://localhost:11434}}"
MODEL="${OLLAMA_MODEL:-${MODEL:-phi3:mini}}"

# Parse flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    -u|--base-url)
      [[ $# -ge 2 ]] || { echo "Missing value for $1" >&2; usage; exit 2; }
      BASE_URL="$2"; shift 2 ;;
    -m|--model)
      [[ $# -ge 2 ]] || { echo "Missing value for $1" >&2; usage; exit 2; }
      MODEL="$2"; shift 2 ;;
    -h|--help)
      usage; exit 0 ;;
    *)
      echo "Unknown argument: $1" >&2; usage; exit 2 ;;
  esac
done

# Basic checks
if ! command -v curl >/dev/null 2>&1; then
  echo "Error: curl is required but not installed." >&2
  exit 127
fi

PULL_URL="${BASE_URL%/}/api/pull"

# Informational log to stderr so stdout can carry streaming JSON
echo "Pulling model '${MODEL}' from ${BASE_URL} ..." >&2

# Execute pull (streaming JSON lines)
set +e
curl -sS -N -X POST "$PULL_URL" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${MODEL}\"}"
status=$?
set -e

if [[ $status -ne 0 ]]; then
  echo >&2
  echo "Pull failed (exit $status). URL: $PULL_URL" >&2
  exit $status
fi
