package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// walPendingRelDir is the intent-log directory for transactional IngestTurn.
// Not flock. Not a multi-process lock. Recovered on NewPalaceStoreWithConfig
// even when TransactionalIngest is false (leftover cloud crash).
const walPendingRelDir = "wal/pending"

type pendingIngestRecord struct {
	TurnID string              `json:"turn_id"`
	Files  []pendingIngestFile `json:"files"`
}

type pendingIngestFile struct {
	Dest   string `json:"dest"`
	SHA256 string `json:"sha256"`
}

type ingestCommitItem struct {
	entry MemoryEntry
	dir   string
	dest  string
	temp  string
	data  []byte
}

func (ps *PalaceStore) walPendingDir() string {
	return filepath.Join(ps.BaseDir, filepath.FromSlash(walPendingRelDir))
}

func (ps *PalaceStore) pendingRecordPath(turnID string) string {
	return filepath.Join(ps.walPendingDir(), sanitizeIngestID(turnID)+".json")
}

func ingestSidecarTempPath(destAbs, turnID string) string {
	return destAbs + ".tmp-" + sanitizeIngestID(turnID)
}

func sanitizeIngestID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return "_"
	}
	var b strings.Builder
	b.Grow(len(id))
	for _, r := range id {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" || out == "." || out == ".." {
		return "_"
	}
	return out
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (ps *PalaceStore) destAbsFromRel(rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", fmt.Errorf("empty dest")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("dest %q is absolute", rel)
	}
	cleaned := filepath.Clean(filepath.FromSlash(rel))
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("dest %q escapes palace root", rel)
	}
	base := filepath.Clean(ps.BaseDir)
	abs := filepath.Join(base, cleaned)
	sep := string(os.PathSeparator)
	if abs != base && !strings.HasPrefix(abs, base+sep) {
		return "", fmt.Errorf("dest %q escapes palace root", rel)
	}
	return abs, nil
}

func (ps *PalaceStore) writePendingRecord(rec pendingIngestRecord) error {
	dir := ps.walPendingDir()
	if err := palaceMkdirAll(dir); err != nil {
		return fmt.Errorf("mkdir wal/pending: %w", err)
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal wal/pending: %w", err)
	}
	dest := ps.pendingRecordPath(rec.TurnID)
	tmp := dest + ".tmp"
	if err := palaceWriteFileNamed(tmp, data); err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename wal/pending: %w", err)
	}
	if err := palaceFsyncDir(dir); err != nil {
		return fmt.Errorf("fsync wal/pending: %w", err)
	}
	return nil
}

func (ps *PalaceStore) abortPendingIngest(turnID string, temps []string) {
	for _, tmp := range temps {
		_ = os.Remove(tmp)
	}
	_ = os.Remove(ps.pendingRecordPath(turnID))
}

// recoverPendingIngest completes leftover IngestTurn intent logs.
// If a sidecar temp exists and its SHA-256 matches the pending record, the
// rename is redone. Otherwise dests are left untouched and the pending
// record is discarded. Runs on open even when TransactionalIngest is false.
func (ps *PalaceStore) recoverPendingIngest() {
	if ps == nil || ps.BaseDir == "" {
		return
	}
	dir := ps.walPendingDir()
	ents, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range ents {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		ps.recoverOnePending(filepath.Join(dir, name))
	}
}

func (ps *PalaceStore) recoverOnePending(path string) {
	defer os.Remove(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var rec pendingIngestRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return
	}
	turnID := rec.TurnID
	if turnID == "" {
		turnID = strings.TrimSuffix(filepath.Base(path), ".json")
	}
	for _, f := range rec.Files {
		destAbs, err := ps.destAbsFromRel(f.Dest)
		if err != nil {
			continue
		}
		tmp := ingestSidecarTempPath(destAbs, turnID)
		raw, err := os.ReadFile(tmp)
		if err != nil {
			continue
		}
		if !strings.EqualFold(sha256Hex(raw), strings.TrimSpace(f.SHA256)) {
			_ = os.Remove(tmp)
			continue
		}
		if err := palaceMkdirAll(filepath.Dir(destAbs)); err != nil {
			continue
		}
		if err := os.Rename(tmp, destAbs); err != nil {
			continue
		}
		_ = palaceFsyncDir(filepath.Dir(destAbs))
	}
}

func (ps *PalaceStore) commitIngestTurnEntries(turnID string, entries []MemoryEntry) error {
	if len(entries) == 0 {
		return nil
	}
	id := sanitizeIngestID(turnID)
	items := make([]ingestCommitItem, 0, len(entries))
	rec := pendingIngestRecord{TurnID: id}
	for _, e := range entries {
		if strings.TrimSpace(e.ID) == "" {
			return fmt.Errorf("ingest dest missing id")
		}
		if e.Version == 0 {
			e.Version = 1
		}
		prev, hasPrev := ps.loadSameID(e.ID, e.Tier)
		e = ps.preparePersistedEntry(e, prev, hasPrev)
		data, err := json.MarshalIndent(e, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal failed: %w", err)
		}
		dir := ps.getTierDir(e.Tier)
		if err := palaceMkdirAll(dir); err != nil {
			return fmt.Errorf("mkdir ingest dest: %w", err)
		}
		destAbs := filepath.Join(dir, e.ID+".json")
		rel, err := filepath.Rel(ps.BaseDir, destAbs)
		if err != nil {
			return fmt.Errorf("ingest dest rel: %w", err)
		}
		items = append(items, ingestCommitItem{
			entry: e,
			dir:   dir,
			dest:  destAbs,
			temp:  ingestSidecarTempPath(destAbs, id),
			data:  data,
		})
		rec.Files = append(rec.Files, pendingIngestFile{
			Dest:   filepath.ToSlash(rel),
			SHA256: sha256Hex(data),
		})
	}

	if err := ps.writePendingRecord(rec); err != nil {
		return err
	}

	writtenTemps := make([]string, 0, len(items))
	abort := func() {
		ps.abortPendingIngest(id, writtenTemps)
	}

	for i := range items {
		if err := palaceWriteFileNamed(items[i].temp, items[i].data); err != nil {
			abort()
			return fmt.Errorf("write ingest temp: %w", err)
		}
		writtenTemps = append(writtenTemps, items[i].temp)
	}

	seenDirs := make(map[string]struct{}, len(items))
	for i := range items {
		if _, ok := seenDirs[items[i].dir]; ok {
			continue
		}
		seenDirs[items[i].dir] = struct{}{}
		if err := palaceFsyncDir(items[i].dir); err != nil {
			abort()
			return fmt.Errorf("fsync ingest dir: %w", err)
		}
	}

	for i := range items {
		if err := os.Rename(items[i].temp, items[i].dest); err != nil {
			// Leave pending + remaining temps for recover-on-open.
			return fmt.Errorf("rename ingest dest: %w", err)
		}
	}
	for dir := range seenDirs {
		_ = palaceFsyncDir(dir)
	}

	if err := os.Remove(ps.pendingRecordPath(id)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove ingest wal: %w", err)
	}
	_ = palaceFsyncDir(ps.walPendingDir())

	for i := range items {
		_ = ps.archiveToVersions(items[i].entry)
		ps.upsertMetaIndex(items[i].entry, items[i].dest)
	}

	// Post-ack ONNX rewrite matches Write: PersistEmbeddings default off;
	// hash vectors are never stored.
	for i := range items {
		e := items[i].entry
		if !ps.shouldComputePersistedEmbedding(e) {
			continue
		}
		vec, ok := ps.safeEmbedEntry(e)
		if !ok {
			continue
		}
		e.Content.Embedding = vec
		e.Content.EmbeddingModel = strings.TrimSpace(ps.Config.EmbeddingModel)
		e.Content.EmbeddingDim = len(vec)
		_ = ps.persistEntryFile(e)
		_ = ps.archiveToVersions(e)
	}
	return nil
}
