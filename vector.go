package memory

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
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

// StoreVector upserts a dense vector with payload
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

// StoreSparseVector upserts a sparse vector (H-Mem style exploration)
// Sparse vectors are useful for high-dimensional sparse embeddings (e.g., BM25-style or keyword vectors)
func (vs *VectorStore) StoreSparseVector(id string, indices []uint32, values []float32, payload map[string]interface{}) error {
	if !vs.Enabled {
		return nil
	}
	sparseVec := &qdrant.SparseVector{
		Indices: indices,
		Values:  values,
	}
	_, err := vs.Client.Upsert(context.Background(), vs.Collection, &qdrant.PointStructs{
		Points: []*qdrant.PointStruct{
			{
				ID:      qdrant.NewID(id),
				Vector:  qdrant.NewVectorsSparse(sparseVec),
				Payload: qdrant.NewPayloadFromMap(payload),
			},
		},
	})
	return err
}

// BatchUpsert performs efficient batch upserts of multiple vectors
func (vs *VectorStore) BatchUpsert(points []*qdrant.PointStruct) error {
	if !vs.Enabled {
		return nil
	}
	_, err := vs.Client.Upsert(context.Background(), vs.Collection, &qdrant.PointStructs{
		Points: points,
	})
	return err
}

// SearchResult represents a single similarity search result
type SearchResult struct {
	ID      string
	Score   float64
	Payload map[string]interface{}
}

// SearchSimilar performs dense vector similarity search using official client
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

// SearchSparse performs sparse vector similarity search (H-Mem exploration)
func (vs *VectorStore) SearchSparse(queryIndices []uint32, queryValues []float32, limit int, filter map[string]interface{}, withPayload bool) ([]SearchResult, error) {
	if !vs.Enabled {
		return nil, nil
	}
	sparseQuery := &qdrant.SparseVector{
		Indices: queryIndices,
		Values:  queryValues,
	}
	searchParams := &qdrant.SearchPoints{
		CollectionName: vs.Collection,
		Vector:         qdrant.NewVectorsSparse(sparseQuery),
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

// SearchByText is a convenience method that generates a simple dense embedding and searches.
func (vs *VectorStore) SearchByText(text string, limit int, filter map[string]interface{}, withPayload bool) ([]SearchResult, error) {
	vec := GenerateSimpleEmbedding(text, 768)
	return vs.SearchSimilar(vec, limit, filter, withPayload)
}

// CreateCollection helper for Qdrant using official client (dense by default)
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

// CreateSparseCollection creates a collection with sparse vector support (H-Mem KG style)
func (vs *VectorStore) CreateSparseCollection() error {
	if !vs.Enabled {
		return nil
	}
	_, err := vs.Client.CreateCollection(context.Background(), vs.Collection, &qdrant.CreateCollection{
		VectorsConfig: qdrant.NewVectorsConfig(
			&qdrant.SparseVectorParams{
				Index: &qdrant.SparseIndexParams{},
			},
		),
	})
	return err
}

const (
	defaultCollectionCreationConcurrency = 8
	defaultMaxRetries                    = 3
	defaultRetryBackoff                  = 500 * time.Millisecond
)

// CreateBatchSparseCollections creates multiple sparse vector collections concurrently
// using a bounded worker pool.
//
// It uses context.Background(). For cancellation/timeout support, use
// CreateBatchSparseCollectionsContext instead.
func (vs *VectorStore) CreateBatchSparseCollections(names []string) error {
	return vs.CreateBatchSparseCollectionsContext(context.Background(), names)
}

// CreateBatchSparseCollectionsContext is the context-aware version of CreateBatchSparseCollections.
// It supports cancellation and deadlines. When the context is cancelled, in-flight and
// pending collection creations are stopped as quickly as possible.
func (vs *VectorStore) CreateBatchSparseCollectionsContext(ctx context.Context, names []string) error {
	return vs.createBatchSparseCollections(ctx, names, defaultCollectionCreationConcurrency)
}

// CreateBatchSparseCollectionsWithConcurrency lets you control concurrency.
// It uses context.Background(). Prefer the Context variant for cancellation support.
func (vs *VectorStore) CreateBatchSparseCollectionsWithConcurrency(names []string, concurrency int) error {
	return vs.CreateBatchSparseCollectionsWithConcurrencyContext(context.Background(), names, concurrency)
}

// CreateBatchSparseCollectionsWithConcurrencyContext is the full context-aware + configurable
// concurrency version. Recommended for production use when you need cancellation or custom limits.
func (vs *VectorStore) CreateBatchSparseCollectionsWithConcurrencyContext(ctx context.Context, names []string, concurrency int) error {
	if concurrency <= 0 {
		concurrency = defaultCollectionCreationConcurrency
	}
	return vs.createBatchSparseCollections(ctx, names, concurrency)
}

// createCollectionWithRetry attempts to create a collection with retry logic and exponential backoff + jitter.
// It respects context cancellation between retries.
func createCollectionWithRetry(ctx context.Context, client *qdrant.Client, name string, maxRetries int) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			backoff := defaultRetryBackoff * time.Duration(1<<uint(attempt-1))
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff + jitter):
			}
		}

		_, err := client.CreateCollection(ctx, name, &qdrant.CreateCollection{
			VectorsConfig: qdrant.NewVectorsConfig(
				&qdrant.SparseVectorParams{
				Index: &qdrant.SparseIndexParams{},
			},
		),
		})
		if err == nil {
			return nil // Success
		}
		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return lastErr
}

func (vs *VectorStore) createBatchSparseCollections(ctx context.Context, names []string, concurrency int) error {
	if !vs.Enabled || len(names) == 0 {
		return nil
	}

	jobs := make(chan string, len(names))
	for _, name := range names {
		jobs <- name
	}
	close(jobs)

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	sem := make(chan struct{}, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case name, ok := <-jobs:
					if !ok {
						return
					}
					sem <- struct{}{} // acquire

					err := createCollectionWithRetry(ctx, vs.Client, name, defaultMaxRetries)
					<-sem // release

					if err != nil {
						// Only send first error
						select {
						case errCh <- fmt.Errorf("failed to create sparse collection %s: %w", name, err):
						default:
						}
						return
					}
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	if err, ok := <-errCh; ok {
		return err
	}

	// Check if we were cancelled even if no creation error occurred
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

// CollectionCreateResult represents the outcome of a single collection creation attempt.
type CollectionCreateResult struct {
	Name  string
	Error error
}

// BatchCollectionError aggregates all failures from a batch collection creation operation.
// It implements error and supports errors.Is / errors.As / errors.Unwrap for rich inspection.
type BatchCollectionError struct {
	Failures []CollectionCreateResult
}

func (e *BatchCollectionError) Error() string {
	if len(e.Failures) == 0 {
		return "batch collection creation succeeded"
	}
	if len(e.Failures) == 1 {
		return fmt.Sprintf("failed to create collection %s: %v", e.Failures[0].Name, e.Failures[0].Error)
	}
	return fmt.Sprintf("failed to create %d collections (first: %s: %v)",
		len(e.Failures), e.Failures[0].Name, e.Failures[0].Error)
}

// Unwrap returns all individual errors (Go 1.20+ multi-error support).
func (e *BatchCollectionError) Unwrap() []error {
	errs := make([]error, len(e.Failures))
	for i, f := range e.Failures {
		errs[i] = f.Error
	}
	return errs
}

// Is implements errors.Is support.
func (e *BatchCollectionError) Is(target error) bool {
	for _, f := range e.Failures {
		if errors.Is(f.Error, target) {
			return true
		}
	}
	return false
}

// CreateBatchSparseCollectionsWithResults attempts to create all collections and returns
// detailed results for every collection (success or failure). This is ideal when you want
// partial results — e.g. some collections were created before a timeout or cancellation.
//
// It uses the default concurrency (8). Use the WithConcurrency variant for custom limits.
// The returned error is the first error encountered (or nil). The results slice always
// contains an entry for every input name.
func (vs *VectorStore) CreateBatchSparseCollectionsWithResults(ctx context.Context, names []string) ([]CollectionCreateResult, error) {
	return vs.createBatchSparseCollectionsWithResults(ctx, names, defaultCollectionCreationConcurrency)
}

// CreateBatchSparseCollectionsWithResultsAndConcurrency is the full version with
// configurable concurrency + detailed per-collection results.
func (vs *VectorStore) CreateBatchSparseCollectionsWithResultsAndConcurrency(ctx context.Context, names []string, concurrency int) ([]CollectionCreateResult, error) {
	if concurrency <= 0 {
		concurrency = defaultCollectionCreationConcurrency
	}
	return vs.createBatchSparseCollectionsWithResults(ctx, names, concurrency)
}

func (vs *VectorStore) createBatchSparseCollectionsWithResults(ctx context.Context, names []string, concurrency int) ([]CollectionCreateResult, error) {
	if !vs.Enabled || len(names) == 0 {
		return nil, nil
	}

	jobs := make(chan string, len(names))
	for _, name := range names {
		jobs <- name
	}
	close(jobs)

	var wg sync.WaitGroup
	resultsCh := make(chan CollectionCreateResult, len(names))
	sem := make(chan struct{}, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case name, ok := <-jobs:
					if !ok {
						return
					}
					sem <- struct{}{}

					err := createCollectionWithRetry(ctx, vs.Client, name, defaultMaxRetries)
					<-sem

					resultsCh <- CollectionCreateResult{Name: name, Error: err}
				}
			}
		}()
	}

	wg.Wait()
	close(resultsCh)

	// Collect results
	var results []CollectionCreateResult
	var failures []CollectionCreateResult

	for res := range resultsCh {
		results = append(results, res)
		if res.Error != nil {
			failures = append(failures, res)
	}
	}

	var finalErr error
	if len(failures) > 0 {
		finalErr = &BatchCollectionError{Failures: failures}
	} else if ctx.Err() != nil {
		finalErr = ctx.Err()
	}
	return results, finalErr
}

// GenerateSimpleEmbedding provides a deterministic pseudo-embedding (for demo / fallback use).
// In production, replace with a real embedding model (e.g. via external service or ONNX).
func GenerateSimpleEmbedding(text string, dim int) []float32 {
	if dim <= 0 {
		dim = 768
	}
	h := sha256.Sum256([]byte(text))
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		vec[i] = float32(h[i%32]) / 255.0
	}
	// L2 normalize
	var norm float32
	for _, v := range vec {
		norm += v * v
	}
	if norm > 0 {
		norm = float32(math.Sqrt(float64(norm)))
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec
}
