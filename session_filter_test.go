package memory

import (
	"strconv"
	"strings"
	"testing"
)

func TestConvTag(t *testing.T) {
	if got := ConvTag("  conv-1  "); got != "conv:conv-1" {
		t.Fatalf("ConvTag = %q", got)
	}
	if ConvTag(" ") != "" || ConvTag("") != "" {
		t.Fatal("blank ConvTag must be empty")
	}
}

func TestEntryMatchesSessionFilter_ConvTag(t *testing.T) {
	e := MemoryEntry{
		SessionID: "hay-1",
		Content:   MemoryContent{Tags: []string{ConvTag("qa-9")}},
	}
	if !entryMatchesSessionFilter(e, "qa-9", nil) {
		t.Fatal("conv:qa-9 must match SessionID=qa-9")
	}
	if !entryMatchesSessionFilter(e, "", []string{"hay-1"}) {
		t.Fatal("SessionIDs hay-1 must match inner SessionID")
	}
	if entryMatchesSessionFilter(e, "other", nil) {
		t.Fatal("other conv must not match")
	}
	if !entryMatchesSessionFilter(e, "", nil) {
		t.Fatal("empty filter matches all")
	}
}

func TestDiversifyBySession_Interleaves(t *testing.T) {
	var entries []MemoryEntry
	for i := 0; i < 8; i++ {
		entries = append(entries, MemoryEntry{ID: "a" + string(rune('0'+i)), SessionID: "sess-A"})
	}
	entries = append(entries, MemoryEntry{ID: "gold", SessionID: "sess-B", Content: MemoryContent{Summary: "boots from zara"}})
	got := diversifyBySession(entries, 5)
	if len(got) != 5 {
		t.Fatalf("len=%d want 5", len(got))
	}
	found := false
	for _, e := range got {
		if e.ID == "gold" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("gold sess-B buried; ids=%v", idsOf(got))
	}
}

func TestDiversifyBySession_SingleSessionPrefix(t *testing.T) {
	var entries []MemoryEntry
	for i := 0; i < 6; i++ {
		entries = append(entries, MemoryEntry{ID: string(rune('a' + i)), SessionID: "only"})
	}
	got := diversifyBySession(entries, 3)
	if len(got) != 3 || got[0].ID != "a" || got[2].ID != "c" {
		t.Fatalf("single-session must prefix Limit, got %+v", idsOf(got))
	}
}

func TestSearchMemoryWithOptions_ConvTagAndDiversify(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	conv := "lme-qid"
	for i := 0; i < 12; i++ {
		e := MemoryEntry{
			ID:        "fill-" + string(rune('a'+i)),
			Tier:      TierContextual,
			SessionID: "closet",
			Content: MemoryContent{
				Summary: "pick up clothes at the store while organizing closet boxes",
				Full:    "pick up clothes at the store while organizing closet winter boxes dresser drawers",
				Tags:    []string{ConvTag(conv)},
			},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	gold := MemoryEntry{
		ID:        "gold-boots",
		Tier:      TierContextual,
		SessionID: "store-run",
		Content: MemoryContent{
			Summary: "exchanged boots at Zara to pick up",
			Full:    "Need to pick up the new pair of boots exchanged at Zara.",
			Tags:    []string{ConvTag(conv)},
		},
	}
	if err := store.Write(gold); err != nil {
		t.Fatal(err)
	}

	hits := store.SearchMemoryWithOptions("pick up boots clothes store", SearchMemoryOptions{
		SessionID: conv,
		Limit:     8,
	})
	found := false
	for _, h := range hits {
		if h.ID == "gold-boots" {
			found = true
		}
		if h.SessionID != "closet" && h.SessionID != "store-run" {
			t.Fatalf("unexpected session %q", h.SessionID)
		}
	}
	if !found {
		t.Fatalf("gold not in top-8 under conv SessionID; ids=%v", idsOf(hits))
	}

	any := store.SearchMemoryWithOptions("pick up boots", SearchMemoryOptions{
		SessionIDs: []string{"store-run", "closet"},
		Limit:      8,
	})
	found = false
	for _, h := range any {
		if h.ID == "gold-boots" {
			found = true
		}
	}
	if !found {
		t.Fatalf("SessionIDs any-of missed gold; ids=%v", idsOf(any))
	}
}

func TestListFactsAsOf_ConvTag(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	conv := "qa-conv"
	e := MemoryEntry{
		ID:        "f1",
		Tier:      TierSemantic,
		SessionID: "inner",
		Content:   MemoryContent{Summary: "led two data projects", Tags: []string{ConvTag(conv)}},
	}
	if err := store.Write(e); err != nil {
		t.Fatal(err)
	}
	got := store.ListFactsAsOf(FactsAsOfOptions{SessionID: conv, Limit: 10})
	if len(got) != 1 || got[0].ID != "f1" {
		t.Fatalf("facts-as-of conv tag: %+v", idsOf(got))
	}
}

func TestSearchMemoryWithOptions_CountQueryPromotesFacts(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	conv := "proj-conv"
	for i := 0; i < 10; i++ {
		e := MemoryEntry{
			ID:        "chatter-" + string(rune('a'+i)),
			Tier:      TierContextual,
			SessionID: "s1",
			Type:      "conversation_turn",
			Content: MemoryContent{
				Summary: "how many clustering methods elbow silhouette analysis",
				Full:    "how many clustering methods elbow silhouette analysis for the project",
				Tags:    []string{ConvTag(conv)},
			},
		}
		if err := store.Write(e); err != nil {
			t.Fatal(err)
		}
	}
	fact := MemoryEntry{
		ID:        "fact-led-two",
		Tier:      TierSemantic,
		SessionID: "s2",
		Type:      "turn_fact",
		Content: MemoryContent{
			Summary: "Currently leading two data analysis projects",
			Full:    "Currently leading two data analysis projects at work.",
			Tags:    []string{ConvTag(conv), "fact_augmented", "from_turn"},
		},
	}
	if err := store.Write(fact); err != nil {
		t.Fatal(err)
	}
	hits := store.SearchMemoryWithOptions("How many projects have I led?", SearchMemoryOptions{
		SessionID: conv,
		Limit:     5,
	})
	if len(hits) == 0 || hits[0].ID != "fact-led-two" {
		t.Fatalf("count query should lead with turn_fact, ids=%v", idsOf(hits))
	}
}

func TestPromoteFactEntriesForQuery_RanksLedOverChatterFacts(t *testing.T) {
	chatter := MemoryEntry{
		ID: "f-chatter", Type: "turn_fact",
		Content: MemoryContent{Summary: "The elbow method is an excellent choice for clustering."},
	}
	gold := MemoryEntry{
		ID: "f-led", Type: "turn_fact",
		Content: MemoryContent{Summary: "I led the data analysis team on a marketing research class project."},
	}
	got := promoteFactEntriesForQuery([]MemoryEntry{chatter, gold}, "How many projects have I led")
	if len(got) != 2 || got[0].ID != "f-led" {
		t.Fatalf("led fact should rank first among facts, ids=%v", idsOf(got))
	}
}

func TestPromoteFactEntriesForQuery_NamedOutranksFallbackChatter(t *testing.T) {
	chatter := MemoryEntry{
		ID: "f-chatter", Type: "turn_fact",
		Content: MemoryContent{Summary: "I have many projects this semester in clustering class."},
	}
	gold := MemoryEntry{
		ID: "f-led", Type: "turn_fact",
		Content: MemoryContent{Summary: "I led the data analysis team on a marketing research class project."},
	}
	got := promoteFactEntriesForQuery([]MemoryEntry{chatter, gold}, "How many projects have I led")
	if len(got) != 2 || got[0].ID != "f-led" {
		t.Fatalf("named led-project fact should outrank fallback chatter, ids=%v", idsOf(got))
	}
}

func TestExtractAtomicFacts_LedProject(t *testing.T) {
	got := ExtractAtomicFacts(MemoryEntry{Content: MemoryContent{
		Full: "I led the data analysis team on a class project. The elbow method is useful.",
	}})
	found := false
	for _, s := range got {
		if strings.Contains(strings.ToLower(s), "led the data analysis") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected led-project sentence in %v", got)
	}

	leading := ExtractAtomicFacts(MemoryEntry{Content: MemoryContent{
		Full: "I am currently leading a data analysis project at work. Clustering is optional.",
	}})
	found = false
	for _, s := range leading {
		if strings.Contains(strings.ToLower(s), "leading a data analysis") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected leading-project sentence in %v", leading)
	}
}

func TestSearchMemoryWithOptions_CountQueryAssemblesFactsAcrossSessions(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	conv := "proj-count"
	if err := store.IngestTurn(MemoryEntry{
		ID:        "turn-led",
		SessionID: "hay-led",
		Content: MemoryContent{
			Full: "By the way, I've had some experience from my Marketing Research class project, where I led the data analysis team and we did a comprehensive market analysis for a new product launch.",
			Tags: []string{ConvTag(conv)},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.IngestTurn(MemoryEntry{
		ID:        "turn-solo",
		SessionID: "hay-solo",
		Content: MemoryContent{
			Full: "I've been working on a solo project for my Data Mining class, and I'm really interested in applying some of these techniques to my customer purchase data.",
			Tags: []string{ConvTag(conv)},
		},
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 16; i++ {
		if err := store.IngestTurn(MemoryEntry{
			ID:        "turn-noise-" + strconv.Itoa(i),
			SessionID: "hay-noise",
			Content: MemoryContent{
				Full: "The elbow method is an excellent choice for clustering how many analysis techniques appear in project dashboards and number of projects metrics " + strconv.Itoa(i) + ".",
				Tags: []string{ConvTag(conv)},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}

	q := "How many projects have I led"
	hits := store.SearchMemoryWithOptions(q, SearchMemoryOptions{SessionID: conv, Limit: 6})
	if !searchHayContains(hits, "led the data analysis") {
		t.Fatalf("missing led-team gold under small Limit; ids=%v summaries=%v", idsOf(hits), summariesOf(hits))
	}
	if !searchHayContains(hits, "solo project") {
		t.Fatalf("missing solo-project gold under small Limit; ids=%v summaries=%v", idsOf(hits), summariesOf(hits))
	}

	evidence := AssembleCountEvidence(q, hits)
	lower := strings.ToLower(evidence)
	if !strings.Contains(lower, "led the data analysis") {
		t.Fatalf("assembly missing led fact: %q", evidence)
	}
	if !strings.Contains(lower, "solo project") {
		t.Fatalf("assembly missing solo fact: %q", evidence)
	}
}

func TestAssembleCountEvidence_SkipsNonCountQuery(t *testing.T) {
	facts := []MemoryEntry{{
		ID: "f-led", Type: "turn_fact",
		Content: MemoryContent{Summary: "I led the data analysis team on a class project."},
	}}
	if got := AssembleCountEvidence("where do I work", facts); got != "" {
		t.Fatalf("non-count query must not assemble, got %q", got)
	}
}

func TestExtractAtomicFacts_ClothingErrands(t *testing.T) {
	dry := ExtractAtomicFacts(MemoryEntry{Content: MemoryContent{
		Full: "I still need to pick up my dry cleaning for the navy blue blazer. The closet is a mess.",
	}})
	if !factsContain(dry, "dry clean") || !factsContain(dry, "blazer") {
		t.Fatalf("expected dry-clean blazer fact, got %v", dry)
	}

	ret := ExtractAtomicFacts(MemoryEntry{Content: MemoryContent{
		Full: "I need to return some boots to Zara. Clustering notes are unrelated.",
	}})
	if !factsContain(ret, "return") || !factsContain(ret, "boot") {
		t.Fatalf("expected return-boots fact, got %v", ret)
	}

	pick := ExtractAtomicFacts(MemoryEntry{Content: MemoryContent{
		Full: "I still need to pick up the new pair of boots I exchanged at Zara.",
	}})
	if !factsContain(pick, "pick up") || !factsContain(pick, "boot") {
		t.Fatalf("expected pick-up boots fact, got %v", pick)
	}

	poster := "I need to return the poster from the case competition tomorrow afternoon."
	if atomicFactClothRet.MatchString(poster) || atomicFactClothPick.MatchString(poster) || atomicFactDryClean.MatchString(poster) {
		t.Fatalf("poster/case-competition must not match clothing errand extract: %q", poster)
	}
}

func TestAssembleCountEvidence_CompoundReturnAndPickup(t *testing.T) {
	facts := []MemoryEntry{{
		ID: "f-compound", Type: "turn_fact",
		Content: MemoryContent{Summary: "I need to return boots and pick them up."},
	}}
	q := "How many items of clothing do I need to pick up or return from a store?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "[return]") {
		t.Fatalf("compound must emit return cluster: %q", got)
	}
	if !strings.Contains(lower, "[pick-up]") {
		t.Fatalf("compound must emit pick-up cluster: %q", got)
	}
	if strings.Count(got, "\n- ") < 2 {
		t.Fatalf("compound should be two bullets, got %q", got)
	}
}

func TestSearchMemoryWithOptions_CountQueryAssemblesClothingErrands(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	conv := "clothes-count"
	if err := store.IngestTurn(MemoryEntry{
		ID:        "turn-dry",
		SessionID: "hay-dry",
		Content: MemoryContent{
			Full: "I still need to pick up my dry cleaning for the navy blue blazer I wore to a meeting.",
			Tags: []string{ConvTag(conv)},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.IngestTurn(MemoryEntry{
		ID:        "turn-return",
		SessionID: "hay-return",
		Content: MemoryContent{
			Full: "I need to return some boots to Zara that were too small, so I exchanged them.",
			Tags: []string{ConvTag(conv)},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.IngestTurn(MemoryEntry{
		ID:        "turn-pickup",
		SessionID: "hay-pickup",
		Content: MemoryContent{
			Full: "I still need to pick up the new pair of boots I exchanged at Zara.",
			Tags: []string{ConvTag(conv)},
		},
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 16; i++ {
		if err := store.IngestTurn(MemoryEntry{
			ID:        "turn-noise-" + strconv.Itoa(i),
			SessionID: "hay-noise",
			Content: MemoryContent{
				Full: "The closet has how many hanging items and number of store receipts for seasonal clothes " + strconv.Itoa(i) + ".",
				Tags: []string{ConvTag(conv)},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}

	q := "How many items of clothing do I need to pick up or return from a store?"
	hits := store.SearchMemoryWithOptions(q, SearchMemoryOptions{SessionID: conv, Limit: 6})
	if !searchHayContains(hits, "dry clean") {
		t.Fatalf("missing dry-clean gold under small Limit; ids=%v summaries=%v", idsOf(hits), summariesOf(hits))
	}
	if !searchHayContains(hits, "return some boots") {
		t.Fatalf("missing return-boots gold under small Limit; ids=%v summaries=%v", idsOf(hits), summariesOf(hits))
	}
	if !searchHayContains(hits, "pick up the new pair") {
		t.Fatalf("missing pick-up boots gold under small Limit; ids=%v summaries=%v", idsOf(hits), summariesOf(hits))
	}

	evidence := AssembleCountEvidence(q, hits)
	lower := strings.ToLower(evidence)
	if !strings.Contains(lower, "dry-clean") && !strings.Contains(lower, "dry clean") {
		t.Fatalf("assembly missing dry-clean cluster: %q", evidence)
	}
	if !strings.Contains(lower, "[return]") && !strings.Contains(lower, "return") {
		t.Fatalf("assembly missing return cluster: %q", evidence)
	}
	if !strings.Contains(lower, "[pick-up]") && !strings.Contains(lower, "pick up") {
		t.Fatalf("assembly missing pick-up cluster: %q", evidence)
	}
	if !strings.Contains(lower, "blazer") {
		t.Fatalf("assembly missing blazer object: %q", evidence)
	}
	if !strings.Contains(lower, "boot") {
		t.Fatalf("assembly missing boot object: %q", evidence)
	}
}

func factsContain(facts []string, needle string) bool {
	n := strings.ToLower(needle)
	for _, s := range facts {
		if strings.Contains(strings.ToLower(s), n) {
			return true
		}
	}
	return false
}

func searchHayContains(hits []MemoryEntry, needle string) bool {
	n := strings.ToLower(needle)
	for _, h := range hits {
		if strings.Contains(strings.ToLower(entryKeywordHaystack(h)), n) {
			return true
		}
	}
	return false
}

func summariesOf(hits []MemoryEntry) []string {
	out := make([]string, 0, len(hits))
	for _, h := range hits {
		out = append(out, h.Content.Summary)
	}
	return out
}
