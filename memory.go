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
	"sort"
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

// MemoryStats provides observability metrics
type MemoryStats struct {
	WorkingCount    int
	ContextualCount int
	ArchivalCount   int
	TotalEntries    int
	LastCompaction  time.Time
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
			return fmt.Errorf("ensureDirs failed for %s: %w", d, err)
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

// Write persists a MemoryEntry (atomic write) + versioning + error wrapping
func (ps *PalaceStore) Write(entry MemoryEntry) error {
	if err := ps.ensureDirs(); err != nil {
		return fmt.Errorf("ensure dirs failed: %w", err)
	}

	// Activate versioning: archive current version before write
	if entry.Version == 0 {
		entry.Version = 1
	}
	if err := ps.archiveToVersions(entry); err != nil {
		fmt.Printf("[memory] versioning warning: %v\n", err)
	}

	dir := ps.getTierDir(entry.Tier)
	filename := fmt.Sprintf("%s.json", entry.ID)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	tmpFile, err := os.CreateTemp(dir, ".tmp-"+entry.ID+"-*.json")
	if err != nil {
		return fmt.Errorf("create temp failed: %w", err)
	}
	tmpName := tmpFile.Name()
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp failed: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp failed: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename failed: %w", err)
	}
	return nil
}

// archiveToVersions archives the entry as a versioned snapshot
func (ps *PalaceStore) archiveToVersions(entry MemoryEntry) error {
	versionsDir := filepath.Join(ps.BaseDir, "versions", "memory-entries", entry.ID)
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		return fmt.Errorf("mkdir versions failed: %w", err)
	}
	versionPath := filepath.Join(versionsDir, fmt.Sprintf("v%d.json", entry.Version))
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal version failed: %w", err)
	}
	if err := os.WriteFile(versionPath, data, 0644); err != nil {
		return fmt.Errorf("write version failed: %w", err)
	}
	return nil
}

// Load retrieves by ID and tier with better error logging
func (ps *PalaceStore) Load(id string, tier MemoryTier) (MemoryEntry, bool) {
	dir := ps.getTierDir(tier)
	path := filepath.Join(dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("[memory] load read failed for %s: %v\n", path, err)
		return MemoryEntry{}, false
	}
	var entry MemoryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		fmt.Printf("[memory] load unmarshal failed for %s: %v\n", path, err)
		return MemoryEntry{}, false
	}
	return entry, true
}

// GetStats returns observability metrics (Phase 1.2)
func (ps *PalaceStore) GetStats() MemoryStats {
	stats := MemoryStats{}

	for _, tier := range []MemoryTier{TierWorking, TierContextual, TierArchival} {
		dir := ps.getTierDir(tier)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		count := len(files)
		switch tier {
		case TierWorking:
			stats.WorkingCount = count
		case TierContextual:
			stats.ContextualCount = count
		case TierArchival:
			stats.ArchivalCount = count
		}
		stats.TotalEntries += count
	}

	return stats
}

// GenerateMemoryID uses cuid2
func GenerateMemoryID() string {
	return cuid2.Generate()
}

// GenerateSimpleEmbedding (deterministic, to be replaced by semantic later)
func GenerateSimpleEmbedding(text string, dim int) []float32 {
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

// PopulateTemporalTags (cycle-aware)
func PopulateTemporalTags(cycle int) []string {
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

// CosineSimilarity helper
func CosineSimilarity(a, b []float32) float64 {
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

// CalculateRecencyBoost returns recency boost factor
func CalculateRecencyBoost(deltaHours float64) float64 {
	switch {
	case deltaHours < 1:
		return 1.0
	case deltaHours < 6:
		return 0.9
	case deltaHours < 24:
		return 0.75
	case deltaHours < 72:
		return 0.55
	case deltaHours < 168:
		return 0.4
	default:
		return 0.2
	}
}

// CalculateTemporalDecay implements H-Mem style forgetting curve
// Exponential decay based on time since last access (hours)
func CalculateTemporalDecay(entry MemoryEntry) float64 {
	if entry.LastAccessed.IsZero() {
		return 1.0
	}
	hoursSince := time.Since(entry.LastAccessed).Hours()
	return math.Exp(-hoursSince / 168) // decay over ~1 week
}

// CalculateRelevanceScore combines score impact, recency, temporal decay, usage (H-Mem s + t + r)
func CalculateRelevanceScore(entry MemoryEntry) float64 {
	if entry.Metrics.ScoreImpact <= 0 {
		return 0.0
	}
	recency := CalculateRecencyBoost(time.Since(entry.LastAccessed).Hours())
	decay := CalculateTemporalDecay(entry)
	usageBoost := 1.0 + 0.1*float64(entry.Metrics.UsageCount+entry.AccessCount)
	total := entry.Metrics.ScoreImpact * recency * decay * usageBoost
	if total > 1.0 {
		total = 1.0
	}
	return total
}

// MultiFactorScore implements full H-Mem s + t + r scoring with optional query vector for semantic component
// s = semantic cosine if queryVec provided, else uses ScoreImpact
func MultiFactorScore(entry MemoryEntry, queryVec []float32) float64 {
	var semantic float64
	if len(queryVec) > 0 {
		entryVec := GenerateSimpleEmbedding(entry.Content.Summary+" "+entry.Content.Full, len(queryVec))
		semantic = CosineSimilarity(entryVec, queryVec)
	} else {
		semantic = entry.Metrics.ScoreImpact
	}
	temporal := CalculateTemporalDecay(entry)
	robust := 1.0 + 0.05*float64(entry.Metrics.UsageCount)
	return semantic * 0.4 + temporal*0.3 + robust*0.3
}

// listEntriesInTier sorts by relevance score (new)
func (ps *PalaceStore) listEntriesInTier(tier MemoryTier) []MemoryEntry {
	dir := ps.getTierDir(tier)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var entries []MemoryEntry
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			id := strings.TrimSuffix(f.Name(), ".json")
			if entry, ok := ps.Load(id, tier); ok {
				entries = append(entries, entry)
			}
		}
	}

	// Sort by relevance score descending (highest first)
	sort.Slice(entries, func(i, j int) bool {
		return CalculateRelevanceScore(entries[i]) > CalculateRelevanceScore(entries[j])
	})

	return entries
}
