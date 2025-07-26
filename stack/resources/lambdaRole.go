package resources

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	iam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/jsii-runtime-go"
)


func NewLambdaRole( stack awscdk.Stack) awsiam.Role {

	lambdaRole := iam.NewRole(stack, jsii.String("LambdaRole"), &iam.RoleProps{
		AssumedBy: iam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]iam.IManagedPolicy{
			iam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaBasicExecutionRole")),
			iam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaVPCAccessExecutionRole")),
			iam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("NeptuneFullAccess")),
			iam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonS3FullAccess")),
		},
	})

	return lambdaRole
}
