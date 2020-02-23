package bleve

import (
	"context"
	"errors"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	ogsearch "github.com/blevesearch/bleve/search"
	"github.com/jsteenb2/search"
)

type IndexCfg struct {
	Name    string
	Path    string
	Mapping mapping.IndexMapping
}

func (i *IndexCfg) Setup(ctx context.Context) (bleve.Index, error) {
	indexMapping := i.Mapping
	if indexMapping == nil {
		indexMapping = bleve.NewIndexMapping()
	}
	return bleve.New(i.Path, indexMapping)
}

type Index struct {
	name  string
	index bleve.Index
	err   error
}

var _ search.Index = (*Index)(nil)

func (i *Index) Name() string {
	return i.name
}

func (i *Index) Index(ctx context.Context, id string, data interface{}) error {
	if i.err != nil {
		return i.err
	}

	return i.index.Index(id, data)
}

func (i *Index) Search(ctx context.Context, q search.Query) (*search.Result, error) {
	if i.err != nil {
		return nil, i.err
	}

	req := bleve.NewSearchRequest(convertQuery(q))
	if err := req.Validate(); err != nil {
		return nil, err
	}

	res, err := i.index.Search(req)
	if err != nil {
		return nil, err
	}
	if res.Total == 0 {
		return nil, errors.New("no results for provided query")
	}

	//fmt.Println(res)
	return convertSearchResult(res), nil
}

func convertSearchResult(r *bleve.SearchResult) *search.Result {
	s := &search.Result{
		MaxScore: r.MaxScore,
		Took:     r.Took,
		Total:    r.Total,
	}
	if r.Status != nil {
		s.Status = &search.Status{
			Total:      r.Status.Total,
			Failed:     r.Status.Failed,
			Successful: r.Status.Successful,
		}
	}

	s.Hits = make([]search.Hit, 0, len(r.Hits))
	for _, h := range r.Hits {
		hit := search.Hit{
			Index:       h.Index,
			ID:          h.ID,
			Score:       h.Score,
			Explanation: convertExplanation(h.Expl),
			Sort:        h.Sort,
			Fields:      h.Fields,
		}
		s.Hits = append(s.Hits, hit)
	}
	return s
}

func convertExplanation(ex *ogsearch.Explanation) *search.Explanation {
	if ex == nil {
		return nil
	}

	newEx := &search.Explanation{
		Value:   ex.Value,
		Message: ex.Message,
	}
	if len(ex.Children) == 0 {
		return newEx
	}

	newEx.Children = make([]*search.Explanation, len(newEx.Children))
	for _, chExpl := range ex.Children {
		newEx.Children = append(newEx.Children, convertExplanation(chExpl))
	}
	return newEx
}
