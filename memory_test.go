package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewPalaceStoreWithConfig_EmptyBaseDirDefaultsToDotPalace(t *testing.T) {
	t.Chdir(t.TempDir())

	store := NewPalaceStoreWithConfig(PalaceConfig{})
	if store.BaseDir != DefaultPalaceBaseDir {
		t.Fatalf("empty PalaceConfig.BaseDir = %q, want %q", store.BaseDir, DefaultPalaceBaseDir)
	}
	if store.Config.BaseDir != DefaultPalaceBaseDir {
		t.Fatalf("Config.BaseDir = %q, want %q", store.Config.BaseDir, DefaultPalaceBaseDir)
	}
	if strings.Contains(store.BaseDir, ".ossa") {
		t.Fatalf("leftover product default still in use: %q", store.BaseDir)
	}

	legacy := NewPalaceStore("")
	if legacy.BaseDir != DefaultPalaceBaseDir {
		t.Fatalf("NewPalaceStore(\"\") BaseDir = %q, want %q", legacy.BaseDir, DefaultPalaceBaseDir)
	}

	if _, err := os.Stat(DefaultPalaceBaseDir); err != nil {
		t.Fatalf("expected default palace dir under cwd: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".ossa", "kb", "palace")); !os.IsNotExist(err) {
		t.Fatalf("must not create leftover .ossa/kb/palace: err=%v", err)
	}
}

func TestNewPalaceStoreWithConfig_ExplicitBaseDirUnchanged(t *testing.T) {
	dir := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: dir, MaxWorkingEntries: 7})
	if store.BaseDir != dir {
		t.Fatalf("explicit BaseDir = %q, want %q", store.BaseDir, dir)
	}
	if store.Config.MaxWorkingEntries != 7 {
		t.Fatalf("MaxWorkingEntries = %d, want 7", store.Config.MaxWorkingEntries)
	}
}

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

func TestEvictWorkingTier_UnlinksWorking(t *testing.T) {
	store := NewPalaceStore(t.TempDir())
	base := time.Now().Add(-4 * time.Hour)
	for i, id := range []string{"w1", "w2", "w3"} {
		ts := base.Add(time.Duration(i) * time.Hour)
		if err := store.Write(MemoryEntry{
			ID:        id,
			Type:      "note",
			Tier:      TierWorking,
			Version:   1,
			CreatedAt: ts,
			UpdatedAt: ts,
			Content:   MemoryContent{Summary: "working " + id},
		}); err != nil {
			t.Fatal(err)
		}
	}

	store.EvictWorkingTier(1, 1)

	if got := store.ListEntriesInTier(TierWorking); len(got) != 1 {
		t.Fatalf("working after evict = %d, want 1 (newest kept)", len(got))
	}
	if _, ok := store.Load("w1", TierWorking); ok {
		t.Fatal("evicted w1 still in working")
	}
	if _, ok := store.Load("w1", TierContextual); !ok {
		t.Fatal("expected evicted w1 in contextual")
	}
	if _, ok := store.Load("w3", TierWorking); !ok {
		t.Fatal("expected newest w3 to remain in working")
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
		Content:   MemoryContent{Summary: "latent fact"},
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

// Phase 3 SemanticRefine tests
func TestSemanticRefine(t *testing.T) {
	tempDir := t.TempDir()
	store := NewPalaceStore(tempDir)

	entry := MemoryEntry{
		ID:   "cluster-entry-1",
		Type: "event",
		Tier: TierContextual,
		Content: MemoryContent{
			Summary: "Meeting with John Doe on 2025-11-17 about project Phoenix.",
			Full:    "Important meeting with John Doe on 2025-11-17. He mentioned the deadline is strict.",
		},
	}

	cluster := []MemoryEntry{entry}

	err := store.SemanticRefine(cluster)
	if err != nil {
		t.Fatalf("SemanticRefine failed: %v", err)
	}

	// Check that semantic facts were created
	stats := store.GetStats()
	if stats.SemanticCount == 0 {
		t.Error("expected at least one semantic fact to be created")
	}
}

func TestSemanticRefine_StampsValidFrom(t *testing.T) {
	store := NewPalaceStore(t.TempDir())
	ts := time.Now().UTC().Add(-3 * time.Hour).Truncate(time.Second)
	entry := MemoryEntry{
		ID:        "cluster-entry-vf",
		Type:      "event",
		Tier:      TierContextual,
		SessionID: "sess-refine",
		Timestamp: ts,
		Content: MemoryContent{
			Summary: "Meeting with John Doe on 2025-11-17 about project Phoenix.",
			Full:    "Important meeting with John Doe on 2025-11-17. He mentioned the deadline is strict.",
		},
	}
	if err := store.SemanticRefine([]MemoryEntry{entry}); err != nil {
		t.Fatal(err)
	}

	facts := store.ListEntriesInTier(TierSemantic)
	if len(facts) == 0 {
		t.Fatal("expected at least one atomic fact")
	}
	now := time.Now().UTC()
	for _, f := range facts {
		if f.Type != "atomic_fact" {
			t.Fatalf("type = %q, want atomic_fact", f.Type)
		}
		if f.SessionID != "sess-refine" || f.Provenance.SourceStep != "semantic_refine" {
			t.Fatalf("unexpected fact metadata: %+v", f)
		}
		if !f.Timestamp.Equal(ts) {
			t.Fatalf("Timestamp = %v, want %v", f.Timestamp, ts)
		}
		if !hasValidFromTag(f) {
			t.Fatalf("missing valid_from on %s tags=%v", f.ID, f.TemporalTags)
		}
		from, until := ParseValidityWindow(f)
		if from == nil {
			t.Fatalf("ParseValidityWindow from=nil on %s tags=%v", f.ID, f.TemporalTags)
		}
		if until != nil {
			t.Fatalf("did not expect valid_until on new fact %s tags=%v", f.ID, f.TemporalTags)
		}
		if !EntryValidAt(f, now) {
			t.Fatalf("EntryValidAt(now) false for %s from=%v", f.ID, from)
		}
		if !EntryValidAt(f, time.Time{}) {
			t.Fatalf("EntryValidAt(zero→now) false for %s", f.ID)
		}
	}
}

func TestSemanticRefine_FactWriteError(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStore(base)
	semDir := filepath.Join(base, "tier-4-semantic")
	if err := os.Chmod(semDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(semDir, 0755) })

	err := store.SemanticRefine([]MemoryEntry{{
		ID:   "cluster-write-err",
		Type: "event",
		Content: MemoryContent{
			Full: "I graduated from MIT. I live in Seattle now.",
		},
	}})
	if err == nil {
		t.Fatal("expected semantic fact Write error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to write semantic fact") {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := store.ListEntriesInTier(TierSemantic); len(got) != 0 {
		t.Fatalf("semantic facts after failed write = %d, want 0", len(got))
	}
}
