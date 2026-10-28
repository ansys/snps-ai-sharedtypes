package sharedtypes

import (
	"encoding/json"
	"testing"

	"github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQdrantQueryJson(t *testing.T) {
	testCases := map[string]QdrantQuery{
		"Dense": QdrantQuery{Query: qdrant.NewQuery(1, 2, 3)},
	}

	for name, query := range testCases {
		t.Run(name, func(t *testing.T) {
			b, err := json.Marshal(query)
			require.NoError(t, err)

			// require.Fail(t, string(b))

			var new QdrantQuery
			require.NoError(t, json.Unmarshal(b, &new))

			assert.Equal(t, query.Variant, new.Variant)
		})
	}
}
