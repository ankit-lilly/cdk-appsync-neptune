package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ankit-lilly/dtd-go-backend/lambdas/resolver/mutations"
	"github.com/ankit-lilly/dtd-go-backend/lambdas/resolver/query"

	"github.com/aws/aws-lambda-go/lambda"
)

type AppSyncEvent struct {
	Info      AppSyncInfo            `json:"info"`
	Arguments map[string]interface{} `json:"arguments"`
	Source    map[string]interface{} `json:"source"`
}

type AppSyncInfo struct {
	FieldName        string   `json:"fieldName"`
	ParentTypeName   string   `json:"parentTypeName"`
	SelectionSetList []string `json:"selectionSetList"`
}

func handler(ctx context.Context, event AppSyncEvent) (interface{}, error) {
	log.Printf("Received AppSync event: TypeName=%s, FieldName=%s", event.Info.ParentTypeName, event.Info.FieldName)

	switch event.Info.ParentTypeName {
	case "Query":
		switch event.Info.FieldName {
		case "study":
			return query.HandleQueryStudy(ctx, event.Arguments, event.Info.SelectionSetList)
		case "studies":
			return query.HandleQueryStudies(ctx, event.Arguments, event.Info.SelectionSetList)
		case "activities":
			return query.HandleQueryActivities(ctx, event.Arguments, event.Info.SelectionSetList)
		case "encounters":
			return query.HandleQueryEncounters(ctx, event.Arguments, event.Info.SelectionSetList)
		case "graphStats":
			return query.HandleQueryGraphStats(ctx, event.Arguments, event.Info.SelectionSetList)
		default:
			return nil, fmt.Errorf("unknown query field: %s", event.Info.FieldName)
		}
	case "Mutation":
		switch event.Info.FieldName {
		case "deleteStudy":
			return mutations.HandleMutationDeleteStudy(ctx, event.Arguments)
		default:
			return nil, fmt.Errorf("unknown mutation field: %s", event.Info.FieldName)
		}
	default:
		return nil, fmt.Errorf("unsupported type: %s", event.Info.ParentTypeName)
	}
}

func main() {
	lambda.Start(handler)
}
