package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/qdrant/go-client/qdrant"
)

// VectorStore provides optional vector capabilities using official Qdrant Go client
type VectorStore struct {
	Client     *qdrant.Client
	Collection string
	Enabled    bool
}

// NewVectorStore initializes the official Qdrant client
func NewVectorStore(url, collection string) *VectorStore {
	if url == "" {
		return &VectorStore{Enabled: false}
	}
	client, err := qdrant.NewClient(&qdrant.Config{
		URL: url,
	})
	if err != nil {
		return &VectorStore{Enabled: false}
	}
	return &VectorStore{
		Client:     client,
		Collection: collection,
		Enabled:    true,
	}
}

// StoreVector upserts a vector with payload
func (vs *VectorStore) StoreVector(id string, vec []float32, payload map[string]interface{}) error {
	if !vs.Enabled {
		return nil
	}
	_, err := vs.Client.Upsert(context.Background(), vs.Collection, &qdrant.PointStructs{
		Points: []*qdrant.PointStruct{
			{
				ID:      qdrant.NewID(id),
				Vector:  qdrant.NewVectorsFloat32(vec),
				Payload: qdrant.NewPayloadFromMap(payload),
			},
		},
	})
	return err
}

// SearchResult represents a single similarity search result
type SearchResult struct {
	ID      string
	Score   float64
	Payload map[string]interface{}
}

// SearchSimilar performs vector similarity search using official client
func (vs *VectorStore) SearchSimilar(queryVec []float32, limit int, filter map[string]interface{}, withPayload bool) ([]SearchResult, error) {
	if !vs.Enabled {
		return nil, nil
	}
	searchParams := &qdrant.SearchPoints{
		CollectionName: vs.Collection,
		Vector:         qdrant.NewVectorsFloat32(queryVec),
		Limit:          uint64(limit),
		WithPayload:    withPayload,
	}
	if filter != nil {
		searchParams.Filter = qdrant.NewFilterFromMap(filter)
	}
	results, err := vs.Client.Search(context.Background(), searchParams)
	if err != nil {
		return nil, err
	}

	var searchResults []SearchResult
	for _, r := range results {
		searchResults = append(searchResults, SearchResult{
			ID:      r.ID.String(),
			Score:   r.Score,
			Payload: r.Payload,
		})
	}
	return searchResults, nil
}

// SearchByText is a convenience method that generates a simple embedding and searches.
func (vs *VectorStore) SearchByText(text string, limit int, filter map[string]interface{}, withPayload bool) ([]SearchResult, error) {
	vec := GenerateSimpleEmbedding(text, 768)
	return vs.SearchSimilar(vec, limit, filter, withPayload)
}

// CreateCollection helper for Qdrant using official client
func (vs *VectorStore) CreateCollection(dim int) error {
	if !vs.Enabled {
		return nil
	}
	_, err := vs.Client.CreateCollection(context.Background(), vs.Collection, &qdrant.CreateCollection{
		VectorsConfig: qdrant.NewVectorsConfig(
			&qdrant.VectorParams{
				Size:     uint64(dim),
				Distance: qdrant.DistanceCosine,
			},
		),
	})
	return err
}
