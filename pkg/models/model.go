package models

type Study struct {
	ID           string                     `json:"id"`
	Name         *string                    `json:"name,omitempty"`
	Description  *string                    `json:"description,omitempty"`
	Label        *string                    `json:"label,omitempty"`
	Versions     []*StudyVersion            `json:"versions,omitempty"`
	DocumentedBy []*StudyDefinitionDocument `json:"documentedBy,omitempty"`
}

type StudyVersion struct {
	ID                string               `json:"id"`
	VersionIdentifier string               `json:"versionIdentifier"`
	Rationale         *string              `json:"rationale,omitempty"`
	Study             *Study               `json:"study,omitempty"`
	StudyDesigns      []*StudyDesign       `json:"studyDesigns,omitempty"`
	Amendments        []*StudyAmendment    `json:"amendments,omitempty"`
	Interventions     []*StudyIntervention `json:"interventions,omitempty"`
	Organizations     []*Organization      `json:"organizations,omitempty"`
}

type StudyDesign struct {
	ID          string       `json:"id"`
	Name        *string      `json:"name,omitempty"`
	Description *string      `json:"description,omitempty"`
	StudyType   *StudyType   `json:"studyType,omitempty"`
	Arms        []*Arm       `json:"arms,omitempty"`
	Epochs      []*Epoch     `json:"epochs,omitempty"`
	Elements    []*Element   `json:"elements,omitempty"`
	StudyCells  []*StudyCell `json:"studyCells,omitempty"`
	Encounters  []*Encounter `json:"encounters,omitempty"`
	Activities  []*Activity  `json:"activities,omitempty"`
}

type Encounter struct {
	ID            string         `json:"id"`
	Name          *string        `json:"name,omitempty"`
	Label         *string        `json:"label,omitempty"`
	Description   *string        `json:"description,omitempty"`
	Type          *EncounterType `json:"type,omitempty"`
	PreviousID    *string        `json:"previousId,omitempty"`    // ADDED: This field was missing
	NextID        *string        `json:"nextId,omitempty"`        // ADDED: This field was missing
	ScheduledAtID *string        `json:"scheduledAtId,omitempty"` // ADDED: This field was missing
}

type EncounterType struct {
	ID                string `json:"id"`
	Code              string `json:"code"`
	CodeSystem        string `json:"codeSystem"`
	CodeSystemVersion string `json:"codeSystemVersion"`
	Decode            string `json:"decode"`
	InstanceType      string `json:"instanceType"`
}

type Activity struct {
	ID                string              `json:"id"`
	Name              *string             `json:"name,omitempty"`
	Label             *string             `json:"label,omitempty"`
	Description       *string             `json:"description,omitempty"`
	definedProcedures []*DefinedProcedure `json:"definedProcedures,omitempty"`
	instanceType      string              `json:"instanceType"`
}

type DefinedProcedure struct {
	ID                  string  `json:"id"`
	Name                *string `json:"name,omitempty"`
	Label               *string `json:"label,omitempty"`
	Description         *string `json:"description,omitempty"`
	ProcedureType       *string `json:"procedureType,omitempty"`
	Code                *Code   `json:"code,omitempty"`
	StudyInterventionID *string `json:"studyInterventionId,omitempty"`
	instanceType        string  `json:"instanceType"`
}

type Code struct {
	ID                string `json:"id"`
	Code              string `json:"code"`
	CodeSystem        string `json:"codeSystem"`
	CodeSystemVersion string `json:"codeSystemVersion"`
	Decode            string `json:"decode"`
	InstanceType      string `json:"instanceType"`
}

type StudyType struct {
	ID                string `json:"id"`
	Code              string `json:"code"`
	CodeSystem        string `json:"codeSystem"`
	CodeSystemVersion string `json:"codeSystemVersion"`
	Decode            string `json:"decode"`
	InstanceType      string `json:"instanceType"`
}

type Arm struct {
	ID             string             `json:"id"`
	Name           *string            `json:"name,omitempty"`
	Description    *string            `json:"description,omitempty"`
	DataOriginType *ArmDataOriginType `json:"dataOriginType,omitempty"` // FIXED: JSON tag was "type"
	StudyDesign    *StudyDesign       `json:"studyDesign,omitempty"`
}

type ArmDataOriginType struct {
	ID                  string   `json:"id"`
	ExtensionAttributes []string `json:"extensionAttributes,omitempty"`
	Code                string   `json:"code"`
	CodeSystem          string   `json:"codeSystem"`
	CodeSystemVersion   string   `json:"codeSystemVersion"`
	Decode              string   `json:"decode"`
	InstanceType        string   `json:"instanceType"`
}

type Epoch struct {
	ID          string  `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	PreviousID  *string `json:"previousId,omitempty"` // ADDED: This field was missing
	Precedes    *Epoch  `json:"precedes,omitempty"`
	PrecededBy  *Epoch  `json:"precededBy,omitempty"`
}

type Element struct {
	ID          string  `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type StudyCell struct {
	ID       string     `json:"id"`
	Arm      *Arm       `json:"arm,omitempty"`
	Epoch    *Epoch     `json:"epoch,omitempty"`
	Elements []*Element `json:"elements,omitempty"`
}

type StudyAmendment struct {
	ID        string        `json:"id"`
	Name      *string       `json:"name,omitempty"`
	Summary   *string       `json:"summary,omitempty"`
	Rationale *string       `json:"rationale,omitempty"`
	Version   *StudyVersion `json:"version,omitempty"`
}

type StudyIntervention struct {
	ID          string  `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Role        *string `json:"role,omitempty"`
	Type        *string `json:"type,omitempty"`
}

type OrgType struct {
	ID                string `json:"id"`
	Code              string `json:"code"`
	CodeSystem        string `json:"codeSystem"`
	CodeSystemVersion string `json:"codeSystemVersion"`
	Decode            string `json:"decode"`
	InstanceType      string `json:"instanceType"`
}

type Organization struct {
	ID           string        `json:"id"`
	Name         *string       `json:"name,omitempty"`
	Type         *OrgType      `json:"type,omitempty"`
	LegalAddress *LegalAddress `json:"legalAddress,omitempty"`
}

type Country struct {
	ID                string `json:"id"`
	Code              string `json:"code"`
	CodeSystem        string `json:"codeSystem"`
	CodeSystemVersion string `json:"codeSystemVersion"`
	Decode            string `json:"decode"`
	InstanceType      string `json:"instanceType"`
}

type LegalAddress struct {
	ID                  string   `json:"id"`
	ExtensionAttributes []string `json:"extensionAttributes,omitempty"`
	Text                string   `json:"text"`
	Lines               []string `json:"lines,omitempty"`
	City                string   `json:"city,omitempty"`
	District            string   `json:"district,omitempty"`
	State               string   `json:"state,omitempty"`
	PostalCode          string   `json:"postalCode,omitempty"`
	Country             *Country `json:"country,omitempty"`
	InstanceType        string   `json:"instanceType"`
}

type StudyDefinitionDocument struct {
	ID                  string                            `json:"id"`
	ExtensionAttributes []string                          `json:"extensionAttributes,omitempty"`
	Name                *string                           `json:"name,omitempty"`
	Label               *string                           `json:"label,omitempty"`
	Description         *string                           `json:"description,omitempty"`
	Language            map[string]interface{}            `json:"language,omitempty"`
	TemplateName        *string                           `json:"templateName,omitempty"`
	Versions            []*StudyDefinitionDocumentVersion `json:"versions,omitempty"`
	ChildIds            []string                          `json:"childIds,omitempty"`
	Notes               []string                          `json:"notes,omitempty"`
	InstanceType        string                            `json:"instanceType"`
}

type StudyDefinitionDocumentVersion struct {
	ID                  string                 `json:"id"`
	ExtensionAttributes []string               `json:"extensionAttributes,omitempty"`
	Version             string                 `json:"version"`
	Status              map[string]interface{} `json:"status,omitempty"`
	DateValues          []interface{}          `json:"dateValues,omitempty"`
	Contents            []interface{}          `json:"contents,omitempty"`
	Notes               []interface{}          `json:"notes,omitempty"`
	InstanceType        string                 `json:"instanceType"`
}
