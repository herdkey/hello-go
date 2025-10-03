import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as ecr from 'aws-cdk-lib/aws-ecr';
import * as apigatewayv2 from 'aws-cdk-lib/aws-apigatewayv2';
import { HttpLambdaIntegration } from 'aws-cdk-lib/aws-apigatewayv2-integrations';
import { Construct } from 'constructs';

export interface HelloGoStackProps extends cdk.StackProps {
  baseName: string;
  stage: string;
  namespace?: string;
  isEphemeral: boolean;
  ecrImageUri: string;
}

export class HelloGoStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: HelloGoStackProps) {
    super(scope, id, props);

    const { baseName, stage, namespace, isEphemeral, ecrImageUri } = props;

    // Determine removal policy and retention settings
    const removalPolicy = isEphemeral
      ? cdk.RemovalPolicy.DESTROY
      : cdk.RemovalPolicy.RETAIN;

    const logRetention = isEphemeral
      ? logs.RetentionDays.THREE_DAYS
      : logs.RetentionDays.ONE_MONTH;

    // Build resource name prefix
    const lambdaBaseName = `${baseName}-api`;
    const lambdaName = `${lambdaBaseName}-${stage}${namespace ? `-${namespace}` : ''}`;

    // Create CloudWatch Log Group for Lambda with explicit retention
    const lambdaLogGroup = new logs.LogGroup(this, 'LambdaLogGroup', {
      logGroupName: `/aws/lambda/${lambdaName}`,
      retention: logRetention,
      removalPolicy,
    });

    // Create Lambda execution role with minimal IAM permissions
    const lambdaRole = new iam.Role(this, 'LambdaExecutionRole', {
      roleName: `${lambdaName}-execution`,
      assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com'),
      managedPolicies: [
        // CloudWatch Logs access
        iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AWSLambdaBasicExecutionRole'),
      ],
      description: `Execution role for ${lambdaName} Lambda`,
    });

    // TODO: Optional X-Ray support
    // Uncomment the following lines to enable X-Ray tracing:
    // lambdaRole.addManagedPolicy(
    //   iam.ManagedPolicy.fromAwsManagedPolicyName('AWSXRayDaemonWriteAccess')
    // );

    // Then replace the existing parsing code with:
    const { repoName, tag, accountId, region } = parseEcrImageUri(ecrImageUri);

    // Reference existing ECR repository
    const repository = ecr.Repository.fromRepositoryAttributes(this, 'EcrRepository', {
      repositoryName: repoName,
      repositoryArn: `arn:aws:ecr:${region}:${accountId}:repository/${repoName}`,
    });

    // Create Lambda function from container image
    const lambdaFunction = new lambda.DockerImageFunction(this, 'HelloGoLambda', {
      functionName: lambdaName,
      code: lambda.DockerImageCode.fromEcr(repository, { tagOrDigest: tag || 'latest' }),
      role: lambdaRole,
      memorySize: 256,
      timeout: cdk.Duration.seconds(10),
      description: `hello-go API Lambda for ${stage}${namespace ? ` (${namespace})` : ''}`,
      logGroup: lambdaLogGroup,
      // TODO: Optional VPC configuration
      // Uncomment and configure the following to wire Lambda into a VPC:
      // vpc: ec2.Vpc.fromLookup(this, 'Vpc', { vpcId: 'vpc-xxxxx' }),
      // vpcSubnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
      // securityGroups: [/* your security groups */],
    });

    // Create HTTP API Gateway
    const httpApi = new apigatewayv2.HttpApi(this, 'HelloGoHttpApi', {
      apiName: lambdaName,
      description: `HTTP API for hello-go ${stage}${namespace ? ` (${namespace})` : ''}`,
      corsPreflight: {
        allowOrigins: ['*'],
        allowMethods: [
          apigatewayv2.CorsHttpMethod.GET,
          apigatewayv2.CorsHttpMethod.POST,
          apigatewayv2.CorsHttpMethod.PUT,
          apigatewayv2.CorsHttpMethod.PATCH,
          apigatewayv2.CorsHttpMethod.DELETE,
          apigatewayv2.CorsHttpMethod.HEAD,
          apigatewayv2.CorsHttpMethod.OPTIONS,
        ],
        allowHeaders: ['*'],
        maxAge: cdk.Duration.days(1),
      },
    });

    // Create Lambda integration
    const integration = new HttpLambdaIntegration('LambdaIntegration', lambdaFunction);

    // Add proxy route (ANY /{proxy+})
    httpApi.addRoutes({
      path: '/{proxy+}',
      methods: [apigatewayv2.HttpMethod.ANY],
      integration,
    });

    // Add root route as well
    httpApi.addRoutes({
      path: '/',
      methods: [apigatewayv2.HttpMethod.ANY],
      integration,
    });

    // CloudFormation outputs
    new cdk.CfnOutput(this, 'ApiBaseUrl', {
      value: httpApi.url || httpApi.apiEndpoint,
      description: 'Base URL of the HTTP API Gateway',
      exportName: `${lambdaBaseName}-url`,
    });

    new cdk.CfnOutput(this, 'LambdaArn', {
      value: lambdaFunction.functionArn,
      description: 'ARN of the Lambda function',
      exportName: `${lambdaBaseName}-lambda-arn`,
    });

    new cdk.CfnOutput(this, 'LogGroupName', {
      value: lambdaFunction.logGroup.logGroupName,
      description: 'CloudWatch Log Group name',
      exportName: `${lambdaBaseName}-log-group`,
    });

    new cdk.CfnOutput(this, 'LambdaRoleArn', {
      value: lambdaRole.roleArn,
      description: 'ARN of the Lambda execution role',
      exportName: `${lambdaBaseName}-role-arn`,
    });
  }
}

// First, add this interface above the HelloGoStack class
interface EcrImageDetails {
  repoName: string;
  tag: string;
  accountId: string;
  region: string;
}

// Add this function above the HelloGoStack class
function parseEcrImageUri(ecrImageUri: string): EcrImageDetails {

  // Validate or provide default ECR image URI
  if (!ecrImageUri) {
    throw new Error(
      'ECR image URI must be provided via context: -c ecr_image_uri=<uri> or in cdk.json'
    );
  }

  // Format: {account}.dkr.ecr.{region}.amazonaws.com/{repo}:{tag}
  const uriParts = ecrImageUri.split('/');
  const repoAndTag = uriParts[uriParts.length - 1];
  const [repoName, tag] = repoAndTag.split(':');
  const registryParts = uriParts[0].split('.');
  const accountId = registryParts[0];
  const region = registryParts[3];

  return {
    repoName,
    tag,
    accountId,
    region,
  };
}
