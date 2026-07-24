package memory

import (
	"strings"
	"time"
)

// SupersedeEntityFacts finds entries matching entityKey (via EntryEntityKeys /
// entity: tags) that are still open at asOf (EntryValidAt), and writes
// valid_until = asOf (exclusive end matching existing FactsAsOf semantics).
// Does not delete entries.
//
// Matching: normalize entityKey to lower-case trim; an entry matches when any
// key from EntryEntityKeys equals that normalized form (also lower-cased).
// Empty entityKey is a no-op (returns 0, nil).
//
// Honesty: bi-temporal lite supersession — not automatic NLP contradiction
// detection, not full Zep dual-clock KG. Callers pass explicit entity keys.
//
// Returns the count of updated entries.
func (ps *PalaceStore) SupersedeEntityFacts(entityKey string, asOf time.Time) (int, error) {
	return ps.supersedeEntityFactsExcluding(entityKey, asOf, "")
}

// WriteAndSupersede writes entry first (stamping valid_from=now when unset),
// then runs SupersedeEntityFacts for each supersedeKeys entry, excluding the
// newly written entry's ID so it is not closed by its own write.
//
// asOf for supersession is the same UTC now used for valid_from stamping.
func (ps *PalaceStore) WriteAndSupersede(entry MemoryEntry, supersedeKeys []string) error {
	now := time.Now().UTC()
	if !hasValidFromTag(entry) {
		entry.TemporalTags = append(entry.TemporalTags, validFromTagPrefix+now.Format(time.RFC3339))
	}
	if err := ps.Write(entry); err != nil {
		return err
	}
	for _, key := range supersedeKeys {
		if _, err := ps.supersedeEntityFactsExcluding(key, now, entry.ID); err != nil {
			return err
		}
	}
	return nil
}

// supersedeEntityFactsExcluding is the shared implementation. excludeID, when
// non-empty, skips that entry (used by WriteAndSupersede for the new write).
func (ps *PalaceStore) supersedeEntityFactsExcluding(entityKey string, asOf time.Time, excludeID string) (int, error) {
	key := normalizeEntityKey(entityKey)
	if key == "" {
		return 0, nil
	}
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	} else {
		asOf = asOf.UTC()
	}

	// Scan all primary tiers (incl. Archival) so open facts are closed wherever stored.
	tiers := []MemoryTier{TierWorking, TierContextual, TierArchival, TierSemantic}
	updated := 0
	for _, tier := range tiers {
		entries := ps.ListEntriesInTier(tier)
		for _, e := range entries {
			if excludeID != "" && e.ID == excludeID {
				continue
			}
			if !entryHasEntityKey(e, key) {
				continue
			}
			// Only currently-open facts at asOf (covers untagged + open windows;
			// already-closed windows fail EntryValidAt and are skipped).
			if !EntryValidAt(e, asOf) {
				continue
			}
			e.TemporalTags = setValidUntilTag(e.TemporalTags, asOf)
			e.UpdatedAt = asOf
			if err := ps.Write(e); err != nil {
				return updated, err
			}
			updated++
		}
	}
	return updated, nil
}

// normalizeEntityKey trims and lower-cases an entity key for matching.
func normalizeEntityKey(entityKey string) string {
	return strings.ToLower(strings.TrimSpace(entityKey))
}

// entryHasEntityKey reports whether EntryEntityKeys(e) contains entityKey after
// lower-case / trim normalization on both sides.
func entryHasEntityKey(e MemoryEntry, entityKey string) bool {
	want := normalizeEntityKey(entityKey)
	if want == "" {
		return false
	}
	for _, k := range EntryEntityKeys(e) {
		if normalizeEntityKey(k) == want {
			return true
		}
	}
	return false
}

// hasValidFromTag reports whether e already carries a valid_from TemporalTag.
func hasValidFromTag(e MemoryEntry) bool {
	for _, t := range e.TemporalTags {
		if strings.HasPrefix(t, validFromTagPrefix) {
			return true
		}
	}
	return false
}

// setValidUntilTag sets (or replaces) the valid_until TemporalTag, preserving
// all other tags including valid_from. Multiple prior valid_until tags collapse
// to a single replacement.
func setValidUntilTag(tags []string, until time.Time) []string {
	tag := validUntilTagPrefix + until.UTC().Format(time.RFC3339)
	out := make([]string, 0, len(tags)+1)
	replaced := false
	for _, t := range tags {
		if strings.HasPrefix(t, validUntilTagPrefix) {
			if !replaced {
				out = append(out, tag)
				replaced = true
			}
			continue
		}
		out = append(out, t)
	}
	if !replaced {
		out = append(out, tag)
	}
	return out
}
