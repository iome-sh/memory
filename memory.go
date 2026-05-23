package memory

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/knights-analytics/hugot"
	"github.com/nrednav/cuid2"
)

// MemoryTier defines the three-tier hierarchical memory (inspired by ossa Palace + H-Mem ideas)
type MemoryTier int

const (
	TierWorking    MemoryTier = 1
	TierContextual MemoryTier = 2
	TierArchival   MemoryTier = 3
	// TierSemantic is used for high-fidelity atomic facts protected by RecMem Phase 3
	TierSemantic MemoryTier = 4
)

// MemoryEntry is the core unit stored in the Palace.
// Extended for LongMemEval production readiness: explicit turn/session granularity + fact-augmented indexing.
type MemoryEntry struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"`
	Tier         MemoryTier `json:"tier"`
	Version      int        `json:"version"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Cycle        int        `json:"cycle"`
	TemporalTags []string   `json:"temporal_tags,omitempty"`
	AccessCount  int        `json:"access_count"`
	LastAccessed time.Time  `json:"last_accessed,omitempty"`

	// === LongMemEval / turn-level granularity extensions (production implementation) ===
	TurnID         string    `json:"turn_id,omitempty"`         // explicit round/turn identifier
	SessionID      string    `json:"session_id,omitempty"`      // multi-session grouping
	Timestamp      time.Time `json:"timestamp,omitempty"`       // precise event time for temporal reasoning
	ExtractedFacts []string  `json:"extracted_facts,omitempty"` // fact-augmented for better recall
	Keyphrases     []string  `json:"keyphrases,omitempty"`      // keyphrase expansion for indexing
	OriginalText   string    `json:"original_text,omitempty"`   // raw turn text for provenance

	Content    MemoryContent    `json:"content"`
	Provenance MemoryProvenance `json:"provenance"`
	Metrics    MemoryMetrics    `json:"metrics"`
	Relations  MemoryRelations  `json:"relations"`
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
	SemanticCount   int // Phase 3
	TotalEntries    int
	LastCompaction  time.Time
}

// PalaceConfig for configurable PalaceStore creation (Phase 4.3)
type PalaceConfig struct {
	BaseDir            string
	MaxWorkingEntries  int
	MaxWorkingAgeHours int
	CompactionConfig   CompactionConfig
	EmbeddingFunc      EmbeddingFunc `json:"-"` // pluggable (Phase 5.1)
}

// EmbeddingFunc is injectable for semantic embeddings (Phase 5.1)
type EmbeddingFunc func(text string, dim int) []float32

// PalaceStore provides file-backed hierarchical memory storage.
type PalaceStore struct {
	BaseDir string
	Config  PalaceConfig
}

// NewPalaceStoreWithConfig creates PalaceStore with full configuration (Phase 4.3)
func NewPalaceStoreWithConfig(cfg PalaceConfig) *PalaceStore {
	if cfg.BaseDir == "" {
		cfg.BaseDir = ".ossa/kb/palace"
	}
	if cfg.MaxWorkingEntries == 0 {
		cfg.MaxWorkingEntries = 50
	}
	if cfg.MaxWorkingAgeHours == 0 {
		cfg.MaxWorkingAgeHours = 48
	}
	if cfg.EmbeddingFunc == nil {
		cfg.EmbeddingFunc = GenerateSimpleEmbedding
	}
	ps := &PalaceStore{
		BaseDir: cfg.BaseDir,
		Config:  cfg,
	}
	_ = ps.ensureDirs()
	return ps
}

// NewPalaceStore is legacy convenience (uses defaults)
func NewPalaceStore(baseDir string) *PalaceStore {
	cfg := PalaceConfig{BaseDir: baseDir}
	return NewPalaceStoreWithConfig(cfg)
}

func (ps *PalaceStore) ensureDirs() error {
	dirs := []string{
		ps.BaseDir,
		filepath.Join(ps.BaseDir, "tier-0-subconscious"), // RecMem latent buffer (Phase 1)
		filepath.Join(ps.BaseDir, "tier-1-working"),
		filepath.Join(ps.BaseDir, "tier-2-contextual"),
		filepath.Join(ps.BaseDir, "tier-3-archival"),
		filepath.Join(ps.BaseDir, "tier-4-semantic"), // RecMem Phase 3 - high fidelity atomic facts
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
	case TierSemantic:
		return filepath.Join(ps.BaseDir, "tier-4-semantic")
	}
	return filepath.Join(ps.BaseDir, "tier-2-contextual")
}

// listSubconsciousEntries returns all entries currently stored in the RecMem subconscious (latent) buffer.
// Used by AutoRecMemCompaction in Phase 2.
func (ps *PalaceStore) listSubconsciousEntries() []MemoryEntry {
	dir := filepath.Join(ps.BaseDir, "tier-0-subconscious")
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var entries []MemoryEntry
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			id := strings.TrimSuffix(f.Name(), ".json")
			fullPath := filepath.Join(dir, id+".json")
			if entry, ok := ps.loadEntry(fullPath); ok {
				entries = append(entries, entry)
			}
		}
	}
	return entries
}

// loadEntry is an internal helper to load a MemoryEntry from a full filesystem path.
func (ps *PalaceStore) loadEntry(fullPath string) (MemoryEntry, bool) {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return MemoryEntry{}, false
	}
	var entry MemoryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return MemoryEntry{}, false
	}
	return entry, true
}

// WriteLatent stores entry in the subconscious latent buffer (RecMem Phase 1)
// No immediate promotion or LLM compaction triggered.
func (ps *PalaceStore) WriteLatent(entry MemoryEntry) error {
	if err := ps.ensureDirs(); err != nil {
		return fmt.Errorf("ensure dirs failed: %w", err)
	}

	if entry.Version == 0 {
		entry.Version = 1
	}
	if err := ps.archiveToVersions(entry); err != nil {
		// Non-fatal
	}

	dir := filepath.Join(ps.BaseDir, "tier-0-subconscious")
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
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return fmt.Errorf("write temp failed: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return fmt.Errorf("close temp failed: %w", err)
	}
	if err := os.Rename(tmpFile.Name(), path); err != nil {
		os.Remove(tmpFile.Name())
		return fmt.Errorf("rename failed: %w", err)
	}
	return nil
}

// Write persists a MemoryEntry (atomic write) + versioning + error wrapping + lifecycle
func (ps *PalaceStore) Write(entry MemoryEntry) error {
	if err := ps.ensureDirs(); err != nil {
		return fmt.Errorf("ensure dirs failed: %w", err)
	}

	// Activate versioning: archive current version before write
	if entry.Version == 0 {
		entry.Version = 1
	}
	if err := ps.archiveToVersions(entry); err != nil {
		// Non-fatal in production; continue (versioning is best-effort)
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
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return fmt.Errorf("write temp failed: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return fmt.Errorf("close temp failed: %w", err)
	}
	if err := os.Rename(tmpFile.Name(), path); err != nil {
		os.Remove(tmpFile.Name())
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
		return fmt.Errorf("marshal failed: %w", err)
	}
	if err = os.WriteFile(versionPath, data, 0644); err != nil {
		return fmt.Errorf("write version failed: %w", err)
	}
	return nil
}

// Load retrieves by ID and tier. Production: no side-effect prints, clean bool return.
func (ps *PalaceStore) Load(id string, tier MemoryTier) (MemoryEntry, bool) {
	dir := ps.getTierDir(tier)
	path := filepath.Join(dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return MemoryEntry{}, false
	}
	var entry MemoryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return MemoryEntry{}, false
	}
	return entry, true
}

// GetStats returns observability metrics (Phase 1.2)
func (ps *PalaceStore) GetStats() MemoryStats {
	stats := MemoryStats{}

	for _, tier := range []MemoryTier{TierWorking, TierContextual, TierArchival, TierSemantic} {
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
		case TierSemantic:
			stats.SemanticCount = count
		}
		stats.TotalEntries += count
	}

	return stats
}

// SearchMemory provides hybrid retrieval (keyword + vector + temporal) - Phase 4.1
// Supports query, optional tier filter, limit, and vector if VectorStore attached.
// Now uses the configured EmbeddingFunc from PalaceConfig for vector re-ranking (enables pure-Go ONNX swap).
func (ps *PalaceStore) SearchMemory(query string, tier *MemoryTier, limit int, vec []float32) []MemoryEntry {
	if limit <= 0 {
		limit = 10
	}
	var results []MemoryEntry

	// Start with all entries (or tier filtered)
	if tier != nil {
		results = ps.ListEntriesInTier(*tier)
	} else {
		for _, t := range []MemoryTier{TierWorking, TierContextual, TierArchival, TierSemantic} {
			results = append(results, ps.ListEntriesInTier(t)...)
		}
	}

	// Improved keyword filter: token-based (any word overlap)
	queryLower := strings.ToLower(query)
	queryWords := strings.FieldsFunc(queryLower, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	}

	var filtered []MemoryEntry
	for _, e := range results {
		contentLower := strings.ToLower(e.Content.Summary + " " + e.Content.Full)
		match := false
		for _, w := range queryWords {
			if len(w) < 3 {
				continue // skip very short words
			}
			if strings.Contains(contentLower, w) {
				match = true
				break
			}
		}
		if match {
			filtered = append(filtered, e)
		}
	}
	results = filtered

	// Vector re-rank if provided -- uses configured EmbeddingFunc (real ONNX when wired)
	if len(vec) > 0 {
		embedFn := ps.Config.EmbeddingFunc
		if embedFn == nil {
			embedFn = GenerateSimpleEmbedding
		}
		sort.Slice(results, func(i, j int) bool {
			iVec := embedFn(results[i].Content.Summary+" "+results[i].Content.Full, len(vec))
			jVec := embedFn(results[j].Content.Summary+" "+results[j].Content.Full, len(vec))
			return CosineSimilarity(iVec, vec) > CosineSimilarity(jVec, vec)
		})
	}

	// Limit
	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// EvictWorkingTier performs age/size-based eviction from Working tier (Phase 4.2)
func (ps *PalaceStore) EvictWorkingTier(maxAgeHours int, maxCount int) {
	working := ps.ListEntriesInTier(TierWorking)
	if len(working) <= maxCount {
		return
	}

	// Sort by age (oldest first)
	sort.Slice(working, func(i, j int) bool {
		return working[i].CreatedAt.Before(working[j].CreatedAt)
	}

	for i := 0; i < len(working)-maxCount; i++ {
		if time.Since(working[i].CreatedAt).Hours() > float64(maxAgeHours) {
			working[i].Tier = TierContextual
			ps.Write(working[i])
		}
	}
}

// PromoteToContextual promotes high-relevance Working entries (Phase 4.2)
func (ps *PalaceStore) PromoteToContextual(threshold float64) {
	working := ps.ListEntriesInTier(TierWorking)
	for _, e := range working {
		if CalculateRelevanceScore(e) > threshold {
			e.Tier = TierContextual
			ps.Write(e)
		}
	}
}

// EntityGraph for H-Mem KG integration (richer relational graph)
type EntityGraph struct {
	Entities map[string][]string `json:"entities"`
}

// AddEntityRelationship adds normalized entity links (H-Mem style)
func (ps *PalaceStore) AddEntityRelationship(entity, related string) {
	graphPath := filepath.Join(ps.BaseDir, "relations", "entity-graph.json")
	graph := make(map[string][]string)
	if data, err := os.ReadFile(graphPath); err == nil {
		json.Unmarshal(data, &graph)
	}
	if graph[entity] == nil {
		graph[entity] = []string{}
	}
	for _, r := range graph[entity] {
		if r == related {
			return
		}
	}
	graph[entity] = append(graph[entity], related)
	data, _ := json.MarshalIndent(graph, "", "  ")
	os.WriteFile(graphPath, data, 0644)
}

// GetRelatedEntities returns related entities for graph traversal
func (ps *PalaceStore) GetRelatedEntities(entity string) []string {
	graphPath := filepath.Join(ps.BaseDir, "relations", "entity-graph.json")
	graph := make(map[string][]string)
	if data, err := os.ReadFile(graphPath); err == nil {
		json.Unmarshal(data, &graph)
	}
	return graph[entity]

}

// GenerateMemoryID uses cuid2
func GenerateMemoryID() string {
	return cuid2.Generate()
}

// NewGONNXEmbeddingFunc returns a production-grade EmbeddingFunc powered by hugot.
// Fixed to use the correct session-based API for hugot v0.3+.
func NewGONNXEmbeddingFunc(modelPath string) (EmbeddingFunc, error) {
	if modelPath == "" {
		return GenerateSimpleEmbedding, nil
	}

	if _, err := os.Stat(modelPath); err != nil {
		return nil, fmt.Errorf("ONNX model not found at %s: %w", modelPath, err)
	}

	// Create hugot session
	session, err := hugot.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create hugot session: %w", err)
	}

	config := hugot.FeatureExtractionConfig{
		ModelPath: modelPath,
	}

	// Use the session-based constructor
	pipeline, err := hugot.NewPipeline(session, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create hugot pipeline: %w", err)
	}

	return func(text string, dim int) []float32 {
		if text == "" {
			return make([]float32, dim)
		}

		result, err := pipeline.Run([]string{text})
		if err != nil || result == nil {
			return GenerateSimpleEmbedding(text, dim)
		}

		// result.Embeddings is the standard field in hugot FeatureExtraction output
		if len(result.Embeddings) == 0 {
			return GenerateSimpleEmbedding(text, dim)
		}

		embedding := result.Embeddings[0]

		// L2 normalize
		var norm float32
		for _, v := range embedding {
			norm += v * v
		}
		if norm > 0 {
			norm = float32(math.Sqrt(float64(norm)))
			for i := range embedding {
				embedding[i] /= norm
			}
		}

		return embedding
	}, nil
}

// GenerateSimpleEmbedding (deterministic hash-based fallback)
func GenerateSimpleEmbedding(text string, dim int) []float32 {
	if dim <= 0 {
		dim = 768
	}
	vec := make([]float32, dim)
	if text == "" {
		return vec
	}
	h := sha256.Sum256([]byte(strings.ToLower(text)))
	seed := binary.BigEndian.Uint64(h[:8])
	r := rand.New(rand.NewPCG(seed, seed>>32))
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
func CalculateTemporalDecay(entry MemoryEntry) float64 {
	if entry.LastAccessed.IsZero() {
		return 1.0
	}
	hoursSince := time.Since(entry.LastAccessed).Hours()
	return math.Exp(-hoursSince / 168)
}

// CalculateRelevanceScore combines score impact, recency, temporal decay, usage
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

// MultiFactorScore implements full H-Mem scoring
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
	return semantic*0.4 + temporal*0.3 + robust*0.3
}

// ListEntriesInTier returns all entries in the given tier, sorted by relevance score (descending).
func (ps *PalaceStore) ListEntriesInTier(tier MemoryTier) []MemoryEntry {
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

	sort.Slice(entries, func(i, j int) bool {
		return CalculateRelevanceScore(entries[i]) > CalculateRelevanceScore(entries[j])
	})

	return entries
}

// ExtractAtomicFacts extracts high-value personal facts from a memory entry.
func ExtractAtomicFacts(entry MemoryEntry) []string {
	text := entry.Content.Full
	if text == "" {
		text = entry.Content.Summary
	}
	if text == "" {
		return nil
	}

	var facts []string
	sentences := strings.Split(text, ". ")

	factPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(graduated|degree|studied|university|college|bachelor|master|phd|major in)`),
		regexp.MustCompile(`(?i)(my name (is|was)|last name|changed my name|used to be called)`),
		regexp.MustCompile(`(?i)(live in|moved to|from .*? (city|town|state|country)|grew up in)`),
		regexp.MustCompile(`(?i)(favorite|love|hate|prefer|always .*? (eat|drink|listen|watch|read|wear))`),
		regexp.MustCompile(`(?i)(bought|got a new|own|just purchased|added to my collection)`),
		regexp.MustCompile(`(?i)(spent .*? (on|for)|paid .*? dollars|cost me)`),
		regexp.MustCompile(`(?i)(\d+\s*(hours?|days?|weeks?|months?|years?|dollars?|bucks?|items?|shirts?|bikes?|plants?))`),
		regexp.MustCompile(`(?i)(on .*? (birthday|anniversary|trip|vacation|wedding)|last (month|week|year)|this (month|year))`),
		regexp.MustCompile(`(?i)(work at|job at|occupation|previous job|used to work)`),
	}

	for _, s := range sentences {
		s := strings.TrimSpace(s)
		if len(s) < 8 {
			continue
		}

		matched := false
		for _, re := range factPatterns {
			if re.MatchString(s) {
				facts = append(facts, s)
				matched = true
				break
			}
		}

		if !matched {
			lower := strings.ToLower(s)
			if (strings.Contains(lower, "i ") || strings.ContainsAny(s, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")) &&
				len(strings.Fields(s)) >= 5 {
				facts = append(facts, s)
			}
		}
	}

	seen := make(map[string]bool)
	unique := make([]string, 0, len(facts))
	for _, f := range facts {
		if !seen[f] {
			seen[f] = true
			unique = append(unique, f)
		}
	}
	return unique
}

// extractKeyphrases is a lightweight rule-based extractor for keyphrases.
func extractKeyphrases(text string) []string {
	if text == "" {
		return nil
	}
	words := strings.Fields(text)
	var phrases []string
	seen := make(map[string]bool)
	for i := 0; i < len(words)-1; i++ {
		w1 := strings.Trim(words[i], ".,;:!?\"")
		w2 := strings.Trim(words[i+1], ".,;:!?\"")
		if len(w1) > 3 && unicode.IsUpper(rune(w1[0])) && len(w2) > 2 {
			phrase := w1 + " " + w2
			if !seen[phrase] {
			seen[phrase] = true
			phrases = append(phrases, phrase)
		}
	}
	for _, w := range words {
		clean := strings.Trim(w, ".,;:!?\"")
		if len(clean) > 4 && unicode.IsUpper(rune(clean[0])) && !seen[clean] {
		seen[clean] = true
		phrases = append(phrases, clean)
	}
	if len(phrases) > 12 {
		phrases = phrases[:12]
	}
	return phrases
}

// IngestTurn is the primary production entry point for LongMemEval benchmark and ego online session processing.
func (ps *PalaceStore) IngestTurn(turn MemoryEntry) error {
	if err := ps.ensureDirs(); err != nil {
		return fmt.Errorf("ensure dirs failed: %w", err)
	}

	if turn.Timestamp.IsZero() {
		turn.Timestamp = time.Now()
	}
	if turn.TurnID == "" {
		turn.TurnID = GenerateMemoryID()
	}
	if turn.ID == "" {
		turn.ID = GenerateMemoryID()
	}

	if len(turn.ExtractedFacts) == 0 {
		turn.ExtractedFacts = ExtractAtomicFacts(turn)
	}

	if len(turn.Keyphrases) == 0 {
		combined := turn.Content.Full
		if combined == "" {
			combined = turn.Content.Summary
		}
		if combined == "" {
			combined = turn.OriginalText
		}
		turn.Keyphrases = extractKeyphrases(combined)
	}

	if turn.OriginalText == "" {
		turn.OriginalText = turn.Content.Full
		if turn.OriginalText == "" {
			turn.OriginalText = turn.Content.Summary
		}
	}

	if turn.Type == "" {
		turn.Type = "turn"
	}

	if err := ps.Write(turn); err != nil {
		return fmt.Errorf("failed to write turn entry: %w", err)
	}

	for _, factText := range turn.ExtractedFacts {
		if strings.TrimSpace(factText) == "" {
			continue
		}
		now := time.Now()
		factID := GenerateMemoryID()

		factEntry := MemoryEntry{
			ID:        factID,
			Type:      "turn_fact",
			Tier:      TierSemantic,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
			Timestamp: turn.Timestamp,
			TurnID:    turn.TurnID,
			SessionID: turn.SessionID,
			Content: MemoryContent{
				Summary: truncate(factText, 280),
				Full:    factText,
				Tags:    []string{"fact_augmented", "from_turn", "longmemeval"},
			},
			Provenance: MemoryProvenance{
				SourceStep: "ingest_turn_fact",
				ParentIDs:  []string{turn.ID},
			},
			Metrics: MemoryMetrics{
				ScoreImpact: 0.92,
				UsageCount:  1,
			},
		}

		_ = ps.Write(factEntry)
	}

	return nil
}

// SemanticRefine (RecMem Phase 3) protects high-stake atomic facts from clusters.
func (ps *PalaceStore) SemanticRefine(cluster []MemoryEntry) error {
	if len(cluster) == 0 {
		return nil
	}

	for _, entry := range cluster {
		facts := ExtractAtomicFacts(entry)
		for _, factText := range facts {
			now := time.Now()
			factID := GenerateMemoryID()

			factEntry := MemoryEntry{
				ID:        factID,
				Type:      "atomic_fact",
				Tier:      TierSemantic,
				Version:   1,
				CreatedAt: now,
				UpdatedAt: now,
				Content: MemoryContent{
					Summary: truncate(factText, 200),
					Full:    factText,
					Tags:    []string{"semantic", "protected", "atomic_fact"},
			},
			Provenance: MemoryProvenance{
				SourceStep: "semantic_refine",
				ParentIDs:  []string{entry.ID},
			},
			Metrics: MemoryMetrics{
				ScoreImpact: 0.95,
				UsageCount:  1,
			},
		}

		if err := ps.Write(factEntry); err != nil {
			return fmt.Errorf("failed to write semantic fact: %w", err)
		}
	}
	return nil
}

// ReadWithChainOfNote formats retrieved MemoryEntry items into a Chain-of-Note style prompt.
func (ps *PalaceStore) ReadWithChainOfNote(retrieved []MemoryEntry, query string) (string, error) {
	if len(retrieved) == 0 {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("You are a precise, memory-augmented reasoner.\n\n")
	sb.WriteString("Original Query: " + query + "\n\n")
	sb.WriteString("Retrieved Context (use ONLY these items; list them explicitly first):\n\n")

	for i, entry := range retrieved {
		id := entry.TurnID
		if id == "" {
			id = entry.ID
		}
		if id == "" {
			id = fmt.Sprintf("entry-%d", i)
		}

		ts := ""
		if !entry.Timestamp.IsZero() {
			ts = entry.Timestamp.Format(time.RFC3339)
		} else if !entry.CreatedAt.IsZero() {
			ts = entry.CreatedAt.Format(time.RFC3339)
		}

		content := entry.Content.Full
		if content == "" {
			content = entry.Content.Summary
		}
		if len(entry.ExtractedFacts) > 0 {
			content = strings.Join(entry.ExtractedFacts, " | ")
		}

		sb.WriteString(fmt.Sprintf("- [Tier:%d] (id=%s, ts=%s)\n  %s\n\n",
			entry.Tier, id, ts, strings.TrimSpace(content)))
	}

	sb.WriteString("Instructions (Chain-of-Note):\n")
	sb.WriteString("1. First, explicitly list every relevant fact/turn above with its ID and timestamp.\n")
	sb.WriteString("2. Then reason step-by-step ONLY from the listed items. Do not use external knowledge.\n")
	sb.WriteString("3. If the information is insufficient, output exactly: I don't know\n")
	sb.WriteString("4. At the very end, add a confidence score from 0.0 to 1.0.\n\n")
	sb.WriteString("Output strictly as JSON:\n")
	sb.WriteString("{\n  \"reasoning\": \"step-by-step reasoning here\",\n  \"answer\": \"final answer or I don't know\",\n  \"confidence\": 0.85\n}\n")

	return sb.String(), nil
}
