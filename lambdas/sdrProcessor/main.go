package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ankit-lilly/dtd-go-backend/internal/neptunedb"
	"github.com/ankit-lilly/dtd-go-backend/pkg/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type UsdmPayload struct {
	Study         models.Study `json:"study"`
	UsdmVersion   string       `json:"usdmVersion"`
	SystemName    string       `json:"systemName"`
	SystemVersion string       `json:"systemVersion"`
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

	qStudyAndVersions := `
    MERGE (s:Study {id: $study.id})
    SET
        s.name = $study.name,
        s.description = $study.description,
        s.label = $study.label
    WITH s
    UNWIND $study.versions AS v
    MERGE (sv:StudyVersion {id: v.id})
    SET
        sv.versionIdentifier = v.versionIdentifier,
        sv.rationale = v.rationale
    MERGE (s)-[:HAS_VERSION]->(sv)`

	qDesigns := `
    UNWIND $study.versions AS v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.studyDesigns AS d
    MERGE (sd:StudyDesign {id: d.id})
    SET
        sd.name = d.name,
        sd.description = d.description,
        sd.studyType  = d.studyType.decode
    MERGE (sv)-[:INCLUDES_DESIGN]->(sd)`

	qArms := `
    UNWIND $study.versions AS v
    UNWIND v.studyDesigns AS d
    MATCH (sd:StudyDesign {id: d.id})
    UNWIND d.arms AS a
    MERGE (arm:Arm {id: a.id})
    SET
        arm.name = a.name,
        arm.description = a.description
    MERGE (sd)-[:HAS_ARM]->(arm)

    // Process Arm Type correctly
    WITH arm, a
    WHERE a.dataOriginType IS NOT NULL
    MERGE (dot:ArmDataOriginType {id: a.dataOriginType.id})
    SET
        dot.code = a.dataOriginType.code,
        dot.codeSystem = a.dataOriginType.codeSystem,
        dot.codeSystemVersion = a.dataOriginType.codeSystemVersion,
        dot.decode = a.dataOriginType.decode,
        dot.instanceType = a.dataOriginType.instanceType
    MERGE (arm)-[:HAS_DATA_ORIGIN_TYPE]->(dot)`

	qEncounters := `
    UNWIND $study.versions AS v
    UNWIND v.studyDesigns AS d
    MATCH (sd:StudyDesign {id: d.id})
    UNWIND d.encounters AS a
    MERGE (enc:Encounter {id: a.id})
    SET
        enc.name = a.name,
        enc.description = a.description,
	enc.label = a.label,
	enc.description = a.description,
	enc.scheduledAtId = a.scheduledAtId
    MERGE (sd)-[:HAS_ENCOUNTER]->(enc)
    MERGE (ect:EncounterType {id: a.type.id})
    SET
        ect.code = a.type.code,
	ect.codeSystem = a.type.codeSystem,
	ect.codeSystemVersion = a.type.codeSystemVersion,
	ect.decode = a.type.decode,
	ect.instanceType = a.type.instanceType
    MERGE (enc)-[:HAS_ENCOUNTER_TYPE]->(ect)
    `

	qActivities := `
    UNWIND $study.versions AS v
    UNWIND v.studyDesigns AS d
    MATCH (sd:StudyDesign {id: d.id})
    UNWIND d.activities AS a
    MERGE (act:Activity {id: a.id})
    SET
        act.name = a.name,
        act.description = a.description,
        act.label = a.label,
        act.instanceType = a.instanceType
    MERGE (sd)-[:HAS_ACTIVITY]->(act)
    WITH act, a
    UNWIND a.definedProcedures AS dp
    MERGE (proc:DefinedProcedure {id: dp.id})
    SET
        proc.name = dp.name,
        proc.description = dp.description,
        proc.label = dp.label,
        proc.procedureType = dp.procedureType,
        proc.studyInterventionId = dp.studyInterventionId,
        proc.instanceType = dp.instanceType
    MERGE (act)-[:HAS_DEFINED_PROCEDURE]->(proc)
    WITH proc, dp
    WHERE dp.code IS NOT NULL
    MERGE (c:Code {id: dp.code.id})
    SET
        c.code = dp.code.code,
        c.codeSystem = dp.code.codeSystem,
        c.codeSystemVersion = dp.code.codeSystemVersion,
        c.decode = dp.code.decode,
        c.instanceType = dp.code.instanceType
    MERGE (proc)-[:HAS_CODE]->(c)`

	qEpochs := `
    UNWIND $study.versions AS v
    UNWIND v.studyDesigns AS d
    MATCH (sd:StudyDesign {id: d.id})
    UNWIND d.epochs AS e
    MERGE (ep:Epoch {id: e.id})
    SET
        ep.name = e.name,
        ep.description = e.description
    MERGE (sd)-[:HAS_EPOCH]->(ep)

    WITH ep, e.previousId AS prevId
    WHERE prevId IS NOT NULL AND prevId <> ""
    MATCH (prev:Epoch {id: prevId})
    MERGE (prev)-[:PRECEDES]->(ep)`

	qAmendments := `
    UNWIND $study.versions AS v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.amendments AS a
    MERGE (am:StudyAmendment {id: a.id})
    SET
        am.name = a.name,
        am.summary = a.summary,
        am.rationale = a.rationale
    MERGE (sv)-[:HAS_AMENDMENT]->(am)`

	qDocuments := `
    UNWIND $study.documentedBy AS d
    MATCH (s:Study {id: $study.id})
    MERGE (doc:StudyDefinitionDocument {id: d.id})
    SET
        doc.name = CASE WHEN d.name IS NOT NULL THEN d.name ELSE '' END,
        doc.description = CASE WHEN d.description IS NOT NULL THEN d.description ELSE '' END,
        doc.label = CASE WHEN d.label IS NOT NULL THEN d.label ELSE '' END
    MERGE (s)-[:DOCUMENTED_BY]->(doc)`

	qOrganizationsAndAddresses := `
    UNWIND $study.versions AS v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.organizations AS o
    MERGE (org:Organization {id: o.id})
    SET
        org.name = o.name
    MERGE (sv)-[:HAS_ORGANIZATION]->(org)

    WITH org, o
    WHERE o.type IS NOT NULL
    MERGE (ot:OrganizationType {id: o.type.id})
    SET
        ot.code = o.type.code,
        ot.codeSystem = o.type.codeSystem,
        ot.codeSystemVersion = o.type.codeSystemVersion,
        ot.decode = o.type.decode,
        ot.instanceType = o.type.instanceType
    MERGE (org)-[:HAS_ORGANIZATION_TYPE]->(ot)

    WITH org, o
    WHERE o.legalAddress IS NOT NULL
    MERGE (la:LegalAddress {id: o.legalAddress.id})
    SET
        la.text = o.legalAddress.text,
        la.city = o.legalAddress.city,
        la.district = o.legalAddress.district,
        la.state = o.legalAddress.state,
        la.postalCode = o.legalAddress.postalCode,
        la.instanceType = o.legalAddress.instanceType
    MERGE (org)-[:HAS_LEGAL_ADDRESS]->(la)

    WITH la, o
    WHERE o.legalAddress.country IS NOT NULL
    MERGE (c:Country {id: o.legalAddress.country.id})
    SET
        c.code = o.legalAddress.country.code,
        c.codeSystem = o.legalAddress.country.codeSystem,
        c.codeSystemVersion = o.legalAddress.country.codeSystemVersion,
        c.decode = o.legalAddress.country.decode,
        c.instanceType = o.legalAddress.country.instanceType
    MERGE (la)-[:LOCATED_IN]->(c)`

	driver := neptunedb.GetDriver()
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		for _, q := range []string{
			qStudyAndVersions,
			qDesigns,
			qArms,
			qEpochs,
			qAmendments,
			qOrganizationsAndAddresses,
			qDocuments,
			qEncounters,
			qActivities,
		} {
			result, err := tx.Run(ctx, q, params)
			if err != nil {
				log.Printf("Error executing query part: %v. Query was: %s", err, q)
				return nil, fmt.Errorf("error executing query part: %w", err)
			}
			if err = result.Err(); err != nil {
				log.Printf("Error from result: %v. Query was: %s", err, q)
				return nil, fmt.Errorf("error from result: %w", err)
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
