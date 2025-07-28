package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sdrProcessor/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	driver neo4j.DriverWithContext
)

type UsdmPayload struct {
	Study         models.Study `json:"study"`
	UsdmVersion   string       `json:"usdmVersion"`
	SystemName    string       `json:"systemName"`
	SystemVersion string       `json:"systemVersion"`
}

func init() {
	neptuneEndpoint := os.Getenv("NEPTUNE_ENDPOINT")
	if neptuneEndpoint == "" {
		log.Fatal("NEPTUNE_ENDPOINT environment variable must be set.")
	}
	uri := fmt.Sprintf("bolt+s://%s:8182", neptuneEndpoint)

	var err error
	driver, err = neo4j.NewDriverWithContext(uri, neo4j.NoAuth())
	if err != nil {
		log.Fatalf("Failed to establish Neo4j driver connection: %v", err)
	}
	log.Println("Neptune openCypher connection established in Lambda init().")
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

		if err := processStudy(ctx, study); err != nil {
			log.Printf("Error processing study %s: %v", study.ID, err)
			return err
		}
	}
	return nil
}

func processStudy(ctx context.Context, study models.Study) error {
	var studyMap map[string]interface{}
	studyJSON, _ := json.Marshal(study)
	json.Unmarshal(studyJSON, &studyMap)

	params := map[string]interface{}{
		"study": studyMap,
	}

	// --- ALL QUERIES BELOW SHOULD REPLACE THE EXISTING ONES ---

	qStudyAndVersions := `
    MERGE (s:Study {id: $study.id})
    ON CREATE SET
        s.name = $study.name,
        s.description = $study.description,
        s.label = $study.label
    ON MATCH SET
        s.name = $study.name,
        s.description = $study.description,
        s.label = $study.label

    WITH s, $study.versions AS versions
    UNWIND versions AS v
    MERGE (sv:StudyVersion {id: v.id})
    ON CREATE SET
        sv.versionIdentifier = v.versionIdentifier,
        sv.rationale = v.rationale
    MERGE (s)-[:HAS_VERSION]->(sv)`

	qDesigns := `
    UNWIND $study.versions AS v
    UNWIND v.studyDesigns AS d
    MATCH (sv:StudyVersion {id: v.id})
    MERGE (sd:StudyDesign {id: d.id})
    ON CREATE SET
        sd.name = d.name,
        sd.description = d.description,
        sd.studyType  = d.studyType.decode
    ON MATCH SET
        sd.name = d.name,
        sd.description = d.description,
        sd.studyType  = d.studyType.decode
    MERGE (sv)-[:INCLUDES_DESIGN]->(sd)`

	qArms := `
    UNWIND $study.versions AS v
    UNWIND v.studyDesigns AS d
    UNWIND d.arms AS a
    MATCH (sd:StudyDesign {id: d.id})
    MERGE (arm:Arm {id: a.id})
    ON CREATE SET
        arm.name = a.name,
        arm.description = a.description,
        arm.type = a.dataOriginType.decode
    ON MATCH SET
        arm.name = a.name,
        arm.description = a.description,
        arm.type = a.dataOriginType.decode
    MERGE (sd)-[:HAS_ARM]->(arm)`

	qEpochs := `
    UNWIND $study.versions AS v
    UNWIND v.studyDesigns AS d
    UNWIND d.epochs AS e
    MATCH (sd:StudyDesign {id: d.id})
    MERGE (ep:Epoch {id: e.id})
    ON CREATE SET
        ep.name = e.name,
        ep.description = e.description
    ON MATCH SET
        ep.name = e.name,
        ep.description = e.description
    MERGE (sd)-[:HAS_EPOCH]->(ep)

    WITH ep, e.previousId AS prevId
    WHERE prevId IS NOT NULL AND prevId <> ""
    MATCH (prev:Epoch {id: prevId})
    MERGE (prev)-[:PRECEDES]->(ep)`

	qAmendments := `
    UNWIND $study.versions AS v
    UNWIND v.amendments AS a
    MATCH (sv:StudyVersion {id: v.id})
    MERGE (am:StudyAmendment {id: a.id})
    ON CREATE SET
        am.name = a.name,
        am.summary = a.summary,
        am.rationale = a.rationale
    ON MATCH SET
        am.name = a.name,
        am.summary = a.summary,
        am.rationale = a.rationale
    MERGE (sv)-[:HAS_AMENDMENT]->(am)`

	qDocuments := `
    UNWIND $study.documentedBy AS d
    MATCH (s:Study {id: $study.id})
    MERGE (doc:StudyDefinitionDocument {id: d.id})
    ON CREATE SET
        doc.name = d.name
    ON MATCH SET
        doc.name = d.name
    MERGE (s)-[:DOCUMENTED_BY]->(doc)`

	// This is the corrected qOrganizations query from Step 2
	qOrganizations := `
    UNWIND $study.versions AS v
    UNWIND v.organizations AS o
    MATCH (s:Study {id: $study.id})
    MERGE (org:Organization {id: o.id})
    ON CREATE SET
        org.name = o.name,
        org.type = o.type.decode
    ON MATCH SET
        org.name = o.name,
        org.type = o.type.decode
    MERGE (s)-[:HAS_ORGANIZATION]->(org)`

	legalAddress := `
		UNWIND $study.versions AS v
		UNWIND v.organizations AS o
		MATCH (org:Organization {id: o.id})

		MERGE (la:LegalAddress {id: o.legalAddress.id})
		ON CREATE SET
			la.extensionAttributes = o.legalAddress.extensionAttributes,
			la.text = o.legalAddress.text,
			la.lines = o.legalAddress.lines,
			la.city = o.legalAddress.city,
			la.district = o.legalAddress.district,
			la.state = o.legalAddress.state,
			la.postalCode = o.legalAddress.postalCode,
			la.instanceType = o.legalAddress.instanceType
		ON MATCH SET
			la.extensionAttributes = o.legalAddress.extensionAttributes,
			la.text = o.legalAddress.text,
			la.lines = o.legalAddress.lines,
			la.city = o.legalAddress.city,
			la.district = o.legalAddress.district,
			la.state = o.legalAddress.state,
			la.postalCode = o.legalAddress.postalCode,
			la.instanceType = o.LegalAddress.instanceType
		MERGE (org)-[:HAS_LEGAL_ADDRESS]->(la)
		MERGE (c:Country {id: o.legalAddress.country.id})
		ON CREATE SET
			c.code = o.legalAddress.country.code,
			c.codeSystem = o.legalAddress.country.codeSystem,
			c.codeSystemVersion = o.legalAddress.country.codeSystemVersion,
			c.decode = o.legalAddress.country.decode,
			c.instanceType = o.legalAddress.country.instanceType
		ON MATCH SET
			c.code = o.legalAddress.country.code,
			c.codeSystem = o.legalAddress.country.codeSystem,
			c.codeSystemVersion = o.legalAddress.country.codeSystemVersion,
			c.decode = o.legalAddress.country.decode,
			c.instanceType = o.legalAddress.country.instanceType
		MERGE (la)-[:LOCATED_IN]->(c)`

	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		for _, q := range []string{
			qStudyAndVersions,
			qDesigns,
			qArms,
			qEpochs,
			qAmendments,
			qOrganizations,
			qDocuments,
			legalAddress,
		} {
			result, err := tx.Run(ctx, q, params)
			if err != nil {
				log.Printf("Error executing query part: %v. Query was: %s", err, q)
				return nil, fmt.Errorf("error executing query part: %w", err)
			}
			if err = result.Err(); err != nil {
				log.Printf("Error consuming result: %v. Query was: %s", err, q)
				return nil, fmt.Errorf("error consuming result: %w", err)
			}
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("failed to execute study upsert queries: %w", err)
	}

	log.Printf("Successfully upserted study %s and its components.", study.ID)
	return nil
}

func main() {
	lambda.Start(handler)
}
