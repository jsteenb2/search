package bleve

import (
	"strconv"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/jsteenb2/search"
)

func convertQuery(q search.Query) query.Query {
	qp := q.QueryPlan()
	switch qp.Type {
	case search.QueryTypeBoolean:
		return newBoolQuery(qp)
	case search.QueryTypeMatchAll:
		q := query.NewMatchAllQuery()
		if qp.BoostVal != nil {
			q.SetBoost(float64(*qp.BoostVal))
		}
		return q
	case search.QueryTypeIDs:
		q := query.NewDocIDQuery(qp.Matches)
		if qp.BoostVal != nil {
			q.SetBoost(float64(*qp.BoostVal))
		}
		return q
	case search.QueryTypeMatch:
		return newMatchQuery(qp)
	case search.QueryTypeTerm:
		return newTermQuery(qp)
	default:
		panic("unexpected query type: " + strconv.Itoa(int(qp.Type)))
	}
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
