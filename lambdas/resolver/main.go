package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	driver neo4j.DriverWithContext
)

func init() {
	neptuneEndpoint := os.Getenv("NEPTUNE_ENDPOINT")
	uri := fmt.Sprintf("bolt+s://%s:8182", neptuneEndpoint)

	if neptuneEndpoint == "" {
		log.Fatal("NEPTUNE_ENDPOINT environment variable must be set.")
	}

	var err error
	driver, err = neo4j.NewDriverWithContext(uri, neo4j.NoAuth())
	if err != nil {
		log.Fatalf("Failed to establish Neo4j driver connection: %v", err)
	}
	log.Println("Neptune openCypher connection established in Lambda init().")
}

func executeReadQuery(ctx context.Context, query string, params map[string]interface{}) ([]*neo4j.Record, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		log.Printf("Executing Query: %s\nWith Params: %+v", query, params)
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		return res.Collect(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]*neo4j.Record), nil
}

func executeWriteQuery(ctx context.Context, query string, params map[string]interface{}) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		log.Printf("Executing Write Query: %s\nWith Params: %+v", query, params)
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		return res.Consume(ctx)
	})
	return err
}

func hasField(field string, selectionSet []string) bool {
	for _, s := range selectionSet {
		if strings.HasPrefix(s, field) {
			return true
		}
	}
	return false
}

func handleMutationDeleteStudy(ctx context.Context, args map[string]interface{}) (bool, error) {
	studyID, ok := args["id"].(string)
	if !ok || studyID == "" {
		return false, fmt.Errorf("study ID is required for deletion")
	}

	// This query finds all nodes connected to the study (at any depth) and the study itself.
	// DETACH DELETE then removes all these nodes and their relationships.
	query := `
	MATCH (s:Study {id: $id})
	OPTIONAL MATCH (s)-[*]->(descendant)
	DETACH DELETE s, descendant
	`
	params := map[string]interface{}{"id": studyID}

	err := executeWriteQuery(ctx, query, params)
	if err != nil {
		log.Printf("Error deleting study %s: %v", studyID, err)
		return false, err
	}

	log.Printf("Successfully deleted study %s and its descendants", studyID)
	return true, nil
}

func handleQueryStudy(ctx context.Context, args map[string]interface{}, selectionSet []string) (*Study, error) {
	studyID, ok := args["id"].(string)
	if !ok || studyID == "" {
		return nil, fmt.Errorf("study ID is required")
	}

	var projectionParts []string

	// Dynamically build projection for top-level Study fields
	if hasField("id", selectionSet) {
		projectionParts = append(projectionParts, ".id")
	}
	if hasField("name", selectionSet) {
		projectionParts = append(projectionParts, ".name")
	}
	if hasField("description", selectionSet) {
		projectionParts = append(projectionParts, ".description")
	}
	if hasField("label", selectionSet) {
		projectionParts = append(projectionParts, ".label")
	}

	// Logic for 'versions' and its nested fields
	if hasField("versions", selectionSet) {
		var versionSubProjection []string
		if hasField("versions/id", selectionSet) {
			versionSubProjection = append(versionSubProjection, ".id")
		}
		if hasField("versions/rationale", selectionSet) {
			versionSubProjection = append(versionSubProjection, ".rationale")
		}

		if hasField("versions/studyDesigns", selectionSet) {
			var designSubProjection []string
			if hasField("versions/studyDesigns/id", selectionSet) {
				designSubProjection = append(designSubProjection, ".id")
			}
			if hasField("versions/studyDesigns/name", selectionSet) {
				designSubProjection = append(designSubProjection, ".name")
			}

			if hasField("versions/studyDesigns/arms", selectionSet) {
				armsProjection := "arms: [(d)-[:HAS_ARM]->(a:Arm) | a { .id, .name, .description, .type }]"
				designSubProjection = append(designSubProjection, armsProjection)
			}

			if hasField("versions/studyDesigns/epochs", selectionSet) {
				epochsProjection := "epochs: [(d)-[:HAS_EPOCH]->(e:Epoch) | e { .id, .name, .description }]"
				designSubProjection = append(designSubProjection, epochsProjection)
			}

			if hasField("versions/organizations", selectionSet) {
				orgsProjection := "organizations: [(d)-[:HAS_ORGANIZATION]->(o:Organization) | o { .id, .name, .type, .legalAddress }]"
				designSubProjection = append(designSubProjection, orgsProjection)

				if hasField("versions/organizations/legalAddress", selectionSet) {
					// `head()` is used because we expect one legal address per organization.
					legalAddressProjection := "legalAddress: head([(o)-[:HAS_LEGAL_ADDRESS]->(la:LegalAddress) | la { .*, country: head([(la)-[:LOCATED_IN]->(c:Country) | c { .* }]) }])"
					designSubProjection = append(designSubProjection, legalAddressProjection)
				}
			}

			if len(designSubProjection) > 0 {
				designsProjection := fmt.Sprintf("studyDesigns: [(v)-[:INCLUDES_DESIGN]->(d:StudyDesign) | d { %s }]", strings.Join(designSubProjection, ", "))
				versionSubProjection = append(versionSubProjection, designsProjection)
			}
		}

		if hasField("versions/amendments", selectionSet) {
			amendmentsProjection := "amendments: [(v)-[:HAS_AMENDMENT]->(am:StudyAmendment) | am { .id, .name, .summary, .rationale }]"
			versionSubProjection = append(versionSubProjection, amendmentsProjection)
		}

		if len(versionSubProjection) > 0 {
			versionsProjection := fmt.Sprintf("versions: [(s)-[:HAS_VERSION]->(v:StudyVersion) | v { %s }]", strings.Join(versionSubProjection, ", "))
			projectionParts = append(projectionParts, versionsProjection)
		}
	}

	if hasField("organizations", selectionSet) {
		// This assumes organizations are linked directly to the Study node `s`
		orgsProjection := "organizations: [(s)-[:HAS_ORGANIZATION]->(o:Organization) | o { .id, .name, .type, .legalAddress }]"
		projectionParts = append(projectionParts, orgsProjection)
	}

	if hasField("documentedBy", selectionSet) {
		// This projects only the fields that are actually being saved in the ingestion step.
		docsProjection := "documentedBy: [(s)-[:DOCUMENTED_BY]->(d:StudyDefinitionDocument) | d { .id, .name }]"
		projectionParts = append(projectionParts, docsProjection)
	}

	if len(projectionParts) == 0 {
		projectionParts = append(projectionParts, ".id") // Default projection
	}

	finalQuery := fmt.Sprintf(
		"MATCH (s:Study {id: $id}) RETURN s { %s } AS study",
		strings.Join(projectionParts, ", "),
	)

	// --- The rest of the function remains the same ---
	params := map[string]interface{}{"id": studyID}
	records, err := executeReadQuery(ctx, finalQuery, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for study %s: %w", studyID, err)
	}

	if len(records) == 0 {
		return nil, nil // Not found
	}

	record := records[0]
	studyData, ok := record.Get("study")
	if !ok {
		return nil, fmt.Errorf("could not find 'study' in result record")
	}

	var study Study
	jsonBytes, err := json.Marshal(studyData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal study data: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &study); err != nil {
		return nil, fmt.Errorf("failed to unmarshal study data into struct: %w", err)
	}

	return &study, nil
}

func handler(ctx context.Context, event AppSyncEvent) (interface{}, error) {
	log.Printf("Received AppSync event: TypeName=%s, FieldName=%s", event.Info.ParentTypeName, event.Info.FieldName)

	switch event.Info.ParentTypeName {
	case "Query":
		switch event.Info.FieldName {
		case "study":
			return handleQueryStudy(ctx, event.Arguments, event.Info.SelectionSetList)
		default:
			return nil, fmt.Errorf("unknown query field: %s", event.Info.FieldName)
		}
	case "Mutation":
		switch event.Info.FieldName {
		case "deleteStudy":
			return handleMutationDeleteStudy(ctx, event.Arguments)
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
