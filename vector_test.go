package memory

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestVectorStore_Disabled verifies that VectorStore gracefully degrades
// when no Qdrant URL is provided (common in unit tests and non-vector deployments).
func TestVectorStore_Disabled(t *testing.T) {
	vs := NewVectorStore("", "test-collection")
	if vs.Enabled {
		t.Fatal("expected VectorStore to be disabled when URL is empty")
	}

	// All operations should be no-ops and return nil error
	if err := vs.StoreVector("id-1", []float32{0.1, 0.2}, nil); err != nil {
		t.Errorf("StoreVector on disabled store should succeed (no-op), got: %v", err)
	}

	if err := vs.StoreSparseVector("id-2", []uint32{0}, []float32{0.5}, nil); err != nil {
		t.Errorf("StoreSparseVector on disabled store should succeed (no-op), got: %v", err)
	}

	if err := vs.BatchUpsert(nil); err != nil {
		t.Errorf("BatchUpsert on disabled store should succeed (no-op), got: %v", err)
	}

	results, err := vs.SearchSimilar([]float32{0.1}, 10, nil, true)
	if err != nil || results != nil {
		t.Errorf("SearchSimilar on disabled store should return nil, nil; got err=%v, results=%v", err, results)
	}

	results, err = vs.SearchSparse([]uint32{0}, []float32{0.1}, 5, nil, false)
	if err != nil || results != nil {
		t.Errorf("SearchSparse on disabled store should return nil, nil; got err=%v, results=%v", err, results)
	}

	results, err = vs.SearchByText("hello", 3, nil, true)
	if err != nil || results != nil {
		t.Errorf("SearchByText on disabled store should return nil, nil; got err=%v, results=%v", err, results)
	}

	if err := vs.CreateCollection(768); err != nil {
		t.Errorf("CreateCollection on disabled store should succeed (no-op), got: %v", err)
	}

	if err := vs.CreateSparseCollection(); err != nil {
		t.Errorf("CreateSparseCollection on disabled store should succeed (no-op), got: %v", err)
	}
}

// TestVectorStore_EnabledFalsePaths exercises additional disabled paths
// and error conditions that do not require a running Qdrant.
func TestVectorStore_EnabledFalsePaths(t *testing.T) {
	vs := NewVectorStore("", "test")

	// Empty query should be handled gracefully (disabled path returns nil, nil)
	_, err := vs.SearchSimilar(nil, 10, nil, true)
	if err != nil {
		t.Errorf("SearchSimilar on disabled store should return nil error, got: %v", err)
	}

	// We mainly test that no panic occurs on disabled store
	_ = vs.CreateBatchSparseCollections([]string{"c1", "c2"})
}

// envQdrantTestURL points the Qdrant integration test at an already running
// Qdrant (gRPC, e.g. http://localhost:6334), such as a CI service container.
// When it is set, an unreachable or unready Qdrant fails the test instead of
// skipping it.
const envQdrantTestURL = "MEMORY_QDRANT_TEST_URL"

// qdrantUnavailable skips the test, or fails it when envQdrantTestURL is set.
func qdrantUnavailable(t *testing.T, format string, args ...any) {
	t.Helper()
	if os.Getenv(envQdrantTestURL) != "" {
		t.Fatalf(format+" (%s is set, so this is a failure, not a skip)", append(args, envQdrantTestURL)...)
	}
	t.Skipf(format, args...)
}

// waitQdrantReady polls the gRPC health check until it answers or the timeout passes.
func waitQdrantReady(rawURL string, timeout time.Duration) error {
	vs := NewVectorStore(rawURL, "readiness")
	if !vs.Enabled {
		return fmt.Errorf("could not create Qdrant client for %s", rawURL)
	}
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, lastErr = vs.Client.HealthCheck(ctx)
		cancel()
		if lastErr == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("qdrant at %s not ready after %s: %v", rawURL, timeout, lastErr)
}

// startTemporaryQdrant returns a gRPC URL for a Qdrant instance. It uses
// envQdrantTestURL when set; otherwise it starts a temporary Podman container
// (REST on 6333, gRPC on 6334) and skips if Podman is unavailable.
func startTemporaryQdrant(t *testing.T) string {
	t.Helper()

	if external := os.Getenv(envQdrantTestURL); external != "" {
		if err := waitQdrantReady(external, 45*time.Second); err != nil {
			qdrantUnavailable(t, "%v", err)
		}
		return external
	}

	if _, err := exec.LookPath("podman"); err != nil {
		qdrantUnavailable(t, "podman not found in PATH and %s unset - skipping Qdrant integration test", envQdrantTestURL)
	}

	containerName := fmt.Sprintf("test-qdrant-%d", time.Now().UnixNano())
	cmd := exec.Command("podman", "run", "-d", "--rm",
		"--name", containerName,
		"-p", "6333:6333",
		"-p", "6334:6334",
		"qdrant/qdrant:v1.19.2")
	if output, err := cmd.CombinedOutput(); err != nil {
		qdrantUnavailable(t, "failed to start temporary Qdrant container: %v\n%s", err, output)
	}
	t.Cleanup(func() { _ = exec.Command("podman", "rm", "-f", containerName).Run() })

	url := "http://localhost:6334"
	if err := waitQdrantReady(url, 45*time.Second); err != nil {
		qdrantUnavailable(t, "%v", err)
	}
	return url
}

// TestVectorStore_WithTemporaryQdrant runs integration tests against a real
// Qdrant: the one named by MEMORY_QDRANT_TEST_URL, or a temporary Podman container.
func TestVectorStore_WithTemporaryQdrant(t *testing.T) {
	qdrantURL := startTemporaryQdrant(t)

	collection := fmt.Sprintf("test_integration_%d", time.Now().UnixNano())
	vs := NewVectorStore(qdrantURL, collection)
	if !vs.Enabled {
		t.Fatal("expected VectorStore to be enabled when connecting to Qdrant")
	}
	if err := vs.CreateCollection(768); err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}
	t.Cleanup(func() { _ = vs.Client.DeleteCollection(context.Background(), collection) })

	// Qdrant point IDs must be UUIDs.
	seeds := []struct{ id, text string }{
		{"550e8400-e29b-41d4-a716-446655440000", "integration test vector for qdrant"},
		{"550e8400-e29b-41d4-a716-446655440001", "the user moved to Lisbon in March"},
		{"550e8400-e29b-41d4-a716-446655440002", "favourite colour is green"},
	}
	for _, s := range seeds {
		vec := GenerateSimpleEmbedding(s.text, 768)
		if err := vs.StoreVector(s.id, vec, map[string]interface{}{"text": s.text}); err != nil {
			t.Fatalf("StoreVector(%s) failed: %v", s.id, err)
		}
	}

	// Each seed's own vector must come back as the top hit with its payload.
	for _, want := range seeds {
		query := GenerateSimpleEmbedding(want.text, 768)
		var results []SearchResult
		var err error
		deadline := time.Now().Add(10 * time.Second)
		for {
			results, err = vs.SearchSimilar(query, len(seeds), nil, true)
			if err != nil {
				t.Fatalf("SearchSimilar failed: %v", err)
			}
			if len(results) == len(seeds) || time.Now().After(deadline) {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		if len(results) != len(seeds) {
			t.Fatalf("SearchSimilar returned %d results, want %d", len(results), len(seeds))
		}
		if results[0].ID != want.id {
			t.Errorf("top hit for %q = %s, want %s", want.text, results[0].ID, want.id)
		}
		if got := results[0].Payload["text"]; got != want.text {
			t.Errorf("top hit payload text = %v, want %q", got, want.text)
		}
	}
}
