# hello-go CDK Infrastructure

CDK v2 TypeScript infrastructure for deploying the **hello-go** containerized Lambda behind an HTTP API Gateway.

## Architecture

- **Lambda Function**: Container-based Lambda from ECR image (256 MB, 10s timeout)
- **HTTP API Gateway**: Single proxy route (`ANY /{proxy+}`) with permissive CORS
- **IAM Role**: Dedicated Lambda execution role with minimal permissions (CloudWatch Logs)
- **CloudWatch Logs**: Environment-aware retention (7 days test, 30 days live)

## Prerequisites

1. **AWS CLI** configured with credentials
2. **Volta** for Node.js version management
3. **pnpm** package manager
4. **ECR image** already pushed to ECR
5. **CDK bootstrapped** in target account/region

## Installation

```bash
cd infra/cdk
pnpm install
```

Node.js version (20.18.1) and pnpm version (9.15.0) are managed by Volta as specified in `package.json`.

## Bootstrap (first time only)

```bash
pnpx cdk bootstrap
```

## Deployment Modes

### Test Environment (Ephemeral, Namespaced)

Deploy ephemeral stacks with `RemovalPolicy.DESTROY` for automatic cleanup:

```bash
pnpx cdk deploy \
  -c stage=test \
  -c namespace=pr-123 \
  -c pr=123 \
  -c sha=abc1234 \
  -c expires_at=2025-10-10T00:00:00Z \
  -c ecr_image_uri=123456789012.dkr.ecr.us-east-1.amazonaws.com/hello-go:abc1234
```

**Context Parameters**:
- `stage`: `test` (required)
- `namespace`: Unique identifier (e.g., `pr-123`) for stack isolation
- `pr`: Pull request number (optional, for tagging)
- `sha`: Git commit SHA (optional, for tagging)
- `expires_at`: ISO8601 expiration timestamp (optional, for tagging)
- `ecr_image_uri`: Full ECR image URI with tag/digest

**Tags Applied**:
```
ephemeral=true
svc=hello-go
stage=test
pr=123
sha=abc1234
expires_at=2025-10-10T00:00:00Z
```

**Stack Name**: `HelloGo-pr-123`

**Behavior**:
- Resources use `RemovalPolicy.DESTROY` (auto-cleanup on stack deletion)
- CloudWatch Logs: 7-day retention
- One stack per namespace (reused across commits in same PR)

### Live Environments (Stable, Permanent)

Deploy stable stacks with permanent resources:

```bash
pnpx cdk deploy \
  -c stage=prod \
  -c ecr_image_uri=123456789012.dkr.ecr.us-east-1.amazonaws.com/hello-go:v1.2.3
```

**Context Parameters**:
- `stage`: `play`, `stage`, or `prod`
- `ecr_image_uri`: Full ECR image URI with tag/digest

**Tags Applied**:
```
svc=hello-go
stage=prod
```

**Stack Name**: `HelloGoStack`

**Behavior**:
- Resources use `RemovalPolicy.RETAIN` (persisted on stack deletion)
- CloudWatch Logs: 30-day retention
- Single stable stack per environment
- No ephemeral or TTL tags

## Outputs

After deployment, the stack outputs:

- **ApiBaseUrl**: HTTP API Gateway base URL (use for curl)
- **LambdaArn**: Lambda function ARN
- **LogGroupName**: CloudWatch Log Group name
- **LambdaRoleArn**: Lambda execution role ARN

View outputs:

```bash
pnpx cdk deploy --outputs-file outputs.json
cat outputs.json
```

Or via AWS CLI:

```bash
aws cloudformation describe-stacks \
  --stack-name HelloGo-pr-123 \
  --query 'Stacks[0].Outputs'
```

## Testing the Endpoint

```bash
# Get the API URL from outputs
API_URL=$(aws cloudformation describe-stacks \
  --stack-name HelloGo-pr-123 \
  --query 'Stacks[0].Outputs[?OutputKey==`ApiBaseUrl`].OutputValue' \
  --output text)

# Test the endpoint
curl $API_URL/

# Test a path
curl $API_URL/health
```

## Cleanup

### Ephemeral Stacks

```bash
pnpx cdk destroy -c stage=test -c namespace=pr-123
```

Resources are automatically removed due to `RemovalPolicy.DESTROY`.

### Live Stacks

```bash
pnpx cdk destroy -c stage=prod
```

Note: Some resources may be retained based on `RemovalPolicy.RETAIN`.

## Optional Features

### X-Ray Tracing

Uncomment the following in `lib/hello-go-stack.ts`:

```typescript
lambdaRole.addManagedPolicy(
  iam.ManagedPolicy.fromAwsManagedPolicyName('AWSXRayDaemonWriteAccess')
);
```

### VPC Wiring

**TODO**: Wire Lambda into VPC for private resource access.

Uncomment and configure in `lib/hello-go-stack.ts`:

```typescript
const lambdaFunction = new lambda.DockerImageFunction(this, 'HelloGoLambda', {
  // ... existing config
  vpc: ec2.Vpc.fromLookup(this, 'Vpc', { vpcId: 'vpc-xxxxx' }),
  vpcSubnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
  securityGroups: [/* your security groups */],
});
```

## Configuration Defaults

| Setting | Test (Ephemeral) | Live (Stable) |
|---------|------------------|---------------|
| Memory | 256 MB | 256 MB |
| Timeout | 10s | 10s |
| Log Retention | 7 days | 30 days |
| Removal Policy | DESTROY | RETAIN |
| Tags | `ephemeral=true`, `pr`, `sha`, `expires_at` | No ephemeral tags |

## IAM Role

The Lambda execution role includes:
- **AWSLambdaBasicExecutionRole**: CloudWatch Logs write access
- **Explicit separation**: No shared roles between functions

## CORS Configuration

Permissive CORS by default:
- **Origins**: `*`
- **Methods**: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`
- **Headers**: `*`
- **Max Age**: 1 day

## Development

```bash
# Build TypeScript
pnpm build

# Watch mode
pnpm watch

# Run tests
pnpm test

# Run tests in watch mode
pnpm test:watch

# Run tests with coverage
pnpm test:coverage

# Synthesize CloudFormation
pnpx cdk synth

# Diff against deployed stack
pnpx cdk diff -c stage=test -c namespace=pr-123
```
