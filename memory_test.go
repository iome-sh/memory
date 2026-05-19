package memory

import (
	"os"
	"path/filepath"
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
	vec := generateSimpleEmbedding("test text", 4)
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
	tags := populateTemporalTags(42)
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
	if s := cosineSimilarity(a, b); s != 1.0 {
		t.Errorf("expected 1.0, got %f", s)
	}

	c := []float32{0, 1, 0}
	if s := cosineSimilarity(a, c); s != 0 {
		t.Errorf("expected 0, got %f", s)
	}
}
