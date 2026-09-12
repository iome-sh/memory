// Package writerprobe is last-write-wins evidence for two OS processes on one
// palace root. Multi-process writers remain unsupported. Not a lock, not flock,
// not tenancy, not Memory GA.
package writerprobe

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iome-sh/memory"
)

const (
	DefaultSharedID = "probe-shared"
	DefaultN        = 24
	HubEntity       = "probe-hub"
)

// UniqueID is the per-writer entry id that must not collide across processes.
func UniqueID(writer string, i int) string {
	return fmt.Sprintf("probe-%s-%d", writer, i)
}

// Spoke is the per-writer entity-graph edge from HubEntity.
func Spoke(writer string, i int) string {
	return fmt.Sprintf("spoke-%s-%d", writer, i)
}

func WriterTag(writer string) string {
	return "probe-writer:" + writer
}

// RunWriter writes n unique ids, n times over sharedID, and n hub spokes.
// PersistEmbeddings stays off (hash default). Errors are returned; callers
// that are probes still exit 0 at the script layer.
func RunWriter(baseDir, writer, sharedID string, n int) error {
	baseDir = strings.TrimSpace(baseDir)
	writer = strings.TrimSpace(writer)
	sharedID = strings.TrimSpace(sharedID)
	if baseDir == "" {
		return fmt.Errorf("base dir required")
	}
	if writer == "" {
		return fmt.Errorf("writer id required")
	}
	if sharedID == "" {
		sharedID = DefaultSharedID
	}
	if n <= 0 {
		n = DefaultN
	}

	store := memory.NewPalaceStoreWithConfig(memory.PalaceConfig{BaseDir: baseDir})
	var first error
	note := func(err error) {
		if err != nil && first == nil {
			first = err
		}
	}
	tag := WriterTag(writer)
	// List once so the durable event-time snapshot is created; later Writes
	// patch+persist it. That is the shared-file last-write-wins path.
	_ = store.ListMemoryWithOptions(memory.ListMemoryOptions{Limit: 1})
	for i := 0; i < n; i++ {
		note(store.Write(memory.MemoryEntry{
			ID:        UniqueID(writer, i),
			Tier:      memory.TierContextual,
			SessionID: "probe-session-" + writer,
			Content: memory.MemoryContent{
				Summary: fmt.Sprintf("unique writer=%s seq=%d", writer, i),
				Tags:    []string{tag, "probe-unique"},
			},
		}))
		note(store.Write(memory.MemoryEntry{
			ID:        sharedID,
			Tier:      memory.TierContextual,
			SessionID: "probe-shared-session",
			Content: memory.MemoryContent{
				Summary: fmt.Sprintf("shared last-write-wins writer=%s seq=%d", writer, i),
				Tags:    []string{tag, "probe-shared"},
			},
		}))
		store.AddEntityRelationship(HubEntity, Spoke(writer, i))
		store.AddEntityRelationship("entity-"+writer, fmt.Sprintf("rel-%s-%d", writer, i))
	}
	_ = store.ListMemoryWithOptions(memory.ListMemoryOptions{Limit: 1})
	return first
}

// Report is last-write-wins / JSON-validity evidence. Lost graph edges or
// event-time rows are expected under unsupported multi-process writers.
type Report struct {
	BaseDir  string
	SharedID string
	N        int

	SharedExists    bool
	SharedValidJSON bool
	SharedWinner    string
	SharedSummary   string

	UniqueByWriter map[string]int

	GraphExists    bool
	GraphValidJSON bool
	GraphHubSpokes int
	GraphSpokes    map[string]int

	EventTimeExists    bool
	EventTimeValidJSON bool
	EventTimeJSONCount int
	EventTimeEntries   int
	TierJSONFiles      int
}

type inspectEntry struct {
	ID      string `json:"id"`
	Content struct {
		Summary string   `json:"summary"`
		Tags    []string `json:"tags"`
	} `json:"content"`
}

type inspectEventTime struct {
	Version   int               `json:"version"`
	JSONCount int               `json:"json_count"`
	Entries   []json.RawMessage `json:"entries"`
}

// Inspect reads the palace after concurrent writers. Missing files and lost
// updates are evidence, not a product failure.
func Inspect(baseDir, sharedID string, n int, writers []string) Report {
	baseDir = strings.TrimSpace(baseDir)
	sharedID = strings.TrimSpace(sharedID)
	if sharedID == "" {
		sharedID = DefaultSharedID
	}
	if n <= 0 {
		n = DefaultN
	}
	if len(writers) == 0 {
		writers = []string{"A", "B"}
	}

	r := Report{
		BaseDir:        baseDir,
		SharedID:       sharedID,
		N:              n,
		UniqueByWriter: make(map[string]int, len(writers)),
		GraphSpokes:    make(map[string]int, len(writers)),
	}

	sharedPath := filepath.Join(baseDir, "tier-2-contextual", sharedID+".json")
	if raw, err := os.ReadFile(sharedPath); err == nil {
		r.SharedExists = true
		var e inspectEntry
		if json.Unmarshal(raw, &e) == nil {
			r.SharedValidJSON = true
			r.SharedSummary = e.Content.Summary
			r.SharedWinner = winnerFromEntry(e, writers)
		}
	}

	for _, w := range writers {
		got := 0
		for i := 0; i < n; i++ {
			p := filepath.Join(baseDir, "tier-2-contextual", UniqueID(w, i)+".json")
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				got++
			}
		}
		r.UniqueByWriter[w] = got
	}

	graphPath := filepath.Join(baseDir, "relations", "entity-graph.json")
	if raw, err := os.ReadFile(graphPath); err == nil {
		r.GraphExists = true
		var graph map[string][]string
		if json.Unmarshal(raw, &graph) == nil {
			r.GraphValidJSON = true
			hub := graph[HubEntity]
			r.GraphHubSpokes = len(hub)
			for _, w := range writers {
				prefix := "spoke-" + w + "-"
				nSpoke := 0
				for _, s := range hub {
					if strings.HasPrefix(s, prefix) {
						nSpoke++
					}
				}
				r.GraphSpokes[w] = nSpoke
			}
		}
	}

	idxPath := filepath.Join(baseDir, "indexes", "event-time.json")
	if raw, err := os.ReadFile(idxPath); err == nil {
		r.EventTimeExists = true
		var snap inspectEventTime
		if json.Unmarshal(raw, &snap) == nil {
			r.EventTimeValidJSON = true
			r.EventTimeJSONCount = snap.JSONCount
			r.EventTimeEntries = len(snap.Entries)
		}
	}
	r.TierJSONFiles = countTierJSON(baseDir)
	return r
}

func winnerFromEntry(e inspectEntry, writers []string) string {
	for _, w := range writers {
		tag := WriterTag(w)
		for _, t := range e.Content.Tags {
			if t == tag {
				return w
			}
		}
	}
	sum := e.Content.Summary
	for _, w := range writers {
		if strings.Contains(sum, "writer="+w) {
			return w
		}
	}
	return "unknown"
}

func countTierJSON(baseDir string) int {
	tiers := []string{
		"tier-1-working",
		"tier-2-contextual",
		"tier-3-archival",
		"tier-4-semantic",
	}
	n := 0
	for _, t := range tiers {
		ents, err := os.ReadDir(filepath.Join(baseDir, t))
		if err != nil {
			continue
		}
		for _, e := range ents {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".json") || strings.HasPrefix(name, ".tmp-") {
				continue
			}
			n++
		}
	}
	return n
}

// WriteReport prints last-write-wins / JSON validity. Does not dump entry Full.
func WriteReport(w io.Writer, r Report) {
	fmt.Fprintln(w, "two-process-writer-probe: last-write-wins evidence (not a lock; flock is not shipped)")
	fmt.Fprintf(w, "palace: %s\n", r.BaseDir)
	fmt.Fprintf(w, "shared id: %s\n", r.SharedID)
	fmt.Fprintf(w, "  exists=%t valid_json=%t winner=%s\n", r.SharedExists, r.SharedValidJSON, r.SharedWinner)
	if r.SharedSummary != "" {
		fmt.Fprintf(w, "  summary: %s\n", r.SharedSummary)
	}
	fmt.Fprint(w, "unique files:")
	for _, wtr := range writerOrder(r.UniqueByWriter) {
		fmt.Fprintf(w, " %s=%d", wtr, r.UniqueByWriter[wtr])
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "entity-graph.json: exists=%t valid_json=%t hub_spokes=%d", r.GraphExists, r.GraphValidJSON, r.GraphHubSpokes)
	for _, wtr := range writerOrder(r.GraphSpokes) {
		fmt.Fprintf(w, " spokes_%s=%d", wtr, r.GraphSpokes[wtr])
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "event-time.json: exists=%t valid_json=%t json_count=%d entries=%d tier_json_files=%d\n",
		r.EventTimeExists, r.EventTimeValidJSON, r.EventTimeJSONCount, r.EventTimeEntries, r.TierJSONFiles)
	fmt.Fprintln(w, "honesty: multi-process writers remain unsupported · probe ≠ flock · probe ≠ Memory GA · dual_write OFF · this does not invent tenancy")
}

func writerOrder(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
