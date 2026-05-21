package memory

import (
	"context"
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

// NOTE on real Qdrant testing:
// For integration tests against a real Qdrant instance you can use Podman:
//
//   podman run -d --rm --name test-qdrant -p 6333:6333 -p 6334:6334 qdrant/qdrant
//
// Then construct VectorStore with "http://localhost:6334".
// A future helper (startTemporaryQdrant) can be added here that uses
// os/exec + podman to spin up a temporary container for the duration of the test
// and cleans it up automatically. This keeps unit tests fast while allowing
// optional integration coverage when the environment supports it.
