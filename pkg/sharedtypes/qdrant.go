package sharedtypes

import (
	"encoding/json"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

type QdrantQuery struct {
	*qdrant.Query
}

const (
	qdrantQueryContext = iota
	qdrantQueryDiscover
	qdrantQueryFormula
	qdrantQueryFusion
	qdrantQueryNearest
	qdrantQueryNearestWithMmr
	qdrantQueryOrderBy
	qdrantQueryRecommend
	qdrantQueryRelevanceFeedback
	qdrantQueryRrf
	qdrantQuerySample
	qdrantQueryText
)

func (qq *QdrantQuery) UnmarshalJSON(data []byte) error {
	var temp struct{ Variant map[string]json.RawMessage }
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp.Variant) != 1 {
		return fmt.Errorf("expected 1-key variant: %#v", temp.Variant)
	}
	var variant string
	var raw json.RawMessage
	for v, r := range temp.Variant {
		variant = v
		raw = r
		break
	}

	qq.Query = &qdrant.Query{}
	switch variant {
	case "Nearest":
		var q qdrant.Query_Nearest
		if err := json.Unmarshal(raw, &q); err != nil {
			return err
		}
		qq.Variant = &q
	default:
		return fmt.Errorf("unknown variant %q", variant)
	}

	return nil
}

type QdrantScoredPoint struct {
	*qdrant.ScoredPoint
}
