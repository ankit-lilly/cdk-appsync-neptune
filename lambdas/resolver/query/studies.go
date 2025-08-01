package query

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ankit-lilly/dtd-go-backend/internal/neptunedb/gremlin"
	"github.com/ankit-lilly/dtd-go-backend/pkg/models"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

var schemaMappings = map[string]string{
	"versions":      "HAS_VERSION",
	"organizations": "HAS_ORGANIZATION",
	"amendments":    "HAS_AMENDMENT",
	"studyDesigns":  "INCLUDES_DESIGN",
	"epochs":        "HAS_EPOCH",
	"arms":          "HAS_ARM",
	"legalAddress":  "HAS_LEGAL_ADDRESS",
	"documentedBy":  "DOCUMENTED_BY",
	"activities":    "HAS_ACTIVITY",
	"encounters":    "HAS_ENCOUNTER",
}

var toOneRelations = map[string]bool{
	"legalAddress": true,
	"documentedBy": false,
	"studyType":    true,
	"country":      true,
	"type":         true,
}

func HandleQueryStudies(ctx context.Context, args map[string]interface{}, selectionSet []string) ([]*models.Study, error) {

	graphSource := gremlin.GetReaderGraphTraversalSource()
	if graphSource == nil {
		return nil, fmt.Errorf("graph source is not initialized")
	}

	parsedFields := parseSelectionSet(selectionSet)

	projectionTraversal := buildProjection(parsedFields)

	finalTraversal := graphSource.V().HasLabel("Study").Map(projectionTraversal)

	results, err := finalTraversal.ToList()
	if err != nil {
		return nil, fmt.Errorf("failed to query studies: %w", err)
	}

	var processedResults []interface{}
	for _, result := range results {
		processedResults = append(processedResults, convertMap(result.Data))
	}

	jsonBytes, err := json.Marshal(processedResults)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal results to JSON: %w", err)
	}

	var studies []*models.Study
	if err := json.Unmarshal(jsonBytes, &studies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to structs: %w", err)
	}

	log.Printf("Successfully converted %d studies to structs.", len(studies))

	return studies, nil

}

func parseSelectionSet(selectionSet []string) map[string]interface{} {
	root := make(map[string]interface{})
	for _, path := range selectionSet {
		parts := strings.Split(path, "/")
		currentMap := root

		for i, part := range parts {
			isLastPart := (i == len(parts)-1)

			if isLastPart {
				if _, ok := currentMap[part].(map[string]interface{}); !ok {
					currentMap[part] = true
				}
			} else {
				nextMap, ok := currentMap[part].(map[string]interface{})

				if !ok {
					nextMap = make(map[string]interface{})
					currentMap[part] = nextMap
				}

				currentMap = nextMap
			}
		}
	}
	log.Printf("Parsed selection set: %v", root)
	return root
}

func buildProjection(fields map[string]interface{}) *gremlingo.GraphTraversal {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}

	projectArgs := make([]interface{}, len(keys))
	for i, k := range keys {
		projectArgs[i] = k
	}
	t := gremlingo.T__.Project(projectArgs...)

	for _, fieldName := range keys {
		subFields := fields[fieldName]

		subMap, isMap := subFields.(map[string]interface{})

		if isMap {
			edgeLabel, ok := schemaMappings[fieldName]
			if !ok {
				t = t.By(gremlingo.T__.Constant(nil))
				continue
			}

			traversal := gremlingo.T__.Out(edgeLabel).Map(buildProjection(subMap))

			if !toOneRelations[fieldName] {
				traversal = traversal.Fold()
			}
			t = t.By(traversal)

		} else {
			if fieldName == "id" {
				t = t.By(gremlingo.T__.Id())
			} else {
				t = t.By(gremlingo.T__.Values(fieldName).Unfold())
			}
		}
	}
	log.Printf("Built projection: %v", t)
	return t
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
