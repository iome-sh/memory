package memory

import (
	"sort"
	"strings"
	"time"
)

// ListMemoryOptions configures event-time ordered listing with optional filters (K2 / s611).
//
// Filters are applied before Limit so windowed/session timelines do not underfill
// when many out-of-scope entries exist (same underfill class as SearchMemoryWithOptions K1).
//
// Complexity: FS Palace is source of truth. When the best-effort meta index is enabled
// (default), ListMemoryWithOptions filters on lightweight entryMeta and loads full JSON
// only for survivors (s1066). A durable snapshot at indexes/event-time.json (#44)
// avoids re-parsing every tier JSON on a fresh process when the stamp matches.
// With DisableMetaIndex, falls back to O(n) full entry scan.
type ListMemoryOptions struct {
	SessionID string
	// TimeFrom / TimeTo filter by entry event time (see entryEventTime). Both inclusive when set.
	TimeFrom *time.Time
	TimeTo   *time.Time
	// Tag exact-matches any of TemporalTags or Content.Tags.
	Tag string
	// TagPrefix matches tag strings with strings.HasPrefix (e.g. "subject:", "session_seq:").
	TagPrefix string
	// Query optional case-insensitive substring on Summary / Full / OriginalText.
	Query string
	// Limit caps results (default 50 when <= 0 — timeline-friendly; not search's 10).
	Limit int
	// Tier when non-nil: only that tier. When nil: Working+Contextual+Semantic
	// (exclude Archival by default to match aion MCP timeline).
	Tier *MemoryTier
	// IncludeArchival, when true and Tier==nil, also includes Archival.
	IncludeArchival bool
	// Ascending: false = newest first (default); true = oldest first.
	Ascending bool
}

// EntryHasTag reports whether e has an exact tag match in TemporalTags or Content.Tags.
func EntryHasTag(e MemoryEntry, tag string) bool {
	if tag == "" {
		return false
	}
	for _, t := range e.TemporalTags {
		if t == tag {
			return true
		}
	}
	for _, t := range e.Content.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// EntryHasTagPrefix reports whether e has a TemporalTags or Content.Tags entry with the given prefix.
func EntryHasTagPrefix(e MemoryEntry, prefix string) bool {
	if prefix == "" {
		return false
	}
	for _, t := range e.TemporalTags {
		if strings.HasPrefix(t, prefix) {
			return true
		}
	}
	for _, t := range e.Content.Tags {
		if strings.HasPrefix(t, prefix) {
			return true
		}
	}
	return false
}

// ListMemoryWithOptions returns entries ordered by event time with optional filters applied
// before Limit (K2 / s611 timeline surface; s1066 meta index when enabled).
//
// Order of operations: collect candidates → session → time → tag filters → query substring →
// sort by entryEventTime → limit.
func (ps *PalaceStore) ListMemoryWithOptions(opts ListMemoryOptions) []MemoryEntry {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	var results []MemoryEntry
	if ps.Config.DisableMetaIndex {
		results = ps.listMemoryScan(opts)
	} else {
		results = ps.listMemoryViaIndex(opts)
	}

	// Sort by event time
	ascending := opts.Ascending
	sort.SliceStable(results, func(i, j int) bool {
		ti := entryEventTime(results[i])
		tj := entryEventTime(results[j])
		if ascending {
			if !ti.Equal(tj) {
				return ti.Before(tj)
			}
			// Stable tie-break by ID for determinism
			return results[i].ID < results[j].ID
		}
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

// listMemoryViaIndex filters on the best-effort meta index, then loads full entries
// only for survivors. Rebuilds lazily when dirty.
func (ps *PalaceStore) listMemoryViaIndex(opts ListMemoryOptions) []MemoryEntry {
	ps.metaMu.Lock()
	ps.ensureMetaIndexLocked()
	// Copy filtered meta under lock so concurrent invalidation cannot race rebuild mid-filter.
	filtered := filterMetaIndex(ps.metaIndex, opts)
	// Shallow-copy paths needed for load (meta rows are value copies already).
	rows := make([]entryMeta, len(filtered))
	copy(rows, filtered)
	ps.metaMu.Unlock()

	return ps.loadEntriesFromMeta(rows)
}

// listMemoryScan is the original O(n) full-JSON path (DisableMetaIndex / parity baseline).
func (ps *PalaceStore) listMemoryScan(opts ListMemoryOptions) []MemoryEntry {
	var results []MemoryEntry
	if opts.Tier != nil {
		results = ps.ListEntriesInTier(*opts.Tier)
	} else {
		tiers := []MemoryTier{TierWorking, TierContextual, TierSemantic}
		if opts.IncludeArchival {
			// Insert Archival with Working/Contextual/Semantic (stable tier order).
			tiers = []MemoryTier{TierWorking, TierContextual, TierArchival, TierSemantic}
		}
		for _, t := range tiers {
			results = append(results, ps.ListEntriesInTier(t)...)
		}
	}

	// Session filter
	if opts.SessionID != "" {
		var filtered []MemoryEntry
		for _, e := range results {
			if e.SessionID == opts.SessionID {
				filtered = append(filtered, e)
			}
		}
		results = filtered
	}

	// Time-window filter (inclusive bounds on preferred event time)
	if opts.TimeFrom != nil || opts.TimeTo != nil {
		var filtered []MemoryEntry
		for _, e := range results {
			et := entryEventTime(e)
			if opts.TimeFrom != nil && et.Before(*opts.TimeFrom) {
				continue
			}
			if opts.TimeTo != nil && et.After(*opts.TimeTo) {
				continue
			}
			filtered = append(filtered, e)
		}
		results = filtered
	}

	// Tag exact match
	if opts.Tag != "" {
		var filtered []MemoryEntry
		for _, e := range results {
			if EntryHasTag(e, opts.Tag) {
				filtered = append(filtered, e)
			}
		}
		results = filtered
	}

	// Tag prefix match
	if opts.TagPrefix != "" {
		var filtered []MemoryEntry
		for _, e := range results {
			if EntryHasTagPrefix(e, opts.TagPrefix) {
				filtered = append(filtered, e)
			}
		}
		results = filtered
	}

	// Query substring (case-insensitive) on Summary / Full / OriginalText
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

	return results
}
