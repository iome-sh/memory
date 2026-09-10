package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestPalaceStore_ConcurrentSharedFileWrites(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStore(base)
	if err := store.Write(MemoryEntry{
		ID:      "seed",
		Tier:    TierContextual,
		Content: MemoryContent{Summary: "seed for clean meta index"},
	}); err != nil {
		t.Fatal(err)
	}
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})

	const n = 40
	var wg sync.WaitGroup
	errCh := make(chan error, n*2)
	for i := 0; i < n; i++ {
		i := i
		wg.Add(2)
		go func() {
			defer wg.Done()
			err := store.Write(MemoryEntry{
				ID:        fmt.Sprintf("w-%d", i),
				Tier:      TierContextual,
				SessionID: "sess-concurrent",
				Content:   MemoryContent{Summary: fmt.Sprintf("concurrent write %d", i)},
			})
			if err != nil {
				errCh <- err
			}
		}()
		go func() {
			defer wg.Done()
			if err := store.IngestTurn(MemoryEntry{
				ID:             fmt.Sprintf("t-%d", i),
				SessionID:      "sess-ingest",
				Content:        MemoryContent{Full: fmt.Sprintf("ingest turn %d unique token", i)},
				ExtractedFacts: []string{fmt.Sprintf("fact from ingest %d", i)},
			}); err != nil {
				errCh <- err
			}
		}()
	}
	for i := 0; i < n; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.AddEntityRelationship(fmt.Sprintf("e%d", i), fmt.Sprintf("r%d", i))
			store.AddEntityRelationship("hub", fmt.Sprintf("spoke-%d", i))
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}

	graphPath := filepath.Join(base, "relations", "entity-graph.json")
	raw, err := os.ReadFile(graphPath)
	if err != nil {
		t.Fatal(err)
	}
	var graph map[string][]string
	if err := json.Unmarshal(raw, &graph); err != nil {
		t.Fatalf("entity-graph.json torn or invalid JSON: %v\n%s", err, raw)
	}
	if len(graph["hub"]) != n {
		t.Fatalf("hub spokes = %d, want %d (lost concurrent graph edges)", len(graph["hub"]), n)
	}
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("e%d", i)
		if len(graph[key]) != 1 || graph[key][0] != fmt.Sprintf("r%d", i) {
			t.Fatalf("graph[%s] = %v", key, graph[key])
		}
	}
	if info, err := os.Stat(graphPath); err != nil {
		t.Fatal(err)
	} else if got := info.Mode().Perm(); got != palaceFileMode {
		t.Fatalf("entity-graph.json mode %04o, want %04o", got, palaceFileMode)
	}

	idxPath := filepath.Join(base, "indexes", "event-time.json")
	idxRaw, err := os.ReadFile(idxPath)
	if err != nil {
		t.Fatal(err)
	}
	var snap durableEventTimeIndex
	if err := json.Unmarshal(idxRaw, &snap); err != nil {
		t.Fatalf("event-time.json torn or invalid JSON: %v\n%s", err, idxRaw)
	}
	if snap.Version != durableEventTimeIndexVersion {
		t.Fatalf("event-time version = %d", snap.Version)
	}
	if snap.JSONCount < n {
		t.Fatalf("event-time JSONCount = %d, want at least %d writes", snap.JSONCount, n)
	}
	if info, err := os.Stat(idxPath); err != nil {
		t.Fatal(err)
	} else if got := info.Mode().Perm(); got != palaceFileMode {
		t.Fatalf("event-time.json mode %04o, want %04o", got, palaceFileMode)
	}

	if got := store.GetRelatedEntities("hub"); len(got) != n {
		t.Fatalf("GetRelatedEntities(hub) = %d, want %d", len(got), n)
	}
}
