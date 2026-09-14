package memory

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const maxTemporalEvidenceSnippets = 12

var (
	reQuotedEvent = regexp.MustCompile(`['"]([^'"]{3,80})['"]`)
	reBetweenAnd  = regexp.MustCompile(`(?i)between\s+(.+?)\s+and\s+(.+?)(?:\?|$|\.)`)
	reMonthDay    = regexp.MustCompile(`(?i)\b(january|february|march|april|may|june|july|august|september|october|november|december)\s+(\d{1,2})(?:st|nd|rd|th)?\b`)
	reNAgo        = regexp.MustCompile(`(?i)\b(\d+|` + wordNumberAlt + `)\s+(days?|weeks?|months?|years?)\s+ago\b`)
	reLastWeekday = regexp.MustCompile(`(?i)\blast\s+(monday|tuesday|wednesday|thursday|friday|saturday|sunday|week|month|year)\b`)
	reSlashDate   = regexp.MustCompile(`\b(\d{1,2})/(\d{1,2})(?:/(\d{2,4}))?\b`)
	reScaleFrac   = regexp.MustCompile(`(?i)1/\d+\s*scale`)
)

var monthNameToNum = map[string]time.Month{
	"january": time.January, "february": time.February, "march": time.March,
	"april": time.April, "may": time.May, "june": time.June, "july": time.July,
	"august": time.August, "september": time.September, "october": time.October,
	"november": time.November, "december": time.December,
}

var weekdayNameToNum = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
	"wednesday": time.Wednesday, "thursday": time.Thursday, "friday": time.Friday,
	"saturday": time.Saturday,
}

// temporalNeedleStop drops question-framing tokens so event names remain.
var temporalNeedleStop = map[string]struct{}{
	"how": {}, "many": {}, "much": {}, "have": {}, "has": {}, "had": {},
	"did": {}, "does": {}, "the": {}, "and": {}, "for": {}, "from": {},
	"with": {}, "that": {}, "this": {}, "are": {}, "was": {}, "were": {},
	"been": {}, "being": {}, "about": {}, "into": {}, "just": {}, "also": {},
	"than": {}, "then": {}, "they": {}, "them": {}, "you": {}, "your": {},
	"our": {}, "any": {}, "all": {}, "can": {}, "could": {}, "would": {},
	"should": {}, "will": {}, "what": {}, "when": {}, "which": {}, "who": {},
	"whom": {}, "whose": {}, "why": {}, "need": {}, "do": {}, "days": {},
	"day": {}, "passed": {}, "pass": {}, "take": {}, "took": {}, "event": {},
	"events": {}, "attend": {}, "attended": {}, "first": {}, "between": {},
	"after": {}, "before": {}, "using": {}, "find": {},
	"finding": {}, "combined": {}, "total": {}, "spend": {}, "spent": {},
	"hours": {}, "hour": {}, "week": {}, "weeks": {}, "month": {}, "months": {}, "year": {}, "ago": {},
	"last": {}, "past": {}, "since": {}, "until": {}, "apart": {},
}

// isTemporalOrderQuery is true for “which … first” / before / after event-order
// questions. Ranking by MemoryEntry.Timestamp alone is wrong when the gold
// order lives in relative dates in the text.
func isTemporalOrderQuery(query string) bool {
	q := strings.ToLower(query)
	if strings.Contains(q, "which") && strings.Contains(q, "first") {
		return true
	}
	if strings.Contains(q, "which event") {
		return true
	}
	if strings.Contains(q, "did i attend first") || strings.Contains(q, "did i do first") {
		return true
	}
	if strings.Contains(q, " before ") || strings.HasPrefix(q, "before ") {
		return true
	}
	if strings.Contains(q, " after ") || strings.HasPrefix(q, "after ") {
		return true
	}
	return false
}

// isDatedSpanQuery is true for “how many days/weeks/months between / did it
// take / ago” questions. These stay out of calendar-window filters
// (classifyTemporalIntent already skips how-many). Count-evidence clustering
// is skipped so day-quantities do not hide dated event bullets.
func isDatedSpanQuery(query string) bool {
	q := strings.ToLower(query)
	if !strings.Contains(q, "how many") {
		return false
	}
	if strings.Contains(q, "between") {
		return true
	}
	if strings.Contains(q, "did it take") || strings.Contains(q, "does it take") {
		return true
	}
	if strings.Contains(q, "days") && (strings.Contains(q, "passed") || strings.Contains(q, "apart") ||
		strings.Contains(q, "until") || strings.Contains(q, "since") || strings.Contains(q, "ago")) {
		return true
	}
	if (strings.Contains(q, "weeks") || strings.Contains(q, "months")) && datedSpanUnitCue(q) {
		return true
	}
	return false
}

func datedSpanUnitCue(q string) bool {
	return strings.Contains(q, "between") || strings.Contains(q, "passed") ||
		strings.Contains(q, "since") || strings.Contains(q, "apart") ||
		strings.Contains(q, "until") || strings.Contains(q, "have i been") ||
		strings.Contains(q, "had passed") || strings.Contains(q, "have i been taking") ||
		strings.Contains(q, "ago")
}

func isTemporalEvidenceQuery(query string) bool {
	return isTemporalOrderQuery(query) || isDatedSpanQuery(query)
}

// AssembleTemporalEvidence lists unique dated-event bullets for temporal-order
// or dated-span queries. Text date phrases are labeled separately from ingest
// Timestamp (RFC3339) so the reader can see relative/absolute dates vs session
// time. Sorted by parsed text time when available, else Timestamp. Multiple
// bullets are prefixed with the cluster count ("N distinct events"). For
// dated-span queries with two or more parsed text times, extra lines report
// the UTC calendar-day difference (`text dates N days apart (phrase → phrase)`)
// and, when the query asks how-many-weeks or how-many-months, a floor week
// delta and a calendar-month delta. When the query contains "ago" and
// questionDate is non-zero, a further line reports the latest parsed text
// time versus that instant (`text date PHRASE is N days before question_date
// YYYY-MM-DD`), plus floor weeks / calendar months when asked. Temporal-order
// queries with two or more parsed text times append `text dates earliest:
// PHRASE · latest: PHRASE`. Span delta is emitted first; question_date next;
// extrema only when isTemporalOrderQuery is still true. Arithmetic only; not
// a gold answer; never ingest Timestamp. questionDate is evidence-only (not a
// retrieve TimeTo/AsOf filter). Empty when the query is not temporal or no
// dated snippets match. Not persisted.
func AssembleTemporalEvidence(query string, entries []MemoryEntry) string {
	return assembleTemporalEvidence(query, entries, time.Time{})
}

// AssembleTemporalEvidenceAt is AssembleTemporalEvidence with a LongMemEval
// question_date. Zero questionDate is the same as AssembleTemporalEvidence.
func AssembleTemporalEvidenceAt(query string, entries []MemoryEntry, questionDate time.Time) string {
	return assembleTemporalEvidence(query, entries, questionDate)
}

func assembleTemporalEvidence(query string, entries []MemoryEntry, questionDate time.Time) string {
	if !isTemporalEvidenceQuery(query) {
		return ""
	}
	needles := queryEventNeedles(query)
	seen := make(map[string]struct{}, 8)
	list := make([]temporalBullet, 0, 8)
	n := 0
	for _, e := range entries {
		hay := entryKeywordHaystack(e)
		if strings.TrimSpace(hay) == "" {
			continue
		}
		for _, sent := range splitEvidenceSentences(hay) {
			if !sentenceFitsTemporalQuery(sent, needles) {
				continue
			}
			hits := extractDatedPhrases(sent, e.Timestamp)
			if len(hits) == 0 {
				continue
			}
			snip := strings.TrimSpace(sent)
			if len(snip) > 280 {
				snip = snip[:280]
			}
			eventKey := datedEventKey(sent, needles)
			for _, h := range hits {
				key := temporalBulletKey(h, eventKey)
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				list = append(list, temporalBullet{
					snippet:     formatTemporalLabel(h.phrase, e.Timestamp) + " " + snip,
					textPhrase:  h.phrase,
					textTime:    h.when,
					hasTextTime: !h.when.IsZero(),
					ingest:      e.Timestamp,
					order:       n,
				})
				n++
			}
		}
	}
	if len(list) == 0 {
		return ""
	}
	sort.SliceStable(list, func(i, j int) bool {
		ti := temporalSortTime(list[i].hasTextTime, list[i].textTime, list[i].ingest)
		tj := temporalSortTime(list[j].hasTextTime, list[j].textTime, list[j].ingest)
		if !ti.Equal(tj) {
			return ti.Before(tj)
		}
		return list[i].order < list[j].order
	})
	if len(list) > maxTemporalEvidenceSnippets {
		list = list[:maxTemporalEvidenceSnippets]
	}
	out := make([]string, 0, len(list))
	for _, b := range list {
		out = append(out, b.snippet)
	}
	var body string
	if len(out) > 1 {
		body = formatEvidenceBlock("Temporal evidence", len(out), "events", out)
	} else {
		body = "Temporal evidence:\n- " + strings.Join(out, "\n- ")
	}
	var extras []string
	if delta := datedSpanTextDelta(query, list); delta != "" {
		extras = append(extras, delta)
	}
	if ago := datedSpanQuestionDateDelta(query, list, questionDate); ago != "" {
		extras = append(extras, ago)
	}
	if extrema := temporalOrderTextExtrema(query, list); extrema != "" {
		extras = append(extras, extrema)
	}
	if len(extras) > 0 {
		return body + "\n" + strings.Join(extras, "\n")
	}
	return body
}

type temporalBullet struct {
	snippet     string
	textPhrase  string
	textTime    time.Time
	hasTextTime bool
	ingest      time.Time
	order       int
}

// temporalOrderTextExtrema labels the earliest and latest parsed text dates
// for which-first / before / after queries. Uses the already-sorted bullet
// list; never ingest Timestamp; not a gold answer.
func temporalOrderTextExtrema(query string, bullets []temporalBullet) string {
	if !isTemporalOrderQuery(query) {
		return ""
	}
	dated := make([]temporalBullet, 0, len(bullets))
	for _, b := range bullets {
		if b.hasTextTime {
			dated = append(dated, b)
		}
	}
	if len(dated) < 2 {
		return ""
	}
	earliest, latest := dated[0], dated[len(dated)-1]
	from := textDatePhrase(earliest.textPhrase, earliest.textTime)
	to := textDatePhrase(latest.textPhrase, latest.textTime)
	return "text dates earliest: " + from + " · latest: " + to
}

func datedSpanTextDelta(query string, bullets []temporalBullet) string {
	if !isDatedSpanQuery(query) {
		return ""
	}
	dated := make([]temporalBullet, 0, len(bullets))
	for _, b := range bullets {
		if b.hasTextTime {
			dated = append(dated, b)
		}
	}
	if len(dated) < 2 {
		return ""
	}
	earliest, latest := dated[0], dated[len(dated)-1]
	from := textDatePhrase(earliest.textPhrase, earliest.textTime)
	to := textDatePhrase(latest.textPhrase, latest.textTime)
	days := utcCalendarDayDelta(earliest.textTime, latest.textTime)
	lines := []string{"text dates " + strconv.Itoa(days) + " days apart (" + from + " → " + to + ")"}
	q := strings.ToLower(query)
	if strings.Contains(q, "weeks") {
		n, r := days/7, days%7
		week := "text dates " + strconv.Itoa(n) + " weeks apart (floor days/7"
		if r != 0 {
			week += "; remainder " + strconv.Itoa(r) + " days"
		}
		week += ")"
		lines = append(lines, week)
	}
	if strings.Contains(q, "months") {
		m := utcCalendarMonthDelta(earliest.textTime, latest.textTime)
		lines = append(lines, "text dates "+strconv.Itoa(m)+" calendar months apart ("+from+" → "+to+")")
	}
	return strings.Join(lines, "\n")
}

// datedSpanQuestionDateDelta reports the latest parsed text time versus
// questionDate for dated-span "ago" queries. Uses text times only; never
// ingest Timestamp; not a gold answer. Empty when questionDate is zero,
// the query lacks "ago", or no dated bullet exists.
func datedSpanQuestionDateDelta(query string, bullets []temporalBullet, questionDate time.Time) string {
	if questionDate.IsZero() || !isDatedSpanQuery(query) {
		return ""
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "ago") {
		return ""
	}
	var latest temporalBullet
	found := false
	for _, b := range bullets {
		if !b.hasTextTime {
			continue
		}
		if !found || b.textTime.After(latest.textTime) {
			latest = b
			found = true
		}
	}
	if !found {
		return ""
	}
	phrase := textDatePhrase(latest.textPhrase, latest.textTime)
	days := utcCalendarDayDelta(latest.textTime, questionDate)
	rel := "before"
	if calendarDayUTC(latest.textTime).After(calendarDayUTC(questionDate)) {
		rel = "after"
	}
	qd := calendarDayUTC(questionDate).Format("2006-01-02")
	lines := []string{"text date " + phrase + " is " + strconv.Itoa(days) + " days " + rel + " question_date " + qd}
	if strings.Contains(q, "weeks") {
		n, r := days/7, days%7
		week := "text date " + phrase + " is " + strconv.Itoa(n) + " weeks " + rel + " question_date " + qd + " (floor days/7"
		if r != 0 {
			week += "; remainder " + strconv.Itoa(r) + " days"
		}
		week += ")"
		lines = append(lines, week)
	}
	if strings.Contains(q, "months") {
		m := utcCalendarMonthDelta(latest.textTime, questionDate)
		lines = append(lines, "text date "+phrase+" is "+strconv.Itoa(m)+" calendar months "+rel+" question_date "+qd)
	}
	return strings.Join(lines, "\n")
}

func utcCalendarDayDelta(a, b time.Time) int {
	ad := calendarDayUTC(a)
	bd := calendarDayUTC(b)
	if ad.After(bd) {
		ad, bd = bd, ad
	}
	return int(bd.Sub(ad) / (24 * time.Hour))
}

func utcCalendarMonthDelta(a, b time.Time) int {
	au := a.UTC()
	bu := b.UTC()
	am := au.Year()*12 + int(au.Month())
	bm := bu.Year()*12 + int(bu.Month())
	if am > bm {
		am, bm = bm, am
	}
	return bm - am
}

func calendarDayUTC(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

func textDatePhrase(phrase string, when time.Time) string {
	if p := strings.TrimSpace(phrase); p != "" {
		return p
	}
	if when.IsZero() {
		return "unknown"
	}
	return when.UTC().Format("2006-01-02")
}

func temporalSortTime(hasText bool, textTime, ingest time.Time) time.Time {
	if hasText {
		return textTime
	}
	return ingest
}

func temporalBulletKey(h datedPhrase, eventKey string) string {
	if !h.when.IsZero() {
		return h.when.Format("2006-01-02") + "\t" + eventKey
	}
	return strings.ToLower(strings.TrimSpace(h.phrase)) + "\t" + eventKey
}

func formatTemporalLabel(phrase string, ingest time.Time) string {
	phrase = strings.TrimSpace(phrase)
	if phrase == "" {
		if ingest.IsZero() {
			return "[text: unknown]"
		}
		return "[text: unknown | ingest: " + ingest.UTC().Format(time.RFC3339) + "]"
	}
	if ingest.IsZero() {
		return "[text: " + phrase + "]"
	}
	return "[text: " + phrase + " | ingest: " + ingest.UTC().Format(time.RFC3339) + "]"
}

func sentenceFitsTemporalQuery(sent string, needles []string) bool {
	if !isFirstPersonFact(sent) && len(needles) == 0 {
		return false
	}
	if len(needles) == 0 {
		return true
	}
	return sentenceMentionsNeedles(sent, needles)
}

func sentenceMentionsNeedles(sent string, needles []string) bool {
	lower := strings.ToLower(sent)
	for _, n := range needles {
		if n != "" && strings.Contains(lower, n) {
			return true
		}
	}
	return false
}

func datedEventKey(sent string, needles []string) string {
	lower := strings.ToLower(sent)
	for _, n := range needles {
		if n != "" && strings.Contains(lower, n) {
			return n
		}
	}
	return ""
}

func queryEventNeedles(query string) []string {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil
	}
	seen := make(map[string]struct{}, 16)
	var out []string
	add := func(s string) {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.Trim(s, " .,;:!?")
		if len(s) < 3 {
			return
		}
		if _, stop := temporalNeedleStop[s]; stop {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	for _, m := range reQuotedEvent.FindAllStringSubmatch(q, -1) {
		add(m[1])
	}
	if m := reBetweenAnd.FindStringSubmatch(q); len(m) == 3 {
		add(m[1])
		add(m[2])
	}
	for _, t := range keywordTokens(q) {
		if _, stop := temporalNeedleStop[t]; stop {
			continue
		}
		if len(t) >= 4 {
			add(t)
		}
	}
	return out
}

func entryMentionsAnyNeedle(e MemoryEntry, needles []string) bool {
	if len(needles) == 0 {
		return false
	}
	hay := strings.ToLower(entryKeywordHaystack(e))
	for _, n := range needles {
		if n != "" && strings.Contains(hay, n) {
			return true
		}
	}
	return false
}

// unionTemporalEventEntries adds candidates that mention query event names
// (quoted titles, between-X-and-Y, leftover content tokens) so Limit cannot
// drop the other event. Additive: does not filter the existing hit list.
func unionTemporalEventEntries(hits, candidates []MemoryEntry, query string) []MemoryEntry {
	if !isTemporalEvidenceQuery(query) {
		return hits
	}
	needles := queryEventNeedles(query)
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
		if !entryMentionsAnyNeedle(e, needles) {
			continue
		}
		if e.ID != "" {
			seen[e.ID] = struct{}{}
		}
		out = append(out, e)
	}
	return out
}

// promoteTemporalEventEntries puts entries that mention either event name
// ahead of the rest (after keyword/vector). DiversifyBySession still runs
// before Limit.
func promoteTemporalEventEntries(results []MemoryEntry, query string) []MemoryEntry {
	if len(results) < 2 {
		return results
	}
	needles := queryEventNeedles(query)
	if len(needles) == 0 {
		return results
	}
	hit := make([]MemoryEntry, 0, len(results))
	rest := make([]MemoryEntry, 0, len(results))
	for _, e := range results {
		if entryMentionsAnyNeedle(e, needles) {
			hit = append(hit, e)
		} else {
			rest = append(rest, e)
		}
	}
	if len(hit) == 0 {
		return results
	}
	return append(hit, rest...)
}

type datedPhrase struct {
	phrase string
	when   time.Time
}

func extractDatedPhrases(text string, ref time.Time) []datedPhrase {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	type span struct {
		start, end int
		hit        datedPhrase
	}
	var spans []span
	add := func(start, end int, phrase string, when time.Time) {
		if start < 0 || end <= start {
			return
		}
		phrase = strings.TrimSpace(phrase)
		if phrase == "" {
			return
		}
		spans = append(spans, span{start: start, end: end, hit: datedPhrase{phrase: phrase, when: when}})
	}

	for _, m := range reNAgo.FindAllStringSubmatchIndex(text, -1) {
		phrase := text[m[0]:m[1]]
		n := parseWordOrDigit(text[m[2]:m[3]])
		unit := strings.ToLower(text[m[4]:m[5]])
		when := time.Time{}
		if !ref.IsZero() {
			when = ref.AddDate(agoDelta(n, unit))
		}
		add(m[0], m[1], phrase, when)
	}
	for _, m := range reLastWeekday.FindAllStringSubmatchIndex(text, -1) {
		phrase := text[m[0]:m[1]]
		word := strings.ToLower(text[m[2]:m[3]])
		add(m[0], m[1], phrase, lastWeekdayOrUnit(ref, word))
	}
	for _, m := range reMonthDay.FindAllStringSubmatchIndex(text, -1) {
		phrase := text[m[0]:m[1]]
		month := monthNameToNum[strings.ToLower(text[m[2]:m[3]])]
		day := parseWordOrDigit(text[m[4]:m[5]])
		add(m[0], m[1], phrase, dateOnRefYear(ref, month, day))
	}

	skip := reScaleFrac.FindAllStringIndex(text, -1)
	for _, m := range reSlashDate.FindAllStringSubmatchIndex(text, -1) {
		if indexInSpans(m[0], skip) {
			continue
		}
		month := parseWordOrDigit(text[m[2]:m[3]])
		day := parseWordOrDigit(text[m[4]:m[5]])
		if month < 1 || month > 12 || day < 1 || day > 31 {
			continue
		}
		year := 0
		if m[6] >= 0 && m[7] > m[6] {
			year = parseWordOrDigit(text[m[6]:m[7]])
			if year < 100 {
				year += 2000
			}
		}
		phrase := text[m[0]:m[1]]
		when := slashDateOnRef(ref, month, day, year)
		add(m[0], m[1], phrase, when)
	}

	if len(spans) == 0 {
		return nil
	}
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].start != spans[j].start {
			return spans[i].start < spans[j].start
		}
		return (spans[i].end - spans[i].start) > (spans[j].end - spans[j].start)
	})
	used := make([]bool, len(text))
	out := make([]datedPhrase, 0, len(spans))
	seenPhrase := make(map[string]struct{}, len(spans))
	for _, s := range spans {
		overlap := false
		for i := s.start; i < s.end && i < len(used); i++ {
			if used[i] {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}
		key := strings.ToLower(s.hit.phrase)
		if _, ok := seenPhrase[key]; ok {
			continue
		}
		seenPhrase[key] = struct{}{}
		for i := s.start; i < s.end && i < len(used); i++ {
			used[i] = true
		}
		out = append(out, s.hit)
	}
	return out
}

func indexInSpans(i int, spans [][]int) bool {
	for _, s := range spans {
		if len(s) >= 2 && i >= s[0] && i < s[1] {
			return true
		}
	}
	return false
}

func agoDelta(n int, unit string) (years, months, days int) {
	if n <= 0 {
		n = 1
	}
	switch {
	case strings.HasPrefix(unit, "year"):
		return -n, 0, 0
	case strings.HasPrefix(unit, "month"):
		return 0, -n, 0
	case strings.HasPrefix(unit, "week"):
		return 0, 0, -7 * n
	default:
		return 0, 0, -n
	}
}

func lastWeekdayOrUnit(ref time.Time, word string) time.Time {
	if ref.IsZero() {
		return time.Time{}
	}
	switch word {
	case "week":
		return ref.AddDate(0, 0, -7)
	case "month":
		return ref.AddDate(0, -1, 0)
	case "year":
		return ref.AddDate(-1, 0, 0)
	}
	wd, ok := weekdayNameToNum[word]
	if !ok {
		return time.Time{}
	}
	d := ref
	for i := 0; i < 7; i++ {
		d = d.AddDate(0, 0, -1)
		if d.Weekday() == wd {
			return d
		}
	}
	return time.Time{}
}

func dateOnRefYear(ref time.Time, month time.Month, day int) time.Time {
	if month == 0 || day < 1 || day > 31 {
		return time.Time{}
	}
	year := 0
	if !ref.IsZero() {
		year = ref.Year()
	}
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if !ref.IsZero() && t.After(ref) {
		t = time.Date(year-1, month, day, 0, 0, 0, 0, time.UTC)
	}
	return t
}

func slashDateOnRef(ref time.Time, month, day, year int) time.Time {
	if year <= 0 {
		return dateOnRefYear(ref, time.Month(month), day)
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func splitEvidenceSentences(text string) []string {
	parts := strings.Split(text, ". ")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{strings.TrimSpace(text)}
	}
	return out
}
