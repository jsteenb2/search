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

var simpleDocs = []struct {
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
	{
		id: "nested bit",
		v: map[string]interface{}{
			"nest": map[string]string{
				"second": "bit",
				"third":  "lift it up",
			},
		},
	},
}

func TestSearchQueries(t *testing.T, engineInitFn InitFn) {
	queryTests := []struct {
		name   string
		testFn func(t *testing.T, engineInitFn InitFn)
	}{
		{
			name:   "bool field",
			testFn: TestBoolFieldQuery,
		},
		{
			name:   "match",
			testFn: TestMatchQuery,
		},
		{
			name:   "match all",
			testFn: TestMatchAllQuery,
		},
		{
			name:   "match none",
			testFn: TestMatchNoneQuery,
		},
		{
			name:   "match phrase",
			testFn: TestMatchPhraseQuery,
		},
	}

	for _, tt := range queryTests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFn(t, engineInitFn)
		})
	}
}

func TestBoolFieldQuery(t *testing.T, engineInitFn InitFn) {
	t.Helper()

	engine, indexName, cleanup := engineInitFn(t)
	defer cleanup()

	docs := []struct {
		id string
		v  interface{}
	}{
		{
			id: "1t",
			v:  map[string]interface{}{"bar": true},
		},
		{
			id: "2f",
			v:  map[string]interface{}{"baz": false},
		},
		{
			id: "1f",
			v:  map[string]interface{}{"bar": false},
		},
		{
			id: "nestedF",
			v: map[string]interface{}{
				"nest": map[string]interface{}{
					"first": false,
				}},
		},
		{
			id: "nestedT",
			v: map[string]interface{}{
				"nest": map[string]interface{}{
					"first": true,
				}},
		},
	}

	seedIndex(t, engine, indexName, docs...)

	tests := []struct {
		name     string
		query    search.Query
		expected []string
	}{
		{
			name:     "basic true",
			query:    search.NewQueryBoolField(true),
			expected: []string{"1t", "nestedT"},
		},
		{
			name:     "basic false",
			query:    search.NewQueryBoolField(false),
			expected: []string{"1f", "2f", "nestedF"},
		},
		{
			name: "nested true",
			query: search.
				NewQueryBoolField(true).
				SetField("nest.first"),
			expected: []string{"nestedT"},
		},
		{
			name: "nested false",
			query: search.
				NewQueryBoolField(false).
				SetField("nest.first"),
			expected: []string{"nestedF"},
		},
	}

	for _, tt := range tests {
		fn := func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			result, err := engine.
				Index(indexName).
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

func TestMatchQuery(t *testing.T, engineInitFn InitFn) {
	t.Helper()

	engine, indexName, cleanup := engineInitFn(t)
	defer cleanup()

	seedIndex(t, engine, indexName, simpleDocs...)

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
		{
			name: "nested field",
			query: search.
				NewQueryMatch("bit").
				SetField("nest.second"),
			expected: []string{"nested bit"},
		},
		{
			name: "fuzzy 1 off",
			query: search.
				NewQueryMatch("fobar").
				SetFuzziness(1),
			expected: []string{"baz"},
		},
		{
			name: "fuzzy 3 off",
			query: search.
				NewQueryMatch("fobarhm").
				SetFuzziness(3),
			expected: []string{"baz"},
		},
		{
			name: "fuzzy with prefix",
			query: search.
				NewQueryMatch("fooba").
				SetPrefix(4).
				SetFuzziness(1),
			expected: []string{"baz"},
		},
	}

	for _, tt := range tests {
		fn := func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			result, err := engine.
				Index(indexName).
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

func TestMatchAllQuery(t *testing.T, engineInitFn InitFn) {
	t.Helper()

	engine, indexName, cleanup := engineInitFn(t)
	defer cleanup()

	seedIndex(t, engine, indexName, simpleDocs...)

	var expectedIDs []string
	for _, d := range simpleDocs {
		expectedIDs = append(expectedIDs, d.id)
	}

	tests := []struct {
		name     string
		query    search.Query
		expected []string
	}{
		{
			name:     "basic match all",
			query:    search.NewQueryMatchAll(),
			expected: expectedIDs,
		},
	}

	for _, tt := range tests {
		fn := func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			result, err := engine.
				Index(indexName).
				Search(ctx, tt.query)
			require.NoError(t, err)

			require.Len(t, result.Hits, len(tt.expected))
			for _, h := range result.Hits {
				assert.Contains(t, tt.expected, h.ID)
			}
		}
		t.Run(tt.name, fn)
	}
}

func TestMatchNoneQuery(t *testing.T, engineInitFn InitFn) {
	t.Helper()

	engine, indexName, cleanup := engineInitFn(t)
	defer cleanup()

	seedIndex(t, engine, indexName, simpleDocs...)

	tests := []struct {
		name     string
		query    search.Query
		expected []string
	}{
		{
			name:     "basic match none",
			query:    search.NewQueryMatchNone(),
			expected: []string{"baz"},
		},
	}

	for _, tt := range tests {
		fn := func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			result, err := engine.
				Index(indexName).
				Search(ctx, tt.query)
			require.NoError(t, err)

			require.Empty(t, result.Hits)
		}
		t.Run(tt.name, fn)
	}
}

func TestMatchPhraseQuery(t *testing.T, engineInitFn InitFn) {
	t.Helper()

	engine, indexName, cleanup := engineInitFn(t)
	defer cleanup()

	seedIndex(t, engine, indexName, simpleDocs...)

	tests := []struct {
		name     string
		query    search.Query
		expected []string
	}{
		{
			name:     "basic phrase",
			query:    search.NewQueryMatchPhrase("bar bug"),
			expected: []string{"foo1"},
		},
		{
			name:     "basic phrase",
			query:    search.NewQueryMatchPhrase("bar bug"),
			expected: []string{"foo1"},
		},
		{
			name: "basic phrase",
			query: search.
				NewQueryMatchPhrase("lift it ").
				SetField("nest.third"), // extra space on purpose
			expected: []string{"nested bit"},
		},
	}

	for _, tt := range tests {
		fn := func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			result, err := engine.
				Index(indexName).
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

func seedIndex(t *testing.T, engine search.Engine, indexName string, docs ...struct {
	id string
	v  interface{}
}) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for _, doc := range docs {
		err := engine.
			Index(indexName).
			Index(ctx, doc.id, doc.v)
		require.NoError(t, err)
	}
}
