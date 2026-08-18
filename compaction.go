package memory

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// CompactionStrategy types (from ossa, kept for compatibility)
type CompactionStrategy string

const (
	StrategySimpleSummary     CompactionStrategy = "simple_summary"
	StrategyPatternExtraction CompactionStrategy = "pattern_extraction"
	StrategyCorePrinciple     CompactionStrategy = "core_principle"
)

// CompactionConfig (extended with RecMem phase-transition parameters)
type CompactionConfig struct {
	Tier2Strategy      CompactionStrategy `json:"tier2_strategy"`
	Tier3Strategy      CompactionStrategy `json:"tier3_strategy"`
	LastEvaluatedCycle int                `json:"last_evaluated_cycle"`
	AvgScoreBefore     float64            `json:"avg_score_before"`
	AvgScoreAfter      float64            `json:"avg_score_after"`
	Improvement        float64            `json:"improvement"`
	// H-Mem inspired
	TemporalWindowSize  int     `json:"temporal_window_size"` // beta-like (e.g. cycles or months)
	SimilarityThreshold float64 `json:"similarity_threshold"` // alpha
	// RecMem phase-transition (Phase 1)
	DataSim   float64 `json:"data_sim"`   // geometric similarity radius (default 0.7)
	DataCount int     `json:"data_count"` // critical recurrence count (default 5)
	// LongMemEval production hardening
	ProtectHighScoreFacts bool    `json:"protect_high_score_facts"` // default true
	FactScoreThreshold    float64 `json:"fact_score_threshold"`     // default 0.90
}

var DefaultCompactionConfig = CompactionConfig{
	Tier2Strategy:         StrategyPatternExtraction,
	Tier3Strategy:         StrategyCorePrinciple,
	TemporalWindowSize:    12,
	SimilarityThreshold:   0.75,
	DataSim:               0.7,
	DataCount:             5,
	ProtectHighScoreFacts: true,
	FactScoreThreshold:    0.90,
}

// VectorStoreCallback allows optional vector integration (e.g. Qdrant)
type VectorStoreCallback func(id string, vec []float32, payload map[string]interface{}) error

// isProtectedFactEntry returns true for high-value fact-augmented entries
// that should be preferentially kept during compaction (LongMemEval production rule).
func isProtectedFactEntry(entry MemoryEntry, cfg CompactionConfig) bool {
	if !cfg.ProtectHighScoreFacts {
		return false
	}
	if entry.Tier == TierSemantic {
		return true
	}
	if (entry.Type == "turn_fact" || entry.Type == "atomic_fact") &&
		entry.Metrics.ScoreImpact >= cfg.FactScoreThreshold {
		return true
	}
	return false
}

// PerformCompaction runs agent-managed compaction with H-Mem temporal window + alpha constraints.
// Now respects LongMemEval fact protection and turn granularity.
// Product Write errors from SUMMARIZE / MERGE / CREATE_CORE_PRINCIPLE / ARCHIVE are returned.
func (ps *PalaceStore) PerformCompaction(
	targetTier MemoryTier,
	cfg CompactionConfig,
	generateFn func(prompt string) string,
	vectorCallback VectorStoreCallback,
) error {
	entries := ps.ListEntriesInTier(targetTier)
	if len(entries) == 0 {
		return nil
	}

	// Separate protected facts from candidates
	var protected []MemoryEntry
	var candidates []MemoryEntry

	for _, e := range entries {
		if isProtectedFactEntry(e, cfg) {
			protected = append(protected, e)
		} else {
			candidates = append(candidates, e)
		}
	}

	// H-Mem style: filter candidates to recent temporal window
	windowed := filterByTemporalWindow(candidates, cfg.TemporalWindowSize)

	// Alpha constraint
	if len(windowed) > 0 && cfg.SimilarityThreshold > 0 {
		avgVec := averageEmbedding(windowed)
		var filtered []MemoryEntry
		for _, e := range windowed {
			entryVec := GenerateSimpleEmbedding(e.Content.Summary+" "+e.Content.Full, len(avgVec))
			sim := CosineSimilarity(entryVec, avgVec)
			if sim >= cfg.SimilarityThreshold {
				filtered = append(filtered, e)
			}
		}
		windowed = filtered
	}

	sort.Slice(windowed, func(i, j int) bool {
		if windowed[i].Metrics.ScoreImpact != windowed[j].Metrics.ScoreImpact {
			return windowed[i].Metrics.ScoreImpact > windowed[j].Metrics.ScoreImpact
		}
		return windowed[i].UpdatedAt.After(windowed[j].UpdatedAt)
	})

	if len(windowed) > 10 {
		windowed = windowed[:10]
	}

	var memList strings.Builder
	for _, e := range windowed {
		memList.WriteString(fmt.Sprintf("ID:%s Type:%s Summary:%s Score:%.1f\n", e.ID, e.Type, truncate(e.Content.Summary, 200), e.Metrics.ScoreImpact))
	}

	prompt := fmt.Sprintf("Compact tier %s within temporal window. Candidates:\n%s\nProtected facts kept: %d\nOutput JSON array of actions...", getTierName(targetTier), memList.String(), len(protected))

	raw := generateFn(prompt)
	actions := parseCompactionActions(raw)

	for _, act := range actions {
		act.Action = strings.ToUpper(strings.TrimSpace(act.Action))
		if !ps.verifyAction(act, targetTier) {
			continue
		}
		var err error
		switch act.Action {
		case "SUMMARIZE":
			err = ps.handleSummarize(act.TargetIDs, targetTier, cfg, vectorCallback)
		case "CREATE_CORE_PRINCIPLE":
			err = ps.handleCreateCorePrinciple(act.TargetIDs, targetTier, cfg, vectorCallback)
		case "ARCHIVE":
			err = ps.handleArchive(act.TargetIDs, targetTier, cfg)
		case "MERGE":
			err = ps.handleMerge(act.TargetIDs, targetTier, cfg, vectorCallback)
		}
		if err != nil {
			ps.lastCompaction = time.Now()
			return err
		}
	}
	ps.lastCompaction = time.Now()
	return nil
}

// averageEmbedding for alpha check
func averageEmbedding(entries []MemoryEntry) []float32 {
	if len(entries) == 0 {
		return []float32{0}
	}
	dim := 768
	avg := make([]float32, dim)
	count := 0
	for _, e := range entries {
		vec := GenerateSimpleEmbedding(e.Content.Summary+" "+e.Content.Full, dim)
		for i := range vec {
			avg[i] += vec[i]
		}
		count++
	}
	for i := range avg {
		avg[i] /= float32(count)
	}
	return avg
}

// verifyAction rejects unknown actions and missing / blank / unknown target IDs.
func (ps *PalaceStore) verifyAction(act CompactionAction, tier MemoryTier) bool {
	minIDs := 1
	switch strings.ToUpper(strings.TrimSpace(act.Action)) {
	case "SUMMARIZE", "CREATE_CORE_PRINCIPLE", "ARCHIVE":
	case "MERGE":
		minIDs = 2
	default:
		return false
	}
	if len(act.TargetIDs) < minIDs {
		return false
	}
	for _, id := range act.TargetIDs {
		if strings.TrimSpace(id) == "" {
			return false
		}
		if _, found := ps.Load(id, tier); !found {
			return false
		}
	}
	return true
}

func filterByTemporalWindow(entries []MemoryEntry, windowSize int) []MemoryEntry {
	if windowSize <= 0 {
		return entries
	}
	if len(entries) == 0 {
		return entries
	}
	maxCycle := 0
	for _, e := range entries {
		if e.Cycle > maxCycle {
			maxCycle = e.Cycle
		}
	}
	cutoff := maxCycle - windowSize
	var result []MemoryEntry
	for _, e := range entries {
		if e.Cycle >= cutoff {
			result = append(result, e)
		}
	}
	return result
}

func getTierName(tier MemoryTier) string {
	switch tier {
	case TierWorking:
		return "working"
	case TierContextual:
		return "contextual"
	case TierArchival:
		return "archival"
	}
	return "contextual"
}

type CompactionAction struct {
	Action    string
	TargetIDs []string
	Reason    string
}

// parseCompactionActions now supports JSON fallback for structured output
func parseCompactionActions(output string) []CompactionAction {
	var jsonActions []struct {
		Action string   `json:"action"`
		Target []string `json:"target"`
		Reason string   `json:"reason"`
	}
	if err := json.Unmarshal([]byte(output), &jsonActions); err == nil {
		var actions []CompactionAction
		for _, ja := range jsonActions {
			actions = append(actions, CompactionAction{
				Action:    ja.Action,
				TargetIDs: ja.Target,
				Reason:    ja.Reason,
			})
		}
		return actions
	}

	var actions []CompactionAction
	lines := strings.Split(output, "\n")
	var current *CompactionAction
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		upper := strings.ToUpper(line)
		if strings.HasPrefix(upper, "ACTION:") {
			if current != nil {
				actions = append(actions, *current)
			}
			current = &CompactionAction{}
			act := strings.TrimSpace(line[7:])
			current.Action = strings.ToUpper(strings.TrimSpace(act))
		} else if current != nil && strings.HasPrefix(upper, "TARGET:") {
			tgts := strings.TrimSpace(line[7:])
			ids := strings.Split(tgts, ",")
			for _, id := range ids {
				id = strings.TrimSpace(id)
				if id != "" {
					current.TargetIDs = append(current.TargetIDs, id)
				}
			}
		} else if current != nil && strings.HasPrefix(upper, "REASON:") {
			current.Reason = strings.TrimSpace(line[7:])
		}
	}
	if current != nil {
		actions = append(actions, *current)
	}
	return actions
}

func (ps *PalaceStore) handleSummarize(ids []string, tier MemoryTier, cfg CompactionConfig, vectorCb VectorStoreCallback) error {
	if len(ids) == 0 {
		return nil
	}
	var contents []string
	var parents []string
	var firstParent MemoryEntry
	var totalScore float64

	for _, id := range ids {
		if entry, ok := ps.Load(id, tier); ok {
			if isProtectedFactEntry(entry, cfg) {
				continue // never summarize protected facts
			}
			if len(parents) == 0 {
				firstParent = entry
			}
			contents = append(contents, entry.Content.Full)
			parents = append(parents, id)
			totalScore += entry.Metrics.ScoreImpact
		}
	}
	if len(contents) == 0 {
		return nil
	}

	combined := strings.Join(contents, "\n\n---\n\n")
	condensed := truncate(combined, 500)

	now := time.Now().UTC()
	newID := GenerateMemoryID()
	newEntry := MemoryEntry{
		ID:        newID,
		Type:      "summary",
		Tier:      TierContextual,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		Content: MemoryContent{
			Summary: truncate(condensed, 140),
			Full:    condensed,
			Tags:    []string{"compacted", "summary"},
		},
		Provenance: MemoryProvenance{
			SourceStep: "compaction",
			ParentIDs:  parents,
		},
		Metrics: MemoryMetrics{
			ScoreImpact: totalScore / float64(len(parents)),
			UsageCount:  1,
		},
	}
	applyParentSessionAndValidFrom(&newEntry, firstParent, now)
	if err := ps.Write(newEntry); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	if vectorCb != nil {
		vec := GenerateSimpleEmbedding(condensed, 768)
		payload := map[string]interface{}{
			"type":      "summary",
			"compacted": true,
		}
		_ = vectorCb(newID, vec, payload)
	}

	for _, id := range ids {
		if entry, ok := ps.Load(id, tier); ok {
			if !isProtectedFactEntry(entry, cfg) {
				from := entry.Tier
				if from == 0 {
					from = tier
				}
				entry.Tier = TierArchival
				if err := ps.writeLeavingTier(entry, from); err != nil {
					return fmt.Errorf("failed to archive summarized entry: %w", err)
				}
			}
		}
	}
	return nil
}

func (ps *PalaceStore) handleCreateCorePrinciple(ids []string, tier MemoryTier, cfg CompactionConfig, vectorCb VectorStoreCallback) error {
	if len(ids) == 0 {
		return nil
	}
	var contents []string
	var parents []string
	var firstParent MemoryEntry
	var totalScore float64

	for _, id := range ids {
		if entry, ok := ps.Load(id, tier); ok {
			if isProtectedFactEntry(entry, cfg) {
				continue
			}
			if len(parents) == 0 {
				firstParent = entry
			}
			contents = append(contents, entry.Content.Full)
			parents = append(parents, id)
			totalScore += entry.Metrics.ScoreImpact
		}
	}
	if len(contents) == 0 {
		return nil
	}

	combined := strings.Join(contents, "\n\n")
	principle := truncate(combined, 400)

	now := time.Now().UTC()
	newID := GenerateMemoryID()
	newEntry := MemoryEntry{
		ID:        newID,
		Type:      "core_principle",
		Tier:      TierArchival,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		Content: MemoryContent{
			Summary: truncate(principle, 140),
			Full:    principle,
			Tags:    []string{"core", "compacted"},
		},
		Provenance: MemoryProvenance{SourceStep: "compaction", ParentIDs: parents},
		Metrics:    MemoryMetrics{ScoreImpact: totalScore/float64(len(parents)) + 1.0, UsageCount: 1},
	}
	applyParentSessionAndValidFrom(&newEntry, firstParent, now)
	if err := ps.Write(newEntry); err != nil {
		return fmt.Errorf("failed to write core principle: %w", err)
	}

	if vectorCb != nil {
		vec := GenerateSimpleEmbedding(principle, 768)
		payload := map[string]interface{}{
			"type":      "core_principle",
			"compacted": true,
		}
		_ = vectorCb(newID, vec, payload)
	}

	for _, id := range ids {
		if entry, ok := ps.Load(id, tier); ok {
			if !isProtectedFactEntry(entry, cfg) {
				from := entry.Tier
				if from == 0 {
					from = tier
				}
				entry.Tier = TierArchival
				if err := ps.writeLeavingTier(entry, from); err != nil {
					return fmt.Errorf("failed to archive core-principle source: %w", err)
				}
			}
		}
	}
	return nil
}

func (ps *PalaceStore) handleArchive(ids []string, tier MemoryTier, cfg CompactionConfig) error {
	for _, id := range ids {
		if entry, ok := ps.Load(id, tier); ok {
			if isProtectedFactEntry(entry, cfg) {
				continue // never archive protected facts
			}
			from := entry.Tier
			if from == 0 {
				from = tier
			}
			entry.Tier = TierArchival
			if err := ps.writeLeavingTier(entry, from); err != nil {
				return fmt.Errorf("failed to archive entry: %w", err)
			}
		}
	}
	return nil
}

func (ps *PalaceStore) handleMerge(ids []string, tier MemoryTier, cfg CompactionConfig, vectorCb VectorStoreCallback) error {
	if len(ids) < 2 {
		return nil
	}
	var contents []string
	var parents []string
	var firstParent MemoryEntry
	var totalScore float64

	for _, id := range ids {
		if entry, ok := ps.Load(id, tier); ok {
			if isProtectedFactEntry(entry, cfg) {
				continue // do not merge protected facts
			}
			if len(parents) == 0 {
				firstParent = entry
			}
			contents = append(contents, entry.Content.Full)
			parents = append(parents, id)
			totalScore += entry.Metrics.ScoreImpact
		}
	}
	if len(contents) == 0 {
		return nil
	}

	combined := strings.Join(contents, "\n\n---\n\n")
	merged := truncate(combined, 500)

	now := time.Now().UTC()
	newID := GenerateMemoryID()
	newEntry := MemoryEntry{
		ID:         newID,
		Type:       "merged",
		Tier:       TierContextual,
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
		Content:    MemoryContent{Summary: truncate(merged, 140), Full: merged, Tags: []string{"merged", "compacted"}},
		Provenance: MemoryProvenance{SourceStep: "compaction", ParentIDs: parents},
		Metrics:    MemoryMetrics{ScoreImpact: totalScore / float64(len(parents)), UsageCount: 1},
	}
	applyParentSessionAndValidFrom(&newEntry, firstParent, now)
	if err := ps.Write(newEntry); err != nil {
		return fmt.Errorf("failed to write merge: %w", err)
	}

	if vectorCb != nil {
		vec := GenerateSimpleEmbedding(merged, 768)
		payload := map[string]interface{}{
			"type": "merged",
		}
		_ = vectorCb(newID, vec, payload)
	}

	for _, id := range ids {
		if entry, ok := ps.Load(id, tier); ok {
			if !isProtectedFactEntry(entry, cfg) {
				from := entry.Tier
				if from == 0 {
					from = tier
				}
				entry.Tier = TierArchival
				if err := ps.writeLeavingTier(entry, from); err != nil {
					return fmt.Errorf("failed to archive merged entry: %w", err)
				}
			}
		}
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// clusterBySimilarity groups entries whose content embeddings are within the configured similarity threshold.
func clusterBySimilarity(entries []MemoryEntry, threshold float64) [][]MemoryEntry {
	if len(entries) == 0 {
		return nil
	}
	avg := averageEmbedding(entries)
	cluster := make([]MemoryEntry, 0, len(entries))
	for _, e := range entries {
		vec := GenerateSimpleEmbedding(e.Content.Summary+" "+e.Content.Full, 768)
		if CosineSimilarity(vec, avg) >= threshold {
			cluster = append(cluster, e)
		}
	}
	if len(cluster) == 0 {
		return nil
	}
	return [][]MemoryEntry{cluster}
}

// shouldTriggerPhaseTransition returns true when a cluster has both sufficient count and density.
func shouldTriggerPhaseTransition(entries []MemoryEntry, cfg CompactionConfig) bool {
	if len(entries) < cfg.DataCount {
		return false
	}
	avg := averageEmbedding(entries)
	for _, e := range entries {
		vec := GenerateSimpleEmbedding(e.Content.Summary+" "+e.Content.Full, 768)
		if CosineSimilarity(vec, avg) < cfg.DataSim {
			return false
		}
	}
	return true
}

// AutoRecMemCompaction is the production entry point for automatic RecMem formation.
func (ps *PalaceStore) AutoRecMemCompaction(generateFn func(prompt string) string, vectorCb VectorStoreCallback) {
	cfg := ps.Config.CompactionConfig
	sub := ps.listSubconsciousEntries()
	if len(sub) == 0 {
		return
	}
	clusters := clusterBySimilarity(sub, cfg.DataSim)
	for _, cluster := range clusters {
		if shouldTriggerPhaseTransition(cluster, cfg) {
			if err := ps.PerformCompaction(TierContextual, cfg, generateFn, vectorCb); err != nil {
				return
			}
			if err := ps.SemanticRefine(cluster); err != nil {
				return
			}
		}
	}
}
