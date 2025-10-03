#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { HelloGoStack } from '../lib/hello-go-stack';

const app = new cdk.App();

// Read context values
const stage = app.node.tryGetContext('stage') || 'test';
const namespace = app.node.tryGetContext('namespace');
const pr = app.node.tryGetContext('pr');
const sha = app.node.tryGetContext('sha');
const expiresAt = app.node.tryGetContext('expires_at');
const ecrImageUri = app.node.tryGetContext('ecr_image_uri');

// Determine if this is an ephemeral deployment (test env with namespace)
const isEphemeral = stage === 'test' && !!namespace;

// Build stack name
let stackName = 'HelloGoStack';
if (isEphemeral) {
  stackName = `HelloGo-${namespace}`;
}

// Build tags
const tags: Record<string, string> = {
  svc: 'hello-go',
  stage,
};

if (isEphemeral) {
  tags.ephemeral = 'true';
  if (pr) tags.pr = pr;
  if (sha) tags.sha = sha;
  if (expiresAt) tags.expires_at = expiresAt;
}

new HelloGoStack(app, stackName, {
  baseName: 'hello-go',
  stage,
  namespace: isEphemeral ? namespace : undefined,
  isEphemeral,
  ecrImageUri,
  tags,
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION,
  },
});
