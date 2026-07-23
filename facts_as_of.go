package memory

import (
	"sort"
	"strings"
	"time"
)

// Validity tag prefixes written by hosts (e.g. aion applyTemporalToEntry).
// Format: "valid_from:<RFC3339>" / "valid_until:<RFC3339>".
const (
	validFromTagPrefix  = "valid_from:"
	validUntilTagPrefix = "valid_until:"
	entityTagPrefix     = "entity:"
)

// ParseValidityWindow reads valid_from / valid_until from TemporalTags.
// Missing or unparseable bounds are open-ended (nil).
//
// Tag format: "valid_from:<RFC3339>" and "valid_until:<RFC3339>" (time.RFC3339 /
// time.RFC3339Nano accepted via time.Parse).
func ParseValidityWindow(e MemoryEntry) (from, until *time.Time) {
	for _, t := range e.TemporalTags {
		if strings.HasPrefix(t, validFromTagPrefix) {
			if ts, err := time.Parse(time.RFC3339Nano, strings.TrimPrefix(t, validFromTagPrefix)); err == nil {
				from = &ts
			} else if ts, err := time.Parse(time.RFC3339, strings.TrimPrefix(t, validFromTagPrefix)); err == nil {
				from = &ts
			}
		}
		if strings.HasPrefix(t, validUntilTagPrefix) {
			if ts, err := time.Parse(time.RFC3339Nano, strings.TrimPrefix(t, validUntilTagPrefix)); err == nil {
				until = &ts
			} else if ts, err := time.Parse(time.RFC3339, strings.TrimPrefix(t, validUntilTagPrefix)); err == nil {
				until = &ts
			}
		}
	}
	return from, until
}

// hasValidityTags reports whether e carries any valid_from / valid_until TemporalTags
// (regardless of whether they parse).
func hasValidityTags(e MemoryEntry) bool {
	for _, t := range e.TemporalTags {
		if strings.HasPrefix(t, validFromTagPrefix) || strings.HasPrefix(t, validUntilTagPrefix) {
			return true
		}
	}
	return false
}

// EntryValidAt reports whether e is considered valid at asOf.
//
// Rules (bi-temporal lite — validity window on entries, not full dual clocks):
//
//  1. Zero asOf is treated as time.Now().UTC().
//  2. When valid_from / valid_until tags are present:
//     - if valid_from is set and asOf.Before(from) → false
//     - if valid_until is set and !asOf.Before(until) → false
//     (valid_until is an exclusive end: asOf == until is invalid)
//  3. When NO validity tags at all: fall back to "known by asOf" —
//     valid if entryEventTime is zero OR !entryEventTime.After(asOf)
//     (entry exists / was recorded by asOf).
//
// This is not a full temporal knowledge graph (no transaction time, no edge
// validity). Hosts that write validity tags get windowed facts; untagged entries
// remain historically "known once recorded."
func EntryValidAt(e MemoryEntry, asOf time.Time) bool {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}

	if !hasValidityTags(e) {
		et := entryEventTime(e)
		if et.IsZero() {
			return true
		}
		// Known by asOf: event time must not be after asOf.
		return !et.After(asOf)
	}

	from, until := ParseValidityWindow(e)
	if from != nil && asOf.Before(*from) {
		return false
	}
	// Exclusive end: valid for asOf in [from, until).
	if until != nil && !asOf.Before(*until) {
		return false
	}
	return true
}

// FactsAsOfOptions configures as-of validity listing (K4 first slice / s616).
//
// Filters apply before Limit so many invalid-at-asOf entries do not underfill.
type FactsAsOfOptions struct {
	// AsOf is the validity instant (zero = Now UTC).
	AsOf time.Time
	// Query optional case-insensitive substring on Summary / Full / OriginalText.
	Query string
	// SessionID, when non-empty, keeps only matching SessionID.
	SessionID string
	// Entity filters TemporalTags:
	//   - if Entity contains ':', exact match on "entity:<value>" (or Entity itself
	//     when it already has the "entity:" prefix);
	//   - otherwise any TemporalTag with prefix "entity:" that contains Entity.
	Entity string
	// Limit caps results (default 50 when <= 0).
	Limit int
	// Tier when non-nil: only that tier. When nil: Working+Contextual+Semantic
	// (exclude Archival unless IncludeArchival).
	Tier *MemoryTier
	// IncludeArchival, when true and Tier==nil, also includes Archival.
	IncludeArchival bool
}

// entryMatchesEntity implements FactsAsOfOptions.Entity matching.
func entryMatchesEntity(e MemoryEntry, entity string) bool {
	if entity == "" {
		return true
	}
	if strings.Contains(entity, ":") {
		want := entity
		if !strings.HasPrefix(want, entityTagPrefix) {
			want = entityTagPrefix + entity
		}
		for _, t := range e.TemporalTags {
			if t == want {
				return true
			}
		}
		return false
	}
	for _, t := range e.TemporalTags {
		if strings.HasPrefix(t, entityTagPrefix) && strings.Contains(t, entity) {
			return true
		}
	}
	return false
}

// tierSemanticRank prefers Semantic over other tiers for facts-as-of ordering.
// Lower rank sorts first.
func tierSemanticRank(t MemoryTier) int {
	if t == TierSemantic {
		return 0
	}
	return 1
}

// ListFactsAsOf lists entries valid at AsOf with optional filters.
// Filters apply before Limit (underfill class, same as K1/K2).
//
// Order of operations: collect candidates → session → entity → query →
// EntryValidAt(asOf) → sort (Semantic first, then event time desc) → limit.
//
// Default tiers when Tier == nil: Working + Contextual + Semantic
// (+ Archival if IncludeArchival).
//
// Honesty: bi-temporal lite (validity window tags). Not full Graphiti dual
// clocks + graph. FS Palace remains O(n) over tier files.
func (ps *PalaceStore) ListFactsAsOf(opts FactsAsOfOptions) []MemoryEntry {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	asOf := opts.AsOf
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}

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

	if opts.SessionID != "" {
		var filtered []MemoryEntry
		for _, e := range results {
			if e.SessionID == opts.SessionID {
				filtered = append(filtered, e)
			}
		}
		results = filtered
	}

	if opts.Entity != "" {
		var filtered []MemoryEntry
		for _, e := range results {
			if entryMatchesEntity(e, opts.Entity) {
				filtered = append(filtered, e)
			}
		}
		results = filtered
	}

	if opts.Query != "" {
		q := strings.ToLower(opts.Query)
		var filtered []MemoryEntry
		for _, e := range results {
			hay := strings.ToLower(e.Content.Summary + " " + e.Content.Full + " " + e.OriginalText)
			if strings.Contains(hay, q) {
				filtered = append(filtered, e)
			}
		}
		results = filtered
	}

	// Validity filter before limit (underfill class).
	{
		var filtered []MemoryEntry
		for _, e := range results {
			if EntryValidAt(e, asOf) {
				filtered = append(filtered, e)
			}
		}
		results = filtered
	}

	// Semantic-heavy ordering: Semantic first, then event time descending within rank.
	sort.SliceStable(results, func(i, j int) bool {
		ri := tierSemanticRank(results[i].Tier)
		rj := tierSemanticRank(results[j].Tier)
		if ri != rj {
			return ri < rj
		}
		ti := entryEventTime(results[i])
		tj := entryEventTime(results[j])
		if !ti.Equal(tj) {
			return ti.After(tj)
		}
		return results[i].ID < results[j].ID
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results
}
