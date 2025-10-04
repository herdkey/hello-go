#!/usr/bin/env bash
set -euo pipefail

# Expect dev to have a [test-admin] profile in ~/.aws/config and to have run `aws sso login`
export AWS_PROFILE="test-admin"

# Consuyme 2 required arguments, then shift the rest of the arguments to pass to CDK via $@
CDK_ACTION="${1:?Usage: $0 <cdk-action> <image-tag> [additional-cdk-args...]}"
IMAGE_TAG="${2:?Usage: $0 <cdk-action> <image-tag> [additional-cdk-args...]}"
shift 2

COMMIT_HASH="$(git describe --tags --dirty --long --always)"
NAMESPACE="${USER}-local"

echo "Running cdk $CDK_ACTION with image tag $IMAGE_TAG and commit hash $COMMIT_HASH"
echo "Namespace: $NAMESPACE"

pnpm cdk "$CDK_ACTION" \
  -c ecrImageTag="$IMAGE_TAG" \
  -c commitHash="$COMMIT_HASH" \
  -c stage="test" \
  -c isEphemeral=true \
  -c namespace="$NAMESPACE" \
  "$@"
