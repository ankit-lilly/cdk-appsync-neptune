package query

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ankit-lilly/dtd-go-backend/internal/neptunedb/cypher"
	"github.com/ankit-lilly/dtd-go-backend/pkg/models"
	"log"
)

func HandleQueryActivities(ctx context.Context, args map[string]interface{}, selectionSet []string) ([]*models.Activity, error) {

	finalQuery := `
		MATCH (a:Activity)
		OPTIONAL MATCH (a)-[:HAS_DEFINED_PROCEDURE]->(p:DefinedProcedure)
		WITH a, collect(p) AS procedures
		RETURN a {
			.id,
			.name,
			.label,
			.description,
			definedProcedures: procedures
		} AS activity`

	records, err := cypher.ExecuteReadQuery(ctx, finalQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for activities: %w", err)
	}

	var activities []*models.Activity

	for _, record := range records {
		activityData, ok := record.Get("activity")
		if !ok {
			log.Println("Warning: found a record without an 'activity' field")
			continue
		}
		log.Printf("Processing activity data: %v", activityData)

		var activity models.Activity
		jsonBytes, err := json.Marshal(activityData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal activity data: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &activity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal activity data into struct: %w", err)
		}
		activities = append(activities, &activity)
	}

	return activities, nil
}
