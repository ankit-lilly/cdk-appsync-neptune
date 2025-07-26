package models

type CommentAnnotation struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Codes Code   `json:"codes"`
}

type StudyVersion struct {
	ID                       string              `json:"id"`
	VersionIdentifier        string              `json:"versionIdentifier"`
	BusinessTherapeuticAreas []Code                `json:"businessTherapeuticAreas"`
	Rationale                string              `json:"rationale"`
	Notes                    []CommentAnnotation `json:"notes"`
	Abbreviations            []Abbreviation      `json:"abbreviations"`
}

type Code struct {
	ID                string `json:"id"`
	Code              string `json:"code"`
	CodeSystem        string `json:"codeSystem"`
	CodeSystemVersion string `json:"codeSystemVersion"`
	Decode            string `json:"decode"`
}

type Study struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Label       string         `json:"label"`
	Versions    []StudyVersion `json:"versions"`
}

type Abbreviation struct {
	ID              string            `json:"id"`
	AbbreviatedText string            `json:"abbreviatedText"`
	ExpandedText    string            `json:"expandedText"`
	Notes           CommentAnnotation `json:"notes"`
}
