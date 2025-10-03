#!/usr/bin/env bash
set -euo pipefail

# Usage: ./deploy-ephemeral.sh <namespace> [pr] [sha] [expires_at]
# Example: ./deploy-ephemeral.sh my-feature 123 abc1234 2025-10-10T00:00:00Z

#if [ $# -lt 1 ]; then
#  echo "Usage: $0 <namespace> [pr] [sha] [expires_at]"
#  echo "Example: $0 my-feature 123 abc1234 2025-10-10T00:00:00Z"
#  exit 1
#fi

INFRA_ACCOUNT_ID="073835883885"

#NAMESPACE="$1"
NAMESPACE="fcb8742"
#PR="${2:-}"
PR="32"
#SHA="${3:-}"
SHA="$NAMESPACE"

if date --version >/dev/null 2>&1; then
  # GNU date (Linux)
  EXPIRES_AT=$(date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ")
else
  # BSD date (macOS)
  EXPIRES_AT=$(date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ")
fi
#EXPIRES_AT="${4:-}"

# Build context arguments
CONTEXT_ARGS=(
  -c "stage=test"
  -c "ecr_image_uri=${INFRA_ACCOUNT_ID}.dkr.ecr.us-west-2.amazonaws.com/test/hello-go/lambda:${NAMESPACE}"
  -c "namespace=$NAMESPACE"
)

if [ -n "$PR" ]; then
  CONTEXT_ARGS+=(-c "pr=$PR")
fi

if [ -n "$SHA" ]; then
  CONTEXT_ARGS+=(-c "sha=$SHA")
fi

if [ -n "$EXPIRES_AT" ]; then
  CONTEXT_ARGS+=(-c "expires_at=$EXPIRES_AT")
fi

# Run CDK deploy with context
exec pnpx cdk deploy "${CONTEXT_ARGS[@]}" "$@"
