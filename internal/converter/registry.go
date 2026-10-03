package converter

import (
	"errors"
	"sync"

	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

var ErrNoConverter = errors.New("no converter available")

type Registry struct {
	mu         sync.RWMutex
	converters []Converter
}

func NewRegistry() *Registry {
	return &Registry{
		converters: make([]Converter, 0),
	}
}

func (r *Registry) Register(converter Converter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.converters = append(r.converters, converter)
}

func (r *Registry) Find(input, output formats.Format) (Converter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, converter := range r.converters {
		if converter.Supports(input, output) {
			return converter, nil
		}
	}

	return nil, ErrNoConverter
}

func (r *Registry) All() []Converter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Converter, len(r.converters))
	copy(result, r.converters)

	return result
}
