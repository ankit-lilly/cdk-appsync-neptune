package resources


import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/jsii-runtime-go"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
)

type LambdaConfig struct {
	Name         string
	Handler      string
	Runtime      awslambda.Runtime
	CodePath     string
	Timeout      awscdk.Duration
	Environment  map[string]*string
	VPC          awsec2.IVpc
	Role         awsiam.IRole
	MemorySize   *float64
	Description  *string
}

func DefaultLambdaConfig() *LambdaConfig {
	return &LambdaConfig{
		Handler:    "main",
		Runtime:    awslambda.Runtime_PROVIDED_AL2023(),
		Timeout:    awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize: jsii.Number(128),
	}
}

type LambdaFactory struct {
	scope constructs.Construct
	vpc   awsec2.IVpc
	role  awsiam.IRole
}

func NewLambdaFactory(scope constructs.Construct, vpc awsec2.IVpc, role awsiam.IRole) *LambdaFactory {
	return &LambdaFactory{
		scope: scope,
		vpc:   vpc,
		role:  role,
	}
}


func (f *LambdaFactory) CreateFunction(config *LambdaConfig) awslambda.Function {
	props := &awslambda.FunctionProps{
		FunctionName: jsii.String(config.Name),
		Runtime:      config.Runtime,
		Handler:      jsii.String(config.Handler),
		Code:         awslambda.AssetCode_FromAsset(jsii.String(config.CodePath), nil),
		Timeout:      config.Timeout,
		MemorySize:   config.MemorySize,
		Vpc:          f.vpc,
		Role:         f.role,
	}

	if config.Environment != nil {
		props.Environment = &config.Environment
	}

	if config.Description != nil {
		props.Description = config.Description
	}

	if config.VPC != nil {
		props.Vpc = config.VPC
	}
	if config.Role != nil {
		props.Role = config.Role
	}

	return awslambda.NewFunction(f.scope, jsii.String(config.Name), props)
}

func (f *LambdaFactory) CreateGoFunction(name, codePath string, env map[string]*string) awslambda.Function {
	config := DefaultLambdaConfig()
	config.Name = name
	config.CodePath = codePath
	config.Environment = env
	config.Description = jsii.String("Go Lambda function: " + name)
	
	return f.CreateFunction(config)
}



