package bleve

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/jsteenb2/search"
)

func convertQuery(q search.Query) query.Query {
	qp := q.QueryPlan()
	switch qp.Type {
	case search.QueryTypeBoolField:
		return newBoolFieldQuery(qp)
	case search.QueryTypeBoolean:
		return newBoolQuery(qp)
	case search.QueryTypeDateRange:
		return newDataRangeQuery(qp)
	case search.QueryTypeIDs:
		q := query.NewDocIDQuery(qp.Matches)
		if qp.BoostVal != nil {
			q.SetBoost(float64(*qp.BoostVal))
		}
		return q
	case search.QueryTypeMatch:
		return newMatchQuery(qp)
	case search.QueryTypeMatchAll:
		q := query.NewMatchAllQuery()
		if qp.BoostVal != nil {
			q.SetBoost(float64(*qp.BoostVal))
		}
		return q
	case search.QueryTypeMatchNone:
		q := query.NewMatchNoneQuery()
		if qp.BoostVal != nil {
			q.SetBoost(float64(*qp.BoostVal))
		}
		return q
	case search.QueryTypeMatchPhrase:
		return newMatchPhraseQuery(qp)
	case search.QueryTypeTerm:
		return newTermQuery(qp)
	default:
		panic("unexpected query type: " + qp.Type.String())
	}
}

func newBoolFieldQuery(qp search.QueryPlan) *query.BoolFieldQuery {
	q := query.NewBoolFieldQuery(qp.Bool)
	if qp.FieldVal != "" {
		q.SetField(qp.FieldVal)
	}
	if qp.BoostVal != nil {
		q.SetBoost(float64(*qp.BoostVal))
	}
	return q
}

func newBoolQuery(qp search.QueryPlan) *query.BooleanQuery {
	q := bleve.NewBooleanQuery()
	if qp.BoostVal != nil {
		q.SetBoost(float64(*qp.BoostVal))
	}
	for _, must := range qp.Must {
		q.AddMust(convertQuery(must))
	}
	for _, should := range qp.Should {
		q.AddShould(convertQuery(should))
	}
	for _, mustNot := range qp.MustNot {
		q.AddMustNot(convertQuery(mustNot))
	}
	return q
}

func newDataRangeQuery(qp search.QueryPlan) *query.DateRangeQuery {
	start, end := search.BoundDate(qp.Min), search.BoundDate(qp.Max)
	q := query.NewDateRangeQuery(start, end)
	q.InclusiveEnd = &qp.InclusiveMax
	q.InclusiveStart = &qp.InclusiveMin
	if qp.FieldVal != "" {
		q.SetField(qp.FieldVal)
	}
	if qp.BoostVal != nil {
		q.SetBoost(float64(*qp.BoostVal))
	}
	return q
}

func newMatchQuery(qp search.QueryPlan) *query.MatchQuery {
	q := bleve.NewMatchQuery(qp.Matches[0])
	q.Operator = query.MatchQueryOperator(qp.Operator)
	if qp.FieldVal != "" {
		q.SetField(qp.FieldVal)
	}
	if qp.Analyzer != "" {
		q.Analyzer = qp.Analyzer
	}
	if qp.BoostVal != nil {
		q.SetBoost(float64(*qp.BoostVal))
	}
	if qp.Prefix > 0 {
		q.SetPrefix(qp.Prefix)
	}
	if qp.Fuzziness > 0 {
		q.SetFuzziness(qp.Fuzziness)
	}
	return q
}

func newMatchPhraseQuery(qp search.QueryPlan) *query.MatchPhraseQuery {
	q := bleve.NewMatchPhraseQuery(qp.Matches[0])
	if qp.FieldVal != "" {
		q.SetField(qp.FieldVal)
	}
	if qp.Analyzer != "" {
		q.Analyzer = qp.Analyzer
	}
	if qp.BoostVal != nil {
		q.SetBoost(float64(*qp.BoostVal))
	}
	return q
}

func newTermQuery(qp search.QueryPlan) *query.TermQuery {
	q := &query.TermQuery{
		Term:     qp.Matches[0],
		FieldVal: qp.FieldVal,
	}
	if qp.BoostVal != nil {
		q.SetBoost(float64(*qp.BoostVal))
	}
	return q
}
