package memory

import (
	"context"
	"fmt"
	"net/http"
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

	// Empty query should be handled gracefully
	_, err := vs.SearchSimilar(nil, 10, nil, true)
	if err == nil {
		t.Error("SearchSimilar with empty vector should return error when enabled, but since disabled it returns nil, nil")
	}

	// We mainly test that no panic occurs on disabled store
	_ = vs.CreateBatchSparseCollections([]string{"c1", "c2"})
}

// startTemporaryQdrant starts a temporary Qdrant instance using Podman.
// It returns the connection URL (gRPC preferred) and a cleanup function.
// The test is skipped if Podman is not available or the container fails to start.
// This enables real integration tests for VectorStore when the environment supports it.
func startTemporaryQdrant(t *testing.T) (url string, cleanup func()) {
	t.Helper()

	if _, err := exec.LookPath("podman"); err != nil {
		t.Skip("podman not found in PATH - skipping temporary Qdrant integration test")
	}

	containerName := fmt.Sprintf("test-qdrant-%d", time.Now().UnixNano())

	// Start Qdrant container (REST on 6333, gRPC on 6334)
	cmd := exec.Command("podman", "run", "-d", "--rm",
		"--name", containerName,
		"-p", "6333:6333",
		"-p", "6334:6334",
		"qdrant/qdrant")

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to start temporary Qdrant container: %v\n%s", err, output)
	}

	url = "http://localhost:6334" // gRPC endpoint (Qdrant client uses this)

	// Wait for Qdrant to become ready (poll REST health)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	restURL := "http://localhost:6333/collections"
	ready := false
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timed out waiting for temporary Qdrant to become ready")
		default:
		}

		resp, err := http.Get(restURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	if !ready {
		t.Fatal("Qdrant container started but never became ready")
	}

	cleanup = func() {
		exec.Command("podman", "rm", "-f", containerName).Run()
	}

	t.Cleanup(cleanup)
	return url, cleanup
}

// TestVectorStore_WithTemporaryQdrant runs integration tests against a real
// temporary Qdrant instance started via Podman (when available).
func TestVectorStore_WithTemporaryQdrant(t *testing.T) {
	qdrantURL, cleanup := startTemporaryQdrant(t)
	defer cleanup()

	vs := NewVectorStore(qdrantURL, "test_integration_collection")
	if !vs.Enabled {
		t.Fatal("expected VectorStore to be enabled when connecting to temporary Qdrant")
	}

	// Basic smoke test
	if err := vs.CreateCollection(768); err != nil {
		t.Fatalf("CreateCollection failed against temporary Qdrant: %v", err)
	}

	vec := []float32{0.1, 0.2, 0.3}
	if err := vs.StoreVector("vec-1", vec, map[string]interface{}{"type": "test"}); err != nil {
		t.Fatalf("StoreVector failed: %v", err)
	}

	results, err := vs.SearchSimilar(vec, 5, nil, true)
	if err != nil {
		t.Fatalf("SearchSimilar failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result from SearchSimilar")
	}
}
