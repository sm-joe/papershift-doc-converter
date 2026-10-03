package converter

import "github.com/sm-joe/papershift-doc-converter/internal/formats"

type Capability struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Engine string `json:"engine"`
}

func Capabilities(registry *Registry) []Capability {
	var result []Capability

	for _, converter := range registry.All() {
		_ = converter
	}

	// Capabilities will be populated by concrete converters.
	return result
}

func formatExists(id string) bool {
	_, ok := formats.Get(id)
	return ok
}
