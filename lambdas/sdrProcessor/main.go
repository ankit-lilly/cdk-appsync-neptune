package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ankit-lilly/dtd-go-backend/pkg/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type UsdmPayload struct {
	Study         models.Study `json:"study"`
	UsdmVersion   string       `json:"usdmVersion"`
	SystemName    string       `json:"systemName"`
	SystemVersion string       `json:"systemVersion"`
}

func parseStudyData(data string) (models.Study, error) {
	var payload UsdmPayload
	err := json.Unmarshal([]byte(data), &payload)
	if err != nil {
		return models.Study{}, fmt.Errorf("failed to parse usdm payload: %w", err)
	}
	return payload.Study, nil
}

func handler(ctx context.Context, event events.SQSEvent) error {
	for _, message := range event.Records {
		log.Printf("Processing message ID: %s", message.MessageId)

		study, err := parseStudyData(message.Body)
		if err != nil {
			log.Printf("Error parsing study data: %v", err)
			continue
		}

		if err := SaveStudyToGraph(ctx, study); err != nil {
			log.Printf("Error processing study %s: %v", study.ID, err)
			return err
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
