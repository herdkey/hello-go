#!/usr/bin/env node

import * as cdk from 'aws-cdk-lib';
import { HelloGoStack } from '../lib/hello-go-stack';

// Hardcoded infra constants
const INFRA_ACCOUNT_ID = '073835883885';
const INFRA_ECR_REGION = 'us-west-2';
const BASE_NAME = 'hello-go';
const EPHEMERAL_HOURS = 1;

/**
 * Application context parsed from CDK context parameters
 */
export interface AppContext {
  stage: string;
  isEphemeral: boolean;
  namespace?: string;
  commitHash?: string;
  ecrImageTagOverride?: string;
  ecrRepoOverride?: string;
}

/**
 * Configuration object for creating the CDK stack
 */
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

/**
 * Reads and parses application context from CDK app
 * @param app - The CDK app instance
 * @returns Parsed application context with defaults
 */
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

/**
 * Validates required context parameters
 * @param context - The application context to validate
 * @throws Error if required parameters are missing
 */
export function validateContext(context: AppContext): void {
  if (!context.commitHash) {
    throw new Error('commit_hash is required');
  }
  if (context.isEphemeral && !context.namespace) {
    throw new Error('namespace is required when is_ephemeral is true');
  }
}

/**
 * Calculates expiration date for ephemeral stacks
 * @param isEphemeral - Whether this is an ephemeral deployment
 * @param hoursFromNow - Number of hours until expiration (default: 1)
 * @param now - Current date (defaults to now, injectable for testing)
 * @returns ISO date string (YYYY-MM-DD) if ephemeral, undefined otherwise
 */
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

/**
 * Builds the full ECR image URI for the Lambda function
 * @param context - Application context with image tag settings
 * @param infraAccountId - AWS account ID for ECR registry
 * @param infraEcrRegion - AWS region for ECR registry
 * @param baseName - Base name for the service
 * @returns Full ECR image URI with tag
 */
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

/**
 * Builds the CloudFormation stack name
 * @param isEphemeral - Whether this is an ephemeral deployment
 * @param namespace - Namespace for ephemeral deployments
 * @returns Stack name (includes namespace if ephemeral)
 */
export function buildStackName(
  isEphemeral: boolean,
  namespace?: string,
): string {
  return isEphemeral ? `HelloGo-${namespace}` : 'HelloGo';
}

/**
 * Builds resource tags for the stack
 * @param context - Application context
 * @param baseName - Base service name
 * @param expiresAt - Expiration date for ephemeral stacks
 * @returns Map of tag key-value pairs
 */
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

/**
 * Builds the complete stack configuration from context
 * @param context - Application context
 * @param baseName - Base service name
 * @returns Complete stack configuration object
 */
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

/**
 * Main entry point - creates and configures the CDK app
 */
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
