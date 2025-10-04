#!/usr/bin/env bash
set -euo pipefail

# Expect dev to have a [test-admin] profile in ~/.aws/config and to have run `aws sso login`
export AWS_PROFILE="test-admin"

IMAGE_TAG="${1:?Usage: $0 <image-tag>}"
COMMIT_HASH="$(git describe --tags --dirty --long --always)"
NAMESPACE="${USER}-local"

echo "Deploying test stack with image tag $IMAGE_TAG and commit hash $COMMIT_HASH"
echo "Namespace: $NAMESPACE"

pnpm cdk deploy \
  -c ecrImageTag="$IMAGE_TAG" \
  -c commitHash="$COMMIT_HASH" \
  -c stage="test" \
  -c isEphemeral=true \
  -c namespace="$NAMESPACE"
