package memory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// VectorStore provides optional vector capabilities (Qdrant by default)
type VectorStore struct {
	URL        string
	Collection string
	Enabled    bool
}

func NewVectorStore(url, collection string) *VectorStore {
	return &VectorStore{
		URL:        url,
		Collection: collection,
		Enabled:    url != "",
	}
}

// StoreVector upserts a vector with payload (optional)
func (vs *VectorStore) StoreVector(id string, vec []float32, payload map[string]interface{}) error {
	if !vs.Enabled || vs.URL == "" {
		return nil
	}
	pointsURL := fmt.Sprintf("%s/collections/%s/points", vs.URL, vs.Collection)
	point := map[string]interface{}{
		"id":      id,
		"vector":  vec,
		"payload": payload,
	}
	body, _ := json.Marshal(map[string]interface{}{"points": []interface{}{point}})
	req, _ := http.NewRequest("PUT", pointsURL, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// SearchResult represents a single similarity search result
type SearchResult struct {
	ID      string
	Score   float64
	Payload map[string]interface{}
}

// SearchSimilar performs vector similarity search.
// Returns structured results with optional payloads.
func (vs *VectorStore) SearchSimilar(queryVec []float32, limit int, filter map[string]interface{}, withPayload bool) ([]SearchResult, error) {
	if !vs.Enabled || vs.URL == "" {
		return nil, nil
	}

	searchURL := fmt.Sprintf("%s/collections/%s/points/search", vs.URL, vs.Collection)

	reqBody := map[string]interface{}{
		"vector":       queryVec,
		"limit":        limit,
		"with_payload": withPayload,
		"params": map[string]interface{}{
			"hnsw_ef": 64,
		},
	}
	if filter != nil {
		reqBody["filter"] = filter
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", searchURL, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			ID      string                 `json:"id"`
			Score   float64                `json:"score"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(result.Result))
	for i, r := range result.Result {
		results[i] = SearchResult{
			ID:      fmt.Sprintf("%v", r.ID),
			Score:   r.Score,
			Payload: r.Payload,
		}
	}
	return results, nil
}

// SearchByText is a convenience method that generates a simple embedding and searches.
// For production use, replace with semantic embedding generation.
func (vs *VectorStore) SearchByText(text string, limit int, filter map[string]interface{}, withPayload bool) ([]SearchResult, error) {
	vec := GenerateSimpleEmbedding(text, 768)
	return vs.SearchSimilar(vec, limit, filter, withPayload)
}

// CreateCollection helper for Qdrant
func (vs *VectorStore) CreateCollection(dim int) error {
	if !vs.Enabled {
		return nil
	}
	url := fmt.Sprintf("%s/collections/%s", vs.URL, vs.Collection)
	payload := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     dim,
			"distance": "Cosine",
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", url, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// QdrantClientInitExample shows recommended initialization
// Example:
// vs := memory.NewVectorStore("http://localhost:6333", "my_collection")
// _ = vs.CreateCollection(768)
// store.Vector = vs
func QdrantClientInitExample() {
	// This is documentation only
	// Full Qdrant client initialization is handled via NewVectorStore + CreateCollection
}
