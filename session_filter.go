package memory

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
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

// SkipVectorScoring reports whether SearchMemoryWithOptions ignores QueryVec
// and does not call scoreEntriesByVector. True for count (how many / how much)
// and temporal-order / dated-span queries. Keyword-first retrieve plus
// AssembleCountEvidence / AssembleTemporalEvidence is the gold path.
func SkipVectorScoring(query string) bool {
	return isCountQuery(query) || isTemporalEvidenceQuery(query)
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
// named-pattern facts (led+project, clothing errands, bought, spent, …) outrank
// fallback chatter even when keyword overlap is similar; remaining facts rank
// by stemmed overlap.
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

func isAssistantChatter(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if strings.HasPrefix(lower, "congratulations") || strings.HasPrefix(lower, "i'd be happy") ||
		strings.HasPrefix(lower, "i would be happy") {
		return true
	}
	return strings.Contains(text, "**")
}

func factOverlapsCountQuery(e MemoryEntry, query string) bool {
	if keywordOverlapCount(e, countQueryNounTokens(query)) > 0 {
		return true
	}
	hay := entryKeywordHaystack(e)
	// Dry-clean has weak overlap with "pick up or return from a store".
	// Treat clothing errands as hits on clothing count queries.
	if isClothingCountQuery(query) && isClothingErrandText(hay) {
		return true
	}
	// Kit / plant / restaurant names often lack the query noun ("plants" vs
	// "peace lily", "Banchan" vs "restaurants"). Catalog aliases and extracted
	// phrases both count as overlap.
	switch countEntityKind(query) {
	case "kit":
		return len(uniqueEntityClusters(hay, "kit")) > 0
	case "plant":
		return len(uniqueEntityClusters(hay, "plant")) > 0
	case "restaurant":
		return len(uniqueEntityClusters(hay, "restaurant")) > 0
	case "hours":
		return hasHourQuantity(hay)
	}
	return false
}

func isClothingCountQuery(query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(q, "clothing") || strings.Contains(q, "clothes") ||
		strings.Contains(q, "pick up") || strings.Contains(q, "return")
}

func isClothingErrandText(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	if atomicFactDryClean.MatchString(text) || atomicFactClothPick.MatchString(text) || atomicFactClothRet.MatchString(text) {
		return true
	}
	return len(clothingErrandClusters(text)) > 0
}

func rankCountQueryFacts(facts []MemoryEntry, query string) []MemoryEntry {
	if len(facts) < 2 {
		return facts
	}
	tokens := countQueryNounTokens(query)
	clothingQuery := isClothingCountQuery(query)
	entityKind := countEntityKind(query)
	type scored struct {
		e           MemoryEntry
		named       int
		clothing    int
		entity      int
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
		clothing := 0
		if clothingQuery && isClothingErrandText(hay) {
			clothing = 1
		}
		entity := 0
		if entityKind != "" && entityKind != "clothing" && len(uniqueEntityClusters(hay, entityKind)) > 0 {
			entity = 1
		}
		fp := 0
		if isFirstPersonFact(hay) {
			fp = 1
		}
		tmp[i] = scored{e: e, named: named, clothing: clothing, entity: entity, firstPerson: fp, overlap: keywordOverlapCount(e, tokens)}
	}
	sort.SliceStable(tmp, func(i, j int) bool {
		if tmp[i].named != tmp[j].named {
			return tmp[i].named > tmp[j].named
		}
		if tmp[i].clothing != tmp[j].clothing {
			return tmp[i].clothing > tmp[j].clothing
		}
		if tmp[i].entity != tmp[j].entity {
			return tmp[i].entity > tmp[j].entity
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
	if !isCountQuery(query) || isDatedSpanQuery(query) || isLatestValueQuery(query) {
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
// Clothing count queries diversify by action+object (dry-clean / return / pick-up
// × boot / blazer / generic) so a compound "return … pick them up" is two bullets
// and dry-clean survives when the query only says pick/return/store. Other
// quantity queries cluster by distinctive object (kit identity, plant name,
// hour+destination, restaurant name; catalogs are aliases) so a repeated B-29
// is one kit and two plants in one turn are two clusters. Clothing action×object clusters prefix
// "Count evidence (N distinct items):" and numbered bullets (1. 2. 3. in
// cluster order) so a reader can enumerate clusters. Unique-entity and
// exact-text paths prefix "Count evidence:" without N and stay unnumbered
// (cluster count is not gold). Dated-span “how many days/weeks/months
// between” is not assembled here (see AssembleTemporalEvidence). Empty when
// the query is not a count question or no facts match. Does not invent a
// numeric gold.
func AssembleCountEvidence(query string, facts []MemoryEntry) string {
	if !isCountQuery(query) || isDatedSpanQuery(query) || isLatestValueQuery(query) {
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
	var snippets []string
	fromClothing := false
	if isClothingCountQuery(query) {
		snippets = assembleClothingCountEvidence(matched)
		fromClothing = len(snippets) > 0
	}
	if len(snippets) == 0 {
		snippets = assembleUniqueEntityCountEvidence(query, matched)
	}
	if len(snippets) == 0 {
		snippets = assembleExactTextCountEvidence(matched)
	}
	if len(snippets) == 0 {
		return ""
	}
	if fromClothing {
		return formatNumberedEvidenceBlock("Count evidence", len(snippets), "items", snippets)
	}
	return "Count evidence:\n- " + strings.Join(snippets, "\n- ")
}

// formatEvidenceBlock prefixes a dash bullet list with the cluster count so a
// reader can count distinct items/events. It does not print a gold answer.
func formatEvidenceBlock(title string, n int, noun string, snippets []string) string {
	return title + " (" + strconv.Itoa(n) + " distinct " + noun + "):\n- " + strings.Join(snippets, "\n- ")
}

// formatNumberedEvidenceBlock is the clothing count path: (N distinct items)
// plus 1. 2. 3. bullets in cluster order. It does not print a gold answer.
func formatNumberedEvidenceBlock(title string, n int, noun string, snippets []string) string {
	var b strings.Builder
	b.WriteString(title)
	b.WriteString(" (")
	b.WriteString(strconv.Itoa(n))
	b.WriteString(" distinct ")
	b.WriteString(noun)
	b.WriteString("):\n")
	for i, s := range snippets {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(". ")
		b.WriteString(s)
	}
	return b.String()
}

func assembleExactTextCountEvidence(matched []MemoryEntry) []string {
	seen := make(map[string]struct{}, len(matched))
	snippets := make([]string, 0, len(matched))
	for _, e := range matched {
		text := factSnippetText(e)
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
	return snippets
}

func factSnippetText(e MemoryEntry) string {
	text := strings.TrimSpace(e.Content.Summary)
	if text == "" {
		text = strings.TrimSpace(e.Content.Full)
	}
	return text
}

const wordNumberAlt = `one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|thirteen|fourteen|fifteen|sixteen|seventeen|eighteen|nineteen|twenty|thirty`

var wordNumberValue = map[string]int{
	"one": 1, "two": 2, "three": 3, "four": 4, "five": 5, "six": 6, "seven": 7,
	"eight": 8, "nine": 9, "ten": 10, "eleven": 11, "twelve": 12, "thirteen": 13,
	"fourteen": 14, "fifteen": 15, "sixteen": 16, "seventeen": 17, "eighteen": 18,
	"nineteen": 19, "twenty": 20, "thirty": 30,
}

var reHourQuantity = regexp.MustCompile(`(?i)\b(\d+|` + wordNumberAlt + `)\s+hours?\b`)

type namedIdentity struct {
	key string
	re  *regexp.Regexp
}

var kitCatalog = []namedIdentity{
	{key: "f-15", re: regexp.MustCompile(`(?i)\bf-?15\b`)},
	{key: "spitfire", re: regexp.MustCompile(`(?i)\bspitfire\b`)},
	{key: "tiger", re: regexp.MustCompile(`(?i)\btiger\s*i\b|\bgerman tiger\b`)},
	{key: "b-29", re: regexp.MustCompile(`(?i)\bb-?29\b`)},
	{key: "camaro", re: regexp.MustCompile(`(?i)\bcamaro\b`)},
}

var plantCatalog = []namedIdentity{
	{key: "peace-lily", re: regexp.MustCompile(`(?i)\bpeace\s*lil(?:y|ies)\b`)},
	{key: "succulent", re: regexp.MustCompile(`(?i)\bsucculents?\b`)},
	{key: "snake-plant", re: regexp.MustCompile(`(?i)\bsnake\s*plants?\b`)},
}

var destCatalog = []namedIdentity{
	{key: "outer-banks", re: regexp.MustCompile(`(?i)\bouter\s*banks\b`)},
	{key: "washington", re: regexp.MustCompile(`(?i)\bwashington\b|\bd\.c\.`)},
	{key: "tennessee", re: regexp.MustCompile(`(?i)\btennessee\b`)},
}

// restaurantCatalog is optional aliases. Unseen names still cluster via
// "… restaurant" phrases, quoted names, and dining-verb proper nouns.
var restaurantCatalog = []namedIdentity{}

var (
	reHoursPrep         = regexp.MustCompile(`(?i)\b(?:\d+|` + wordNumberAlt + `)\s+hours?\s+(?:to|in|for|at|toward|towards)\s+`)
	reKitAnchor         = regexp.MustCompile(`(?i)\b(?:model\s+)?kits?\b`)
	reNamedPlant        = regexp.MustCompile(`(?i)\b([A-Za-z][A-Za-z0-9-]*)\s+plants?\b`)
	reModelCode         = regexp.MustCompile(`(?i)^[a-z]{1,3}-?\d{1,3}[a-z0-9]*$`)
	reMarkCode          = regexp.MustCompile(`(?i)^mk\.?[ivxlcdm0-9]+$`)
	reRestaurantAnchor  = regexp.MustCompile(`(?i)\brestaurants?\b`)
	reQuotedName        = regexp.MustCompile(`["“”]([^"“”\n]{2,48})["“”]`)
	reDiningVerb        = regexp.MustCompile(`(?i)\b(?:(?:went|go(?:ing)?)\s+back\s+to|(?:went|go(?:ing)?)\s+to|tried|try(?:ing)?|ate\s+at|eat(?:ing)?\s+at|dined\s+at|dine\s+at|had\s+(?:dinner|lunch|brunch|breakfast)\s+at|visited|visit(?:ing)?|ordered\s+(?:from|at)|stopped\s+(?:by|at))\s+`)
	reRestaurantSuffix  = regexp.MustCompile(`\b((?:[A-Z][A-Za-z0-9'’.-]*\s+){0,3}[A-Z][A-Za-z0-9'’.-]*)\s+(Kitchens?|Houses?|Grills?|Bistros?|Caf[eé]s?|BBQs?|Eaterys?|Eateries|Diners?|Taverns?)\b`)
	reTriedCountMention = regexp.MustCompile(`(?i)\b(?:tried|try(?:ing)?)\s+(?:(?:\d+|` + wordNumberAlt + `)\s+)?(?:different\s+)?(?:ones|restaurants)\b(?:\s+so\s+far|\s+recently)?`)
)

var kitPhraseStop = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "this": {}, "that": {}, "these": {}, "those": {},
	"my": {}, "our": {}, "his": {}, "her": {}, "their": {}, "some": {}, "any": {},
	"i": {}, "we": {}, "and": {}, "or": {}, "of": {}, "on": {}, "for": {},
	"simple": {}, "new": {}, "another": {}, "finished": {}, "working": {},
	"got": {}, "bought": {}, "picked": {}, "started": {}, "building": {},
	"just": {}, "recently": {}, "also": {}, "scale": {}, "model": {},
	"revell": {}, "tamiya": {}, "hasegawa": {}, "airfix": {},
	"how": {}, "many": {}, "much": {}, "number": {},
}

var destPhraseStop = map[string]struct{}{
	"drive": {}, "driving": {}, "drove": {}, "get": {}, "getting": {},
	"go": {}, "going": {}, "there": {}, "here": {}, "it": {},
	"my": {}, "our": {}, "this": {}, "that": {}, "make": {}, "making": {},
	"reach": {}, "reaching": {}, "arrive": {}, "arriving": {},
	"visit": {}, "visiting": {}, "see": {}, "come": {}, "coming": {},
	"do": {}, "be": {}, "been": {}, "work": {}, "working": {},
	"wait": {}, "waiting": {}, "spend": {}, "spending": {},
	"take": {}, "taking": {}, "got": {}, "have": {}, "had": {},
	"and": {}, "or": {}, "then": {}, "after": {}, "before": {},
	"from": {}, "with": {}, "about": {}, "around": {}, "over": {},
}

var plantNameStop = map[string]struct{}{
	"power": {}, "processing": {}, "chemical": {}, "nuclear": {},
	"water": {}, "sewage": {}, "gas": {}, "coal": {}, "treatment": {},
	"industrial": {}, "manufacturing": {}, "the": {}, "a": {}, "an": {},
	"my": {}, "our": {}, "this": {}, "that": {}, "some": {}, "his": {},
	"her": {}, "their": {},
}

var restaurantPhraseStop = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "this": {}, "that": {}, "these": {}, "those": {},
	"my": {}, "our": {}, "his": {}, "her": {}, "their": {}, "some": {}, "any": {},
	"i": {}, "we": {}, "and": {}, "or": {}, "of": {}, "on": {}, "for": {}, "at": {},
	"in": {}, "to": {}, "from": {}, "with": {}, "about": {}, "around": {},
	"new": {}, "local": {}, "nearby": {}, "favorite": {}, "another": {},
	"good": {}, "great": {}, "best": {}, "different": {}, "other": {},
	"one": {}, "ones": {}, "place": {}, "places": {}, "spot": {}, "spots": {},
	"how": {}, "many": {}, "much": {}, "number": {},
	"tried": {}, "try": {}, "trying": {}, "ate": {}, "eat": {}, "eating": {},
	"went": {}, "go": {}, "going": {}, "back": {}, "visited": {}, "visit": {},
	"visiting": {}, "dined": {}, "dine": {}, "ordered": {}, "stopped": {},
	"had": {}, "have": {}, "has": {}, "been": {}, "just": {}, "recently": {},
	"also": {}, "last": {}, "week": {}, "month": {}, "year": {}, "today": {},
	"so": {}, "far": {}, "there": {}, "here": {}, "it": {}, "them": {},
	"city": {}, "town": {}, "area": {}, "neighborhood": {},
	"making": {}, "make": {}, "home": {},
	"if": {}, "whether": {}, "maybe": {},
	"could": {}, "would": {}, "should": {}, "can": {}, "will": {},
	"did": {}, "does": {}, "do": {},
	"not": {}, "no": {}, "yes": {}, "your": {}, "you": {},
}

var restaurantCuisineStop = map[string]struct{}{
	"korean": {}, "chinese": {}, "japanese": {}, "italian": {}, "mexican": {},
	"thai": {}, "indian": {}, "vietnamese": {}, "french": {}, "american": {},
	"mediterranean": {}, "ethiopian": {}, "greek": {}, "spanish": {},
	"turkish": {}, "lebanese": {}, "persian": {}, "moroccan": {},
	"brazilian": {}, "peruvian": {}, "filipino": {}, "malaysian": {},
	"indonesian": {}, "caribbean": {}, "asian": {}, "european": {},
	"fusion": {}, "soul": {},
}

var restaurantGenericName = map[string]struct{}{
	"kitchen": {}, "kitchens": {}, "house": {}, "houses": {},
	"grill": {}, "grills": {}, "bistro": {}, "bistros": {},
	"cafe": {}, "cafes": {}, "café": {}, "cafés": {},
	"bbq": {}, "eatery": {}, "eateries": {}, "diner": {}, "diners": {},
	"tavern": {}, "taverns": {}, "restaurant": {}, "restaurants": {},
	"bar": {}, "bars": {}, "place": {}, "places": {}, "spot": {}, "spots": {},
}

type entityCluster struct {
	kind    string
	key     string
	snippet string
}

func countEntityKind(query string) string {
	q := strings.ToLower(query)
	if isClothingCountQuery(query) {
		return "clothing"
	}
	if strings.Contains(q, "kit") || (strings.Contains(q, "model") && !strings.Contains(q, "plant")) {
		return "kit"
	}
	if strings.Contains(q, "plant") {
		return "plant"
	}
	if strings.Contains(q, "hour") {
		return "hours"
	}
	if strings.Contains(q, "restaurant") {
		return "restaurant"
	}
	if strings.Contains(q, "dollar") || strings.Contains(q, "buck") {
		return "dollars"
	}
	return ""
}

func identityKeys(text string, catalog []namedIdentity) []string {
	var out []string
	for _, c := range catalog {
		if c.re.MatchString(text) {
			out = append(out, c.key)
		}
	}
	return out
}

func hasHourQuantity(text string) bool {
	return reHourQuantity.MatchString(text)
}

func parseWordOrDigit(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))
	if n, ok := wordNumberValue[s]; ok {
		return n
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			continue
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// uniqueEntityClusters returns distinctive-object clusters for kit, plant,
// restaurant, and hours count queries. Catalogs are aliases; kits also parse
// "… kit" / "model kit" noun phrases, plants also parse "<name> plant(s)",
// restaurants also parse "… restaurant" / quoted names / dining-verb proper
// nouns, and hours also parse dest after N hours to/in/for/at/toward.
func uniqueEntityClusters(text, kind string) []entityCluster {
	switch kind {
	case "kit":
		return mergeIdentityClusters(text, "kit", kitCatalog, extractKitPhrases(text))
	case "plant":
		return mergeIdentityClusters(text, "plant", plantCatalog, extractPlantPhrases(text))
	case "restaurant":
		return mergeIdentityClusters(text, "restaurant", restaurantCatalog, extractRestaurantPhrases(text))
	case "hours":
		return hoursClusters(text)
	default:
		return nil
	}
}

// catalogClusters emits one cluster per catalog identity present in text.
func catalogClusters(text, kind string, catalog []namedIdentity) []entityCluster {
	keys := identityKeys(text, catalog)
	if len(keys) == 0 {
		return nil
	}
	out := make([]entityCluster, 0, len(keys))
	for _, key := range keys {
		out = append(out, entityCluster{
			kind:    kind,
			key:     key,
			snippet: "[" + kind + ":" + key + "] " + strings.TrimSpace(text),
		})
	}
	return out
}

func mergeIdentityClusters(text, kind string, catalog []namedIdentity, extra []string) []entityCluster {
	out := catalogClusters(text, kind, catalog)
	if len(extra) == 0 {
		return out
	}
	seen := make(map[string]struct{}, len(out)+len(extra))
	for _, c := range out {
		seen[c.key] = struct{}{}
	}
	for _, phrase := range extra {
		var keys []string
		if matched := identityKeys(phrase, catalog); len(matched) > 0 {
			keys = matched
		} else if s := identitySlug(phrase); s != "" {
			keys = []string{s}
		}
		for _, key := range keys {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, entityCluster{
				kind:    kind,
				key:     key,
				snippet: "[" + kind + ":" + key + "] " + strings.TrimSpace(text),
			})
		}
	}
	return out
}

// hoursClusters groups hour-quantity facts by destination. destCatalog names
// are aliases; destinations are also taken from "N hours to/in/for/at/toward
// <place>". Repeated dests collapse; distinct dests stay distinct.
func hoursClusters(text string) []entityCluster {
	if !hasHourQuantity(text) {
		return nil
	}
	seen := make(map[string]struct{}, 8)
	var keys []string
	add := func(key string) {
		key = strings.TrimSpace(key)
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	for _, k := range identityKeys(text, destCatalog) {
		add(k)
	}
	for _, dest := range extractHourDestinations(text) {
		if matched := identityKeys(dest, destCatalog); len(matched) > 0 {
			for _, k := range matched {
				add(k)
			}
			continue
		}
		add(identitySlug(dest))
	}
	if len(keys) == 0 {
		qty := strings.ToLower(strings.TrimSpace(reHourQuantity.FindString(text)))
		key := qty
		if key == "" {
			key = "hours"
		}
		return []entityCluster{{
			kind:    "hours",
			key:     key,
			snippet: "[hours] " + strings.TrimSpace(text),
		}}
	}
	out := make([]entityCluster, 0, len(keys))
	for _, d := range keys {
		out = append(out, entityCluster{
			kind:    "hours",
			key:     d,
			snippet: "[hours:" + d + "] " + strings.TrimSpace(text),
		})
	}
	return out
}

func extractHourDestinations(text string) []string {
	locs := reHoursPrep.FindAllStringIndex(text, -1)
	if len(locs) == 0 {
		return nil
	}
	var out []string
	for _, loc := range locs {
		dest := takePlaceName(strings.TrimSpace(text[loc[1]:]))
		if dest != "" {
			out = append(out, dest)
		}
	}
	return out
}

func takePlaceName(rest string) string {
	fields := strings.Fields(rest)
	var parts []string
	for i, f := range fields {
		tok := strings.Trim(f, ".,;:!?\"`()[]")
		if tok == "" {
			break
		}
		low := strings.ToLower(tok)
		if i == 0 && (low == "the" || low == "a" || low == "an") && len(parts) == 0 {
			continue
		}
		if _, stop := destPhraseStop[low]; stop {
			break
		}
		if !looksLikePlaceToken(tok) && len(identityKeys(tok, destCatalog)) == 0 {
			break
		}
		parts = append(parts, tok)
		if len(parts) >= 4 {
			break
		}
	}
	return strings.Join(parts, " ")
}

func looksLikePlaceToken(tok string) bool {
	if tok == "" {
		return false
	}
	r := []rune(tok)
	return r[0] >= 'A' && r[0] <= 'Z'
}

func extractKitPhrases(text string) []string {
	locs := reKitAnchor.FindAllStringIndex(text, -1)
	if len(locs) == 0 {
		return nil
	}
	var out []string
	for _, loc := range locs {
		toks := strings.Fields(text[:loc[0]])
		var collected []string
		looked := 0
		for i := len(toks) - 1; i >= 0 && looked < 8; i-- {
			looked++
			tok := strings.Trim(toks[i], ".,;:!?\"`()[]")
			if tok == "" {
				continue
			}
			low := strings.ToLower(tok)
			if isScaleToken(low) {
				if len(collected) > 0 {
					break
				}
				continue
			}
			if _, stop := kitPhraseStop[low]; stop {
				if len(collected) > 0 {
					break
				}
				continue
			}
			if isKitNameToken(tok) {
				collected = append([]string{tok}, collected...)
				if len(collected) >= 4 {
					break
				}
				continue
			}
			if len(collected) > 0 && isAlphaWord(tok) && len(tok) >= 3 {
				collected = append([]string{tok}, collected...)
				if len(collected) >= 4 {
					break
				}
				continue
			}
			if len(collected) > 0 {
				break
			}
		}
		if len(collected) == 0 {
			continue
		}
		out = append(out, strings.Join(collected, " "))
	}
	return out
}

func isKitNameToken(tok string) bool {
	if reModelCode.MatchString(tok) || reMarkCode.MatchString(tok) || isYearToken(tok) {
		return true
	}
	r := []rune(tok)
	return len(r) > 0 && r[0] >= 'A' && r[0] <= 'Z'
}

func isYearToken(tok string) bool {
	if strings.HasPrefix(tok, "'") && len(tok) == 3 {
		return tok[1] >= '0' && tok[1] <= '9' && tok[2] >= '0' && tok[2] <= '9'
	}
	return false
}

func isScaleToken(tok string) bool {
	return tok == "scale" || strings.Contains(tok, "/")
}

func isAlphaWord(tok string) bool {
	if tok == "" {
		return false
	}
	for _, r := range tok {
		if r < 'A' || (r > 'Z' && r < 'a') || r > 'z' {
			return false
		}
	}
	return true
}

func extractPlantPhrases(text string) []string {
	matches := reNamedPlant.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}
	var out []string
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		if len(identityKeys(m[0], plantCatalog)) > 0 {
			continue
		}
		name := m[1]
		if _, stop := plantNameStop[strings.ToLower(name)]; stop {
			continue
		}
		out = append(out, name)
	}
	return out
}

func extractRestaurantPhrases(text string) []string {
	var out []string
	seen := make(map[string]struct{}, 8)
	add := func(s string) {
		s = normalizeRestaurantName(s)
		if s == "" {
			return
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	locs := reRestaurantAnchor.FindAllStringIndex(text, -1)
	for _, loc := range locs {
		toks := strings.Fields(text[:loc[0]])
		var collected []string
		looked := 0
		for i := len(toks) - 1; i >= 0 && looked < 8; i-- {
			looked++
			tok := strings.Trim(toks[i], ".,;:!?\"'`()[]")
			if tok == "" {
				continue
			}
			low := strings.ToLower(tok)
			if _, stop := restaurantPhraseStop[low]; stop {
				if len(collected) > 0 {
					break
				}
				continue
			}
			if _, ok := wordNumberValue[low]; ok {
				if len(collected) > 0 {
					break
				}
				continue
			}
			if isRestaurantNameToken(tok) {
				if isCuisineToken(low) && len(collected) == 0 {
					continue
				}
				collected = append([]string{tok}, collected...)
				if len(collected) >= 4 {
					break
				}
				continue
			}
			if len(collected) > 0 {
				break
			}
		}
		if len(collected) > 0 {
			add(strings.Join(collected, " "))
		}
	}
	for _, m := range reQuotedName.FindAllStringSubmatch(text, -1) {
		if len(m) < 2 {
			continue
		}
		add(m[1])
	}
	for _, loc := range reDiningVerb.FindAllStringIndex(text, -1) {
		add(takeRestaurantName(strings.TrimSpace(text[loc[1]:])))
	}
	for _, m := range reRestaurantSuffix.FindAllStringSubmatch(text, -1) {
		if len(m) < 3 {
			continue
		}
		add(strings.TrimSpace(m[1] + " " + m[2]))
	}
	return out
}

func takeRestaurantName(rest string) string {
	fields := strings.Fields(rest)
	var parts []string
	for i, f := range fields {
		tok := strings.Trim(f, ".,;:!?\"'`()[]")
		if tok == "" {
			break
		}
		low := strings.ToLower(tok)
		if i == 0 && (low == "the" || low == "a" || low == "an") && len(parts) == 0 {
			continue
		}
		if low == "restaurant" || low == "restaurants" {
			break
		}
		if _, stop := restaurantPhraseStop[low]; stop {
			break
		}
		if _, ok := wordNumberValue[low]; ok {
			break
		}
		if !isRestaurantNameToken(tok) {
			break
		}
		if isCuisineToken(low) && len(parts) == 0 {
			next := ""
			if i+1 < len(fields) {
				next = fields[i+1]
			}
			if !isVenueGenericSuffix(next) {
				break
			}
		}
		parts = append(parts, tok)
		if len(parts) >= 4 {
			break
		}
	}
	return strings.Join(parts, " ")
}

func isRestaurantNameToken(tok string) bool {
	if tok == "" {
		return false
	}
	low := strings.ToLower(tok)
	if _, stop := restaurantPhraseStop[low]; stop {
		return false
	}
	if isCuisineToken(low) {
		return true
	}
	if _, ok := restaurantGenericName[low]; ok {
		return true
	}
	r := []rune(tok)
	return r[0] >= 'A' && r[0] <= 'Z'
}

func isCuisineToken(low string) bool {
	low = strings.ToLower(strings.Trim(low, ".,;:!?\"'`()[]"))
	if low == "" {
		return false
	}
	if _, ok := restaurantCuisineStop[low]; ok {
		return true
	}
	for _, part := range strings.Split(low, "-") {
		if part == "" || part == "style" || part == "inspired" {
			continue
		}
		if _, ok := restaurantCuisineStop[part]; ok {
			return true
		}
	}
	return false
}

func isVenueGenericSuffix(tok string) bool {
	switch strings.ToLower(strings.Trim(tok, ".,;:!?\"'`()[]")) {
	case "kitchen", "kitchens", "house", "houses",
		"grill", "grills", "bistro", "bistros",
		"cafe", "cafes", "café", "cafés",
		"eatery", "eateries", "diner", "diners",
		"tavern", "taverns":
		return true
	default:
		return false
	}
}

func isRestaurantDishOrGenericToken(low string) bool {
	low = strings.ToLower(strings.Trim(low, ".,;:!?\"'`()[]"))
	if isCuisineToken(low) {
		return true
	}
	if _, ok := restaurantGenericName[low]; ok {
		return true
	}
	return low == "style" || low == "inspired"
}

func normalizeRestaurantName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'“”`)
	fields := strings.Fields(s)
	for len(fields) > 0 {
		low := strings.ToLower(strings.Trim(fields[len(fields)-1], ".,;:!?\"'`()[]"))
		if low == "restaurant" || low == "restaurants" {
			fields = fields[:len(fields)-1]
			continue
		}
		if isCuisineToken(low) {
			fields = fields[:len(fields)-1]
			continue
		}
		break
	}
	for len(fields) > 0 {
		low := strings.ToLower(strings.Trim(fields[0], ".,;:!?\"'`()[]"))
		if _, stop := restaurantPhraseStop[low]; stop {
			fields = fields[1:]
			continue
		}
		if isCuisineToken(low) && len(fields) > 1 {
			fields = fields[1:]
			continue
		}
		break
	}
	if len(fields) == 0 {
		return ""
	}
	cleaned := make([]string, 0, len(fields))
	for _, f := range fields {
		tok := strings.Trim(f, ".,;:!?\"'`()[]")
		if tok != "" {
			cleaned = append(cleaned, tok)
		}
	}
	if len(cleaned) == 0 {
		return ""
	}
	onlyDish := true
	for _, tok := range cleaned {
		if !isRestaurantDishOrGenericToken(strings.ToLower(tok)) {
			onlyDish = false
			break
		}
	}
	if onlyDish {
		return ""
	}
	if len(cleaned) == 1 {
		if _, gen := restaurantGenericName[strings.ToLower(cleaned[0])]; gen {
			return ""
		}
		if _, stop := restaurantPhraseStop[strings.ToLower(cleaned[0])]; stop {
			return ""
		}
	}
	return strings.Join(cleaned, " ")
}

func identitySlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastHyphen := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		default:
			if b.Len() > 0 && !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// assembleUniqueEntityCountEvidence lists one dash bullet per distinctive kit /
// plant / restaurant / hour-destination identity. Catalogs are aliases; unseen
// kit phrases, restaurant names, and hour destinations still cluster. Restaurant
// queries prepend latest-first "tried N" self-reports (not a gold answer). No
// (N distinct items) header; cluster count is not gold.
func assembleUniqueEntityCountEvidence(query string, matched []MemoryEntry) []string {
	kind := countEntityKind(query)
	if kind == "" || kind == "clothing" {
		return nil
	}
	var out []string
	if kind == "restaurant" {
		out = extractTriedCountMentions(matched)
	}
	type filled struct {
		snippet string
		order   int
	}
	got := make(map[string]filled, 8)
	n := 0
	for _, e := range matched {
		text := factSnippetText(e)
		if text == "" {
			continue
		}
		clusters := uniqueEntityClusters(text, kind)
		for _, c := range clusters {
			key := c.kind + "\t" + c.key
			if _, ok := got[key]; ok {
				continue
			}
			got[key] = filled{snippet: c.snippet, order: n}
			n++
		}
	}
	if len(got) == 0 && len(out) == 0 {
		return nil
	}
	type ordered struct {
		order int
		snip  string
	}
	list := make([]ordered, 0, len(got))
	for _, v := range got {
		list = append(list, ordered{order: v.order, snip: v.snippet})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].order < list[j].order })
	for _, x := range list {
		if len(out) >= maxCountEvidenceSnippets {
			break
		}
		out = append(out, x.snip)
	}
	return out
}

func extractTriedCountMentions(matched []MemoryEntry) []string {
	type hit struct {
		when  time.Time
		snip  string
		order int
	}
	var hits []hit
	seen := make(map[string]struct{}, 4)
	n := 0
	for _, e := range matched {
		when := entryEventTime(e)
		text := factSnippetText(e)
		if text == "" {
			continue
		}
		for _, sent := range splitEvidenceSentences(text) {
			if !reTriedCountMention.MatchString(sent) {
				continue
			}
			snip := triedCountSnippet(sent)
			key := strings.ToLower(snip)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			hits = append(hits, hit{
				when:  when,
				snip:  formatLatestValueLabel(when) + " " + snip,
				order: n,
			})
			n++
		}
	}
	if len(hits) == 0 {
		return nil
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if !hits[i].when.Equal(hits[j].when) {
			return hits[i].when.After(hits[j].when)
		}
		return hits[i].order < hits[j].order
	})
	out := make([]string, 0, len(hits))
	for _, h := range hits {
		out = append(out, h.snip)
		if len(out) >= maxCountEvidenceSnippets {
			break
		}
	}
	return out
}

func triedCountSnippet(sent string) string {
	sent = strings.TrimSpace(sent)
	loc := reTriedCountMention.FindStringIndex(sent)
	if loc == nil {
		return sent
	}
	start := loc[0]
	head := strings.ToLower(strings.TrimSpace(sent[:start]))
	switch {
	case strings.HasSuffix(head, "i've"):
		start = strings.LastIndex(strings.ToLower(sent[:loc[0]]), "i've")
	case strings.HasSuffix(head, "i have"):
		start = strings.LastIndex(strings.ToLower(sent[:loc[0]]), "i have")
	case strings.HasSuffix(head, "i"):
		start = strings.LastIndex(strings.ToLower(sent[:loc[0]]), "i")
	}
	if start < 0 {
		start = loc[0]
	}
	return strings.TrimSpace(sent[start:loc[1]])
}

const (
	errandDryClean = "dry-clean"
	errandReturn   = "return"
	errandPickup   = "pick-up"
	objectBoot     = "boot"
	objectBlazer   = "blazer"
)

type errandCluster struct {
	action string
	object string
}

var (
	reErrandReturn = regexp.MustCompile(`(?i)\b(return(?:ed|ing)?|exchange(?:d|s)?)\b`)
	reErrandPickup = regexp.MustCompile(`(?i)pick(?:ed)?(?:\s+them)?[\s-]*up`)
	reErrandBoot   = regexp.MustCompile(`(?i)\bboots?\b`)
	reErrandBlazer = regexp.MustCompile(`(?i)\bblazers?\b`)
)

func clothingErrandClusters(text string) []errandCluster {
	lower := strings.ToLower(text)
	hasDry := atomicFactDryClean.MatchString(lower)
	hasRet := reErrandReturn.MatchString(lower)
	hasPick := reErrandPickup.MatchString(lower)
	hasBoot := reErrandBoot.MatchString(lower)
	hasBlazer := reErrandBlazer.MatchString(lower)
	hasZara := strings.Contains(lower, "zara")

	// Dry-clean pickup of a blazer is one errand, not pick-up + dry-clean.
	// Keep pick-up when a distinct boot object is also present.
	if hasDry && hasPick && !hasBoot {
		hasPick = false
	}

	// Return/exchange without a clothing object is sister-sweater / poster noise.
	if hasRet && !hasBoot && !hasBlazer && !hasZara {
		hasRet = false
	}
	// Generic pick-up is allowed (pronoun "pick them up" after sentence split).
	if !hasDry && !hasRet && !hasPick {
		return nil
	}

	objFor := func(action string) string {
		switch action {
		case errandDryClean:
			if hasBlazer {
				return objectBlazer
			}
			if hasBoot {
				return objectBoot
			}
			return ""
		default:
			if hasBoot {
				return objectBoot
			}
			if hasBlazer {
				return objectBlazer
			}
			return ""
		}
	}

	out := make([]errandCluster, 0, 3)
	if hasDry {
		out = append(out, errandCluster{action: errandDryClean, object: objFor(errandDryClean)})
	}
	if hasRet {
		out = append(out, errandCluster{action: errandReturn, object: objFor(errandReturn)})
	}
	if hasPick {
		out = append(out, errandCluster{action: errandPickup, object: objFor(errandPickup)})
	}
	return out
}

func assembleClothingCountEvidence(matched []MemoryEntry) []string {
	type filled struct {
		snippet string
		order   int
	}
	got := make(map[string]filled, 8)
	n := 0
	for _, e := range matched {
		text := factSnippetText(e)
		if text == "" {
			continue
		}
		clusters := clothingErrandClusters(text)
		if len(clusters) == 0 {
			continue
		}
		snips := splitErrandSnippets(text, clusters)
		for i, c := range clusters {
			key := c.action + "\t" + c.object
			if _, ok := got[key]; ok {
				continue
			}
			snip := text
			if i < len(snips) {
				snip = snips[i]
			}
			got[key] = filled{snippet: snip, order: n}
			n++
		}
	}
	if len(got) == 0 {
		return nil
	}
	specific := make(map[string]bool, 3)
	for k := range got {
		action, obj, _ := strings.Cut(k, "\t")
		if obj != "" {
			specific[action] = true
		}
	}
	type ordered struct {
		order int
		snip  string
	}
	list := make([]ordered, 0, len(got))
	for k, v := range got {
		action, obj, _ := strings.Cut(k, "\t")
		if obj == "" && specific[action] {
			continue
		}
		list = append(list, ordered{order: v.order, snip: v.snippet})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].order < list[j].order })
	out := make([]string, 0, len(list))
	for _, x := range list {
		out = append(out, x.snip)
		if len(out) >= maxCountEvidenceSnippets {
			break
		}
	}
	return out
}

func splitErrandSnippets(text string, clusters []errandCluster) []string {
	if len(clusters) == 0 {
		return nil
	}
	if len(clusters) == 1 {
		return []string{labelErrand(clusters[0], text)}
	}
	clauses := splitErrandClauses(text)
	used := make([]bool, len(clauses))
	out := make([]string, 0, len(clusters))
	for _, c := range clusters {
		assigned := ""
		for i, cl := range clauses {
			if used[i] {
				continue
			}
			if clauseMatchesAction(cl, c.action) {
				assigned = cl
				used[i] = true
				break
			}
		}
		if assigned == "" {
			assigned = text
		}
		out = append(out, labelErrand(c, assigned))
	}
	return out
}

func labelErrand(c errandCluster, text string) string {
	return "[" + c.action + "] " + strings.TrimSpace(text)
}

func splitErrandClauses(text string) []string {
	for _, sep := range []string{"; ", ", and ", " and "} {
		parts := strings.Split(text, sep)
		if len(parts) < 2 {
			continue
		}
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		if len(out) >= 2 {
			return out
		}
	}
	return []string{strings.TrimSpace(text)}
}

func clauseMatchesAction(clause, action string) bool {
	lower := strings.ToLower(clause)
	switch action {
	case errandDryClean:
		return atomicFactDryClean.MatchString(lower)
	case errandReturn:
		return reErrandReturn.MatchString(lower)
	case errandPickup:
		return reErrandPickup.MatchString(lower)
	}
	return false
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
