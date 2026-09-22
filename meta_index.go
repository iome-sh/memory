package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// entryMeta is a lightweight in-memory record for ListMemoryWithOptions filtering
// without re-loading full MemoryEntry JSON for every candidate (K2 / s1066).
//
// FS Palace remains the source of truth. Write/unlink patches a clean index in
// place; a dirty/missing index rebuilds lazily. Correctness must match the
// full-scan ListMemory path.
type entryMeta struct {
	ID        string
	Tier      MemoryTier
	EventTime time.Time
	SessionID string
	// Tags is the union of TemporalTags and Content.Tags (order not significant).
	// Not used for validity or ListFactsAsOf entity match: a content tag may
	// look like valid_until: or entity: without being a temporal tag.
	Tags []string
	Path string
	// queryHay is lowercased Summary + Full + OriginalText for substring Query.
	queryHay string
	// hasValidity / validFrom / validUntil are computed from TemporalTags only,
	// same rules as hasValidityTags / ParseValidityWindow / EntryValidAt.
	hasValidity bool
	validFrom   *time.Time
	validUntil  *time.Time
	// entityTags are TemporalTags with the entity: prefix (not Content.Tags).
	entityTags []string
	// entityKeys are normalized EntryEntityKeys. entityKeysKnown is false when
	// the row cannot prove a non-match; SupersedeEntityFacts then loads the file.
	entityKeys      []string
	entityKeysKnown bool
}

// metaFromEntry builds an entryMeta from a loaded MemoryEntry and its FS path.
func metaFromEntry(e MemoryEntry, path string) entryMeta {
	n := len(e.TemporalTags) + len(e.Content.Tags)
	tags := make([]string, 0, n)
	tags = append(tags, e.TemporalTags...)
	tags = append(tags, e.Content.Tags...)
	hay := strings.ToLower(e.Content.Summary + " " + e.Content.Full + " " + e.OriginalText)
	has, from, until := validityFromTemporalTags(e.TemporalTags)
	return entryMeta{
		ID:              e.ID,
		Tier:            e.Tier,
		EventTime:       entryEventTime(e),
		SessionID:       e.SessionID,
		Tags:            tags,
		Path:            path,
		queryHay:        hay,
		hasValidity:     has,
		validFrom:       from,
		validUntil:      until,
		entityTags:      temporalEntityTags(e.TemporalTags),
		entityKeys:      normalizedEntityKeys(e),
		entityKeysKnown: true,
	}
}

// validityFromTemporalTags mirrors EntryValidAt's tag rules. Content.Tags are
// not consulted: a content "valid_until:" tag must not close the meta row.
func validityFromTemporalTags(tags []string) (has bool, from, until *time.Time) {
	e := MemoryEntry{TemporalTags: tags}
	has = hasValidityTags(e)
	from, until = ParseValidityWindow(e)
	return has, cloneTimePtr(from), cloneTimePtr(until)
}

func cloneTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	c := t.UTC()
	return &c
}

// temporalEntityTags copies TemporalTags that use the entity: prefix.
// Content.Tags are omitted so they cannot widen ListFactsAsOf's entity filter.
func temporalEntityTags(tags []string) []string {
	var out []string
	for _, t := range tags {
		if strings.HasPrefix(t, entityTagPrefix) {
			out = append(out, t)
		}
	}
	return out
}

// normalizedEntityKeys is EntryEntityKeys lower-cased and trimmed, first-seen order.
func normalizedEntityKeys(e MemoryEntry) []string {
	raw := EntryEntityKeys(e)
	if len(raw) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, k := range raw {
		n := normalizeEntityKey(k)
		if n == "" {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

// metaValidAt is EntryValidAt on the cached temporal window / event time.
func metaValidAt(m entryMeta, asOf time.Time) bool {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	if !m.hasValidity {
		if m.EventTime.IsZero() {
			return true
		}
		return !m.EventTime.After(asOf)
	}
	if m.validFrom != nil && asOf.Before(*m.validFrom) {
		return false
	}
	if m.validUntil != nil && !asOf.Before(*m.validUntil) {
		return false
	}
	return true
}

// metaMatchesEntity is entryMatchesEntity over temporal entity tags only.
func metaMatchesEntity(m entryMeta, entity string) bool {
	if entity == "" {
		return true
	}
	return tagsMatchEntity(m.entityTags, entity)
}

// metaCouldMatchEntityKey reports whether m might satisfy entryHasEntityKey.
// A row that does not know its keys cannot prove a non-match, so the caller loads the file.
func metaCouldMatchEntityKey(m entryMeta, entityKey string) bool {
	want := normalizeEntityKey(entityKey)
	if want == "" {
		return false
	}
	if !m.entityKeysKnown {
		return true
	}
	for _, k := range m.entityKeys {
		if normalizeEntityKey(k) == want {
			return true
		}
	}
	return false
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
	Relations struct {
		RelatedConcepts []string `json:"related_concepts"`
	} `json:"relations"`
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
		Relations: MemoryRelations{
			RelatedConcepts: m.Relations.RelatedConcepts,
		},
	}
	if e.Tier == 0 {
		e.Tier = dirTier
	}
	return metaFromEntry(e, path)
}

// invalidateMetaIndex marks the in-memory metadata index dirty.
// Fallback when a mutation cannot be patched; next list rebuilds lazily.
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

// MetaIndexRebuilds returns how many full tier-JSON walks rebuildMetaIndexLocked
// has run on this store. Tests use this to prove Write/unlink patched in place.
// Not a product API.
func (ps *PalaceStore) MetaIndexRebuilds() uint64 {
	ps.metaMu.Lock()
	defer ps.metaMu.Unlock()
	return ps.metaIndexRebuilds
}

// upsertMetaIndex patches a clean in-memory (and optional durable) index after Write.
// If the index is dirty or unbuilt, it stays dirty so the next list rebuilds from FS.
func (ps *PalaceStore) upsertMetaIndex(entry MemoryEntry, path string) {
	if ps.Config.DisableMetaIndex || path == "" {
		return
	}
	path = filepath.Clean(path)
	if entry.Tier == 0 {
		entry.Tier = inferTierFromPath(path)
	}
	m := metaFromEntry(entry, path)

	ps.metaMu.Lock()
	defer ps.metaMu.Unlock()
	if ps.metaIndexDirty || ps.metaIndex == nil {
		return
	}
	replaced := false
	for i := range ps.metaIndex {
		if ps.metaIndex[i].Path == path {
			ps.metaIndex[i] = m
			replaced = true
			break
		}
	}
	if !replaced {
		ps.metaIndex = append(ps.metaIndex, m)
	}
	ps.metaIndexGen++
	ps.persistDurableMetaLocked()
}

// removeMetaIndex drops the row for path from a clean index and persists the
// optional durable snapshot. Dirty/unbuilt indexes stay dirty.
func (ps *PalaceStore) removeMetaIndex(path string) {
	if ps.Config.DisableMetaIndex || path == "" {
		return
	}
	path = filepath.Clean(path)

	ps.metaMu.Lock()
	defer ps.metaMu.Unlock()
	if ps.metaIndexDirty || ps.metaIndex == nil {
		return
	}
	kept := ps.metaIndex[:0]
	removed := false
	for _, row := range ps.metaIndex {
		if row.Path == path {
			removed = true
			continue
		}
		kept = append(kept, row)
	}
	if !removed {
		return
	}
	// Release tail references when shrinking.
	for i := len(kept); i < len(ps.metaIndex); i++ {
		ps.metaIndex[i] = entryMeta{}
	}
	ps.metaIndex = kept
	ps.metaIndexGen++
	ps.persistDurableMetaLocked()
}

// unlinkEntry removes a listed-tier JSON file and patches the meta index.
// Missing files are not an error. FS remains source of truth.
func (ps *PalaceStore) unlinkEntry(id string, tier MemoryTier) error {
	if id == "" {
		return nil
	}
	path := filepath.Join(ps.getTierDir(tier), id+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		ps.invalidateMetaIndex()
		return err
	}
	ps.removeMetaIndex(path)
	return nil
}

func inferTierFromPath(path string) MemoryTier {
	switch filepath.Base(filepath.Dir(path)) {
	case "tier-1-working":
		return TierWorking
	case "tier-2-contextual":
		return TierContextual
	case "tier-3-archival":
		return TierArchival
	case "tier-4-semantic":
		return TierSemantic
	default:
		return TierContextual
	}
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
// Caller must hold ps.metaMu. Incremental Write/unlink avoids this path.
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
			path := filepath.Clean(filepath.Join(dir, name))
			ps.noteEntryJSONRead(path)
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
	ps.metaIndexRebuilds++
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
		if (opts.SessionID != "" || len(opts.SessionIDs) > 0) && !metaMatchesSessionFilter(m, opts.SessionID, opts.SessionIDs) {
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
