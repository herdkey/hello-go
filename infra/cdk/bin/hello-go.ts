#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { HelloGoStack } from '../lib/hello-go-stack';

const app = new cdk.App();

// Hardcoded infra constants
const infraAccountId = '073835883885';
const infraEcrRegion = 'us-west-2';
const baseName = 'hello-go';

// Read required context values
const stage = app.node.tryGetContext('stage') as string || "test";
const isEphemeral = app.node.tryGetContext('is_ephemeral') as boolean || false;
const namespace = app.node.tryGetContext('namespace') as string | undefined;
const commitHash = app.node.tryGetContext('commit_hash') as string | undefined;

// Optional overrides
const ecrImageTagOverride = app.node.tryGetContext('ecrImageTag') as string | undefined;
const ecrRepoOverride = app.node.tryGetContext('ecrRepo') as string | undefined;

// Validate required params
if (!commitHash) {
  throw new Error('commit_hash is required');
}
if (isEphemeral && !namespace) {
  throw new Error('namespace is required when is_ephemeral is true');
}

// Calculate expires_at (30 days from now) for ephemeral deployments
const expiresAt = isEphemeral
  ? new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  : undefined;

// Build ECR image URI
const ecrImageTag = ecrImageTagOverride || (isEphemeral ? commitHash : 'latest');
const ecrRepo = ecrRepoOverride || `${infraAccountId}.dkr.ecr.${infraEcrRegion}.amazonaws.com/${stage}/${baseName}/lambda`;
const ecrImageUri = `${ecrRepo}:${ecrImageTag}`;

// Build stack name
const stackName = isEphemeral ? `HelloGo-${namespace}` : 'HelloGo';

// Build tags
const tags: Record<string, string> = {
  svc: baseName,
  stage,
};

if (isEphemeral) {
  tags.ephemeral = 'true';
  tags.namespace = namespace;
  tags.sha = commitHash;
  tags.expires_at = String(expiresAt);
}

new HelloGoStack(app, stackName, {
  baseName,
  stage,
  namespace,
  isEphemeral,
  ecrImageUri,
  tags,
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION,
  },
});
