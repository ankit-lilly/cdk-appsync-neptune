package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ankit-lilly/dtd-go-backend/internal/neptunedb/cypher"
	"github.com/ankit-lilly/dtd-go-backend/pkg/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func SaveStudyToGraph(ctx context.Context, study models.Study) error {

	var studyMap map[string]any

	studyJSON, err := json.Marshal(study)

	if err != nil {
		return fmt.Errorf("failed to marshal study: %w", err)
	}

	err = json.Unmarshal(studyJSON, &studyMap)

	if err != nil {
		return fmt.Errorf("failed to unmarshal study JSON: %w", err)
	}

	params := map[string]any{
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
			sd.label = d.label,
			sd.description = d.description,
			sd.rationale = d.rationale,
			sd.instanceType = d.instanceType
	MERGE (sv)-[:INCLUDES_DESIGN]->(sd)
	WITH sd, d
	WHERE d.studyType IS NOT NULL
	MERGE (st:Code {id: d.studyType.id})
	SET
			st.code = d.studyType.code,
			st.codeSystem = d.studyType.codeSystem,
			st.codeSystemVersion = d.studyType.codeSystemVersion,
			st.decode = d.studyType.decode,
			st.instanceType = d.studyType.instanceType
	MERGE (sd)-[:HAS_TYPE]->(st)`

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
					am.description = a.description,
					am.label = a.label,
					am.number = a.number,
					am.summary = a.summary,
					am.instanceType = a.instanceType
			MERGE (sv)-[:HAS_AMENDMENT]->(am)

			WITH am, a
			WHERE a.primaryReason IS NOT NULL
			MERGE (reason:StudyAmendmentReason {id: a.primaryReason.id})
			SET reason.instanceType = a.primaryReason.instanceType
			MERGE (am)-[:HAS_PRIMARY_REASON]->(reason)
			WITH reason, a, am
			MERGE (reasonCode:Code {id: a.primaryReason.code.id})
			SET
					reasonCode.code = a.primaryReason.code.code,
					reasonCode.codeSystem = a.primaryReason.code.codeSystem,
					reasonCode.codeSystemVersion = a.primaryReason.code.codeSystemVersion,
					reasonCode.decode = a.primaryReason.code.decode,
					reasonCode.instanceType = a.primaryReason.code.instanceType
			MERGE (reason)-[:HAS_CODE]->(reasonCode)

			WITH am, a
			UNWIND a.enrollments AS enrollment
			MERGE (enroll:SubjectEnrollment {id: enrollment.id})
			SET
					enroll.name = enrollment.name,
					enroll.instanceType = enrollment.instanceType
			MERGE (am)-[:HAS_ENROLLMENT]->(enroll)

			WITH enroll, enrollment
			WHERE enrollment.quantity IS NOT NULL
			MERGE (eq:Quantity {id: enrollment.quantity.id})
			SET
					eq.value = enrollment.quantity.value,
					eq.unit = enrollment.quantity.unit,
					eq.instanceType = enrollment.quantity.instanceType
			MERGE (enroll)-[:HAS_QUANTITY]->(eq)
		`

	qDocuments := `
    UNWIND $study.documentedBy AS d
    MATCH (s:Study {id: $study.id})
    MERGE (doc:StudyDefinitionDocument {id: d.id})
    SET
        doc.name = CASE WHEN d.name IS NOT NULL THEN d.name ELSE '' END,
        doc.description = CASE WHEN d.description IS NOT NULL THEN d.description ELSE '' END,
        doc.label = CASE WHEN d.label IS NOT NULL THEN d.label ELSE '' END
    MERGE (s)-[:DOCUMENTED_BY]->(doc)`

	qBioMedicalConcepts := `
    UNWIND $study.versions AS v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.biomedicalConcepts AS o
    MERGE (biom:BioMedicalConcept {id: o.id})
		SET
			biom.name = o.name,
			biom.label = o.label,
			biom.description = o.description,
			biom.instanceType = o.instanceType
	  MERGE (sv)-[:HAS_BIO_MEDICAL_CONCEPT]->(biom)
		MERGE (biomCode:BioMedicalConceptCode {id: o.code.id})
		SET
			biomCode.code = o.code.code,
			biomCode.codeSystem = o.code.codeSystem,
			biomCode.codeSystemVersion = o.code.codeSystemVersion,
			biomCode.decode = o.code.decode,
			biomCode.instanceType = o.code.instanceType
		MERGE (biom)-[:HAS_BIO_MEDICAL_CONCEPT_CODE]->(biomCode)`

	qBCSurrogate := `
    UNWIND $study.versions AS v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.bcSurrogates AS bs
    MERGE (bcs: BCSurrogates {id: bs.id})
		SET
			bcs.name = bs.name,
			bcs.label = bs.label,
			bcs.description = bs.description,
			bcs.reference = bs.reference,
			bcs.instanceType = bs.instanceType
	  MERGE (sv)-[:HAS_BC_SURROGATE]->(bcs)
	 `

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

	qStudyInterventions := `
    UNWIND $study.versions as v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.studyInterventions as si
    MERGE (intervention:StudyIntervention {id: si.id})
    SET
        intervention.name = si.name,
        intervention.label = si.label,
        intervention.description = si.description,
        intervention.instanceType = si.instanceType
    MERGE (sv)-[:HAS_INTERVENTION]->(intervention)

    WITH intervention, si
    WHERE si.type IS NOT NULL
    MERGE (it:Code {id: si.type.id})
    SET
        it.code = si.type.code,
        it.codeSystem = si.type.codeSystem,
        it.codeSystemVersion = si.type.codeSystemVersion,
        it.decode = si.type.decode,
        it.instanceType = si.type.instanceType
    MERGE (intervention)-[:HAS_TYPE]->(it)

    WITH intervention, si
    WHERE si.role IS NOT NULL
    MERGE (ir:Code {id: si.role.id})
    SET
        ir.code = si.role.code,
        ir.codeSystem = si.role.codeSystem,
        ir.codeSystemVersion = si.role.codeSystemVersion,
        ir.decode = si.role.decode,
        ir.instanceType = si.role.instanceType
    MERGE (intervention)-[:HAS_ROLE]->(ir)

    WITH intervention, si
    UNWIND si.administrations as admin
    MERGE (adm:Administration {id: admin.id})
    SET
        adm.name = admin.name,
        adm.label = admin.label,
        adm.description = admin.description,
        adm.instanceType = admin.instanceType
    MERGE (intervention)-[:HAS_ADMINISTRATION]->(adm)

    WITH adm, admin
    WHERE admin.dose IS NOT NULL
    MERGE (dose:Quantity {id: admin.dose.id})
    SET
        dose.value = admin.dose.value,
        dose.unit = admin.dose.unit,
        dose.instanceType = admin.dose.instanceType
    MERGE (adm)-[:HAS_DOSE]->(dose)

    WITH adm, admin
    WHERE admin.route IS NOT NULL
    MERGE (route:Code {id: admin.route.id})
    SET
        route.code = admin.route.code,
        route.codeSystem = admin.route.codeSystem,
        route.decode = admin.route.decode,
        route.instanceType = admin.route.instanceType
    MERGE (adm)-[:HAS_ROUTE]->(route)`

	qConditions := `
    UNWIND $study.versions as v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.conditions as c
    MERGE (cond:Condition {id: c.id})
    SET
        cond.name = c.name,
        cond.label = c.label,
        cond.description = c.description,
        cond.text = c.text,
        cond.instanceType = c.instanceType
    MERGE (sv)-[:HAS_CONDITION]->(cond)`

	qTitles := `
    UNWIND $study.versions as v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.titles as t
    MERGE (title:StudyTitle {id: t.id})
    SET
        title.text = t.text,
        title.instanceType = t.instanceType
    MERGE (sv)-[:HAS_TITLE]->(title)

    WITH title, t
    WHERE t.type IS NOT NULL
    MERGE (tt:Code {id: t.type.id})
    SET
        tt.code = t.type.code,
        tt.decode = t.type.decode,
        tt.codeSystem = t.type.codeSystem,
        tt.instanceType = t.type.instanceType
    MERGE (title)-[:HAS_TYPE]->(tt)`

	qStudyIdentifiers := `
    UNWIND $study.versions as v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.studyIdentifiers as si
    MERGE (ident:StudyIdentifier {id: si.id})
    SET
        ident.text = si.text,
        ident.scopeId = si.scopeId,
        ident.instanceType = si.instanceType
    MERGE (sv)-[:HAS_IDENTIFIER]->(ident)`

	qEligibility := `
    UNWIND $study.versions as v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.eligibilityCriterionItems as ec
    MERGE (crit:EligibilityCriterionItem {id: ec.id})
    SET
        crit.name = ec.name,
        crit.text = ec.text,
        crit.instanceType = ec.instanceType
    MERGE (sv)-[:HAS_ELIGIBILITY_CRITERION]->(crit)`

	qNarratives := `
    UNWIND $study.versions as v
    MATCH (sv:StudyVersion {id: v.id})
    UNWIND v.narrativeContentItems as nc
    MERGE (narr:NarrativeContentItem {id: nc.id})
    SET
        narr.name = nc.name,
        narr.text = nc.text,
        narr.instanceType = nc.instanceType
    MERGE (sv)-[:HAS_NARRATIVE_CONTENT]->(narr)`

	driver := cypher.GetDriver()
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		for _, q := range []string{
			qStudyAndVersions,
			qBioMedicalConcepts,
			qBCSurrogate,
			qDocuments,
			qDesigns,
			qEligibility,
			qTitles,
			qStudyIdentifiers,
			qStudyInterventions,
			qArms,
			qEpochs,
			qConditions,
			qAmendments,
			qOrganizationsAndAddresses,
			qEncounters,
			qActivities,
			qNarratives,
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
