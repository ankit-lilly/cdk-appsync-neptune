package query

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ankit-lilly/dtd-go-backend/internal/neptunedb/cypher"
	"github.com/ankit-lilly/dtd-go-backend/pkg/models"
	"log"
	"strings"
)

func hasField(field string, selectionSet []string) bool {
	for _, s := range selectionSet {
		if strings.HasPrefix(s, field) {
			return true
		}
	}
	return false
}

func HandleQueryStudy(ctx context.Context, args map[string]interface{}, selectionSet []string) (*models.Study, error) {
	studyID, ok := args["id"].(string)
	if !ok || studyID == "" {
		return nil, fmt.Errorf("study ID is required")
	}

	var projectionParts []string

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

			if hasField("versions/studyDesigns/encounters", selectionSet) {
				encountersProjection := "encounters: [(d)-[:HAS_ENCOUNTER]->(e:Encounter) | e { .id, .name, .label, .description }]"
				designSubProjection = append(designSubProjection, encountersProjection)
			}

			if hasField("versions/studyDesigns/activities", selectionSet) {
				activitiesProjection := "activities: [(d)-[:HAS_ACTIVITY]->(a:Activity) | a { .id, .name, .label, .description }]"
				designSubProjection = append(designSubProjection, activitiesProjection)
			}

			if hasField("versions/studyDesigns/epochs", selectionSet) {
				epochsProjection := "epochs: [(d)-[:HAS_EPOCH]->(e:Epoch) | e { .id, .name, .description }]"
				designSubProjection = append(designSubProjection, epochsProjection)
			}

			if len(designSubProjection) > 0 {
				designsProjection := fmt.Sprintf("studyDesigns: [(v)-[:INCLUDES_DESIGN]->(d:StudyDesign) | d { %s }]", strings.Join(designSubProjection, ", "))
				versionSubProjection = append(versionSubProjection, designsProjection)
			}
		}

		if hasField("versions/organizations", selectionSet) {
			var orgSubProjection []string
			if hasField("versions/organizations/id", selectionSet) {
				orgSubProjection = append(orgSubProjection, ".id")
			}
			if hasField("versions/organizations/name", selectionSet) {
				orgSubProjection = append(orgSubProjection, ".name")
			}

			if hasField("versions/organizations/legalAddress", selectionSet) {
				legalAddressProjection := `legalAddress: head([(o)-[:HAS_LEGAL_ADDRESS]->(la:LegalAddress) | la { .*, country: head([(la)-[:LOCATED_IN]->(c:Country) | c {.*}]) }])`
				orgSubProjection = append(orgSubProjection, legalAddressProjection)
			}

			orgsProjection := fmt.Sprintf("organizations: [(v)-[:HAS_ORGANIZATION]->(o:Organization) | o { %s }]", strings.Join(orgSubProjection, ", "))
			versionSubProjection = append(versionSubProjection, orgsProjection)
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

	if hasField("documentedBy", selectionSet) {
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

	log.Printf("FinalQuery: %s\nWith Params: %+v", finalQuery, studyID)
	params := map[string]interface{}{"id": studyID}
	records, err := cypher.ExecuteReadQuery(ctx, finalQuery, params)
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

	var study models.Study
	jsonBytes, err := json.Marshal(studyData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal study data: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, &study); err != nil {
		return nil, fmt.Errorf("failed to unmarshal study data into struct: %w", err)
	}

	return &study, nil
}
