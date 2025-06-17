package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

func TestNeptuneappStack(t *testing.T) {
	app := awscdk.NewApp(nil)

	stack := NewNeptuneappStack(app, "MyStack", nil)

	template := assertions.Template_FromStack(stack, nil)

	template.HasResourceProperties(jsii.String("AWS::SQS::Queue"), map[string]interface{}{
		"VisibilityTimeout": 300,
	})
}
