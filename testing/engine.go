package testing

import (
	"context"
	"strings"
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
			testFn: TestQueryBoolField,
		},
		{
			name:   "date range",
			testFn: TestQueryDateRange,
		},
		{
			name:   "match",
			testFn: TestQueryMatch,
		},
		{
			name:   "match all",
			testFn: TestQueryMatchAll,
		},
		{
			name:   "match none",
			testFn: TestQueryMatchNone,
		},
		{
			name:   "match phrase",
			testFn: TestQueryMatchPhrase,
		},
		{
			name:   "numeric range",
			testFn: TestQueryNumericRange,
		},
		{
			name:   "prefix",
			testFn: TestQueryPrefix,
		},
		{
			name:   "term",
			testFn: TestQueryTerm,
		},
		{
			name:   "term range",
			testFn: TestQueryTermRange,
		},
	}

	for _, tt := range queryTests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFn(t, engineInitFn)
		})
	}
}

func TestQueryBoolField(t *testing.T, engineInitFn InitFn) {
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

			hasHitIDs(t, result.Hits, tt.expected...)
		}
		t.Run(tt.name, fn)
	}
}

func TestQueryDateRange(t *testing.T, engineInitFn InitFn) {
	t.Helper()

	engine, indexName, cleanup := engineInitFn(t)
	defer cleanup()

	now, day := time.Now(), 24*time.Hour

	docs := []struct {
		id string
		v  interface{}
	}{
		{
			id: "bar 30 days ago",
			v:  map[string]interface{}{"bar": now.Add(-30 * day)},
		},
		{
			id: "bar 20 days ago",
			v:  map[string]interface{}{"bar": now.Add(-20 * day)},
		},
		{
			id: "bar 10 days ago",
			v:  map[string]interface{}{"bar": now.Add(-10 * day)},
		},
		{
			id: "bar today",
			v:  map[string]interface{}{"bar": now},
		},
		{
			id: "baz 10 days ago",
			v:  map[string]interface{}{"baz": now.Add(-10 * day)},
		},
		{
			id: "baz today",
			v:  map[string]interface{}{"baz": now},
		},
		{
			id: "nested 10 days ago",
			v: map[string]interface{}{
				"nested": map[string]interface{}{
					"date": now.Add(-10 * day),
				},
			},
		},
		{
			id: "nested today",
			v: map[string]interface{}{
				"nested": map[string]interface{}{
					"date": now,
				},
			},
		},
	}

	seedIndex(t, engine, indexName, docs...)

	tests := []struct {
		name     string
		query    search.Query
		expected []string
	}{
		{
			name:  "30 day range",
			query: search.NewQueryDataRange(now.Add(-31*day), now.Add(1*day)),
			expected: []string{
				"bar 20 days ago", "bar 30 days ago", "bar 10 days ago", "bar today",
				"baz 10 days ago", "baz today",
				"nested 10 days ago", "nested today",
			},
		},
		{
			name:  "defaults to inclusive start and exclusive end",
			query: search.NewQueryDataRange(now.Add(-30*day), now),
			expected: []string{
				"bar 20 days ago", "bar 30 days ago", "bar 10 days ago",
				"baz 10 days ago",
				"nested 10 days ago",
			},
		},
		{
			name: "30 day range exclusive start and end",
			query: search.
				NewQueryDataRange(now.Add(-30*day), now).
				SetInclusiveStart(false).
				SetInclusiveEnd(false),
			expected: []string{
				"bar 20 days ago", "bar 10 days ago",
				"baz 10 days ago",
				"nested 10 days ago",
			},
		},
		{
			name: "30 day range inclusive start and exclusive end",
			query: search.
				NewQueryDataRange(now.Add(-30*day), now).
				SetInclusiveEnd(false),
			expected: []string{
				"bar 20 days ago", "bar 30 days ago", "bar 10 days ago",
				"baz 10 days ago",
				"nested 10 days ago",
			},
		},
		{
			name: "30 day range inclusive start and exclusive end",
			query: search.
				NewQueryDataRange(now.Add(-30*day), now).
				SetInclusiveStart(true).
				SetInclusiveEnd(false),
			expected: []string{
				"bar 20 days ago", "bar 30 days ago", "bar 10 days ago",
				"baz 10 days ago",
				"nested 10 days ago",
			},
		},
		{
			name: "nested to inclusive start and exclusive end",
			query: search.
				NewQueryDataRange(now.Add(-30*day), now).
				SetField("nested.date"),
			expected: []string{"nested 10 days ago"},
		},
		{
			name: "nested to inclusive start and inclusive end",
			query: search.
				NewQueryDataRange(now.Add(-30*day), now).
				SetField("nested.date").
				SetInclusiveEnd(true),
			expected: []string{"nested 10 days ago", "nested today"},
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

			hasHitIDs(t, result.Hits, tt.expected...)
		}
		t.Run(tt.name, fn)
	}
}

func TestQueryMatch(t *testing.T, engineInitFn InitFn) {
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

			hasHitIDs(t, result.Hits, tt.expected...)
		}
		t.Run(tt.name, fn)
	}
}

func TestQueryMatchAll(t *testing.T, engineInitFn InitFn) {
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

func TestQueryMatchNone(t *testing.T, engineInitFn InitFn) {
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

func TestQueryMatchPhrase(t *testing.T, engineInitFn InitFn) {
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

			hasHitIDs(t, result.Hits, tt.expected...)
		}
		t.Run(tt.name, fn)
	}
}

func TestQueryNumericRange(t *testing.T, engineInitFn InitFn) {
	t.Helper()

	engine, indexName, cleanup := engineInitFn(t)
	defer cleanup()

	docs := []struct {
		id string
		v  interface{}
	}{
		{
			id: "0",
			v:  map[string]interface{}{"foo0": 0},
		},
		{
			id: "1",
			v:  map[string]interface{}{"foo1": 1},
		},
		{
			id: "5",
			v:  map[string]interface{}{"foo5": 5},
		},
		{
			id: "-5",
			v:  map[string]interface{}{"foo-5": -5},
		},
		{
			id: "-1",
			v:  map[string]interface{}{"foo-5": -1},
		},
	}

	seedIndex(t, engine, indexName, docs...)

	tests := []struct {
		name     string
		query    search.Query
		expected []string
	}{
		{
			name: "basic max exclusive",
			query: search.
				NewQueryNumericRange().
				SetMax(-1),
			expected: []string{"-5"},
		},
		{
			name: "basic max inclusive",
			query: search.
				NewQueryNumericRange().
				SetMax(0).
				SetInclusiveMax(true),
			expected: []string{"-1", "-5", "0"},
		},
		{
			name: "basic min exclusive",
			query: search.
				NewQueryNumericRange().
				SetMin(0).
				SetInclusiveMin(false),
			expected: []string{"1", "5"},
		},
		{
			name: "basic max inclusive",
			query: search.
				NewQueryNumericRange().
				SetMin(0),
			expected: []string{"0", "1", "5"},
		},
		{
			name: "min inclusive and max exclusive",
			query: search.
				NewQueryNumericRange().
				SetMin(0).
				SetMax(1),
			expected: []string{"0"},
		},
		{
			name: "min inclusive and max exclusive",
			query: search.
				NewQueryNumericRange().
				SetMin(0).
				SetMax(1).
				SetInclusiveMax(true),
			expected: []string{"0", "1"},
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

			hasHitIDs(t, result.Hits, tt.expected...)
		}
		t.Run(tt.name, fn)
	}
}

func TestQueryPrefix(t *testing.T, engineInitFn InitFn) {
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
			name:     "basic prefix",
			query:    search.NewQueryPrefix("bar"),
			expected: []string{"foo2", "foo1", "fit"},
		},
		{
			name:     "basic prefix spaced",
			query:    search.NewQueryPrefix("b"),
			expected: []string{"foo1", "fit", "bar", "nested bit", "foo2"},
		},
		{
			name: "prefix nested",
			query: search.
				NewQueryPrefix("b").
				SetField("nest.second"),
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

			hasHitIDs(t, result.Hits, tt.expected...)
		}
		t.Run(tt.name, fn)
	}
}

func TestQueryTerm(t *testing.T, engineInitFn InitFn) {
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
			name:     "no matches",
			query:    search.NewQueryTerm("b"),
			expected: []string{},
		},
		{
			name:     "basic term",
			query:    search.NewQueryTerm("bar"),
			expected: []string{"foo2", "foo1", "fit"},
		},
		{
			name:     "basic term with nested",
			query:    search.NewQueryTerm("bit"),
			expected: []string{"nested bit", "fit"},
		},
		{
			name: "nested term",
			query: search.
				NewQueryTerm("bit").
				SetField("nest.second"),
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

			hasHitIDs(t, result.Hits, tt.expected...)
		}
		t.Run(tt.name, fn)
	}
}

func TestQueryTermRange(t *testing.T, engineInitFn InitFn) {
	t.Helper()

	engine, indexName, cleanup := engineInitFn(t)
	defer cleanup()

	docs := []struct {
		id string
		v  interface{}
	}{
		{
			id: "0",
			v:  map[string]interface{}{"foo1": "0"},
		},
		{
			id: "1",
			v:  map[string]interface{}{"foo1": "1"},
		},
		{
			id: "2",
			v:  map[string]interface{}{"foo2": "2"},
		},
		{
			id: "nested1",
			v: map[string]interface{}{
				"nested": map[string]interface{}{
					"first": "1",
				},
			},
		},
	}
	seedIndex(t, engine, indexName, docs...)

	tests := []struct {
		name     string
		query    search.Query
		expected []string
	}{
		{
			name:     "no matches",
			query:    search.NewQueryTermRange("3", ""),
			expected: []string{},
		},
		{
			name:     "min single",
			query:    search.NewQueryTermRange("2", ""),
			expected: []string{"2"},
		},
		{
			name: "min single exclusive min",
			query: search.
				NewQueryTermRange("0", "5").
				SetInclusiveMin(false),
			expected: []string{"2", "1", "nested1"},
		},
		{
			name: "min single inclusive min inclusive max",
			query: search.
				NewQueryTermRange("0", "2").
				SetInclusiveMax(true),
			expected: []string{"0", "2", "1", "nested1"},
		},
		{
			name: "min single exclusive min exclusive max",
			query: search.
				NewQueryTermRange("0", "2").
				SetInclusiveMin(false),
			expected: []string{"1", "nested1"},
		},
		{
			name: "nested",
			query: search.
				NewQueryTermRange("0", "2").
				SetField("nested.first"),
			expected: []string{"nested1"},
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

			hasHitIDs(t, result.Hits, tt.expected...)
		}
		t.Run(tt.name, fn)
	}
}

func hasHitIDs(t *testing.T, hits []search.Hit, expected ...string) {
	t.Helper()

	hitIDs := make([]string, 0, len(expected))
	for _, h := range hits {
		hitIDs = append(hitIDs, h.ID)
	}
	if len(hits) != len(expected) {
		require.Lenf(t, hits, len(expected), "got ids:\t\t[%s]\nexpected ids:\t[%s]", strings.Join(hitIDs, ", "), strings.Join(expected, ", "))
	}
	for i, expected := range expected {
		assert.Equal(t, expected, hitIDs[i])
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
