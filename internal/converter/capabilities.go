package converter

import (
	"sort"

	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

type Capability struct {
	Input  formats.Format `json:"input"`
	Output formats.Format `json:"output"`
	Engine string         `json:"engine"`
}

func Capabilities(registry *Registry) []Capability {
	var result []Capability

	allFormats := formats.All()

	for _, converter := range registry.All() {
		for _, input := range allFormats {
			for _, output := range allFormats {
				if converter.Supports(input, output) {
					result = append(result, Capability{
						Input:  input,
						Output: output,
						Engine: converter.Name(),
					})
				}
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Input.ID != result[j].Input.ID {
			return result[i].Input.ID < result[j].Input.ID
		}

		if result[i].Output.ID != result[j].Output.ID {
			return result[i].Output.ID < result[j].Output.ID
		}

		return result[i].Engine < result[j].Engine
	})

	return result
}
