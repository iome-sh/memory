package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPalaceStore_FreshPalaceModeBits(t *testing.T) {
	base := filepath.Join(t.TempDir(), "palace")
	store := NewPalaceStore(base)

	if err := store.Write(MemoryEntry{
		ID:      "mode-entry",
		Tier:    TierContextual,
		Version: 1,
		Content: MemoryContent{Summary: "fresh palace mode bits"},
	}); err != nil {
		t.Fatal(err)
	}
	store.AddEntityRelationship("person:alice", "org:acme")
	_ = store.ListMemoryWithOptions(ListMemoryOptions{Limit: 10})

	assertPalaceModeBits(t, base)

	entryPath := filepath.Join(base, "tier-2-contextual", "mode-entry.json")
	versionPath := filepath.Join(base, "versions", "memory-entries", "mode-entry", "v1.json")
	graphPath := filepath.Join(base, "relations", "entity-graph.json")
	indexPath := filepath.Join(base, "indexes", "event-time.json")
	for _, p := range []string{entryPath, versionPath, graphPath, indexPath} {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("expected %s: %v", p, err)
		}
		if got := info.Mode().Perm(); got != palaceFileMode {
			t.Errorf("%s mode %04o, want %04o", p, got, palaceFileMode)
		}
	}
}

func assertPalaceModeBits(t *testing.T, root string) {
	t.Helper()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		perm := info.Mode().Perm()
		if info.IsDir() {
			if perm != palaceDirMode {
				t.Errorf("dir %s mode %04o, want %04o", path, perm, palaceDirMode)
			}
			return nil
		}
		if perm != palaceFileMode {
			t.Errorf("file %s mode %04o, want %04o", path, perm, palaceFileMode)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
