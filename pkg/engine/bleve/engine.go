package bleve

import (
	"context"
	"fmt"

	"github.com/blevesearch/bleve"
	"github.com/jsteenb2/search"
)

type Engine struct {
	indices map[string]bleve.Index
}

var _ search.Engine = (*Engine)(nil)

func NewEngine(index IndexCfg, rest ...IndexCfg) (*Engine, error) {
	setupIndices := make(map[string]bleve.Index)
	for _, i := range append(rest, index) {
		index, err := i.Setup(context.TODO())
		if err != nil {
			return nil, err
		}
		setupIndices[i.Name] = index
	}

	return &Engine{
		indices: setupIndices,
	}, nil
}

func (e *Engine) Index(name string) search.Index {
	index, ok := e.indices[name]
	if !ok {
		return &Index{
			err: fmt.Errorf("index does not exist for this engine: %q" + name),
		}
	}
	return &Index{
		name:  name,
		index: index,
	}
}

func (e *Engine) Indices() []search.Index {
	indices := make([]search.Index, 0, len(e.indices))
	for name, index := range e.indices {
		indices = append(indices, &Index{
			name:  name,
			index: index,
		})
	}
	return indices
}
