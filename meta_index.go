package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// entryMeta is a lightweight in-memory record for ListMemoryWithOptions filtering
// without re-loading full MemoryEntry JSON for every candidate (K2 residual / s1066).
//
// FS Palace remains the source of truth. The index is a best-effort cache rebuilt
// lazily when dirty; correctness must match the full-scan ListMemory path.
type entryMeta struct {
	ID        string
	Tier      MemoryTier
	EventTime time.Time
	SessionID string
	// Tags is the union of TemporalTags and Content.Tags (order not significant).
	Tags []string
	Path string
	// queryHay is lowercased Summary + Full + OriginalText for substring Query.
	queryHay string
}

// metaFromEntry builds an entryMeta from a loaded MemoryEntry and its FS path.
func metaFromEntry(e MemoryEntry, path string) entryMeta {
	n := len(e.TemporalTags) + len(e.Content.Tags)
	tags := make([]string, 0, n)
	tags = append(tags, e.TemporalTags...)
	tags = append(tags, e.Content.Tags...)
	hay := strings.ToLower(e.Content.Summary + " " + e.Content.Full + " " + e.OriginalText)
	return entryMeta{
		ID:        e.ID,
		Tier:      e.Tier,
		EventTime: entryEventTime(e),
		SessionID: e.SessionID,
		Tags:      tags,
		Path:      path,
		queryHay:  hay,
	}
}

// entryMetaJSON is a slim unmarshal target for rebuild (avoids full MemoryEntry alloc).
type entryMetaJSON struct {
	ID           string     `json:"id"`
	Tier         MemoryTier `json:"tier"`
	SessionID    string     `json:"session_id"`
	Timestamp    time.Time  `json:"timestamp"`
	CreatedAt    time.Time  `json:"created_at"`
	LastAccessed time.Time  `json:"last_accessed"`
	TemporalTags []string   `json:"temporal_tags"`
	OriginalText string     `json:"original_text"`
	Content      struct {
		Summary string   `json:"summary"`
		Full    string   `json:"full"`
		Tags    []string `json:"tags"`
	} `json:"content"`
}

func (m entryMetaJSON) toMeta(path string, dirTier MemoryTier) entryMeta {
	e := MemoryEntry{
		ID:           m.ID,
		Tier:         m.Tier,
		SessionID:    m.SessionID,
		Timestamp:    m.Timestamp,
		CreatedAt:    m.CreatedAt,
		LastAccessed: m.LastAccessed,
		TemporalTags: m.TemporalTags,
		OriginalText: m.OriginalText,
		Content: MemoryContent{
			Summary: m.Content.Summary,
			Full:    m.Content.Full,
			Tags:    m.Content.Tags,
		},
	}
	if e.Tier == 0 {
		e.Tier = dirTier
	}
	return metaFromEntry(e, path)
}

// invalidateMetaIndex marks the in-memory metadata index dirty (call after writes).
// Safe for concurrent use. FS remains source of truth; next list rebuilds lazily.
func (ps *PalaceStore) invalidateMetaIndex() {
	ps.metaMu.Lock()
	ps.metaIndexDirty = true
	ps.metaMu.Unlock()
}

// InvalidateMetaIndex is the exported test/debug hook for force-rebuild.
func (ps *PalaceStore) InvalidateMetaIndex() {
	ps.invalidateMetaIndex()
}

// MetaIndexLen returns the number of entries currently cached in the meta index
// (0 if dirty/unbuilt). Intended for tests and observability — not a product API.
func (ps *PalaceStore) MetaIndexLen() int {
	ps.metaMu.Lock()
	defer ps.metaMu.Unlock()
	if ps.metaIndexDirty || ps.metaIndex == nil {
		return 0
	}
	return len(ps.metaIndex)
}

// ensureMetaIndexLocked rebuilds the index when dirty or missing.
// Caller must hold ps.metaMu.
func (ps *PalaceStore) ensureMetaIndexLocked() {
	if !ps.metaIndexDirty && ps.metaIndex != nil {
		return
	}
	if ps.tryLoadDurableMetaLocked() {
		return
	}
	ps.rebuildMetaIndexLocked()
}

// rebuildMetaIndexLocked walks all tier dirs once and rebuilds metaIndex.
// Caller must hold ps.metaMu.
func (ps *PalaceStore) rebuildMetaIndexLocked() {
	var meta []entryMeta
	for _, tier := range []MemoryTier{TierWorking, TierContextual, TierArchival, TierSemantic} {
		dir := ps.getTierDir(tier)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			// Skip temp files from atomic writes.
			name := f.Name()
			if strings.HasPrefix(name, ".tmp-") {
				continue
			}
			path := filepath.Join(dir, name)
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var raw entryMetaJSON
			if err := json.Unmarshal(data, &raw); err != nil {
				continue
			}
			if raw.ID == "" {
				// Fall back to filename stem when ID missing in JSON.
				raw.ID = strings.TrimSuffix(name, ".json")
			}
			meta = append(meta, raw.toMeta(path, tier))
		}
	}
	ps.metaIndex = meta
	ps.metaIndexDirty = false
	ps.metaIndexGen++
	ps.persistDurableMetaLocked()
}

// metaHasTag reports exact tag match on entryMeta.Tags.
func metaHasTag(m entryMeta, tag string) bool {
	if tag == "" {
		return false
	}
	for _, t := range m.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// metaHasTagPrefix reports prefix match on entryMeta.Tags.
func metaHasTagPrefix(m entryMeta, prefix string) bool {
	if prefix == "" {
		return false
	}
	for _, t := range m.Tags {
		if strings.HasPrefix(t, prefix) {
			return true
		}
	}
	return false
}

// filterMetaIndex applies ListMemoryOptions filters (except Limit/sort) to meta rows.
// Query uses queryHay; session/time/tag use meta fields. Tier set is enforced here.
func filterMetaIndex(meta []entryMeta, opts ListMemoryOptions) []entryMeta {
	// Tier set
	var wantTiers map[MemoryTier]struct{}
	if opts.Tier != nil {
		wantTiers = map[MemoryTier]struct{}{*opts.Tier: {}}
	} else {
		wantTiers = map[MemoryTier]struct{}{
			TierWorking:    {},
			TierContextual: {},
			TierSemantic:   {},
		}
		if opts.IncludeArchival {
			wantTiers[TierArchival] = struct{}{}
		}
	}

	out := make([]entryMeta, 0, len(meta))
	q := ""
	if opts.Query != "" {
		q = strings.ToLower(opts.Query)
	}

	for _, m := range meta {
		if _, ok := wantTiers[m.Tier]; !ok {
			continue
		}
		if opts.SessionID != "" && m.SessionID != opts.SessionID {
			continue
		}
		if opts.TimeFrom != nil && m.EventTime.Before(*opts.TimeFrom) {
			continue
		}
		if opts.TimeTo != nil && m.EventTime.After(*opts.TimeTo) {
			continue
		}
		if opts.Tag != "" && !metaHasTag(m, opts.Tag) {
			continue
		}
		if opts.TagPrefix != "" && !metaHasTagPrefix(m, opts.TagPrefix) {
			continue
		}
		if q != "" && !strings.Contains(m.queryHay, q) {
			continue
		}
		out = append(out, m)
	}
	return out
}

// loadEntriesFromMeta loads full MemoryEntry values for the given meta rows
// (in the same order). Skips rows that fail to load.
func (ps *PalaceStore) loadEntriesFromMeta(rows []entryMeta) []MemoryEntry {
	out := make([]MemoryEntry, 0, len(rows))
	for _, m := range rows {
		var e MemoryEntry
		var ok bool
		if m.Path != "" {
			e, ok = ps.loadEntry(m.Path)
		}
		if !ok {
			e, ok = ps.Load(m.ID, m.Tier)
		}
		if ok {
			out = append(out, e)
		}
	}
	return out
}
