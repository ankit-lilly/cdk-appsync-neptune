package query

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ankit-lilly/dtd-go-backend/internal/neptunedb/cypher"
	"github.com/ankit-lilly/dtd-go-backend/pkg/models"
	"log"
)

func HandleQueryEncounters(ctx context.Context, args map[string]any, selectionSet []string) ([]*models.Encounter, error) {

	finalQuery := `
		MATCH (e:Encounter)
		OPTIONAL MATCH (e)-[:HAS_ENCOUNTER_TYPE]->(p:EncounterType)
		WITH e, collect(p) AS encounterTypes
		RETURN e {
			.id,
			.name,
			.label,
			.description,
			type: encounterTypes
		} AS encounter`

	records, err := cypher.ExecuteReadQuery(ctx, finalQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for encounters: %w", err)
	}
	log.Println("Query for encounters executed successfully.", records)
	var encounters []*models.Encounter
	for _, record := range records {
		encounterData, ok := record.Get("encounter")
		if !ok {
			log.Println("Warning: found a record without an 'encounter' field")
			continue
		}

		log.Printf("Processing encounter data: %v", encounterData)
		var encounter models.Encounter
		jsonBytes, err := json.Marshal(encounterData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal encounter data: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &encounter); err != nil {
			return nil, fmt.Errorf("failed to unmarshal encounter data into struct: %w", err)
		}
		encounters = append(encounters, &encounter)
	}
	return encounters, nil
}
