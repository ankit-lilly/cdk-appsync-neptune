package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sdrHandler/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type IncomingPayload struct {
	Study models.Study `json:"study"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var payload IncomingPayload

	queueURL := os.Getenv("QUEUE_URL")

	if queueURL == "" {
		log.Fatal("QUEUE_URL environment variable is not set.")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	q := sqs.NewFromConfig(cfg)

	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid request body. Please provide a valid JSON payload.",
		}, nil
	}

	output, err := q.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: &request.Body,
	})

	if err != nil {
		log.Printf("Failed to send message to SQS: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Failed to submit SDR. Please try again later.",
		}, nil
	}

	log.Printf("Message sent to SQS with ID: %s", *output.MessageId)

	return events.APIGatewayProxyResponse{
		StatusCode: 202,
		Body:       "SDR Successfully submitted. You'll receive a confirmation once it's processed.",
	}, nil
}

func main() {
	lambda.Start(handler)
}
