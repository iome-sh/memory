package memory

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// MultiHopOptions configures associative / multi-hop lite retrieval over the
// EntityGraph plus entry entity tags (A2 first slice / s619; hop ranking s1067).
//
// Honesty: competitive multi-hop lite — BFS over GetRelatedEntities and entry
// collect via TemporalTags / Content.Tags / RelatedConcepts. Not a full Zep /
// Graphiti knowledge graph. Hop-distance ranking is path-aware ranking lite
// (prefer shorter BFS hops), not typed-edge / embedding-guided path scoring.
type MultiHopOptions struct {
	// SeedEntity is the starting entity key (normalized). Prefer exact graph node keys.
	SeedEntity string
	// SeedQuery optional: if set, run SearchMemoryWithOptions with Query then derive
	// entity seeds from top hits' TemporalTags entity:* and RelatedConcepts
	// (combined with SeedEntity when both are set).
	SeedQuery string
	// MaxHops default 2; clamped to 1..4. Hop 0 is the seed itself.
	MaxHops int
	// Limit default 20; applied AFTER expansion + entry collect + filters.
	Limit int
	// SessionID, when non-empty, keeps only matching SessionID (before Limit).
	SessionID string
	// AsOf optional EntryValidAt filter (before Limit).
	AsOf *time.Time
	// Tier when non-nil: only that tier. When nil: Working+Contextual+Semantic
	// (exclude Archival unless IncludeArchival).
	Tier *MemoryTier
	// IncludeArchival, when true and Tier==nil, also includes Archival.
	IncludeArchival bool
	// QueryVec optional pass-through when seeding via SeedQuery search.
	QueryVec []float32
	// PreferShorterHops ranks by minimum BFS hop distance from seed (lower first),
	// then event time descending within the same hop. Default true (nil or true).
	// Set to a false pointer to opt out and use legacy seed-match-first sort.
	PreferShorterHops *bool
}

const (
	subjectTagPrefix = "subject:"
)

// preferShorterHops reports whether hop-distance ranking is enabled (default true).
func (opts MultiHopOptions) preferShorterHops() bool {
	if opts.PreferShorterHops == nil {
		return true
	}
	return *opts.PreferShorterHops
}

// EntryEntityKeys extracts entity-like keys from an entry for multi-hop matching.
// Sources:
//   - TemporalTags with "entity:" prefix → value after prefix (e.g. entity:person:alice → person:alice)
//   - TemporalTags with "subject:" prefix → full tag (subject:auth)
//   - Content.Tags with entity:/subject: prefixes (same rules)
//   - Relations.RelatedConcepts (trimmed non-empty strings as-is)
//
// Duplicates are removed; order is stable (first-seen).
func EntryEntityKeys(e MemoryEntry) []string {
	seen := make(map[string]struct{})
	var keys []string
	add := func(k string) {
		k = strings.TrimSpace(k)
		if k == "" {
			return
		}
		if _, ok := seen[k]; ok {
			return
		}
		seen[k] = struct{}{}
		keys = append(keys, k)
	}

	collectFromTags := func(tags []string) {
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			if strings.HasPrefix(t, entityTagPrefix) {
				// entity:type:id → type:id (and keep full form without double-prefix)
				rest := strings.TrimPrefix(t, entityTagPrefix)
				if rest != "" {
					add(rest)
				}
				// Also retain the full "entity:..." form so graph seeds written with
				// the tag prefix still match.
				add(t)
				continue
			}
			if strings.HasPrefix(t, subjectTagPrefix) {
				add(t)
			}
		}
	}

	collectFromTags(e.TemporalTags)
	collectFromTags(e.Content.Tags)
	for _, c := range e.Relations.RelatedConcepts {
		add(c)
	}
	return keys
}

// entryEntityKeySet is a set view of EntryEntityKeys for O(1) membership.
func entryEntityKeySet(e MemoryEntry) map[string]struct{} {
	keys := EntryEntityKeys(e)
	set := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		set[k] = struct{}{}
	}
	return set
}

// entryMatchesExpandedEntity reports whether e is associated with entity via
// TemporalTags entity:*, Content.Tags, or RelatedConcepts (exact key match after
// EntryEntityKeys normalization, plus exact tag / concept equality).
func entryMatchesExpandedEntity(e MemoryEntry, entity string) bool {
	if entity == "" {
		return false
	}
	set := entryEntityKeySet(e)
	if _, ok := set[entity]; ok {
		return true
	}
	// Also accept entity:entity when callers seed with full tag form.
	if strings.HasPrefix(entity, entityTagPrefix) {
		if _, ok := set[strings.TrimPrefix(entity, entityTagPrefix)]; ok {
			return true
		}
	} else if _, ok := set[entityTagPrefix+entity]; ok {
		return true
	}
	// Content.Tags may store bare entity strings without entity: prefix.
	for _, t := range e.Content.Tags {
		if t == entity || t == entityTagPrefix+entity {
			return true
		}
	}
	for _, c := range e.Relations.RelatedConcepts {
		if c == entity || c == entityTagPrefix+entity {
			return true
		}
	}
	for _, t := range e.TemporalTags {
		if t == entity || t == entityTagPrefix+entity {
			return true
		}
		if strings.HasPrefix(t, entityTagPrefix) && strings.TrimPrefix(t, entityTagPrefix) == entity {
			return true
		}
	}
	return false
}

// clampMaxHops clamps hop budget to 1..4 (ExpandRelatedEntities contract).
func clampMaxHops(maxHops int) int {
	if maxHops < 1 {
		return 1
	}
	if maxHops > 4 {
		return 4
	}
	return maxHops
}

// ExpandRelatedEntitiesHops BFS from seed over GetRelatedEntities up to maxHops
// (includes seed at hop 0). maxHops is clamped to 1..4.
//
// Returns entity key → minimum hop distance from seed. Empty seed → nil.
func (ps *PalaceStore) ExpandRelatedEntitiesHops(seed string, maxHops int) map[string]int {
	seed = strings.TrimSpace(seed)
	if seed == "" {
		return nil
	}
	maxHops = clampMaxHops(maxHops)

	type node struct {
		entity string
		hop    int
	}
	hops := map[string]int{seed: 0}
	queue := []node{{entity: seed, hop: 0}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.hop >= maxHops {
			continue
		}
		for _, rel := range ps.GetRelatedEntities(cur.entity) {
			rel = strings.TrimSpace(rel)
			if rel == "" {
				continue
			}
			nextHop := cur.hop + 1
			if prev, ok := hops[rel]; ok && prev <= nextHop {
				continue
			}
			hops[rel] = nextHop
			queue = append(queue, node{entity: rel, hop: nextHop})
		}
	}
	return hops
}

// ExpandRelatedEntities BFS from seed over GetRelatedEntities up to maxHops
// (includes seed at hop 0). maxHops is clamped to 1..4.
//
// Returns unique entity keys in BFS discovery order (seed first).
func (ps *PalaceStore) ExpandRelatedEntities(seed string, maxHops int) []string {
	seed = strings.TrimSpace(seed)
	if seed == "" {
		return nil
	}
	maxHops = clampMaxHops(maxHops)

	type node struct {
		entity string
		hop    int
	}
	seen := map[string]struct{}{seed: {}}
	out := []string{seed}
	queue := []node{{entity: seed, hop: 0}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.hop >= maxHops {
			continue
		}
		for _, rel := range ps.GetRelatedEntities(cur.entity) {
			rel = strings.TrimSpace(rel)
			if rel == "" {
				continue
			}
			if _, ok := seen[rel]; ok {
				continue
			}
			seen[rel] = struct{}{}
			out = append(out, rel)
			queue = append(queue, node{entity: rel, hop: cur.hop + 1})
		}
	}
	return out
}

// entryMinHop among entityHops keys that match e; returns -1 if none match.
func entryMinHop(e MemoryEntry, entityHops map[string]int) int {
	minHop := -1
	for ent, hop := range entityHops {
		if !entryMatchesExpandedEntity(e, ent) {
			continue
		}
		if minHop < 0 || hop < minHop {
			minHop = hop
		}
	}
	return minHop
}

// MultiHopRetrieve performs multi-hop lite associative retrieval:
//  1. Resolve seeds (SeedEntity and/or entities from SeedQuery search hits)
//  2. ExpandRelatedEntitiesHops for each seed (min hop across seeds)
//  3. Collect entries from default tiers matching any expanded entity
//  4. Optional AsOf + SessionID filters BEFORE Limit
//  5. Sort: lower min hop first (default), then event time desc within hop
//     (legacy: seed-match first when PreferShorterHops is explicitly false)
//  6. Limit
//
// Default MaxHops=2 (clamped 1..4), Limit=20. Default tiers: Working +
// Contextual + Semantic (+ Archival if IncludeArchival).
// PreferShorterHops defaults true (path-aware ranking lite / s1067).
func (ps *PalaceStore) MultiHopRetrieve(opts MultiHopOptions) []MemoryEntry {
	maxHops := opts.MaxHops
	if maxHops <= 0 {
		maxHops = 2
	}
	if maxHops < 1 {
		maxHops = 1
	}
	if maxHops > 4 {
		maxHops = 4
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}

	// 1) Resolve seeds
	seedSet := make(map[string]struct{})
	var seeds []string
	addSeed := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seedSet[s]; ok {
			return
		}
		seedSet[s] = struct{}{}
		seeds = append(seeds, s)
	}

	addSeed(opts.SeedEntity)

	if opts.SeedQuery != "" {
		// Derive entity seeds from top search hits. Use a modest hit cap so
		// seed extraction stays cheap; session/as-of/query-vec pass through.
		searchLimit := limit
		if searchLimit < 10 {
			searchLimit = 10
		}
		if searchLimit > 50 {
			searchLimit = 50
		}
		searchOpts := SearchMemoryOptions{
			SessionID: opts.SessionID,
			AsOf:      opts.AsOf,
			Limit:     searchLimit,
			Tier:      opts.Tier,
			QueryVec:  opts.QueryVec,
		}
		hits := ps.SearchMemoryWithOptions(opts.SeedQuery, searchOpts)
		for _, h := range hits {
			for _, k := range EntryEntityKeys(h) {
				addSeed(k)
			}
		}
	}

	if len(seeds) == 0 {
		return nil
	}

	// 2) Expand related entities for each seed; keep min hop across seeds
	entityHops := make(map[string]int)
	for _, s := range seeds {
		for ent, hop := range ps.ExpandRelatedEntitiesHops(s, maxHops) {
			if prev, ok := entityHops[ent]; !ok || hop < prev {
				entityHops[ent] = hop
			}
		}
	}
	if len(entityHops) == 0 {
		return nil
	}

	// 3) Collect candidate entries from tiers
	var results []MemoryEntry
	if opts.Tier != nil {
		results = ps.ListEntriesInTier(*opts.Tier)
	} else {
		tiers := []MemoryTier{TierWorking, TierContextual, TierSemantic}
		if opts.IncludeArchival {
			tiers = []MemoryTier{TierWorking, TierContextual, TierArchival, TierSemantic}
		}
		for _, t := range tiers {
			results = append(results, ps.ListEntriesInTier(t)...)
		}
	}

	// Match any expanded entity; record min hop among matched entity keys
	minHopByIdx := make([]int, 0, len(results))
	{
		var filtered []MemoryEntry
		for _, e := range results {
			hop := entryMinHop(e, entityHops)
			if hop < 0 {
				continue
			}
			filtered = append(filtered, e)
			minHopByIdx = append(minHopByIdx, hop)
		}
		results = filtered
	}

	// 4) Session + AsOf filters before Limit (keep hop slice in sync)
	if opts.SessionID != "" {
		var filtered []MemoryEntry
		var hops []int
		for i, e := range results {
			if e.SessionID == opts.SessionID {
				filtered = append(filtered, e)
				hops = append(hops, minHopByIdx[i])
			}
		}
		results = filtered
		minHopByIdx = hops
	}
	if opts.AsOf != nil {
		asOf := *opts.AsOf
		var filtered []MemoryEntry
		var hops []int
		for i, e := range results {
			if EntryValidAt(e, asOf) {
				filtered = append(filtered, e)
				hops = append(hops, minHopByIdx[i])
			}
		}
		results = filtered
		minHopByIdx = hops
	}

	// 5) Sort: hop-distance ranking lite (default) or legacy seed-match first.
	// Sort indices so hop slice stays aligned with results.
	useHopRank := opts.preferShorterHops()
	seedMatches := func(e MemoryEntry) bool {
		for s := range seedSet {
			if entryMatchesExpandedEntity(e, s) {
				return true
			}
		}
		return false
	}
	order := make([]int, len(results))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		i, j := order[a], order[b]
		if useHopRank {
			hi, hj := minHopByIdx[i], minHopByIdx[j]
			if hi != hj {
				return hi < hj // shorter hop first (hop 0 = seed)
			}
		} else {
			si := seedMatches(results[i])
			sj := seedMatches(results[j])
			if si != sj {
				return si // seed matches first
			}
		}
		ti := entryEventTime(results[i])
		tj := entryEventTime(results[j])
		if !ti.Equal(tj) {
			return ti.After(tj)
		}
		return results[i].ID < results[j].ID
	})
	sorted := make([]MemoryEntry, len(results))
	for k, i := range order {
		sorted[k] = results[i]
	}
	results = sorted

	// 6) Limit
	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

// ensureRelationsDir creates BaseDir/relations if missing (safe for graph writes).
func (ps *PalaceStore) ensureRelationsDir() error {
	return os.MkdirAll(filepath.Join(ps.BaseDir, "relations"), 0755)
}
