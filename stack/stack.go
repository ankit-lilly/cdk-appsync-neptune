package stack

import (
	"github.com/ankit-lilly/dtd-go-backend/stack/resources"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
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
	lambdaFactory := resources.NewLambdaFactory(stack, vpc, lambdaRole)


	sdrHandler :=	lambdaFactory.CreateGoFunction(
		"sdrHandler", 
		"lambdas/sdrHandler/function.zip", 
		map[string]*string{ 
			"QUEUE_URL": queue.QueueUrl(),
		},
  )

	queue.GrantSendMessages(sdrHandler)

	sdrProcessor := lambdaFactory.CreateGoFunction(
		"sdrProcessor", 
		"lambdas/sdrProcessor/function.zip", 
		map[string]*string{
			"NEPTUNE_ENDPOINT": cluster.ClusterEndpoint().Hostname(),
			"NEPTUNE_PORT":     jsii.String("8182"),
	})

	resolverFn :=  lambdaFactory.CreateGoFunction(
		"resolverFunction", 
		"lambdas/resolver/function.zip", 
		map[string]*string{
			"NEPTUNE_ENDPOINT":        cluster.ClusterEndpoint().Hostname(),
			"NEPTUNE_READER_ENDPOINT": cluster.ClusterReadEndpoint().Hostname(),
			"NEPTUNE_PORT":            jsii.String("8182"),
	})


	appSyncAPI := resources.NewAppSyncApi(stack, vpc, resolverFn)

	sdrHandlerIntegration := awsapigateway.NewLambdaIntegration(sdrHandler, nil);
	sdrHandlerResource := apiGateway.Root().AddResource(jsii.String("sdr"), nil)
	sdrHandlerResource.AddMethod(jsii.String("POST"), sdrHandlerIntegration, nil)

	sdrProcessor.AddEventSource(
		awslambdaeventsources.NewSqsEventSource( 
			queue, 
			&awslambdaeventsources.SqsEventSourceProps{
				BatchSize: jsii.Number(10),
				ReportBatchItemFailures: jsii.Bool(true),
			},
	))

	cfnOutput(stack, map[string]*awscdk.CfnOutputProps{
		"NeptuneHostEndpoint": {
			Value: cluster.ClusterEndpoint().Hostname(),
		},
		"NeptuneReaderEndpoint": {
			Value: cluster.ClusterReadEndpoint().Hostname(),
		},
		"AppSyncAPIEndpoint": {
			Value: appSyncAPI.GraphqlUrl(),
		},
		"AppSyncAPIKey": {
			Value: appSyncAPI.ApiKey(),
		},
		"RestApiEndpoint": {
			Value: apiGateway.Url(),
		},
	});

	return stack
}

func cfnOutput(stack awscdk.Stack, outputMap  map[string]*awscdk.CfnOutputProps) {
	for key, value := range outputMap {
		awscdk.NewCfnOutput(stack, jsii.String(key), value)
	}
}

