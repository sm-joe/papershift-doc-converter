package formats

type Family string

const (
	FamilyDocument     Family = "document"
	FamilySpreadsheet  Family = "spreadsheet"
	FamilyPresentation Family = "presentation"
	FamilyImage        Family = "image"
)

type Format struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Extension string   `json:"extension"`
	MIMETypes []string `json:"mime_types"`
	Family    Family   `json:"family"`
}
