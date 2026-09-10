package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// dual_write stays OFF in these tests (host policy, not a kernel product flag).
// No mesh bind, no dual-write path.

func TestClassifyIngestSourceHint(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"mcp_memory_ingest_turn", ""},
		{"source:iomesh-memory-mcp", ""},
		{"iomesh-memory-mcp", ""},
		{"private", ingestSourcePrivate},
		{"source_hint:private", ingestSourcePrivate},
		{"source:private", ingestSourcePrivate},
		{"private_overlay", ingestSourcePrivate},
		{"private_rca", ingestSourcePrivate},
		{"palace", ingestSourcePrivate},
		{"palace_timeline", ingestSourcePrivate},
		{"source_hint:palace_timeline", ingestSourcePrivate},
		{"local", ingestSourcePrivate},
		{"local_palace", ingestSourcePrivate},
		{"mesh", ingestSourceMesh},
		{"source_hint:mesh", ingestSourceMesh},
		{"source:mesh", ingestSourceMesh},
		{"mesh_stream", ingestSourceMesh},
		{"source_hint:mesh_pulse", ingestSourceMesh},
		{"catalog", ""},
		{"grant", ""},
	}
	for _, tc := range cases {
		if got := ClassifyIngestSourceHint(tc.in); got != tc.want {
			t.Fatalf("ClassifyIngestSourceHint(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIngestTurn_PrivateSourceHintOnDisk(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:       base,
		EmbeddingFunc: GenerateSimpleEmbedding,
	})
	// Simulate MCP ingest-dir: process label only, no cite-both class.
	if err := store.IngestTurn(MemoryEntry{
		ID:        "turn-private-1",
		SessionID: "local-overlay",
		Content: MemoryContent{
			Full: "RCA: auth token refresh failed on staging.",
			Tags: []string{"source:iomesh-memory-mcp", "role:user"},
		},
		Provenance: MemoryProvenance{SourceStep: "mcp_memory_ingest_turn"},
		ExtractedFacts: []string{
			"Auth token refresh failed on staging",
		},
	}); err != nil {
		t.Fatal(err)
	}

	parent, ok := store.Load("turn-private-1", TierContextual)
	if !ok {
		// default tier is 0 → contextual
		parent, ok = store.Load("turn-private-1", 0)
	}
	if !ok {
		listed := store.ListMemoryWithOptions(ListMemoryOptions{Limit: 20})
		for _, e := range listed {
			if e.ID == "turn-private-1" {
				parent = e
				ok = true
				break
			}
		}
	}
	if !ok {
		t.Fatal("parent turn missing after IngestTurn")
	}
	assertPrivateSourceOnDisk(t, base, parent)
	if parent.Provenance.SourceStep != "mcp_memory_ingest_turn" {
		t.Fatalf("must keep caller SourceStep; got %q", parent.Provenance.SourceStep)
	}

	facts := store.ListEntriesInTier(TierSemantic)
	if len(facts) != 1 {
		t.Fatalf("facts = %d, want 1", len(facts))
	}
	assertPrivateSourceOnDisk(t, base, facts[0])
	if facts[0].Provenance.SourceHint != SourceHintPrivate {
		t.Fatalf("child SourceHint = %q, want %q", facts[0].Provenance.SourceHint, SourceHintPrivate)
	}

	listed := store.ListMemoryWithOptions(ListMemoryOptions{Tag: FormatSourceHintTag(SourceHintPrivate), Limit: 50})
	if len(listed) != 2 {
		t.Fatalf("list tag=source_hint:private got %d, want parent+child 2", len(listed))
	}
	hits := store.SearchMemory("auth token refresh", nil, 10, nil)
	if len(hits) == 0 {
		t.Fatal("retrieve must find private ingest")
	}
	sawPrivate := false
	for _, h := range hits {
		if h.Provenance.SourceHint == SourceHintPrivate || EntryHasTag(h, FormatSourceHintTag(SourceHintPrivate)) {
			sawPrivate = true
		}
	}
	if !sawPrivate {
		t.Fatal("retrieve hits must carry observable private source class")
	}
}

func TestIngestTurn_MeshSourceHintRemainsDistinct(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:       base,
		EmbeddingFunc: GenerateSimpleEmbedding,
	})
	if err := store.IngestTurn(MemoryEntry{
		ID: "turn-mesh-1",
		Content: MemoryContent{
			Full: "Mesh pulse: deploy finished on prod.",
			Tags: []string{"source:mesh", "source_hint:mesh_stream"},
		},
		Provenance:     MemoryProvenance{SourceHint: "mesh_stream"},
		ExtractedFacts: []string{"Deploy finished on prod"},
	}); err != nil {
		t.Fatal(err)
	}

	parent := mustLoadTurn(t, store, "turn-mesh-1")
	if parent.Provenance.SourceHint != "mesh_stream" {
		t.Fatalf("parent SourceHint = %q, want mesh_stream", parent.Provenance.SourceHint)
	}
	if EntryHasTag(parent, FormatSourceHintTag(SourceHintPrivate)) {
		t.Fatalf("mesh ingest must not stamp private; tags=%v", parent.Content.Tags)
	}
	if ClassifyIngestSourceHint(parent.Provenance.SourceHint) != ingestSourceMesh {
		t.Fatalf("parent hint class = %q, want mesh", ClassifyIngestSourceHint(parent.Provenance.SourceHint))
	}

	facts := store.ListEntriesInTier(TierSemantic)
	if len(facts) != 1 {
		t.Fatalf("facts = %d, want 1", len(facts))
	}
	if facts[0].Provenance.SourceHint != "mesh_stream" {
		t.Fatalf("child SourceHint = %q, want mesh_stream", facts[0].Provenance.SourceHint)
	}
	if !EntryHasTag(facts[0], "source:mesh") || !EntryHasTag(facts[0], "source_hint:mesh_stream") {
		t.Fatalf("child missing mesh tags: %v", facts[0].Content.Tags)
	}
	if EntryHasTag(facts[0], FormatSourceHintTag(SourceHintPrivate)) {
		t.Fatalf("child must not gain private; tags=%v", facts[0].Content.Tags)
	}

	raw, err := os.ReadFile(filepath.Join(base, "tier-4-semantic", facts[0].ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var disk MemoryEntry
	if err := json.Unmarshal(raw, &disk); err != nil {
		t.Fatal(err)
	}
	if disk.Provenance.SourceHint != "mesh_stream" {
		t.Fatalf("disk child source_hint = %q", disk.Provenance.SourceHint)
	}
	if EntryHasTag(disk, FormatSourceHintTag(SourceHintPrivate)) {
		t.Fatalf("disk child stamped private; tags=%v", disk.Content.Tags)
	}

	byMesh := store.ListMemoryWithOptions(ListMemoryOptions{TagPrefix: "source_hint:mesh", Limit: 50})
	if len(byMesh) != 2 {
		t.Fatalf("tag_prefix=source_hint:mesh got %d, want 2", len(byMesh))
	}
	byPrivate := store.ListMemoryWithOptions(ListMemoryOptions{Tag: FormatSourceHintTag(SourceHintPrivate), Limit: 50})
	if len(byPrivate) != 0 {
		t.Fatalf("tag=source_hint:private got %d, want 0 on mesh ingest", len(byPrivate))
	}
}

func TestIngestTurn_CallerPrivateHintKept(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{
		BaseDir:       t.TempDir(),
		EmbeddingFunc: GenerateSimpleEmbedding,
	})
	if err := store.IngestTurn(MemoryEntry{
		Content: MemoryContent{
			Full: "Palace overlay note.",
			Tags: []string{"source_hint:private_rca"},
		},
		ExtractedFacts: []string{"Palace overlay note"},
	}); err != nil {
		t.Fatal(err)
	}
	facts := store.ListEntriesInTier(TierSemantic)
	if len(facts) != 1 {
		t.Fatalf("facts = %d, want 1", len(facts))
	}
	if facts[0].Provenance.SourceHint != "private_rca" {
		t.Fatalf("SourceHint = %q, want private_rca", facts[0].Provenance.SourceHint)
	}
	if !EntryHasTag(facts[0], "source_hint:private_rca") {
		t.Fatalf("missing caller private tag: %v", facts[0].Content.Tags)
	}
	// Do not also stamp the default source_hint:private when a private-class hint exists.
	if EntryHasTag(facts[0], FormatSourceHintTag(SourceHintPrivate)) && facts[0].Provenance.SourceHint != SourceHintPrivate {
		t.Fatalf("should not add default private tag beside private_rca; tags=%v", facts[0].Content.Tags)
	}
}

func TestWrite_PersistsCallerSourceHintNoDefaultStamp(t *testing.T) {
	base := t.TempDir()
	store := NewPalaceStore(base)
	entry := MemoryEntry{
		ID:   "write-1",
		Tier: TierContextual,
		Content: MemoryContent{
			Summary: "direct write",
			Tags:    []string{"type:fact"},
		},
	}
	if err := store.Write(entry); err != nil {
		t.Fatal(err)
	}
	loaded, ok := store.Load("write-1", TierContextual)
	if !ok {
		t.Fatal("missing write")
	}
	if loaded.Provenance.SourceHint != "" {
		t.Fatalf("Write must not default-stamp SourceHint; got %q", loaded.Provenance.SourceHint)
	}
	if EntryHasTag(loaded, FormatSourceHintTag(SourceHintPrivate)) {
		t.Fatalf("Write must not default-stamp source_hint:private; tags=%v", loaded.Content.Tags)
	}

	mesh := MemoryEntry{
		ID:   "write-mesh",
		Tier: TierContextual,
		Content: MemoryContent{
			Summary: "mesh fact",
			Tags:    []string{"source_hint:mesh"},
		},
		Provenance: MemoryProvenance{SourceHint: "mesh"},
	}
	if err := store.Write(mesh); err != nil {
		t.Fatal(err)
	}
	loadedMesh, ok := store.Load("write-mesh", TierContextual)
	if !ok {
		t.Fatal("missing mesh write")
	}
	if loadedMesh.Provenance.SourceHint != "mesh" {
		t.Fatalf("Write must persist caller SourceHint; got %q", loadedMesh.Provenance.SourceHint)
	}
}

func assertPrivateSourceOnDisk(t *testing.T, base string, e MemoryEntry) {
	t.Helper()
	if e.Provenance.SourceHint != SourceHintPrivate {
		t.Fatalf("id=%s SourceHint = %q, want %q", e.ID, e.Provenance.SourceHint, SourceHintPrivate)
	}
	if !EntryHasTag(e, FormatSourceHintTag(SourceHintPrivate)) {
		t.Fatalf("id=%s missing tag source_hint:private; tags=%v", e.ID, e.Content.Tags)
	}
	if ClassifyIngestSourceHint(e.Provenance.SourceHint) != ingestSourcePrivate {
		t.Fatalf("id=%s hint %q not private-class", e.ID, e.Provenance.SourceHint)
	}

	dir := ""
	switch e.Tier {
	case TierWorking:
		dir = "tier-1-working"
	case TierContextual, 0:
		dir = "tier-2-contextual"
	case TierArchival:
		dir = "tier-3-archival"
	case TierSemantic:
		dir = "tier-4-semantic"
	default:
		dir = "tier-2-contextual"
	}
	raw, err := os.ReadFile(filepath.Join(base, dir, e.ID+".json"))
	if err != nil {
		// default tier 0 may land in contextual
		raw, err = os.ReadFile(filepath.Join(base, "tier-2-contextual", e.ID+".json"))
		if err != nil {
			t.Fatalf("read disk %s: %v", e.ID, err)
		}
	}
	var disk MemoryEntry
	if err := json.Unmarshal(raw, &disk); err != nil {
		t.Fatal(err)
	}
	if disk.Provenance.SourceHint != SourceHintPrivate {
		t.Fatalf("disk %s source_hint = %q", e.ID, disk.Provenance.SourceHint)
	}
	if !EntryHasTag(disk, FormatSourceHintTag(SourceHintPrivate)) {
		t.Fatalf("disk %s missing source_hint:private; tags=%v", e.ID, disk.Content.Tags)
	}
}

func mustLoadTurn(t *testing.T, store *PalaceStore, id string) MemoryEntry {
	t.Helper()
	for _, tier := range []MemoryTier{TierContextual, 0, TierWorking, TierSemantic} {
		if e, ok := store.Load(id, tier); ok {
			return e
		}
	}
	for _, e := range store.ListMemoryWithOptions(ListMemoryOptions{Limit: 50}) {
		if e.ID == id {
			return e
		}
	}
	t.Fatalf("entry %s not found", id)
	return MemoryEntry{}
}
