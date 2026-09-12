package memory

import (
	"sort"
	"strings"
)

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

// promoteFactEntriesForQuery puts turn_fact children first. On a count query,
// named-pattern facts (led+project, bought, spent, …) outrank fallback chatter
// even when keyword overlap is similar; remaining facts rank by stemmed overlap.
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
		facts = rankCountQueryFacts(facts, query)
	}
	return append(facts, rest...)
}

const maxCountEvidenceSnippets = 12

// countQueryStopwords are query tokens that inflate fallback-fact overlap on
// "how many / how much" questions without identifying the counted objects.
var countQueryStopwords = map[string]struct{}{
	"how": {}, "many": {}, "much": {}, "have": {}, "has": {}, "had": {},
	"did": {}, "does": {}, "the": {}, "and": {}, "for": {}, "from": {},
	"with": {}, "that": {}, "this": {}, "are": {}, "was": {}, "were": {},
	"been": {}, "being": {}, "currently": {}, "about": {}, "into": {},
	"just": {}, "also": {}, "than": {}, "then": {}, "they": {}, "them": {},
	"you": {}, "your": {}, "our": {}, "any": {}, "all": {}, "can": {},
	"could": {}, "would": {}, "should": {}, "will": {}, "what": {},
	"when": {}, "which": {}, "who": {}, "whom": {}, "whose": {}, "why": {},
	"need": {}, "do": {},
}

// countQueryNounTokens are content tokens for count-query fact matching:
// stopwords dropped, simple plural/gerund stems, led/lead/leading folded.
func countQueryNounTokens(query string) []string {
	raw := keywordTokens(query)
	seen := make(map[string]struct{}, len(raw)*3)
	var out []string
	add := func(s string) {
		if len(s) < 3 {
			return
		}
		if _, stop := countQueryStopwords[s]; stop {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	for _, t := range raw {
		if _, stop := countQueryStopwords[t]; stop {
			continue
		}
		add(t)
		if strings.HasSuffix(t, "ing") && len(t) > 6 {
			add(strings.TrimSuffix(t, "ing"))
		}
		if strings.HasSuffix(t, "es") && len(t) > 4 {
			add(strings.TrimSuffix(t, "es"))
		}
		if strings.HasSuffix(t, "s") && !strings.HasSuffix(t, "ss") && len(t) > 3 {
			add(strings.TrimSuffix(t, "s"))
		}
		switch t {
		case "led", "lead", "leading":
			add("led")
			add("lead")
			add("leading")
		}
	}
	return out
}

func isFirstPersonFact(text string) bool {
	lower := strings.ToLower(text)
	if strings.HasPrefix(lower, "i ") || strings.HasPrefix(lower, "i'm") || strings.HasPrefix(lower, "i've") {
		return true
	}
	return strings.Contains(lower, " i ") ||
		strings.Contains(lower, "i'm") ||
		strings.Contains(lower, "i've") ||
		strings.Contains(lower, "i'd") ||
		strings.Contains(lower, "i'll") ||
		strings.Contains(lower, " my ")
}

func factOverlapsCountQuery(e MemoryEntry, query string) bool {
	tokens := countQueryNounTokens(query)
	return keywordOverlapCount(e, tokens) > 0
}

func rankCountQueryFacts(facts []MemoryEntry, query string) []MemoryEntry {
	if len(facts) < 2 {
		return facts
	}
	tokens := countQueryNounTokens(query)
	type scored struct {
		e           MemoryEntry
		named       int
		firstPerson int
		overlap     int
	}
	tmp := make([]scored, len(facts))
	for i, e := range facts {
		hay := entryKeywordHaystack(e)
		named := 0
		if matchesNamedFactPattern(hay) {
			named = 1
		}
		fp := 0
		if isFirstPersonFact(hay) {
			fp = 1
		}
		tmp[i] = scored{e: e, named: named, firstPerson: fp, overlap: keywordOverlapCount(e, tokens)}
	}
	sort.SliceStable(tmp, func(i, j int) bool {
		if tmp[i].named != tmp[j].named {
			return tmp[i].named > tmp[j].named
		}
		if tmp[i].firstPerson != tmp[j].firstPerson {
			return tmp[i].firstPerson > tmp[j].firstPerson
		}
		return tmp[i].overlap > tmp[j].overlap
	})
	out := make([]MemoryEntry, len(tmp))
	for i, s := range tmp {
		out[i] = s.e
	}
	return out
}

// unionCountQueryFacts adds turn_fact / fact_augmented / atomic_fact children
// from the palace/session candidate set whose text overlaps count-query nouns
// (stemmed). Keyword search alone can miss "solo project" when the query says
// "projects".
func unionCountQueryFacts(hits, candidates []MemoryEntry, query string) []MemoryEntry {
	if !isCountQuery(query) {
		return hits
	}
	seen := make(map[string]struct{}, len(hits)+len(candidates))
	out := make([]MemoryEntry, 0, len(hits)+8)
	for _, e := range hits {
		if e.ID != "" {
			if _, ok := seen[e.ID]; ok {
				continue
			}
			seen[e.ID] = struct{}{}
		}
		out = append(out, e)
	}
	for _, e := range candidates {
		if !isFactEntry(e) {
			continue
		}
		if e.ID != "" {
			if _, ok := seen[e.ID]; ok {
				continue
			}
		}
		if !factOverlapsCountQuery(e, query) {
			continue
		}
		if e.ID != "" {
			seen[e.ID] = struct{}{}
		}
		out = append(out, e)
	}
	return out
}

// AssembleCountEvidence lists unique matching fact snippets for a count query
// (named-pattern first, then first-person noun overlap). Deduped, not persisted.
// Empty when the query is not a count question or no facts match.
func AssembleCountEvidence(query string, facts []MemoryEntry) string {
	if !isCountQuery(query) {
		return ""
	}
	matched := make([]MemoryEntry, 0, len(facts))
	for _, e := range facts {
		if !isFactEntry(e) {
			continue
		}
		if !factOverlapsCountQuery(e, query) {
			continue
		}
		hay := entryKeywordHaystack(e)
		if !matchesNamedFactPattern(hay) && !isFirstPersonFact(hay) {
			continue
		}
		matched = append(matched, e)
	}
	if len(matched) == 0 {
		return ""
	}
	matched = rankCountQueryFacts(matched, query)
	seen := make(map[string]struct{}, len(matched))
	snippets := make([]string, 0, len(matched))
	for _, e := range matched {
		text := strings.TrimSpace(e.Content.Summary)
		if text == "" {
			text = strings.TrimSpace(e.Content.Full)
		}
		if text == "" {
			continue
		}
		key := strings.ToLower(text)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		snippets = append(snippets, text)
		if len(snippets) >= maxCountEvidenceSnippets {
			break
		}
	}
	if len(snippets) == 0 {
		return ""
	}
	return "Count evidence:\n- " + strings.Join(snippets, "\n- ")
}

// diversifyBySession round-robins distinct SessionID groups before Limit so
// filler from one haystack session cannot occupy every slot (T1).
// Single-session result sets are unchanged (prefix Limit).
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
