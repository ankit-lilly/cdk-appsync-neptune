package resources

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	appsync "github.com/aws/aws-cdk-go/awscdk/v2/awsappsync"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/jsii-runtime-go"
)

func NewAppSyncApi( stack awscdk.Stack, vpc awsec2.Vpc, resolverFunc awslambda.IFunction) appsync.GraphqlApi {

	appSyncAPI := appsync.NewGraphqlApi(stack, jsii.String("NewsApi"), &appsync.GraphqlApiProps{
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

	ds := appSyncAPI.AddLambdaDataSource(jsii.String("ResolverDS"), resolverFunc, nil)

	for _, field := range []string{"study", "studies", "studyVersion", "organization"}  {
		ds.CreateResolver(&field, &appsync.BaseResolverProps{
			TypeName:  jsii.String("Query"),
			FieldName: jsii.String(field),
		})
	}

	return appSyncAPI
}
