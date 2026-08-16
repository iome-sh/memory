package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	durableEventTimeIndexVersion = 1
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
	ID        string     `json:"id"`
	Tier      MemoryTier `json:"tier"`
	EventTime time.Time  `json:"event_time"`
	SessionID string     `json:"session_id"`
	Tags      []string   `json:"tags"`
	RelPath   string     `json:"rel_path"`
	QueryHay  string     `json:"query_hay"`
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
		ID:        m.ID,
		Tier:      m.Tier,
		EventTime: m.EventTime,
		SessionID: m.SessionID,
		Tags:      m.Tags,
		RelPath:   rel,
		QueryHay:  m.queryHay,
	}
}

func (d durableEntryMeta) toMeta(baseDir string) entryMeta {
	path := d.RelPath
	if path != "" && !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, filepath.FromSlash(d.RelPath))
	}
	return entryMeta{
		ID:        d.ID,
		Tier:      d.Tier,
		EventTime: d.EventTime,
		SessionID: d.SessionID,
		Tags:      d.Tags,
		Path:      path,
		queryHay:  d.QueryHay,
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
	if err := os.MkdirAll(dir, 0755); err != nil {
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
	tmp, err := os.CreateTemp(dir, ".tmp-event-time-*.json")
	if err != nil {
		return
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return
	}
	dest := ps.eventTimeIndexPath()
	if err := os.Rename(tmpName, dest); err != nil {
		os.Remove(tmpName)
	}
}
