package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPalaceConfig_TransactionalIngestDefaultFalse(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	if store.Config.TransactionalIngest {
		t.Fatal("TransactionalIngest default must be false (laptop partial persist)")
	}
	if store.Config.PersistEmbeddings {
		t.Fatal("PersistEmbeddings default must stay off")
	}
}

func TestTransactionalIngestDefaultLeavesTwoProcessProbeValid(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	if store.Config.TransactionalIngest {
		t.Fatal("laptop default TransactionalIngest must stay false so two-process last-write-wins probe remains valid; flock is not shipped")
	}
}

func TestIngestTurn_TransactionalIngest_ParentAndFacts(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:             base,
		TransactionalIngest: true,
	})
	if err := store.IngestTurn(MemoryEntry{
		ID:        "turn-tx-1",
		TurnID:    "turn-tx-1",
		SessionID: "sess-tx",
		Content:   MemoryContent{Full: "I live in Seattle. My name is Alice."},
		ExtractedFacts: []string{
			"I live in Seattle",
			"My name is Alice",
		},
	}); err != nil {
		t.Fatal(err)
	}

	parent, ok := store.Load("turn-tx-1", TierContextual)
	if !ok {
		t.Fatal("parent missing after transactional ingest")
	}
	if parent.Provenance.SourceHint != SourceHintPrivate {
		t.Fatalf("parent source_hint=%q want private", parent.Provenance.SourceHint)
	}
	facts := store.ListEntriesInTier(TierSemantic)
	if len(facts) != 2 {
		t.Fatalf("semantic facts = %d, want 2", len(facts))
	}
	for _, f := range facts {
		if f.Type != "turn_fact" {
			t.Fatalf("type = %q, want turn_fact", f.Type)
		}
		if f.TurnID != "turn-tx-1" {
			t.Fatalf("fact TurnID = %q", f.TurnID)
		}
		if f.Content.Embedding != nil || f.Content.EmbeddingModel != "" || f.Content.EmbeddingDim != 0 {
			t.Fatalf("hash embeddings must not be stored: %+v", f.Content)
		}
	}
	assertNoPendingOrTemps(t, base)
}

func TestIngestTurn_TransactionalIngest_ChildDirErrorDoesNotPersistParent(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:             base,
		TransactionalIngest: true,
	})
	semDir := filepath.Join(base, "tier-4-semantic")
	if err := os.Chmod(semDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(semDir, 0755) })

	err := store.IngestTurn(MemoryEntry{
		ID:             "turn-tx-fail",
		TurnID:         "turn-tx-fail",
		Content:        MemoryContent{Full: "turn body"},
		ExtractedFacts: []string{"I graduated from MIT"},
	})
	if err == nil {
		t.Fatal("expected transactional ingest error, got nil")
	}
	if _, ok := store.Load("turn-tx-fail", TierContextual); ok {
		t.Fatal("parent dest must stay absent when a child temp cannot be written")
	}
	if got := store.ListEntriesInTier(TierSemantic); len(got) != 0 {
		t.Fatalf("semantic facts after failed tx ingest = %d, want 0", len(got))
	}
	assertNoPendingOrTemps(t, base)
}

func TestIngestTurn_TransactionalIngest_TornWriteRecover(t *testing.T) {
	base := t.TempDir()
	_ = NewPalaceStoreWithConfig(PalaceConfig{BaseDir: base})

	parent := MemoryEntry{
		ID:        "turn-torn",
		TurnID:    "turn-torn",
		Type:      "turn",
		Tier:      TierContextual,
		Version:   1,
		SessionID: "sess-torn",
		Content:   MemoryContent{Full: "parent after crash recover", Tags: []string{"source_hint:private"}},
		Provenance: MemoryProvenance{
			SourceStep: "ingest_turn",
			SourceHint: SourceHintPrivate,
		},
	}
	fact := MemoryEntry{
		ID:        "fact-torn",
		TurnID:    "turn-torn",
		Type:      "turn_fact",
		Tier:      TierSemantic,
		Version:   1,
		SessionID: "sess-torn",
		Content:   MemoryContent{Full: "I live in Seattle", Summary: "I live in Seattle", Tags: []string{"fact_augmented", "from_turn"}},
		Provenance: MemoryProvenance{
			SourceStep: "ingest_turn_fact",
			SourceHint: SourceHintPrivate,
			ParentIDs:  []string{"turn-torn"},
		},
	}
	plantPendingIngest(t, base, "turn-torn", []MemoryEntry{parent, fact}, plantTempsIntact)

	if _, err := os.Stat(filepath.Join(base, "tier-2-contextual", "turn-torn.json")); !os.IsNotExist(err) {
		t.Fatalf("dest parent should be absent before recover: %v", err)
	}

	// Recover-on-open even when TransactionalIngest is false (cloud crash leftover).
	reopened := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: base})
	if reopened.Config.TransactionalIngest {
		t.Fatal("reopen used default TransactionalIngest false")
	}
	gotParent, ok := reopened.Load("turn-torn", TierContextual)
	if !ok {
		t.Fatal("parent dest missing after torn-write recover")
	}
	if gotParent.Content.Full != parent.Content.Full {
		t.Fatalf("parent full = %q", gotParent.Content.Full)
	}
	gotFact, ok := reopened.Load("fact-torn", TierSemantic)
	if !ok {
		t.Fatal("fact dest missing after torn-write recover")
	}
	if gotFact.Content.Full != fact.Content.Full {
		t.Fatalf("fact full = %q", gotFact.Content.Full)
	}
	assertNoPendingOrTemps(t, base)
}

func TestIngestTurn_TransactionalIngest_PendingWithoutTempsDiscarded(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: base})
	old := MemoryEntry{
		ID:      "turn-keep",
		Type:    "turn",
		Tier:    TierContextual,
		Version: 1,
		Content: MemoryContent{Full: "already committed parent"},
	}
	if err := store.Write(old); err != nil {
		t.Fatal(err)
	}

	newer := old
	newer.Content.Full = "should not replace dest without temps"
	plantPendingIngest(t, base, "turn-keep", []MemoryEntry{newer}, plantTempsNone)

	reopened := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: base})
	got, ok := reopened.Load("turn-keep", TierContextual)
	if !ok {
		t.Fatal("existing dest disappeared")
	}
	if got.Content.Full != old.Content.Full {
		t.Fatalf("dest overwritten without temps: %q", got.Content.Full)
	}
	assertNoPendingOrTemps(t, base)
}

func TestIngestTurn_TransactionalIngest_TornTempSHAMismatch(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: base})
	old := MemoryEntry{
		ID:      "turn-sha",
		Type:    "turn",
		Tier:    TierContextual,
		Version: 1,
		Content: MemoryContent{Full: "committed before torn temp"},
	}
	if err := store.Write(old); err != nil {
		t.Fatal(err)
	}

	newer := old
	newer.Content.Full = "new payload whose temp is garbage"
	plantPendingIngest(t, base, "turn-sha", []MemoryEntry{newer}, plantTempsCorrupt)

	reopened := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: base})
	got, ok := reopened.Load("turn-sha", TierContextual)
	if !ok {
		t.Fatal("existing dest disappeared")
	}
	if got.Content.Full != old.Content.Full {
		t.Fatalf("dest replaced from torn temp: %q", got.Content.Full)
	}
	assertNoPendingOrTemps(t, base)
}

type plantTempsMode int

const (
	plantTempsIntact plantTempsMode = iota
	plantTempsNone
	plantTempsCorrupt
)

func plantPendingIngest(t *testing.T, base, turnID string, entries []MemoryEntry, mode plantTempsMode) {
	t.Helper()
	rec := pendingIngestRecord{TurnID: sanitizeIngestID(turnID)}
	for _, e := range entries {
		data, err := json.MarshalIndent(e, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		dir := (&PalaceStore{BaseDir: base}).getTierDir(e.Tier)
		if err := palaceMkdirAll(dir); err != nil {
			t.Fatal(err)
		}
		destAbs := filepath.Join(dir, e.ID+".json")
		rel, err := filepath.Rel(base, destAbs)
		if err != nil {
			t.Fatal(err)
		}
		rec.Files = append(rec.Files, pendingIngestFile{
			Dest:   filepath.ToSlash(rel),
			SHA256: sha256Hex(data),
		})
		tmp := ingestSidecarTempPath(destAbs, turnID)
		switch mode {
		case plantTempsIntact:
			if err := palaceWriteFileNamed(tmp, data); err != nil {
				t.Fatal(err)
			}
		case plantTempsCorrupt:
			if err := palaceWriteFileNamed(tmp, []byte("torn-not-json-{")); err != nil {
				t.Fatal(err)
			}
		}
	}
	ps := &PalaceStore{BaseDir: base}
	if err := ps.writePendingRecord(rec); err != nil {
		t.Fatal(err)
	}
}

func assertNoPendingOrTemps(t *testing.T, base string) {
	t.Helper()
	pendingDir := filepath.Join(base, filepath.FromSlash(walPendingRelDir))
	ents, err := os.ReadDir(pendingDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, e := range ents {
		name := e.Name()
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(name, ".json") || strings.Contains(name, ".tmp") {
			t.Fatalf("leftover wal/pending file %s", name)
		}
	}
	tiers := []string{
		"tier-1-working",
		"tier-2-contextual",
		"tier-3-archival",
		"tier-4-semantic",
	}
	for _, tier := range tiers {
		files, err := os.ReadDir(filepath.Join(base, tier))
		if err != nil {
			continue
		}
		for _, f := range files {
			if strings.Contains(f.Name(), ".tmp-") {
				t.Fatalf("leftover ingest temp %s/%s", tier, f.Name())
			}
		}
	}
}
