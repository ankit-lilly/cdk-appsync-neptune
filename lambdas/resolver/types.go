package main

type AppSyncEvent struct {
	Info      AppSyncInfo            `json:"info"`
	Arguments map[string]interface{} `json:"arguments"`
	Source    map[string]interface{} `json:"source"`
}

type AppSyncInfo struct {
	FieldName        string   `json:"fieldName"`
	ParentTypeName   string   `json:"parentTypeName"`
	SelectionSetList []string `json:"selectionSetList"`
}

type Study struct {
	ID          string                     `json:"id"`
	Name        *string                    `json:"name,omitempty"`
	Description *string                    `json:"description,omitempty"`
	Label       *string                    `json:"label,omitempty"`
	Versions    []*StudyVersion            `json:"versions,omitempty"`
	Organizations []*Organization          `json:"organizations,omitempty"`
	DocumentedBy []*StudyDefinitionDocument `json:"documentedBy,omitempty"`
}

// StudyVersion corresponds to the GraphQL type:
// A specific version of a clinical study protocol.
type StudyVersion struct {
	ID                string               `json:"id"`
	VersionIdentifier string               `json:"versionIdentifier"`
	Rationale         *string              `json:"rationale,omitempty"`
	Study             *Study               `json:"study,omitempty"`
	StudyDesigns      []*StudyDesign       `json:"studyDesigns,omitempty"`
	Amendments        []*StudyAmendment    `json:"amendments,omitempty"`
	Interventions     []*StudyIntervention `json:"interventions,omitempty"`
}

// StudyDesign corresponds to the GraphQL type:
// Defines the overall design of the study, including its arms, epochs, and elements.
type StudyDesign struct {
	ID          string       `json:"id"`
	Name        *string      `json:"name,omitempty"`
	Description *string      `json:"description,omitempty"`
	StudyType   *string      `json:"studyType,omitempty"`
	StudyPhase  *string      `json:"studyPhase,omitempty"`
	Arms        []*Arm       `json:"arms,omitempty"`
	Epochs      []*Epoch     `json:"epochs,omitempty"`
	Elements    []*Element   `json:"elements,omitempty"`
	StudyCells  []*StudyCell `json:"studyCells,omitempty"`
}

// Arm corresponds to the GraphQL type:
// A group of subjects in a clinical trial who receive a specific intervention.
type Arm struct {
	ID          string       `json:"id"`
	Name        *string      `json:"name,omitempty"`
	Description *string      `json:"description,omitempty"`
	Type        *string      `json:"type,omitempty"`
	StudyDesign *StudyDesign `json:"studyDesign,omitempty"`
}

// Epoch corresponds to the GraphQL type:
// A period of time in a clinical trial during which subjects are in a consistent state.
type Epoch struct {
	ID          string       `json:"id"`
	Name        *string      `json:"name,omitempty"`
	Description *string      `json:"description,omitempty"`
	Type        *string      `json:"type,omitempty"`
	StudyDesign *StudyDesign `json:"studyDesign,omitempty"`
	Precedes    *Epoch       `json:"precedes,omitempty"`
	PrecededBy  *Epoch       `json:"precededBy,omitempty"`
}

// Element corresponds to the GraphQL type:
// A component of the study design, often representing a specific treatment or assessment.
type Element struct {
	ID          string  `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// StudyCell corresponds to the GraphQL type:
// Represents the combination of an Arm and an Epoch.
type StudyCell struct {
	ID       string     `json:"id"`
	Arm      *Arm       `json:"arm,omitempty"`
	Epoch    *Epoch     `json:"epoch,omitempty"`
	Elements []*Element `json:"elements,omitempty"`
}

// StudyAmendment corresponds to the GraphQL type:
// An amendment to the study protocol.
type StudyAmendment struct {
	ID        string        `json:"id"`
	Name      *string       `json:"name,omitempty"`
	Summary   *string       `json:"summary,omitempty"`
	Rationale *string       `json:"rationale,omitempty"`
	Version   *StudyVersion `json:"version,omitempty"`
}

// StudyIntervention corresponds to the GraphQL type:
// An intervention being investigated in the study (e.g., a drug, device).
type StudyIntervention struct {
	ID          string  `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Role        *string `json:"role,omitempty"`
	Type        *string `json:"type,omitempty"`
}

// Organization corresponds to the GraphQL type:
// An organization involved in the study (e.g., sponsor, CRO).
type Organization struct {
	ID           string  `json:"id"`
	Name         *string `json:"name,omitempty"`
	Type         *string `json:"type,omitempty"`
	LegalAddress *string `json:"legalAddress,omitempty"`
}

// StudyDefinitionDocument corresponds to the GraphQL type:
// A document that provides the definition of the study.
type StudyDefinitionDocument struct {
	ID   string  `json:"id"`
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
}

