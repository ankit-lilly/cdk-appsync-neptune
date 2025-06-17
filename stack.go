package main

import (
	"fmt"
	"strings"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	appsync "github.com/aws/aws-cdk-go/awscdk/v2/awsappsync"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	events "github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	targets "github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	iam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	lambda "github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	s3 "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	neptune "github.com/aws/aws-cdk-go/awscdkneptunealpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type NeptuneappStackProps struct {
	awscdk.StackProps
}

func NewNeptuneappStack(scope constructs.Construct, id string, props *NeptuneappStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	vpc := ec2.NewVpc(stack, jsii.String("nepVPC"), &ec2.VpcProps{
		MaxAzs: jsii.Number(2),
	})

	cluster := neptune.NewDatabaseCluster(stack, jsii.String("nepCluster"), &neptune.DatabaseClusterProps{
		Vpc:                     vpc,
		InstanceType:            neptune.InstanceType_R5_LARGE(),
		AutoMinorVersionUpgrade: jsii.Bool(true),
	})

	cluster.Connections().AllowDefaultPortFromAnyIpv4(jsii.String("Allow access to Neptune from anywhere"))

	bucket := s3.NewBucket(stack, jsii.String("nepBucket"), &s3.BucketProps{
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	lambdaRole := iam.NewRole(stack, jsii.String("LambdaRole"), &iam.RoleProps{
		AssumedBy: iam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]iam.IManagedPolicy{
			iam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaBasicExecutionRole")),
			iam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaVPCAccessExecutionRole")),
			iam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("NeptuneFullAccess")),
			iam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonS3FullAccess")),
		},
	})

	scraperFn := lambda.NewFunction(stack, jsii.String("ScraperFunction"), &lambda.FunctionProps{
		Runtime: lambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Code:    lambda.AssetCode_FromAsset(jsii.String("lambdas/scraper/function.zip"), nil),
		Vpc:     vpc,
		Role:    lambdaRole,
		Environment: &map[string]*string{
			"NEPTUNE_ENDPOINT": cluster.ClusterEndpoint().Hostname(),
			"NEPTUNE_PORT":     
			jsii.String("8182"), 
		},
		Timeout: awscdk.Duration_Minutes(jsii.Number(5)),
	})
	bucket.GrantPut(scraperFn, jsii.String("*"))

	events.NewRule(stack, jsii.String("ScrapeSchedule"), &events.RuleProps{
		Schedule: events.Schedule_Rate(awscdk.Duration_Hours(jsii.Number(3))),
		Targets: &[]events.IRuleTarget{
			targets.NewLambdaFunction(scraperFn, nil),
		},
	})

	api := appsync.NewGraphqlApi(stack, jsii.String("NewsApi"), &appsync.GraphqlApiProps{
		Name:   jsii.String("NewsGraphApi"),
		Schema: appsync.SchemaFile_FromAsset(jsii.String("./schema/schema.graphql")),
		AuthorizationConfig: &appsync.AuthorizationConfig{
			DefaultAuthorization: &appsync.AuthorizationMode{
				AuthorizationType: appsync.AuthorizationType_API_KEY,
				ApiKeyConfig: &appsync.ApiKeyConfig{
					Expires: awscdk.Expiration_After(awscdk.Duration_Days(jsii.Number(365))),
				},
			},
		},
	})

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

	ds := api.AddLambdaDataSource(jsii.String("ResolverDS"), resolverFn, nil)

	for _, field := range []string{"feed", "related", "recommended", "favorites", "article"} {
		ds.CreateResolver(&field, &appsync.BaseResolverProps{
			TypeName:  jsii.String("Query"),
			FieldName: jsii.String(field),
		})
	}

	var favorites string = "favorites"
	mutationResolverId := fmt.Sprintf("Mutation%sResolver", strings.Title(favorites))
	ds.CreateResolver(jsii.String(mutationResolverId), &appsync.BaseResolverProps{
		TypeName:  jsii.String("Mutation"),
		FieldName: jsii.String(favorites),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)
	NewNeptuneappStack(app, "bigBeautifulStack", &NeptuneappStackProps{
		awscdk.StackProps{
			StackName: jsii.String("bigBeautifulStack"),
			Env:       &awscdk.Environment{Region: jsii.String("us-east-2")},
		},
	})
	app.Synth(nil)
}
