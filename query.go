package search

import (
	"fmt"
	"time"
)

type QueryType int

func (q QueryType) String() string {
	if int(q) >= len(queryTypes) {
		return "unknown query type"
	}
	return queryTypes[q] + " query type"
}

const (
	QueryTypeUnknown QueryType = iota
	QueryTypeBoolean
	QueryTypeBoolField
	QueryTypeDateRange
	QueryTypeIDs
	QueryTypeMatch
	QueryTypeMatchAll
	QueryTypeMatchNone
	QueryTypeMatchPhrase
	QueryTypeMultiPhrase
	QueryTypeNumericRange
	QueryTypePrefix
	QueryTypeString
	QueryTypeTerm
	QueryTypeTermRange
	QueryTypeWildcard
)

var queryTypes = [...]string{
	QueryTypeUnknown:      "unknown",
	QueryTypeBoolean:      "boolean",
	QueryTypeBoolField:    "bool field",
	QueryTypeDateRange:    "date range",
	QueryTypeIDs:          "ids",
	QueryTypeMatch:        "match",
	QueryTypeMatchAll:     "match all",
	QueryTypeMatchNone:    "match none",
	QueryTypeMatchPhrase:  "match phrase",
	QueryTypeMultiPhrase:  "multi phrase",
	QueryTypeNumericRange: "numeric range",
	QueryTypePrefix:       "prefix",
	QueryTypeString:       "string",
	QueryTypeTerm:         "term",
	QueryTypeTermRange:    "term range",
	QueryTypeWildcard:     "wildcard",
}

type (
	Query interface {
		QueryPlan() QueryPlan
	}

	QueryPlan struct {
		Type QueryType

		Should  []Query
		Must    []Query
		MustNot []Query

		Analyzer string
		BoostVal *Boost
		FieldVal string

		Bool      bool
		Matches   []string
		Fuzziness int
		Operator  QueryOperator
		Prefix    int
		Terms     [][]string

		Min, Max     Bound
		InclusiveMin bool
		InclusiveMax bool
	}

	QueryDateRange struct {
		Start          time.Time
		End            time.Time
		InclusiveStart *bool
		InclusiveEnd   *bool
		FieldVal       string
		BoostVal       *Boost
	}

	QueryMultiPhrase struct {
		Terms    [][]string
		Field    string
		BoostVal *Boost
	}

	QueryNumericRange struct {
		Min          *float64
		Max          *float64
		InclusiveMin *bool
		InclusiveMax *bool
		FieldVal     string
		BoostVal     *Boost
	}

	QueryPrefix struct {
		Prefix   string
		FieldVal string
		BoostVal *Boost
	}

	QueryRegexp struct {
		Regexp   string
		FieldVal string
		BoostVal *Boost
	}

	QueryString struct {
		Query    string
		BoostVal *Boost
	}

	QueryTermRange struct {
		Min          string
		Max          string
		InclusiveMin *bool
		InclusiveMax *bool
		FieldVal     string
		BoostVal     *Boost
	}

	QueryWildcard struct {
		Wildcard string
		FieldVal string
		BoostVal *Boost
	}
)

type QueryBoolField struct {
	Bool     bool
	BoostVal *Boost
	FieldVal string
}

func NewQueryBoolField(b bool) *QueryBoolField {
	return &QueryBoolField{
		Bool: b,
	}
}

func (q *QueryBoolField) QueryPlan() QueryPlan {
	return QueryPlan{
		Type:     QueryTypeBoolField,
		Bool:     q.Bool,
		BoostVal: q.BoostVal,
		FieldVal: q.FieldVal,
	}
}

func (q *QueryBoolField) SetBoost(b float64) *QueryBoolField {
	boost := Boost(b)
	q.BoostVal = &boost
	return q
}

func (q *QueryBoolField) SetField(field string) *QueryBoolField {
	q.FieldVal = field
	return q
}

type QueryBoolean struct {
	Should   []Query
	Must     []Query
	MustNot  []Query
	BoostVal *Boost
}

func NewQueryBoolean() *QueryBoolean {
	return new(QueryBoolean)
}

func (q *QueryBoolean) QueryPlan() QueryPlan {
	return QueryPlan{
		Type:     QueryTypeBoolean,
		Should:   q.Should,
		Must:     q.Must,
		MustNot:  q.MustNot,
		BoostVal: q.BoostVal,
	}
}

func (q *QueryBoolean) SetBoost(b float64) {
	boost := Boost(b)
	q.BoostVal = &boost
}

func (q *QueryBoolean) AddMust(musts ...Query) *QueryBoolean {
	q.Must = append(q.Must, musts...)
	return q
}

func (q *QueryBoolean) AddMustNot(nots ...Query) *QueryBoolean {
	q.MustNot = append(q.MustNot, nots...)
	return q
}

func (q *QueryBoolean) AddShould(shoulds ...Query) *QueryBoolean {
	q.Should = append(q.Should, shoulds...)
	return q
}

type QueryIDs struct {
	IDs      []string
	BoostVal *Boost
}

func NewQueryIDs(ids []string) *QueryIDs {
	return &QueryIDs{
		IDs: ids,
	}
}

func (q *QueryIDs) QueryPlan() QueryPlan {
	return QueryPlan{
		Type:     QueryTypeIDs,
		Matches:  q.IDs,
		BoostVal: q.BoostVal,
	}
}

func (q *QueryIDs) SetBoost(b float64) {
	boost := Boost(b)
	q.BoostVal = &boost
}

type QueryMatch struct {
	Match     string
	Analyzer  string
	BoostVal  *Boost
	FieldVal  string
	Prefix    int
	Fuzziness int
	Operator  QueryOperator
}

func NewQueryMatch(match string) *QueryMatch {
	return &QueryMatch{
		Match: match,
	}
}

func (q *QueryMatch) QueryPlan() QueryPlan {
	return QueryPlan{
		Type:      QueryTypeMatch,
		Matches:   []string{q.Match},
		Analyzer:  q.Analyzer,
		BoostVal:  q.BoostVal,
		FieldVal:  q.FieldVal,
		Prefix:    q.Prefix,
		Fuzziness: q.Fuzziness,
		Operator:  q.Operator,
	}
}

func (q *QueryMatch) SetAnalyzer(analyzer string) *QueryMatch {
	q.Analyzer = analyzer
	return q
}

func (q *QueryMatch) SetBoost(b float64) *QueryMatch {
	boost := Boost(b)
	q.BoostVal = &boost
	return q
}

func (q *QueryMatch) SetField(field string) *QueryMatch {
	q.FieldVal = field
	return q
}

func (q *QueryMatch) SetFuzziness(fuzz int) *QueryMatch {
	q.Fuzziness = fuzz
	return q
}

func (q *QueryMatch) SetPrefix(prefix int) *QueryMatch {
	q.Prefix = prefix
	return q
}

type QueryMatchAll struct {
	BoostVal *Boost
}

func NewQueryMatchAll() *QueryMatchAll {
	return new(QueryMatchAll)
}

func (q *QueryMatchAll) QueryPlan() QueryPlan {
	return QueryPlan{
		Type: QueryTypeMatchAll,
	}
}

func (q *QueryMatchAll) SetBoost(b float64) *QueryMatchAll {
	boost := Boost(b)
	q.BoostVal = &boost
	return q
}

type QueryMatchNone struct {
	BoostVal *Boost
}

func NewQueryMatchNone() *QueryMatchNone {
	return new(QueryMatchNone)
}

func (q *QueryMatchNone) QueryPlan() QueryPlan {
	return QueryPlan{
		Type: QueryTypeMatchNone,
	}
}

func (q *QueryMatchNone) SetBoost(b float64) *QueryMatchNone {
	boost := Boost(b)
	q.BoostVal = &boost
	return q
}

type QueryMatchPhrase struct {
	MatchPhrase string
	FieldVal    string
	Analyzer    string
	BoostVal    *Boost
}

func NewQueryMatchPhrase(phrase string) *QueryMatchPhrase {
	return &QueryMatchPhrase{
		MatchPhrase: phrase,
	}
}

func (q *QueryMatchPhrase) QueryPlan() QueryPlan {
	return QueryPlan{
		Type:     QueryTypeMatchPhrase,
		Matches:  []string{q.MatchPhrase},
		Analyzer: q.Analyzer,
		BoostVal: q.BoostVal,
		FieldVal: q.FieldVal,
	}
}

func (q *QueryMatchPhrase) SetAnalyzer(analyzer string) *QueryMatchPhrase {
	q.Analyzer = analyzer
	return q
}

func (q *QueryMatchPhrase) SetBoost(b float64) *QueryMatchPhrase {
	boost := Boost(b)
	q.BoostVal = &boost
	return q
}

func (q *QueryMatchPhrase) SetField(field string) *QueryMatchPhrase {
	q.FieldVal = field
	return q
}

type QueryTerm struct {
	Term     string
	FieldVal string
	BoostVal *Boost
}

func NewQueryTerm(term string) *QueryTerm {
	return &QueryTerm{
		Term: term,
	}
}

func (q *QueryTerm) QueryPlan() QueryPlan {
	return QueryPlan{
		Type:     QueryTypeTerm,
		Matches:  []string{q.Term},
		BoostVal: q.BoostVal,
		FieldVal: q.FieldVal,
	}
}

func (q *QueryTerm) SetBoost(b float64) *QueryTerm {
	boost := Boost(b)
	q.BoostVal = &boost
	return q
}

func (q *QueryTerm) SetField(field string) *QueryTerm {
	q.FieldVal = field
	return q
}

type Boost float64

func (b *Boost) Value() float64 {
	if b == nil {
		return 1.0
	}
	return float64(*b)
}

func (b *Boost) GoString() string {
	if b == nil {
		return "boost unspecified"
	}
	return fmt.Sprintf("%f", *b)
}

type QueryOperator int

const (
	// Document must satisfy AT LEAST ONE of term searches.
	MatchQueryOperatorOr = 0
	// Document must satisfy ALL of term searches.
	MatchQueryOperatorAnd = 1
)

type Bound interface{}

func BoundString(b Bound) string {
	s, _ := b.(string)
	return s
}

func BoundFloat64(b Bound) float64 {
	f, _ := b.(float64)
	return f
}

func BoundDate(b Bound) time.Time {
	t, _ := b.(time.Time)
	return t
}
