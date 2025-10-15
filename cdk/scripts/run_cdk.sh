#!/usr/bin/env bash
set -euo pipefail

# Usage function
usage() {
    cat <<EOF
Usage: $0 <cdk-action> [options] [additional-cdk-args...]

Required Arguments:
    cdk-action              CDK command (deploy, synth, destroy, etc.)

Required Options:
    --image-tag <tag>       ECR image tag to deploy
    --stage <stage>         Deployment stage (test, play, stage, or prod)

Optional:
    --instance-ns <name>    Namespace for this instance of the stack (default: \${USER}-local when deploying from local)
    --commit-hash <hash>    Commit hash for tagging (default: auto-detect from git)
    --aws-profile <profile> AWS profile to use (default: test-admin for local, none for CI)
    --ci <enabled>          Running in CI mode (disables AWS_PROFILE, changes defaults)

Examples:
    # Local deployment
    $0 deploy --image-tag my-tag

    # CI deployment
    $0 deploy --image-tag my-tag --ci true --instance-ns my-branch --commit-hash abc123

    # Custom settings
    $0 deploy --image-tag my-tag --instance-ns custom --stage prod
EOF
    exit 1
}

# Parse required arguments
if [[ $# -lt 1 ]]; then
    usage
fi

CDK_ACTION="${1}"
shift 1

# Default values
CI_MODE=false
AWS_PROFILE_NAME="test-admin"
STAGE=""
INSTANCE_NS=""
COMMIT_HASH=""
IMAGE_TAG=""

# Parse optional arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --image-tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        --instance-ns)
            INSTANCE_NS="$2"
            shift 2
            ;;
        --commit-hash)
            COMMIT_HASH="$2"
            shift 2
            ;;
        --stage)
            STAGE="$2"
            shift 2
            ;;
        --aws-profile)
            AWS_PROFILE_NAME="$2"
            shift 2
            ;;
        --ci)
            CI_MODE="$2"
            shift 2
            ;;
        --help|-h)
            usage
            ;;
        *)
            # Remaining args are passed to CDK
            break
            ;;
    esac
done

# Apply defaults for missing/empty values when running locally
if [[ "$CI_MODE" == "false" ]]; then
    # Auto-detect commit hash if not provided
    if [[ -z "$COMMIT_HASH" ]]; then
        COMMIT_HASH=$(git describe --tags --dirty --long --always)
    fi

    # Namespace by the username
    if [[ -z "$INSTANCE_NS" ]]; then
        # Local dev: default to ${USER}-local
        INSTANCE_NS="${USER}-local"
    fi

    # Default stage to "test" if unset and not in CI mode
    if [[ -z "$STAGE" ]]; then
        STAGE="test"
    fi

    # Set AWS profile if not in CI mode
    export AWS_PROFILE="$AWS_PROFILE_NAME"
fi

echo "=== CDK Deployment Configuration ==="
echo "CDK Action:    $CDK_ACTION"
echo "Image Tag:     $IMAGE_TAG"
echo "Commit Hash:   $COMMIT_HASH"
echo "Stage:         $STAGE"
echo "Instance NS:   $INSTANCE_NS"
echo "CI Mode:       $CI_MODE"

if [[ "$CI_MODE" == "false" ]]; then
    echo "AWS Profile:   $AWS_PROFILE"
fi
echo "===================================="

# Build CDK arguments array
CDK_ARGS=(
    "$CDK_ACTION"
    -c "ecrImageTag=$IMAGE_TAG"
    -c "commit=$COMMIT_HASH"
    -c "stage=$STAGE"
    -c "instanceNs=$INSTANCE_NS"
)

# The CDK CLI has some builtin behaviors for CI mode,
# like sending logs to stdout instead of stderr.
if [[ "$CI_MODE" == "true" ]]; then
    CDK_ARGS+=(--ci)
    # Without this, we will get this error in CI:
    # "terminal (TTY) is not attached so we are unable to get a confirmation from the user"
    if [[ "$CDK_ACTION" == "destroy" ]]; then
        CDK_ARGS+=(--force)
    fi
fi

# Add deploy-specific flags
if [[ "$CDK_ACTION" == "deploy" ]]; then
    CDK_ARGS+=(--require-approval never)
    # Output stack outputs to file for downstream consumption
    CDK_ARGS+=(--outputs-file cdk-outputs.json)
    # No-rollback for test stage deployments
    if [[ "$STAGE" == "test" ]]; then
        CDK_ARGS+=(--no-rollback)
    fi
fi

# Pass through additional arguments
CDK_ARGS+=("$@")

pnpm cdk "${CDK_ARGS[@]}"
