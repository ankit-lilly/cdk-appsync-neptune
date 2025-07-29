package resources

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	sqs "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/jsii-runtime-go"
)


func NewSQSQueue ( stack awscdk.Stack, vpc awsec2.Vpc) sqs.Queue {

	deadLetterQueue := sqs.NewQueue(stack, jsii.String("DataIngestionDLQ"), &sqs.QueueProps{
		RetentionPeriod: awscdk.Duration_Days(jsii.Number(14)),
	})

	queue := sqs.NewQueue(stack, jsii.String("DataIngestionQueue"), &sqs.QueueProps{
		VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)), 
		DeadLetterQueue: &sqs.DeadLetterQueue{
			MaxReceiveCount: jsii.Number(5), 
			Queue:           deadLetterQueue,
		},
	})

	return queue
}
