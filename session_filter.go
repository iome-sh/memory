package memory

import "strings"

// ConvTagPrefix groups haystack sessions that belong to one conversation/palace
// retrieve key (T1). Retrieve with SessionID=conv_id matches entries tagged
// conv:<conv_id> even when MemoryEntry.SessionID is the inner haystack session.
const ConvTagPrefix = "conv:"

// ConvTag returns conv:<id>, or empty when id is blank.
func ConvTag(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	return ConvTagPrefix + id
}

// entryMatchesSessionFilter is true when no session filter is set, or the entry
// belongs to any requested session (exact SessionID or conv:<id> tag).
func entryMatchesSessionFilter(e MemoryEntry, sessionID string, sessionIDs []string) bool {
	ids := sessionFilterIDs(sessionID, sessionIDs)
	if len(ids) == 0 {
		return true
	}
	for _, id := range ids {
		if e.SessionID == id || EntryHasTag(e, ConvTag(id)) {
			return true
		}
	}
	return false
}

func metaMatchesSessionFilter(m entryMeta, sessionID string, sessionIDs []string) bool {
	ids := sessionFilterIDs(sessionID, sessionIDs)
	if len(ids) == 0 {
		return true
	}
	for _, id := range ids {
		if m.SessionID == id {
			return true
		}
		want := ConvTag(id)
		for _, t := range m.Tags {
			if t == want {
				return true
			}
		}
	}
	return false
}

func sessionFilterIDs(sessionID string, sessionIDs []string) []string {
	n := 0
	if strings.TrimSpace(sessionID) != "" {
		n++
	}
	for _, s := range sessionIDs {
		if strings.TrimSpace(s) != "" {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	out := make([]string, 0, n)
	if s := strings.TrimSpace(sessionID); s != "" {
		out = append(out, s)
	}
	seen := make(map[string]struct{}, n)
	if len(out) == 1 {
		seen[out[0]] = struct{}{}
	}
	for _, s := range sessionIDs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// diversifyBySession round-robins distinct SessionID groups before Limit so
// filler from one haystack session cannot occupy every slot (T1).
// Single-session result sets are unchanged (prefix Limit).
func isCountQuery(query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(q, "how many") || strings.Contains(q, "how much")
}

func isFactEntry(e MemoryEntry) bool {
	if e.Type == "turn_fact" || e.Type == "atomic_fact" {
		return true
	}
	return EntryHasTag(e, "fact_augmented")
}

// promoteFactEntries puts turn_fact / fact_augmented children first so count
// questions see extracted facts before haystack chatter (T1 assembly).
func promoteFactEntries(results []MemoryEntry) []MemoryEntry {
	return promoteFactEntriesForQuery(results, "")
}

// promoteFactEntriesForQuery puts turn_fact children first, ranked by keyword
// overlap with query so noisy fallback facts cannot bury "led two projects".
func promoteFactEntriesForQuery(results []MemoryEntry, query string) []MemoryEntry {
	if len(results) < 2 {
		return results
	}
	facts := make([]MemoryEntry, 0, len(results))
	rest := make([]MemoryEntry, 0, len(results))
	for _, e := range results {
		if isFactEntry(e) {
			facts = append(facts, e)
		} else {
			rest = append(rest, e)
		}
	}
	if len(facts) == 0 {
		return results
	}
	if strings.TrimSpace(query) != "" {
		facts = rankKeywordHitsByOverlap(facts, query)
	}
	return append(facts, rest...)
}

func diversifyBySession(entries []MemoryEntry, limit int) []MemoryEntry {
	if limit <= 0 || len(entries) <= limit {
		return entries
	}
	order := make([]string, 0, 4)
	by := make(map[string][]MemoryEntry, 4)
	for _, e := range entries {
		k := e.SessionID
		if k == "" {
			k = "_"
		}
		if _, ok := by[k]; !ok {
			order = append(order, k)
		}
		by[k] = append(by[k], e)
	}
	if len(order) < 2 {
		return entries[:limit]
	}
	out := make([]MemoryEntry, 0, limit)
	idx := make([]int, len(order))
	for len(out) < limit {
		progressed := false
		for i, k := range order {
			if idx[i] >= len(by[k]) {
				continue
			}
			out = append(out, by[k][idx[i]])
			idx[i]++
			progressed = true
			if len(out) >= limit {
				break
			}
		}
		if !progressed {
			break
		}
	}
	return out
}
