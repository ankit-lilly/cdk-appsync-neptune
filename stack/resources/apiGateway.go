package resources

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/jsii-runtime-go"
	"github.com/aws/aws-cdk-go/awscdk/v2"
)


func NewApiGateway(stack awscdk.Stack) awsapigateway.RestApi {
	return awsapigateway.NewRestApi(stack, jsii.String("sdrRestApi"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("SDR Rest endpoints"),
		Description: jsii.String("API for SDR operations"),
		DeployOptions: &awsapigateway.StageOptions{
			DataTraceEnabled:   jsii.Bool(true),
			//LoggingLevel:       awsapigateway.MethodLoggingLevel_INFO,
			MetricsEnabled:     jsii.Bool(true),
			ThrottlingRateLimit: jsii.Number(100),
			ThrottlingBurstLimit: jsii.Number(50),
		},
	})
}

