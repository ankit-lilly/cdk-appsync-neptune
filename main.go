package main

import (
	"neptuneapp/stack"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {

	const (
		STACK_NAME = "BigBeautifulStack"
		APP_ID      = "sdrbackend"
		ACCOUNT_ID = "073395114856"
		REGION      = "us-west-1"
	)

	app := awscdk.NewApp(nil)

	stack.NewSDRBackendStack(app, APP_ID, awscdk.StackProps{
			StackName: jsii.String(STACK_NAME),
			Env: &awscdk.Environment{
				Account: jsii.String(ACCOUNT_ID),
				Region:  jsii.String(REGION),
			},
		},
	)
	app.Synth(nil)
}
