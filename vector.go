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

// SearchSimilar performs vector search (optional)
func (vs *VectorStore) SearchSimilar(queryVec []float32, limit int, filter map[string]interface{}) ([]string, []float64, error) {
	if !vs.Enabled || vs.URL == "" {
		return nil, nil, nil
	}
	searchURL := fmt.Sprintf("%s/collections/%s/points/search", vs.URL, vs.Collection)
	reqBody := map[string]interface{}{
		"vector":       queryVec,
		"limit":        limit,
		"with_payload": false,
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
		return nil, nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			ID    string  `json:"id"`
			Score float64 `json:"score"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, err
	}

	ids := make([]string, len(result.Result))
	scores := make([]float64, len(result.Result))
	for i, r := range result.Result {
		ids[i] = fmt.Sprintf("%v", r.ID)
		scores[i] = r.Score
	}
	return ids, scores, nil
}

// CreateCollection (helper for Qdrant)
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
