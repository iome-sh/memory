package memory

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nrednav/cuid2"
)

// MemoryTier defines the three-tier hierarchical memory (inspired by ossa Palace + H-Mem ideas)
type MemoryTier int

const (
	TierWorking    MemoryTier = 1
	TierContextual MemoryTier = 2
	TierArchival   MemoryTier = 3
)

// MemoryEntry is the core unit stored in the Palace.
type MemoryEntry struct {
	ID           string           `json:"id"`
	Type         string           `json:"type"`
	Tier         MemoryTier       `json:"tier"`
	Version      int              `json:"version"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	Cycle        int              `json:"cycle"`
	TemporalTags []string         `json:"temporal_tags,omitempty"`
	AccessCount  int              `json:"access_count"`
	LastAccessed time.Time        `json:"last_accessed,omitempty"`
	Content      MemoryContent    `json:"content"`
	Provenance   MemoryProvenance `json:"provenance"`
	Metrics      MemoryMetrics    `json:"metrics"`
	Relations    MemoryRelations  `json:"relations"`
}

// MemoryContent holds summary/full/tags
type MemoryContent struct {
	Summary string   `json:"summary"`
	Full    string   `json:"full,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// MemoryProvenance tracks origin
type MemoryProvenance struct {
	SourceCycle int      `json:"source_cycle"`
	SourceStep  string   `json:"source_step,omitempty"`
	ParentIDs   []string `json:"parent_ids,omitempty"`
	ToolCalls   []string `json:"tool_calls,omitempty"`
}

// MemoryMetrics for scoring and access
type MemoryMetrics struct {
	ScoreImpact  float64   `json:"score_impact,omitempty"`
	UsageCount   int       `json:"usage_count"`
	LastAccessed time.Time `json:"last_accessed,omitempty"`
}

// MemoryRelations for graph links
type MemoryRelations struct {
	ImprovesUpon    []string `json:"improves_upon,omitempty"`
	RelatedConcepts []string `json:"related_concepts,omitempty"`
	Backlinks       []string `json:"backlinks,omitempty"`
}

// PalaceStore provides file-backed hierarchical memory storage.
type PalaceStore struct {
	BaseDir string
}

func NewPalaceStore(baseDir string) *PalaceStore {
	ps := &PalaceStore{BaseDir: baseDir}
	_ = ps.ensureDirs() // best effort
	return ps
}

func (ps *PalaceStore) ensureDirs() error {
	dirs := []string{
		ps.BaseDir,
		filepath.Join(ps.BaseDir, "tier-1-working"),
		filepath.Join(ps.BaseDir, "tier-2-contextual"),
		filepath.Join(ps.BaseDir, "tier-3-archival"),
		filepath.Join(ps.BaseDir, "versions", "memory-entries"),
		filepath.Join(ps.BaseDir, "relations"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (ps *PalaceStore) getTierDir(tier MemoryTier) string {
	switch tier {
	case TierWorking:
		return filepath.Join(ps.BaseDir, "tier-1-working")
	case TierContextual:
		return filepath.Join(ps.BaseDir, "tier-2-contextual")
	case TierArchival:
		return filepath.Join(ps.BaseDir, "tier-3-archival")
	}
	return filepath.Join(ps.BaseDir, "tier-2-contextual")
}

// Write persists a MemoryEntry (atomic write)
func (ps *PalaceStore) Write(entry MemoryEntry) error {
	if err := ps.ensureDirs(); err != nil {
		return err
	}
	dir := ps.getTierDir(entry.Tier)
	filename := fmt.Sprintf("%s.json", entry.ID)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(dir, ".tmp-"+entry.ID+"-*.json")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// Load retrieves by ID and tier
func (ps *PalaceStore) Load(id string, tier MemoryTier) (MemoryEntry, bool) {
	dir := ps.getTierDir(tier)
	path := filepath.Join(dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return MemoryEntry{}, false
	}
	var entry MemoryEntry
	if json.Unmarshal(data, &entry) != nil {
		return MemoryEntry{}, false
	}
	return entry, true
}

// generateMemoryID uses cuid2
func generateMemoryID() string {
	return cuid2.Generate()
}

// generateSimpleEmbedding (deterministic, to be replaced by semantic later)
func generateSimpleEmbedding(text string, dim int) []float32 {
	if dim <= 0 {
		dim = 768
	}
	vec := make([]float32, dim)
	if text == "" {
		return vec
	}
	h := sha256.Sum256([]byte(strings.ToLower(text)))
	seed := int64(binary.BigEndian.Uint64(h[:8]))
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < dim; i++ {
		vec[i] = float32(r.Float64()*2 - 1)
	}
	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := range vec {
			vec[i] /= float32(norm)
		}
	}
	return vec
}

// populateTemporalTags (cycle-aware)
func populateTemporalTags(cycle int) []string {
	now := time.Now()
	weekday := now.Weekday().String()
	hour := now.Hour()
	timeOfDay := "evening"
	if hour < 6 {
		timeOfDay = "night"
	} else if hour < 12 {
		timeOfDay = "morning"
	} else if hour < 18 {
		timeOfDay = "afternoon"
	}
	return []string{
		fmt.Sprintf("cycle-%d", cycle),
		now.Format("2006-01"),
		weekday,
		timeOfDay,
	}
}

// cosineSimilarity helper
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
