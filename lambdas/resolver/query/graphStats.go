package query

import (
	"context"
	"fmt"
	"log"

	"github.com/ankit-lilly/dtd-go-backend/internal/neptunedb/gremlin"
	"github.com/ankit-lilly/dtd-go-backend/pkg/models"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func HandleQueryGraphStats(ctx context.Context, args map[string]any, selectionSet []string) ([]*models.NodeCount, error) {
	graphSource := gremlin.GetReaderGraphTraversalSource()
	if graphSource == nil {
		return nil, fmt.Errorf("graph source is not initialized")
	}

	log.Println("Executing query for graph stats (node counts)")

	results, err := graphSource.V().GroupCount().By(gremlingo.T.Label).Next()
	if err != nil {
		return nil, fmt.Errorf("failed to query graph stats: %w", err)
	}

	if results == nil {
		return nil, nil
	}

	log.Printf("Received results: %v", results)
	processedResults := convertMap(results.Data)

	log.Printf("Processing results: %v", processedResults)

	var graphStats []*models.NodeCount
	for key, value := range processedResults.(map[string]any) {
		node := &models.NodeCount{
			Label: key,
			Count: value.(int64),
		}

		graphStats = append(graphStats, node)
	}

	return graphStats, nil

}
