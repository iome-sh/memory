package memory

import (
	"strings"
	"testing"
	"time"
)

func TestPalaceStore_WriteLoad(t *testing.T) {
	tempDir := t.TempDir()
	store := NewPalaceStore(tempDir)

	entry := MemoryEntry{
		ID:        "test-id",
		Type:      "test",
		Tier:      TierContextual,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Cycle:     42,
		Content: MemoryContent{
			Summary: "test summary",
			Full:    "full test content",
		},
		Metrics: MemoryMetrics{ScoreImpact: 0.95, UsageCount: 1},
	}

	if err := store.Write(entry); err != nil {
		t.Fatal(err)
	}

	loaded, ok := store.Load("test-id", TierContextual)
	if !ok {
		t.Fatal("failed to load entry")
	}

	if loaded.ID != entry.ID || loaded.Content.Summary != entry.Content.Summary {
		t.Errorf("loaded entry mismatch: got %+v, want %+v", loaded, entry)
	}
}

func TestGenerateSimpleEmbedding(t *testing.T) {
	vec := GenerateSimpleEmbedding("test text", 4)
	if len(vec) != 4 {
		t.Errorf("expected dim 4, got %d", len(vec))
	}
	// check normalized
	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	if norm < 0.99 || norm > 1.01 {
		t.Errorf("vector not normalized, norm = %f", norm)
	}
}

func TestPopulateTemporalTags(t *testing.T) {
	tags := PopulateTemporalTags(42)
	if len(tags) != 4 {
		t.Errorf("expected 4 tags, got %d", len(tags))
	}
	if !strings.HasPrefix(tags[0], "cycle-42") {
		t.Errorf("unexpected first tag: %s", tags[0])
	}
}

func TestCosineSimilarity(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{1, 0, 0}
	if s := CosineSimilarity(a, b); s != 1.0 {
		t.Errorf("expected 1.0, got %f", s)
	}

	c := []float32{0, 1, 0}
	if s := CosineSimilarity(a, c); s != 0 {
		t.Errorf("expected 0, got %f", s)
	}
}

// RecMem Phase 1 tests
func TestCompactionConfig_RecMemDefaults(t *testing.T) {
	cfg := DefaultCompactionConfig
	if cfg.DataSim != 0.7 {
		t.Errorf("expected DataSim 0.7, got %f", cfg.DataSim)
	}
	if cfg.DataCount != 5 {
		t.Errorf("expected DataCount 5, got %d", cfg.DataCount)
	}
}

func TestPalaceStore_WriteLatent(t *testing.T) {
	tempDir := t.TempDir()
	store := NewPalaceStore(tempDir)

	entry := MemoryEntry{
		ID:        "latent-test",
		Type:      "latent",
		Tier:      TierWorking,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Content: MemoryContent{Summary: "latent fact"},
	}

	if err := store.WriteLatent(entry); err != nil {
		t.Fatal(err)
	}

	// Verify it exists in subconscious tier (via stats or direct load if exposed)
	// For now, just ensure no error and file exists conceptually.
	// Use _ for loaded since we only check the ok flag in this placeholder test.
	_, ok := store.Load("latent-test", TierWorking) // fallback check (note: may return false as it's in tier-0-subconscious)
	if ok {
		t.Log("latent entry loaded via fallback")
	}
}

// Phase 2 recurrence trigger tests
// Note: We use identical strings because the deterministic hash-based embedding
// produces perfect similarity only for identical input. This makes the test reliable.
func TestShouldTriggerPhaseTransition(t *testing.T) {
	cfg := DefaultCompactionConfig

	// Recurrent identical entries (high similarity) should trigger
	text := "project phoenix status update"
	similar := []MemoryEntry{
		{Content: MemoryContent{Summary: text}},
		{Content: MemoryContent{Summary: text}},
		{Content: MemoryContent{Summary: text}},
		{Content: MemoryContent{Summary: text}},
		{Content: MemoryContent{Summary: text}},
	}
	if !shouldTriggerPhaseTransition(similar, cfg) {
		t.Error("expected to trigger on recurrent identical cluster")
	}

	// Completely different entries should not trigger
	dissimilar := []MemoryEntry{
		{Content: MemoryContent{Summary: "weather in Vancouver today"}},
		{Content: MemoryContent{Summary: "best recipe for pad thai"}},
	}
	if shouldTriggerPhaseTransition(dissimilar, cfg) {
		t.Error("should not trigger on dissimilar entries")
	}
}

func TestClusterBySimilarity(t *testing.T) {
	text := "alpha sprint planning meeting"
	entries := []MemoryEntry{
		{Content: MemoryContent{Summary: text}},
		{Content: MemoryContent{Summary: text}},
		{Content: MemoryContent{Summary: "unrelated cooking tip"}},
	}
	clusters := clusterBySimilarity(entries, 0.7)
	if len(clusters) == 0 {
		t.Fatal("expected at least one cluster")
	}
	// The identical alpha items should form a cluster of size 2
	if len(clusters[0]) < 2 {
		t.Error("expected cluster to contain the recurrent identical items")
	}
}
