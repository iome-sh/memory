package memory

import (
	"regexp"
	"sort"
	"strings"
	"time"
)

const maxLatestValueEvidenceSnippets = 12

var (
	reDollarAmount = regexp.MustCompile(`\$\s*(?:\d{1,3}(?:,\d{3})+|\d+)(?:\.\d{2})?`)
	reDollarK      = regexp.MustCompile(`(?i)\$\s*\d+(?:\.\d+)?\s*k\b`)
	reNDollars     = regexp.MustCompile(`(?i)\b(\d{1,3}(?:,\d{3})+)\s*dollars?\b`)
)

// latestValueStop drops question-framing tokens so entity nouns remain
// (Wells Fargo, mortgage). "amount" is framing, not an entity.
var latestValueStop = map[string]struct{}{
	"how": {}, "many": {}, "much": {}, "have": {}, "has": {}, "had": {},
	"did": {}, "does": {}, "the": {}, "and": {}, "for": {}, "from": {},
	"with": {}, "that": {}, "this": {}, "are": {}, "was": {}, "were": {},
	"been": {}, "being": {}, "currently": {}, "current": {}, "about": {},
	"into": {}, "just": {}, "also": {}, "than": {}, "then": {}, "they": {},
	"them": {}, "you": {}, "your": {}, "our": {}, "any": {}, "all": {},
	"can": {}, "could": {}, "would": {}, "should": {}, "will": {}, "what": {},
	"when": {}, "which": {}, "who": {}, "whom": {}, "whose": {}, "why": {},
	"need": {}, "do": {}, "got": {}, "get": {}, "amount": {}, "now": {},
}

// isLatestValueQuery is true for amount / pre-approved / how-much-was-I
// questions whose gold is the latest matching scalar. Narrow on purpose:
// "how many kits" stays on the count path; which-first / dated-span stay
// temporal. Does not NLP-supersede; retrieve ranks later Timestamp first.
func isLatestValueQuery(query string) bool {
	if isTemporalEvidenceQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if strings.Contains(q, "how many") {
		return false
	}
	if strings.Contains(q, "pre-approved") || strings.Contains(q, "preapproved") ||
		strings.Contains(q, "pre approved") {
		return true
	}
	if strings.Contains(q, "the amount") {
		return true
	}
	if strings.Contains(q, "how much was i") || strings.Contains(q, "how much am i") ||
		strings.Contains(q, "how much did i") {
		return true
	}
	valueNoun := strings.Contains(q, "amount") || strings.Contains(q, "salary") ||
		strings.Contains(q, "price") || strings.Contains(q, "dollar") ||
		strings.Contains(q, "approved") || strings.Contains(q, "mortgage")
	recency := strings.Contains(q, "currently") || strings.Contains(q, "current") ||
		strings.Contains(q, " now") || strings.HasPrefix(q, "now ")
	return valueNoun && recency
}

func latestValueNeedles(query string) []string {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	seen := make(map[string]struct{}, 16)
	var out []string
	add := func(s string) {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.Trim(s, " .,;:!?")
		if len(s) < 4 {
			return
		}
		if _, stop := latestValueStop[s]; stop {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if strings.Contains(q, "pre-approved") || strings.Contains(q, "preapproved") ||
		strings.Contains(q, "pre approved") {
		add("pre-approved")
	}
	for _, t := range keywordTokens(q) {
		if _, stop := latestValueStop[t]; stop {
			continue
		}
		add(t)
	}
	return out
}

func hasDollarAmount(text string) bool {
	return reDollarAmount.MatchString(text) || reDollarK.MatchString(text) || reNDollars.MatchString(text)
}

func extractDollarAmounts(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	seen := make(map[string]struct{}, 4)
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		key := normalizeAmountKey(s)
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	for _, m := range reDollarAmount.FindAllString(text, -1) {
		add(m)
	}
	for _, m := range reDollarK.FindAllString(text, -1) {
		add(m)
	}
	for _, m := range reNDollars.FindAllString(text, -1) {
		add(m)
	}
	return out
}

func normalizeAmountKey(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if r == 'k' && b.Len() > 0 {
			b.WriteString("000")
		}
	}
	return b.String()
}

func entryMatchesLatestValue(e MemoryEntry, needles []string) bool {
	return latestValueHayMatches(entryKeywordHaystack(e), needles)
}

func latestValueHayMatches(hay string, needles []string) bool {
	if strings.TrimSpace(hay) == "" || len(needles) == 0 {
		return false
	}
	if isAssistantChatter(hay) {
		return false
	}
	if !hasDollarAmount(hay) {
		return false
	}
	return sentenceMentionsNeedles(hay, needles)
}

// unionLatestValueEntries adds session/conv candidates that mention the query
// entity and a dollar (or similar) scalar so a later haystack session still
// reaches Limit. Additive: does not filter the existing hit list.
func unionLatestValueEntries(hits, candidates []MemoryEntry, query string) []MemoryEntry {
	if !isLatestValueQuery(query) {
		return hits
	}
	needles := latestValueNeedles(query)
	if len(needles) == 0 {
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
		if e.ID != "" {
			if _, ok := seen[e.ID]; ok {
				continue
			}
		}
		if !entryMatchesLatestValue(e, needles) {
			continue
		}
		if e.ID != "" {
			seen[e.ID] = struct{}{}
		}
		out = append(out, e)
	}
	return out
}

// promoteLatestValueEntries puts entity+amount matches first, later Timestamp
// before earlier (Nov session before Aug). DiversifyBySession still runs
// before Limit. Does not drop the stale value (not NLP supersession).
func promoteLatestValueEntries(results []MemoryEntry, query string) []MemoryEntry {
	if len(results) < 2 || !isLatestValueQuery(query) {
		return results
	}
	needles := latestValueNeedles(query)
	if len(needles) == 0 {
		return results
	}
	match := make([]MemoryEntry, 0, len(results))
	rest := make([]MemoryEntry, 0, len(results))
	for _, e := range results {
		if entryMatchesLatestValue(e, needles) {
			match = append(match, e)
		} else {
			rest = append(rest, e)
		}
	}
	if len(match) == 0 {
		return results
	}
	sort.SliceStable(match, func(i, j int) bool {
		ti := entryEventTime(match[i])
		tj := entryEventTime(match[j])
		if !ti.Equal(tj) {
			return ti.After(tj)
		}
		return false
	})
	return append(match, rest...)
}

// AssembleLatestValueEvidence lists unique matching amount snippets for a
// latest-value query, labeled with entry Timestamp, latest first. Deduped by
// normalized dollar amount (keep the later mention). Empty when the query is
// not a latest-value question or no matching scalars exist. Not persisted.
func AssembleLatestValueEvidence(query string, entries []MemoryEntry) string {
	if !isLatestValueQuery(query) {
		return ""
	}
	needles := latestValueNeedles(query)
	if len(needles) == 0 {
		return ""
	}
	type bullet struct {
		snippet string
		when    time.Time
		order   int
	}
	byAmount := make(map[string]bullet, 8)
	n := 0
	for _, e := range entries {
		hay := entryKeywordHaystack(e)
		if !latestValueHayMatches(hay, needles) {
			continue
		}
		when := entryEventTime(e)
		for _, sent := range splitEvidenceSentences(hay) {
			if len(needles) > 0 && !sentenceMentionsNeedles(sent, needles) {
				continue
			}
			amounts := extractDollarAmounts(sent)
			if len(amounts) == 0 {
				continue
			}
			snip := strings.TrimSpace(sent)
			if len(snip) > 280 {
				snip = snip[:280]
			}
			labeled := formatLatestValueLabel(when) + " " + snip
			for _, amt := range amounts {
				key := normalizeAmountKey(amt)
				if key == "" {
					continue
				}
				prev, ok := byAmount[key]
				if ok && !when.After(prev.when) {
					continue
				}
				order := n
				if ok {
					order = prev.order
				} else {
					n++
				}
				byAmount[key] = bullet{snippet: labeled, when: when, order: order}
			}
		}
	}
	if len(byAmount) == 0 {
		return ""
	}
	list := make([]bullet, 0, len(byAmount))
	for _, b := range byAmount {
		list = append(list, b)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if !list[i].when.Equal(list[j].when) {
			return list[i].when.After(list[j].when)
		}
		return list[i].order < list[j].order
	})
	out := make([]string, 0, len(list))
	for _, b := range list {
		out = append(out, b.snippet)
		if len(out) >= maxLatestValueEvidenceSnippets {
			break
		}
	}
	return formatEvidenceBlock("Latest-value evidence", len(out), "values", out)
}

func formatLatestValueLabel(ts time.Time) string {
	if ts.IsZero() {
		return "[time: unknown]"
	}
	return "[time: " + ts.UTC().Format(time.RFC3339) + "]"
}
