package stack

import (
	"neptuneapp/stack/resources"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	lambda "github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewSDRBackendStack(scope constructs.Construct, id string, sprops awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &sprops)

	vpc := ec2.NewVpc(stack, jsii.String("nepVPC"), &ec2.VpcProps{
		MaxAzs: jsii.Number(2),
	})

	cluster := resources.NewNeptuneDB(stack, vpc)
	queue := resources.NewSQSQueue(stack, vpc)
	apiGateway := resources.NewApiGateway(stack)
	lambdaRole := resources.NewLambdaRole(stack)


	resolverFn := lambda.NewFunction(stack, jsii.String("ResolverFunction"), &lambda.FunctionProps{
		Runtime: lambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Code:    lambda.AssetCode_FromAsset(jsii.String("lambdas/resolver/function.zip"), nil),
		Vpc:     vpc,
		Role:    lambdaRole,
		Environment: &map[string]*string{
			"NEPTUNE_ENDPOINT":        cluster.ClusterEndpoint().Hostname(),
			"NEPTUNE_READER_ENDPOINT": cluster.ClusterReadEndpoint().Hostname(),
			"NEPTUNE_PORT":            jsii.String("8182"),
		},
		Timeout: awscdk.Duration_Minutes(jsii.Number(5)),
	})

	appSyncAPI := resources.NewAppSyncApi(stack, vpc, resolverFn)

	sdrHandler := lambda.NewFunction(stack, jsii.String("SdrHandler"), &lambda.FunctionProps{
		Runtime: lambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Code: lambda.AssetCode_FromAsset(jsii.String("lambdas/sdrHandler/function.zip"),nil),
		Vpc: vpc,
		Role: lambdaRole,
		Environment: &map[string]*string{
			"QUEUE_URL": queue.QueueUrl(),
		},
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
	})

	queue.GrantSendMessages(sdrHandler)

	sdrProcessor := lambda.NewFunction(stack, jsii.String("CreateSDR"), &lambda.FunctionProps{
		Runtime: lambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Code: lambda.AssetCode_FromAsset(jsii.String("lambdas/sdrProcessor/function.zip"),nil),
		Vpc: vpc,
		Role: lambdaRole,
		Environment: &map[string]*string{
			"NEPTUNE_ENDPOINT": cluster.ClusterEndpoint().Hostname(),
			"NEPTUNE_PORT":     jsii.String("8182"),
		}, 
		Timeout: awscdk.Duration_Minutes(jsii.Number(5)),
	})

	sdrHandlerIntegration := awsapigateway.NewLambdaIntegration(sdrHandler, nil);
	sdrHandlerResource := apiGateway.Root().AddResource(jsii.String("sdr"), nil)
	sdrHandlerResource.AddMethod(jsii.String("POST"), sdrHandlerIntegration, nil)

	sdrProcessor.AddEventSource(awslambdaeventsources.NewSqsEventSource(queue, &awslambdaeventsources.SqsEventSourceProps{
		BatchSize: jsii.Number(10),
	}))

	awscdk.NewCfnOutput(stack, jsii.String("AppSyncAPIEndpoint"), &awscdk.CfnOutputProps{
		Value: appSyncAPI.GraphqlUrl(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("AppSyncAPIKey"), &awscdk.CfnOutputProps{
		Value: appSyncAPI.ApiKey(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("RestApiEndpoint"), &awscdk.CfnOutputProps{
		Value: apiGateway.Url(),
	})

	return stack
}

