package memory

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qdrant/go-client/qdrant"
)

// VectorStore provides optional vector capabilities using official Qdrant Go client v1
type VectorStore struct {
	Client     *qdrant.Client
	Collection string
	Enabled    bool
}

// NewVectorStore initializes the official Qdrant client.
func NewVectorStore(rawURL, collection string) *VectorStore {
	if rawURL == "" {
		return &VectorStore{Enabled: false}
	}

	host := "localhost"
	port := 6334

	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		u, err := url.Parse(rawURL)
		if err == nil {
			host = u.Hostname()
			if p := u.Port(); p != "" {
				if pi, err := strconv.Atoi(p); err == nil {
					port = pi
				}
			}
		}
	} else if strings.Contains(rawURL, ":") {
		parts := strings.Split(rawURL, ":")
		host = parts[0]
		if len(parts) > 1 {
			if pi, err := strconv.Atoi(parts[1]); err == nil {
				port = pi
			}
		}
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		fmt.Printf("[vector] failed to create qdrant client: %v\n", err)
		return &VectorStore{Enabled: false}
	}

	return &VectorStore{
		Client:     client,
		Collection: collection,
		Enabled:    true,
	}
}

// StoreVector upserts a dense vector
func (vs *VectorStore) StoreVector(id string, vec []float32, payload map[string]interface{}) error {
	if !vs.Enabled {
		return nil
	}
	_, err := vs.Client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: vs.Collection,
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewID(id),
				Vectors: qdrant.NewVectors(vec...),
				Payload: nil,
			},
		},
	})
	return err
}

// StoreSparseVector
func (vs *VectorStore) StoreSparseVector(id string, indices []uint32, values []float32, payload map[string]interface{}) error {
	if !vs.Enabled {
		return nil
	}
	_, err := vs.Client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: vs.Collection,
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewID(id),
				Vectors: qdrant.NewVectorsSparse(indices, values),
				Payload: nil,
			},
		},
	})
	return err
}

// BatchUpsert
func (vs *VectorStore) BatchUpsert(points []*qdrant.PointStruct) error {
	if !vs.Enabled {
		return nil
	}
	_, err := vs.Client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: vs.Collection,
		Points:         points,
	})
	return err
}

// SearchResult
type SearchResult struct {
	ID      string
	Score   float64
	Payload map[string]interface{}
}

// SearchSimilar (stub - full implementation via client.Query in future)
func (vs *VectorStore) SearchSimilar(queryVec []float32, limit int, filter map[string]interface{}, withPayload bool) ([]SearchResult, error) {
	if !vs.Enabled {
		return nil, nil
	}
	return nil, nil
}

func (vs *VectorStore) SearchSparse(queryIndices []uint32, queryValues []float32, limit int, filter map[string]interface{}, withPayload bool) ([]SearchResult, error) {
	if !vs.Enabled {
		return nil, nil
	}
	return nil, nil
}

func (vs *VectorStore) SearchByText(text string, limit int, filter map[string]interface{}, withPayload bool) ([]SearchResult, error) {
	return vs.SearchSimilar(nil, limit, filter, withPayload)
}

// CreateCollection
func (vs *VectorStore) CreateCollection(dim int) error {
	if !vs.Enabled {
		return nil
	}
	return vs.Client.CreateCollection(context.Background(), &qdrant.CreateCollection{
		CollectionName: vs.Collection,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     uint64(dim),
			Distance: qdrant.Distance_Cosine,
		}),
	})
}

// CreateSparseCollection
func (vs *VectorStore) CreateSparseCollection() error {
	if !vs.Enabled {
		return nil
	}
	return vs.Client.CreateCollection(context.Background(), &qdrant.CreateCollection{
		CollectionName: vs.Collection,
	})
}

const (
	defaultCollectionCreationConcurrency = 8
	defaultMaxRetries                    = 3
	defaultRetryBackoff                  = 500 * time.Millisecond
)

func (vs *VectorStore) CreateBatchSparseCollections(names []string) error {
	return vs.CreateBatchSparseCollectionsContext(context.Background(), names)
}

func (vs *VectorStore) CreateBatchSparseCollectionsContext(ctx context.Context, names []string) error {
	return vs.createBatchSparseCollections(ctx, names, defaultCollectionCreationConcurrency)
}

func (vs *VectorStore) CreateBatchSparseCollectionsWithConcurrency(names []string, concurrency int) error {
	return vs.CreateBatchSparseCollectionsWithConcurrencyContext(context.Background(), names, concurrency)
}

func (vs *VectorStore) CreateBatchSparseCollectionsWithConcurrencyContext(ctx context.Context, names []string, concurrency int) error {
	if concurrency <= 0 {
		concurrency = defaultCollectionCreationConcurrency
	}
	return vs.createBatchSparseCollections(ctx, names, concurrency)
}

func createCollectionWithRetry(ctx context.Context, client *qdrant.Client, name string, maxRetries int) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := defaultRetryBackoff * time.Duration(1<<uint(attempt-1))
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff + jitter):
			}
		}
		err := client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: name,
		})
		if err == nil {
			return nil
		}
		lastErr = err
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
					sem <- struct{}{}
					err := createCollectionWithRetry(ctx, vs.Client, name, defaultMaxRetries)
					<-sem
					if err != nil {
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
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

type CollectionCreateResult struct {
	Name  string
	Error error
}

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
	return fmt.Sprintf("failed to create %d collections (first: %s: %v)", len(e.Failures), e.Failures[0].Name, e.Failures[0].Error)
}

func (e *BatchCollectionError) Unwrap() []error {
	errs := make([]error, len(e.Failures))
	for i, f := range e.Failures {
		errs[i] = f.Error
	}
	return errs
}

func (e *BatchCollectionError) Is(target error) bool {
	for _, f := range e.Failures {
		if errors.Is(f.Error, target) {
			return true
		}
	}
	return false
}

func (vs *VectorStore) CreateBatchSparseCollectionsWithResults(ctx context.Context, names []string) ([]CollectionCreateResult, error) {
	return vs.createBatchSparseCollectionsWithResults(ctx, names, defaultCollectionCreationConcurrency)
}

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
