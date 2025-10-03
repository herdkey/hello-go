#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { HelloGoStack } from '../lib/hello-go-stack';

const app = new cdk.App();

// Read context values
const stage = app.node.tryGetContext('stage') as string || "test";
const namespace = app.node.tryGetContext('namespace') as string | undefined;
const pr = app.node.tryGetContext('pr') as string | undefined;
const sha = app.node.tryGetContext('sha') as string | undefined;
const expiresAt = app.node.tryGetContext('expires_at') as string | undefined;
const ecrImageUri = app.node.tryGetContext('ecr_image_uri') as string || "073835883885.dkr.ecr.us-west-2.amazonaws.com/test/hello-go/lambda:fcb8742";

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
  if (pr) tags.pr = String(pr);
  if (sha) tags.sha = String(sha);
  if (expiresAt) tags.expires_at = String(expiresAt);
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
