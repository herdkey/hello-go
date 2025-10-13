#!/usr/bin/env node

import * as cdk from 'aws-cdk-lib';
import { HelloGoStack } from '../lib/hello-go-stack';
import { DefaultStackSynthesizer } from 'aws-cdk-lib';

// Constants
const BASE_NAME = 'hello-go';
const INFRA_ACCOUNT_ID = '073835883885';
const INFRA_ECR_REGION = 'us-west-2';
const ACCOUNT_ID = process.env.CDK_DEFAULT_ACCOUNT;
// CDK (local) will assume this role to CRUD the CDK stack.
const DEPLOYER_ROLE_ARN = `arn:aws:iam::${ACCOUNT_ID}:role/cdk-deploy-${BASE_NAME}`;
// CDK passes this role to the CFN stack.
// CFN will assume it to CRUD resources for the stack.
const EXECUTION_ROLE_ARN = `arn:aws:iam::${ACCOUNT_ID}:role/cdk-exec-${BASE_NAME}`;

const NOW = Math.floor(Date.now() / 1000);

/**
 * Application context parsed from CDK context parameters
 */
export interface AppContext {
  stage: string;
  isEphemeral: boolean;
  instanceNs: string;
  commit: string;
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
  instanceNs?: string;
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
  const commit = app.node.tryGetContext('commit') as string | undefined;
  const instanceNs = app.node.tryGetContext('instanceNs') as string | undefined;
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

  // Validate required parameter: commit
  if (!commit) {
    throw new Error('commit is required (pass via -c commit=<hash>)');
  }

  // Validate required parameter: instanceNs
  if (!instanceNs) {
    throw new Error(
      'instanceNs is required for all deployments (pass via -c instanceNs=<name>)',
    );
  }

  // Default: ephemeral if stage is "test", unless explicitly overridden
  const isEphemeral =
    isEphemeralRaw === 'true' ||
    (isEphemeralRaw !== 'false' && stage === 'test');

  return {
    stage,
    isEphemeral,
    instanceNs,
    commit,
    ecrImageTag,
    ecrRepoName: app.node.tryGetContext('ecrRepoName') as string | undefined,
    ecrAccountId: app.node.tryGetContext('ecrAccountId') as string | undefined,
    ecrRegion: app.node.tryGetContext('ecrRegion') as string | undefined,
  };
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
 * Builds resource tags for the stack
 * @param context - Application context
 * @param baseName - Base service name
 * @param now - current timestamp
 * @returns Map of tag key-value pairs
 */
export function buildTags(
  context: AppContext,
  baseName: string = BASE_NAME,
  now: number = NOW,
): Record<string, string> {
  const tags: Record<string, string> = {
    ['savi:stage']: context.stage,
    ['savi:namespace']: baseName,
  };

  if (context.isEphemeral) {
    if (!context.instanceNs) {
      throw new Error('instanceNs is required for ephemeral deployments');
    }
    tags['savi:instance-ns'] = context.instanceNs;
    tags['savi:commit'] = context.commit;
    tags['savi:created-at'] = `${now}`;
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
  const ecrImage = buildEcrImageDetails(context);
  const tags = buildTags(context, baseName);

  return {
    baseName,
    stage: context.stage,
    instanceNs: context.instanceNs,
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
  const stackName = `${BASE_NAME}-${context.instanceNs}`;
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
