package longmemeval

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/iome-sh/memory"
)

func TestLoadV2_Subset(t *testing.T) {
	t.Parallel()
	ds, err := LoadV2(v2SubsetRoot(), "small")
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Questions) != 1 {
		t.Fatalf("questions = %d, want 1", len(ds.Questions))
	}
	if ds.Questions[0].ID != "q-staging-reset" {
		t.Fatalf("id = %q", ds.Questions[0].ID)
	}
	if _, ok := ds.Trajectories["traj-admin-gear"]; !ok {
		t.Fatal("missing trajectory")
	}
	ids := ds.Haystack["q-staging-reset"]
	if len(ids) != 1 || ids[0] != "traj-admin-gear" {
		t.Fatalf("haystack = %#v", ids)
	}
}

func TestPalaceMemory_InsertQueryText(t *testing.T) {
	t.Parallel()
	ds, err := LoadV2(v2SubsetRoot(), "small")
	if err != nil {
		t.Fatal(err)
	}
	store := memory.NewPalaceStoreWithConfig(memory.PalaceConfig{
		BaseDir:       t.TempDir(),
		EmbeddingFunc: memory.GenerateSimpleEmbedding,
	})
	adapter := &PalaceMemory{Store: store, TopK: 5}
	tr := ds.Trajectories["traj-admin-gear"]
	if err := adapter.Insert(tr); err != nil {
		t.Fatal(err)
	}
	items := adapter.Query(ds.Questions[0].Question, "")
	if len(items) == 0 {
		t.Fatal("expected at least one text context item")
	}
	if items[0].Type != "text" {
		t.Fatalf("type = %q", items[0].Type)
	}
	joined := strings.ToLower(items[0].Value)
	if !strings.Contains(joined, "staging-reset") && !strings.Contains(joined, "admin gear") {
		t.Fatalf("query should retrieve staging-reset evidence; got %q", items[0].Value)
	}
}

func v2SubsetRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", "longmemeval_v2_subset")
}
