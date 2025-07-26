package resources

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	neptune "github.com/aws/aws-cdk-go/awscdkneptunealpha/v2"
	"github.com/aws/jsii-runtime-go"
)


func NewNeptuneDB( stack awscdk.Stack, vpc awsec2.Vpc) neptune.DatabaseCluster {

	clusterParameterGroup := neptune.NewClusterParameterGroup(stack, jsii.String("neptuneClusterParamName"), &neptune.ClusterParameterGroupProps{
		Family: neptune.ParameterGroupFamily_NEPTUNE_1_4(),
		Parameters: &map[string]*string{
			"neptune_enable_audit_log": jsii.String("1"),
		},
	})

	cluster := neptune.NewDatabaseCluster(stack, jsii.String("nepCluster"), &neptune.DatabaseClusterProps{
		Vpc:                     vpc,
		InstanceType:            neptune.InstanceType_R5_LARGE(),
		AutoMinorVersionUpgrade: jsii.Bool(true),
		ClusterParameterGroup:   clusterParameterGroup,
		CloudwatchLogsExports: &[]neptune.LogType{
			neptune.LogType_AUDIT(),
		},
	})

	cluster.Connections().AllowDefaultPortFromAnyIpv4(jsii.String("Allow access to Neptune from anywhere"))
	return cluster
}



