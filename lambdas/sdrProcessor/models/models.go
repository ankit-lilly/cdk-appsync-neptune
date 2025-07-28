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
	ID          string       `json:"id"`
	Name        *string      `json:"name,omitempty"`
	Description *string      `json:"description,omitempty"`
	Type        *string      `json:"type,omitempty"`
	StudyDesign *StudyDesign `json:"studyDesign,omitempty"`
}

type Epoch struct {
	ID          string       `json:"id"`
	Name        *string      `json:"name,omitempty"`
	Description *string      `json:"description,omitempty"`
	Type        *string      `json:"type,omitempty"`
	StudyDesign *StudyDesign `json:"studyDesign,omitempty"`
	Precedes    *Epoch       `json:"precedes,omitempty"`
	PrecededBy  *Epoch       `json:"precededBy,omitempty"`
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
	ID   string  `json:"id"`
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
}
