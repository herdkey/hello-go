import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as apigatewayv2 from 'aws-cdk-lib/aws-apigatewayv2';
import { HttpLambdaIntegration } from 'aws-cdk-lib/aws-apigatewayv2-integrations';
import { Construct } from 'constructs';

export interface HelloGoStackProps extends cdk.StackProps {
  stage: string;
  namespace?: string;
  isEphemeral: boolean;
  ecrImageUri?: string;
}

export class HelloGoStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: HelloGoStackProps) {
    super(scope, id, props);

    const { stage, namespace, isEphemeral, ecrImageUri } = props;

    // Determine removal policy and retention settings
    const removalPolicy = isEphemeral ? cdk.RemovalPolicy.DESTROY : cdk.RemovalPolicy.RETAIN;
    const logRetention = isEphemeral
      ? logs.RetentionDays.ONE_WEEK
      : logs.RetentionDays.ONE_MONTH;

    // Build resource name prefix
    const resourcePrefix = namespace ? `hello-go-${namespace}` : `hello-go-${stage}`;

    // Create Lambda execution role with minimal IAM permissions
    const lambdaRole = new iam.Role(this, 'LambdaExecutionRole', {
      roleName: `${resourcePrefix}-lambda-role`,
      assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com'),
      managedPolicies: [
        // CloudWatch Logs access
        iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AWSLambdaBasicExecutionRole'),
      ],
      description: `Execution role for ${resourcePrefix} Lambda`,
    });

    // TODO: Optional X-Ray support
    // Uncomment the following lines to enable X-Ray tracing:
    // lambdaRole.addManagedPolicy(
    //   iam.ManagedPolicy.fromAwsManagedPolicyName('AWSXRayDaemonWriteAccess')
    // );

    // Validate or provide default ECR image URI
    const imageUri = ecrImageUri || this.node.tryGetContext('ecr_image_uri');
    if (!imageUri) {
      throw new Error(
        'ECR image URI must be provided via context: -c ecr_image_uri=<uri> or in cdk.json'
      );
    }

    // Create Lambda function from container image
    const lambdaFunction = new lambda.DockerImageFunction(this, 'HelloGoLambda', {
      functionName: `${resourcePrefix}-api`,
      code: lambda.DockerImageCode.fromEcr(
        lambda.EcrImageCode.fromAssetImage('.').repository,
        { tagOrDigest: imageUri.split(':').pop() || 'latest' }
      ),
      role: lambdaRole,
      memorySize: 256,
      timeout: cdk.Duration.seconds(10),
      description: `hello-go API Lambda for ${stage}${namespace ? ` (${namespace})` : ''}`,
      logRetention,
      // TODO: Optional VPC configuration
      // Uncomment and configure the following to wire Lambda into a VPC:
      // vpc: ec2.Vpc.fromLookup(this, 'Vpc', { vpcId: 'vpc-xxxxx' }),
      // vpcSubnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
      // securityGroups: [/* your security groups */],
    });

    // Create HTTP API Gateway
    const httpApi = new apigatewayv2.HttpApi(this, 'HelloGoHttpApi', {
      apiName: `${resourcePrefix}-api`,
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
      exportName: `${resourcePrefix}-api-url`,
    });

    new cdk.CfnOutput(this, 'LambdaArn', {
      value: lambdaFunction.functionArn,
      description: 'ARN of the Lambda function',
      exportName: `${resourcePrefix}-lambda-arn`,
    });

    new cdk.CfnOutput(this, 'LogGroupName', {
      value: lambdaFunction.logGroup.logGroupName,
      description: 'CloudWatch Log Group name',
      exportName: `${resourcePrefix}-log-group`,
    });

    new cdk.CfnOutput(this, 'LambdaRoleArn', {
      value: lambdaRole.roleArn,
      description: 'ARN of the Lambda execution role',
      exportName: `${resourcePrefix}-role-arn`,
    });
  }
}
