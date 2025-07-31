package query

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ankit-lilly/dtd-go-backend/internal/neptunedb/gremlin"
	"github.com/ankit-lilly/dtd-go-backend/pkg/models"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func HandleQueryStudies(ctx context.Context, args map[string]interface{}, selectionSet []string) (*models.Study, error) {

	graphSource := gremlin.GetReaderGraphTraversalSource()

	if graphSource == nil {
		return nil, fmt.Errorf("graph source is not initialized")
	}

	log.Printf("Executing query for studies with args: %v", selectionSet)


	results, err := graphSource.V().HasLabel("Study").Project("versions", "name", "label", "documentedBy").By(
		gremlingo.T__.ValueMap(true),
	).By(
		gremlingo.T__.Out("HAS_VERSION").ValueMap(true).Fold(),
	).By(
		gremlingo.T__.Out("HAS_VERSION").Out("INCLUDES_DESIGN").Out("HAS_ARM").ValueMap(true).Fold(),
	).ToList()

	if err != nil {
		return nil, fmt.Errorf("failed to query studies: %w", err)
	}

	log.Printf("Traversal results: %v", results)
	res := convertMap(results)

	jsonBytes, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal results: %w", err)
	}

	log.Printf("Traversal results JSON: %s", jsonBytes)

	return nil, fmt.Errorf("not implemented yet")

}

func convertMap(i interface{}) interface{} {
	switch v := i.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for key, val := range v {
			strKey := fmt.Sprintf("%v", key)
			m[strKey] = convertMap(val)
		}
		return m
	case map[string]interface{}:
		for key, val := range v {
			v[key] = convertMap(val)
		}
		return v
	case []interface{}:
		for i, val := range v {
			v[i] = convertMap(val)
		}
		return v
	default:
		return v
	}
}
