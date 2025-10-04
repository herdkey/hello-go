#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { HelloGoStack } from '../lib/hello-go-stack';

// Hardcoded infra constants
const INFRA_ACCOUNT_ID = '073835883885';
const INFRA_ECR_REGION = 'us-west-2';
const BASE_NAME = 'hello-go';
const EPHEMERAL_HOURS = 1;

export interface AppContext {
  stage: string;
  isEphemeral: boolean;
  namespace?: string;
  commitHash?: string;
  ecrImageTagOverride?: string;
  ecrRepoOverride?: string;
}

export interface StackConfig {
  baseName: string;
  stage: string;
  namespace?: string;
  isEphemeral: boolean;
  ecrImageUri: string;
  tags: Record<string, string>;
  env: {
    account: string | undefined;
    region: string | undefined;
  };
}

export function readAppContext(app: cdk.App): AppContext {
  return {
    stage: (app.node.tryGetContext('stage') as string) || 'test',
    isEphemeral: (app.node.tryGetContext('is_ephemeral') as boolean) || false,
    namespace: app.node.tryGetContext('namespace') as string | undefined,
    commitHash: app.node.tryGetContext('commit_hash') as string | undefined,
    ecrImageTagOverride: app.node.tryGetContext('ecrImageTag') as
      | string
      | undefined,
    ecrRepoOverride: app.node.tryGetContext('ecrRepo') as string | undefined,
  };
}

export function validateContext(context: AppContext): void {
  if (!context.commitHash) {
    throw new Error('commit_hash is required');
  }
  if (context.isEphemeral && !context.namespace) {
    throw new Error('namespace is required when is_ephemeral is true');
  }
}

export function calculateExpiresAt(
  isEphemeral: boolean,
  hoursFromNow: number = EPHEMERAL_HOURS,
  now: Date = new Date(),
): string | undefined {
  if (!isEphemeral) {
    return undefined;
  }
  const expirationDate = new Date(
    // calculate expiration time in milliseconds
    now.getTime() + hoursFromNow * 60 * 60 * 1000,
  );
  return expirationDate.toISOString().split('T')[0];
}

export function buildEcrImageUri(
  context: AppContext,
  infraAccountId: string = INFRA_ACCOUNT_ID,
  infraEcrRegion: string = INFRA_ECR_REGION,
  baseName: string = BASE_NAME,
): string {
  const ecrImageTag =
    context.ecrImageTagOverride ||
    (context.isEphemeral ? context.commitHash : 'latest');
  const ecrRepo =
    context.ecrRepoOverride ||
    `${infraAccountId}.dkr.ecr.${infraEcrRegion}.amazonaws.com/${context.stage}/${baseName}/lambda`;
  return `${ecrRepo}:${ecrImageTag}`;
}

export function buildStackName(
  isEphemeral: boolean,
  namespace?: string,
): string {
  return isEphemeral ? `HelloGo-${namespace}` : 'HelloGo';
}

export function buildTags(
  context: AppContext,
  baseName: string = BASE_NAME,
  expiresAt?: string,
): Record<string, string> {
  const tags: Record<string, string> = {
    svc: baseName,
    stage: context.stage,
  };

  if (context.isEphemeral && context.namespace && context.commitHash) {
    tags.ephemeral = 'true';
    tags.namespace = context.namespace;
    tags.sha = context.commitHash;
    if (expiresAt) {
      tags.expires_at = expiresAt;
    }
  }

  return tags;
}

export function buildStackConfig(
  context: AppContext,
  baseName: string = BASE_NAME,
): StackConfig {
  const expiresAt = calculateExpiresAt(context.isEphemeral);
  const ecrImageUri = buildEcrImageUri(context);
  const tags = buildTags(context, baseName, expiresAt);

  return {
    baseName,
    stage: context.stage,
    namespace: context.namespace,
    isEphemeral: context.isEphemeral,
    ecrImageUri,
    tags,
    env: {
      account: process.env.CDK_DEFAULT_ACCOUNT,
      region: process.env.CDK_DEFAULT_REGION,
    },
  };
}

export function main(): void {
  const app = new cdk.App();
  const context = readAppContext(app);
  validateContext(context);
  const stackConfig = buildStackConfig(context, BASE_NAME);
  const stackName = buildStackName(context.isEphemeral, context.namespace);

  new HelloGoStack(app, stackName, stackConfig);
}

// Only run main if this file is executed directly
if (require.main === module) {
  main();
}
