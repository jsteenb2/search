package testing

import (
	"context"
	"testing"
	"time"

	"github.com/jsteenb2/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type InitFn func(*testing.T) (engine search.Engine, name string, cleanup func())

func TestSearchQueries(t *testing.T, engineInitFn InitFn) {
	queryTests := []struct {
		name   string
		testFn func(t *testing.T, engineInitFn InitFn)
	}{
		{
			name:   "match",
			testFn: TestMatchQuery,
		},
	}

	for _, tt := range queryTests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFn(t, engineInitFn)
		})
	}
}

func TestMatchQuery(t *testing.T, engineInitFn InitFn) {
	t.Helper()

	// TODO: add table tests for all the extra fields in
	//  the MatchQuery

	engine, indexName, cleanup := engineInitFn(t)
	defer cleanup()

	docs := []struct {
		id string
		v  interface{}
	}{
		{
			id: "foo1",
			v:  map[string]string{"foo1": "bar bug"},
		},
		{
			id: "foo2",
			v:  map[string]interface{}{"foo2": "bar"},
		},
		{
			id: "bar",
			v:  map[string]string{"bar": "baz"},
		},
		{
			id: "baz",
			v:  map[string]string{"baz": "foobar"},
		},
		{
			id: "fit",
			v:  map[string]string{"fit": "foo bar bit fit"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for _, doc := range docs {
		err := engine.
			Index(indexName).
			Index(ctx, doc.id, doc.v)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		query    search.Query
		expected []string
	}{
		{
			name:     "basic bar",
			query:    search.NewQueryMatch("bar"),
			expected: []string{"foo2", "foo1", "fit"},
		},
		{
			name:     "basic foobar",
			query:    search.NewQueryMatch("foobar"),
			expected: []string{"baz"},
		},
		{
			name:     "multiple",
			query:    search.NewQueryMatch("foobar bar foo"),
			expected: []string{"fit", "baz", "foo2", "foo1"},
		},
	}

	for _, tt := range tests {
		fn := func(t *testing.T) {
			result, err := engine.
				Index("base").
				Search(ctx, tt.query)
			require.NoError(t, err)

			require.Len(t, result.Hits, len(tt.expected))
			for i, expected := range tt.expected {
				assert.Equal(t, expected, result.Hits[i].ID)
			}
		}
		t.Run(tt.name, fn)
	}
}
