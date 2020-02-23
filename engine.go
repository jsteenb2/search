package search

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type (
	Engine interface {
		Index(name string) Index
		Indices() []Index
	}

	Index interface {
		Name() string
		Index(ctx context.Context, id string, data interface{}) error
		Search(ctx context.Context, q Query) (*Result, error)
	}
)

type Result struct {
	Status   *Status
	Hits     []Hit
	Total    uint64
	MaxScore float64
	Took     time.Duration
}

func (r *Result) String() string {
	if len(r.Hits) == 0 {
		return fmt.Sprintf("0 matches, took %s", r.Took)
	}

	hits := make([]string, 0, len(r.Hits))
	for i, h := range r.Hits {
		hits = append(hits, fmt.Sprintf("%d. %s (%f)", i+1, h.ID, h.Score))
	}
	return fmt.Sprintf("%d matches, took %s\n\t%s", len(hits), r.Took, strings.Join(hits, "\n\t"))
}

type Hit struct {
	Index string
	ID    string
	Score float64
	Sort  []string

	Explanation *Explanation

	// Fields contains the values for document fields listed in
	// SearchRequest.Fields. Text fields are returned as strings, numeric
	// fields as float64s and date fields as time.RFC3339 formatted strings.
	Fields map[string]interface{}
}

type Explanation struct {
	Value    float64
	Message  string
	Children []*Explanation
}

type Status struct {
	Total      int
	Failed     int
	Successful int
}
