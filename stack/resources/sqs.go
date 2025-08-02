package resources

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatchactions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"
	sqs "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/jsii-runtime-go"
)


func NewSQSQueue ( stack awscdk.Stack, vpc awsec2.Vpc) sqs.Queue {

	deadLetterQueue := sqs.NewQueue(stack, jsii.String("DataIngestionDLQ"), &sqs.QueueProps{
		RetentionPeriod: awscdk.Duration_Days(jsii.Number(14)),
	})

	dlqMetric := deadLetterQueue.MetricApproximateNumberOfMessagesVisible(&awscloudwatch.MetricOptions{
		Statistic: jsii.String("Sum"),
		Period: awscdk.Duration_Minutes(jsii.Number(5)),
	})

	alarm := awscloudwatch.NewAlarm(stack, jsii.String("DLQMessagesAlarm"), &awscloudwatch.AlarmProps{
		AlarmName: jsii.String("DLQMessagesAlarm"),
		Metric:  dlqMetric,
		Threshold: jsii.Number(1),
		EvaluationPeriods: jsii.Number(1),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		AlarmDescription: jsii.String("Alarm when there are messages in the dead letter queue"),
	})

	topic := awssns.NewTopic(stack, jsii.String("DataIngestionAlarmTopic"), &awssns.TopicProps{
		DisplayName: jsii.String("Data Ingestion Alarm Topic"),
	})

	topic.AddSubscription(awssnssubscriptions.NewEmailSubscription(jsii.String("bahuguna_ankit@lilly.com"),nil))

	alarm.AddAlarmAction(awscloudwatchactions.NewSnsAction(topic))

	queue := sqs.NewQueue(stack, jsii.String("DataIngestionQueue"), &sqs.QueueProps{
		VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)), 
		ReceiveMessageWaitTime: awscdk.Duration_Seconds(jsii.Number(20)),
		DeadLetterQueue: &sqs.DeadLetterQueue{
			MaxReceiveCount: jsii.Number(5), 
			Queue:           deadLetterQueue,
		},
	})

	return queue
}
