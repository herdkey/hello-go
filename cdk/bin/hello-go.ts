#!/usr/bin/env node

import * as cdk from 'aws-cdk-lib';
import { HelloGoStack } from '../lib/hello-go-stack';
import { DefaultStackSynthesizer } from 'aws-cdk-lib';

// Constants
const BASE_NAME = 'hello-go';
const INFRA_ACCOUNT_ID = '073835883885';
const INFRA_ECR_REGION = 'us-west-2';
const EPHEMERAL_HOURS = 1;
const ACCOUNT_ID = process.env.CDK_DEFAULT_ACCOUNT;
// CDK (local) will assume this role to CRUD the CDK stack.
const DEPLOYER_ROLE_ARN = `arn:aws:iam::${ACCOUNT_ID}:role/cdk-deploy-${BASE_NAME}`;
// CDK passes this role to the CFN stack.
// CFN will assume it to CRUD resources for the stack.
const EXECUTION_ROLE_ARN = `arn:aws:iam::${ACCOUNT_ID}:role/cdk-exec-${BASE_NAME}`;

/**
 * Application context parsed from CDK context parameters
 */
export interface AppContext {
  stage: string;
  isEphemeral: boolean;
  ephemeralNs?: string;
  commitHash: string;
  ecrImageTag: string;
  ecrRepoName?: string;
  ecrAccountId?: string;
  ecrRegion?: string;
}

/**
 * ECR image details
 */
export interface EcrImageDetails {
  repoName: string;
  tag: string;
  accountId: string;
  region: string;
}

/**
 * Configuration object for creating the CDK stack
 */
export interface StackConfig {
  baseName: string;
  stage: string;
  ephemeralNs?: string;
  isEphemeral: boolean;
  ecrImage: EcrImageDetails;
  tags: Record<string, string>;
  env: {
    account: string | undefined;
    region: string | undefined;
  };
}

/**
 * Reads and parses application context from CDK app with validation and defaults
 * @param app - The CDK app instance
 * @returns Parsed application context with defaults applied
 * @throws Error if required parameters are missing or invalid
 */
export function readAppContext(app: cdk.App): AppContext {
  // Read raw values from context
  const ecrImageTag = app.node.tryGetContext('ecrImageTag') as
    | string
    | undefined;
  const stage = app.node.tryGetContext('stage') as string;
  const commitHash = app.node.tryGetContext('commitHash') as string | undefined;
  const ephemeralNs = app.node.tryGetContext('ephemeralNs') as string | undefined;
  const isEphemeralRaw = app.node.tryGetContext('isEphemeral') as
    | string
    | boolean
    | undefined;

  // Validate required parameter: ecrImageTag
  if (!ecrImageTag) {
    throw new Error('ecrImageTag is required (pass via -c ecrImageTag=<tag>)');
  }

  // Validate required parameter: stage
  if (!stage) {
    throw new Error('stage is required (pass via -c stage=<stage>)');
  }

  // Validate required parameter: commitHash
  if (!commitHash) {
    throw new Error('commitHash is required (pass via -c commitHash=<hash>)');
  }

  // Default: ephemeral if stage is "test", unless explicitly overridden
  const isEphemeral =
    isEphemeralRaw === 'true' ||
    (isEphemeralRaw !== 'false' && stage === 'test');

  // Validate namespace logic
  if (!isEphemeral) {
    // Non-ephemeral: namespace must NOT be provided
    if (ephemeralNs) {
      throw new Error(
        'namespace is only allowed for ephemeral (test stage) deployments',
      );
    }
  } else {
    // Ephemeral: namespace is required
    if (!ephemeralNs) {
      throw new Error(
        'namespace is required for ephemeral deployments (pass via -c namespace=<name>)',
      );
    }
  }

  return {
    stage,
    isEphemeral,
    ephemeralNs,
    commitHash,
    ecrImageTag,
    ecrRepoName: app.node.tryGetContext('ecrRepoName') as string | undefined,
    ecrAccountId: app.node.tryGetContext('ecrAccountId') as string | undefined,
    ecrRegion: app.node.tryGetContext('ecrRegion') as string | undefined,
  };
}

/**
 * Calculates expiration date for ephemeral stacks
 * @param isEphemeral - Whether this is an ephemeral deployment
 * @param daysFromNow - Number of days until expiration (default: 1)
 * @param now - Current date (defaults to now, injectable for testing)
 * @returns ISO date string (YYYY-MM-DD) if ephemeral, undefined otherwise
 */
export function calculateExpiresAt(
  isEphemeral: boolean,
  daysFromNow: number = EPHEMERAL_HOURS,
  now: Date = new Date(),
): string | undefined {
  if (!isEphemeral) {
    return undefined;
  }
  const expirationDate = new Date(
    // calculate expiration time in milliseconds
    now.getTime() + daysFromNow * 24 * 60 * 60 * 1000,
  );
  return expirationDate.toISOString().split('T')[0];
}

/**
 * Builds ECR image details from context
 * @param context - Application context with image settings
 * @param infraAccountId - AWS account ID for ECR registry
 * @param infraEcrRegion - AWS region for ECR registry
 * @param baseName - Base name for the service
 * @returns ECR image component details
 */
export function buildEcrImageDetails(
  context: AppContext,
  infraAccountId: string = INFRA_ACCOUNT_ID,
  infraEcrRegion: string = INFRA_ECR_REGION,
  baseName: string = BASE_NAME,
): EcrImageDetails {
  return {
    repoName: context.ecrRepoName || `${context.stage}/${baseName}/lambda`,
    tag: context.ecrImageTag,
    accountId: context.ecrAccountId || infraAccountId,
    region: context.ecrRegion || infraEcrRegion,
  };
}

/**
 * Builds the CloudFormation stack name
 * @param isEphemeral - Whether this is an ephemeral deployment
 * @param ephemeralNs - Namespace for ephemeral deployments
 * @returns Stack name (includes ephemeralNs if ephemeral)
 */
export function buildStackName(
  isEphemeral: boolean,
  ephemeralNs?: string,
): string {
  const suffix = isEphemeral ? `-${ephemeralNs}` : '';
  return `${BASE_NAME}${suffix}`;
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
    Stage: context.stage,
    Namespace: baseName,
  };

  if (context.isEphemeral) {
    if (!context.ephemeralNs) {
      throw new Error('ephemeralNs is required for ephemeral deployments');
    }
    if (!expiresAt) {
      throw new Error('expiresAt is required for ephemeral deployments');
    }
    tags.Ephemeral = 'true';
    tags.EphemeralNS = context.ephemeralNs;
    tags.SHA = context.commitHash;
    tags.ExpiresAt = expiresAt;
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
  const ecrImage = buildEcrImageDetails(context);
  const tags = buildTags(context, baseName, expiresAt);

  return {
    baseName,
    stage: context.stage,
    ephemeralNs: context.ephemeralNs,
    isEphemeral: context.isEphemeral,
    ecrImage,
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
  const stackName = buildStackName(context.isEphemeral, context.ephemeralNs);
  const stackConfig = {
    ...buildStackConfig(context, BASE_NAME),
    synthesizer: new DefaultStackSynthesizer({
      cloudFormationExecutionRole: EXECUTION_ROLE_ARN,
      deployRoleArn: DEPLOYER_ROLE_ARN,
    }),
  };

  new HelloGoStack(app, stackName, stackConfig);
}

// Only run main if this file is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}
