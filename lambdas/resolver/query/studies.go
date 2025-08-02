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

var schemaMappings = map[string]map[string]string{
	"Study": {
		"versions":     "HAS_VERSION",
		"documentedBy": "DOCUMENTED_BY",
	},
	"StudyVersion": {
		"organizations":      "HAS_ORGANIZATION",
		"amendments":         "HAS_AMENDMENT",
		"studyDesigns":       "INCLUDES_DESIGN",
		"biomedicalConcepts": "HAS_BIOMEDICAL_CONCEPT",
		"bcSurrogates":       "HAS_BC_SURROGATE",
		"enrollments":        "HAS_ENROLLMENT",
		"conditions":         "HAS_CONDITION",
	},
	"StudyDesign": {
		"arms":       "HAS_ARM",
		"epochs":     "HAS_EPOCH",
		"activities": "HAS_ACTIVITY",
		"encounters": "HAS_ENCOUNTER",
	},
	"Organization": {
		"legalAddress": "HAS_LEGAL_ADDRESS",
		"type":         "HAS_ORGANIZATION_TYPE",
	},
	"LegalAddress": {
		"country": "LOCATED_IN",
	},
	"BioMedicalConcept": {
		"code": "HAS_BM_CODE",
	},
	"DefinedProcedure": {
		"code": "HAS_CODE",
	},
	"StudyAmendment": {
		"primaryReason": "HAS_AMENDMENT_PRIMARY_REASON",
	},
}

// 2. New map to know the label of the children
var fieldToChildLabel = map[string]string{
	"versions":           "StudyVersion",
	"documentedBy":       "StudyDefinitionDocument",
	"organizations":      "Organization",
	"legalAddress":       "LegalAddress",
	"country":            "Country",
	"studyDesigns":       "StudyDesign",
	"arms":               "Arm",
	"epochs":             "Epoch",
	"activities":         "Activity",
	"definedProcedures":  "DefinedProcedure",
	"biomedicalConcepts": "BioMedicalConcept",
	"code":               "Code",
	"title":              "StudyTitle",
}

var toOneRelations = map[string]bool{
	"legalAddress":  true,
	"documentedBy":  false,
	"studyType":     true,
	"country":       true,
	"type":          true,
	"primaryReason": true,
}

func HandleQueryStudies(ctx context.Context, args map[string]any, selectionSet []string) ([]*models.Study, error) {

	graphSource := gremlin.GetReaderGraphTraversalSource()
	if graphSource == nil {
		return nil, fmt.Errorf("graph source is not initialized")
	}

	parsedFields := parseSelectionSet(selectionSet)

	projectionTraversal := buildProjection(parsedFields, "Study")

	finalTraversal := graphSource.V().HasLabel("Study").Map(projectionTraversal)

	results, err := finalTraversal.ToList()
	if err != nil {
		return nil, fmt.Errorf("failed to query studies: %w", err)
	}

	var processedResults []any
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

func parseSelectionSet(selectionSet []string) map[string]any {
	root := make(map[string]any)
	for _, path := range selectionSet {
		parts := strings.Split(path, "/")
		currentMap := root

		for i, part := range parts {
			isLastPart := (i == len(parts)-1)

			if isLastPart {
				if _, ok := currentMap[part].(map[string]any); !ok {
					currentMap[part] = true
				}
			} else {
				nextMap, ok := currentMap[part].(map[string]any)

				if !ok {
					nextMap = make(map[string]any)
					currentMap[part] = nextMap
				}

				currentMap = nextMap
			}
		}
	}
	log.Printf("Parsed selection set: %v", root)
	return root
}

func buildProjection(fields map[string]any, parentLabel string) *gremlingo.GraphTraversal {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}

	projectArgs := make([]any, len(keys))
	for i, k := range keys {
		projectArgs[i] = k
	}
	t := gremlingo.T__.Project(projectArgs...)

	for _, fieldName := range keys {
		subFields := fields[fieldName]
		subMap, isMap := subFields.(map[string]any)

		if isMap {
			edgeLabel, ok := schemaMappings[parentLabel][fieldName]
			if !ok {
				t = t.By(gremlingo.T__.Constant(nil))
				continue
			}

			childLabel := fieldToChildLabel[fieldName]

			traversal := gremlingo.T__.Out(edgeLabel).Map(buildProjection(subMap, childLabel))

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
	return t
}

func convertMap(i any) any {
	switch v := i.(type) {
	case map[any]any:
		m := make(map[string]any)
		for key, val := range v {
			strKey := fmt.Sprintf("%v", key)
			m[strKey] = convertMap(val)
		}
		return m
	case map[string]any:
		for key, val := range v {
			v[key] = convertMap(val)
		}
		return v
	case []any:
		for i, val := range v {
			v[i] = convertMap(val)
		}
		return v
	default:
		return v
	}
}
