package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// durableEventTimeIndexVersion is the on-disk entryMeta schema.
	// v2 stores has_validity / valid_from / valid_until plus temporal entity
	// tags and normalized entity keys. A v1 file fails the version check so
	// rebuildMetaIndexLocked rewrites it. Do not decode v1 as v2: missing
	// has_validity looks like "no validity tags" and would keep closed facts
	// via the known-by event-time rule.
	durableEventTimeIndexVersion = 2
	durableEventTimeIndexRelPath = "indexes/event-time.json"
)

// durableEventTimeIndex is a best-effort on-disk snapshot of entryMeta
// (K2 / #44). FS Palace JSON remains source of truth. A stamp of tier
// JSON count + max mtime detects staleness without parsing every file.
type durableEventTimeIndex struct {
	Version    int                `json:"version"`
	JSONCount  int                `json:"json_count"`
	MaxMtimeNs int64              `json:"max_mtime_ns"`
	Entries    []durableEntryMeta `json:"entries"`
}

type durableEntryMeta struct {
	ID              string     `json:"id"`
	Tier            MemoryTier `json:"tier"`
	EventTime       time.Time  `json:"event_time"`
	SessionID       string     `json:"session_id"`
	Tags            []string   `json:"tags"`
	RelPath         string     `json:"rel_path"`
	QueryHay        string     `json:"query_hay"`
	HasValidity     bool       `json:"has_validity"`
	ValidFrom       *time.Time `json:"valid_from,omitempty"`
	ValidUntil      *time.Time `json:"valid_until,omitempty"`
	EntityTags      []string   `json:"entity_tags,omitempty"`
	EntityKeys      []string   `json:"entity_keys,omitempty"`
	EntityKeysKnown bool       `json:"entity_keys_known"`
}

func (ps *PalaceStore) eventTimeIndexPath() string {
	return filepath.Join(ps.BaseDir, filepath.FromSlash(durableEventTimeIndexRelPath))
}

func (m entryMeta) toDurable(baseDir string) durableEntryMeta {
	rel := m.Path
	if baseDir != "" {
		if r, err := filepath.Rel(baseDir, m.Path); err == nil {
			rel = filepath.ToSlash(r)
		}
	}
	return durableEntryMeta{
		ID:              m.ID,
		Tier:            m.Tier,
		EventTime:       m.EventTime,
		SessionID:       m.SessionID,
		Tags:            m.Tags,
		RelPath:         rel,
		QueryHay:        m.queryHay,
		HasValidity:     m.hasValidity,
		ValidFrom:       cloneTimePtr(m.validFrom),
		ValidUntil:      cloneTimePtr(m.validUntil),
		EntityTags:      m.entityTags,
		EntityKeys:      m.entityKeys,
		EntityKeysKnown: m.entityKeysKnown,
	}
}

func (d durableEntryMeta) toMeta(baseDir string) entryMeta {
	path := d.RelPath
	if path != "" && !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, filepath.FromSlash(d.RelPath))
	}
	return entryMeta{
		ID:              d.ID,
		Tier:            d.Tier,
		EventTime:       d.EventTime,
		SessionID:       d.SessionID,
		Tags:            d.Tags,
		Path:            path,
		queryHay:        d.QueryHay,
		hasValidity:     d.HasValidity,
		validFrom:       cloneTimePtr(d.ValidFrom),
		validUntil:      cloneTimePtr(d.ValidUntil),
		entityTags:      append([]string(nil), d.EntityTags...),
		entityKeys:      append([]string(nil), d.EntityKeys...),
		entityKeysKnown: d.EntityKeysKnown,
	}
}

// palaceJSONStamp is a cheap FS fingerprint: count of tier *.json files
// (skipping .tmp-*) and the max ModTime. No JSON parse.
func (ps *PalaceStore) palaceJSONStamp() (count int, maxMtimeNs int64) {
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
			if strings.HasPrefix(f.Name(), ".tmp-") {
				continue
			}
			count++
			info, err := f.Info()
			if err != nil {
				continue
			}
			ns := info.ModTime().UnixNano()
			if ns > maxMtimeNs {
				maxMtimeNs = ns
			}
		}
	}
	return count, maxMtimeNs
}

// tryLoadDurableMetaLocked loads indexes/event-time.json when the stamp
// matches the live palace. Caller must hold ps.metaMu.
func (ps *PalaceStore) tryLoadDurableMetaLocked() bool {
	if ps.Config.DisableDurableIndex || ps.Config.DisableMetaIndex {
		return false
	}
	data, err := os.ReadFile(ps.eventTimeIndexPath())
	if err != nil {
		return false
	}
	var snap durableEventTimeIndex
	if err := json.Unmarshal(data, &snap); err != nil {
		return false
	}
	if snap.Version != durableEventTimeIndexVersion {
		return false
	}
	count, maxNs := ps.palaceJSONStamp()
	if snap.JSONCount != count || snap.MaxMtimeNs != maxNs {
		return false
	}
	meta := make([]entryMeta, 0, len(snap.Entries))
	for _, d := range snap.Entries {
		if d.ID == "" {
			continue
		}
		meta = append(meta, d.toMeta(ps.BaseDir))
	}
	ps.metaIndex = meta
	ps.metaIndexDirty = false
	ps.metaIndexGen++
	return true
}

// persistDurableMetaLocked writes the in-memory meta snapshot atomically.
// Caller must hold ps.metaMu. Best-effort: errors are ignored.
func (ps *PalaceStore) persistDurableMetaLocked() {
	if ps.Config.DisableDurableIndex || ps.Config.DisableMetaIndex {
		return
	}
	dir := filepath.Join(ps.BaseDir, "indexes")
	if err := palaceMkdirAll(dir); err != nil {
		return
	}
	count, maxNs := ps.palaceJSONStamp()
	snap := durableEventTimeIndex{
		Version:    durableEventTimeIndexVersion,
		JSONCount:  count,
		MaxMtimeNs: maxNs,
		Entries:    make([]durableEntryMeta, 0, len(ps.metaIndex)),
	}
	for _, m := range ps.metaIndex {
		snap.Entries = append(snap.Entries, m.toDurable(ps.BaseDir))
	}
	data, err := json.Marshal(snap)
	if err != nil {
		return
	}
	// Caller holds metaMu. writeMu serializes with entity-graph rewrites (#86).
	ps.writeMu.Lock()
	defer ps.writeMu.Unlock()
	_ = palaceWriteFileAtomic(dir, ps.eventTimeIndexPath(), ".tmp-event-time-*.json", data)
}
