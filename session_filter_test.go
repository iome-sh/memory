package memory

import (
	"strconv"
	"strings"
	"testing"
	"time"
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
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("project unique-entity path must not label N distinct items: %q", evidence)
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
	if !strings.Contains(got, "1. ") || !strings.Contains(got, "2. ") {
		t.Fatalf("compound should be two numbered bullets, got %q", got)
	}
	if strings.Contains(got, "3. ") {
		t.Fatalf("compound should not emit a third numbered bullet, got %q", got)
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
	if !strings.Contains(lower, "3 distinct items") {
		t.Fatalf("assembly must label 3 distinct clothing clusters: %q", evidence)
	}
	for _, mark := range []string{"1. ", "2. ", "3. "} {
		if !strings.Contains(evidence, mark) {
			t.Fatalf("clothing evidence must number bullets, missing %q: %q", mark, evidence)
		}
	}
	if strings.Contains(evidence, "\n- ") {
		t.Fatalf("clothing bullets should be numbered, not dashed: %q", evidence)
	}
	if strings.Contains(lower, "the answer is") {
		t.Fatalf("must not invent a numeric gold: %q", evidence)
	}
}

func TestExtractAtomicFacts_ModelKitsHoursPlantsDates(t *testing.T) {
	kits := ExtractAtomicFacts(MemoryEntry{Content: MemoryContent{
		Full: "I recently finished a simple Revell F-15 Eagle kit that I picked up at the hobby store. Clustering notes are unrelated.",
	}})
	if !factsContain(kits, "f-15") && !factsContain(kits, "revell") {
		t.Fatalf("expected F-15 kit fact, got %v", kits)
	}

	hours := ExtractAtomicFacts(MemoryEntry{Content: MemoryContent{
		Full: "I drove for six hours to Washington D.C. recently. Elbow method chatter.",
	}})
	if !factsContain(hours, "six hours") && !factsContain(hours, "6 hours") {
		t.Fatalf("expected drove-hours fact, got %v", hours)
	}

	plants := ExtractAtomicFacts(MemoryEntry{Content: MemoryContent{
		Full: "I'm trying to care for my peace lily and a succulent I got from the nursery.",
	}})
	if !factsContain(plants, "peace lily") || !factsContain(plants, "succulent") {
		t.Fatalf("expected both plants in extract, got %v", plants)
	}

	dated := ExtractAtomicFacts(MemoryEntry{Content: MemoryContent{
		Full: "I attended the Sunday mass at St. Mary's Church on January 2nd. Unrelated clustering.",
	}})
	if !factsContain(dated, "january 2nd") && !factsContain(dated, "January 2nd") {
		t.Fatalf("expected January 2nd date fact, got %v", dated)
	}

	assist := "Congratulations on completing your Tamiya 1/48 scale Spitfire Mk.V."
	if factsContain(ExtractAtomicFacts(MemoryEntry{Content: MemoryContent{Full: assist}}), "spitfire") {
		t.Fatalf("assistant kit chatter must not extract: %q", assist)
	}
}

func TestAssembleCountEvidence_FiveKitsDedupeB29(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("k1", "s1", "I recently finished a simple Revell F-15 Eagle kit that I picked up at the hobby store."),
		factEntry("k2", "s1", "I'm thinking of working on a 1/72 scale B-29 bomber next."),
		factEntry("k3", "s2", "I recently finished a Tamiya 1/48 scale Spitfire Mk.V."),
		factEntry("k4", "s3", "I started working on a diorama featuring a 1/16 scale German Tiger I tank."),
		factEntry("k5", "s4", "I just got this 1/72 scale B-29 bomber kit and a 1/24 scale '69 Camaro at a model show."),
	}
	q := "How many model kits have I worked on or bought?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Count(lower, "[kit:") != 5 {
		t.Fatalf("want 5 kit clusters, got %q", got)
	}
	if strings.Count(lower, "[kit:b-29]") != 1 {
		t.Fatalf("repeated B-29 must be one cluster, got %q", got)
	}
	for _, key := range []string{"[kit:f-15]", "[kit:spitfire]", "[kit:tiger]", "[kit:camaro]"} {
		if !strings.Contains(lower, key) {
			t.Fatalf("missing %s in %q", key, got)
		}
	}
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("kit unique-entity path must not label N distinct items: %q", got)
	}
	if !strings.Contains(got, "\n- ") {
		t.Fatalf("kit unique-entity path must stay unnumbered dashes: %q", got)
	}
	if strings.Contains(got, "\n1. ") || strings.Contains(got, "\n2. ") || strings.Contains(got, "\n3. ") {
		t.Fatalf("kit unique-entity path must not number bullets: %q", got)
	}
}

func TestAssembleCountEvidence_ThreeHourDestinations(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("h1", "s1", "My recent trip to Outer Banks in North Carolina only took me four hours to drive there."),
		factEntry("h2", "s2", "I drove for six hours to Washington D.C. recently."),
		factEntry("h3", "s3", "On my recent trip to the mountains in Tennessee I drove for five hours to get there."),
	}
	q := "How many hours in total did I spend driving to my three road trip destinations combined?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Count(lower, "[hours:") != 3 {
		t.Fatalf("want 3 hour clusters, got %q", got)
	}
	if !strings.Contains(lower, "outer banks") && !strings.Contains(lower, "outer-banks") {
		t.Fatalf("missing Outer Banks destination: %q", got)
	}
	if !strings.Contains(lower, "washington") {
		t.Fatalf("missing Washington destination: %q", got)
	}
	if !strings.Contains(lower, "tennessee") {
		t.Fatalf("missing Tennessee destination: %q", got)
	}
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("hour unique-entity path must not label N distinct items: %q", got)
	}
}

func TestAssembleCountEvidence_ThreePlantsOneTurnTwoNames(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("p1", "s1", "I'm trying to care for my peace lily and a succulent I got from the nursery."),
		factEntry("p2", "s2", "I got a snake plant from my sister last month."),
	}
	q := "How many plants did I acquire in the last month?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Count(lower, "[plant:") != 3 {
		t.Fatalf("want 3 plant clusters, got %q", got)
	}
	for _, key := range []string{"[plant:peace-lily]", "[plant:succulent]", "[plant:snake-plant]"} {
		if !strings.Contains(lower, key) {
			t.Fatalf("missing %s in %q", key, got)
		}
	}
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("plant unique-entity path must not label N distinct items: %q", got)
	}
}

func TestAssembleCountEvidence_HoursUnseenDestinations(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("h1", "s1", "I drove six hours to Yosemite."),
		factEntry("h2", "s2", "I spent eight hours in Zion."),
		factEntry("h3", "s3", "On my recent trip to the mountains in Tennessee I drove for five hours to get there."),
	}
	q := "How many hours in total did I spend driving to destinations?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Count(lower, "[hours:") != 3 {
		t.Fatalf("want 3 hour clusters (yosemite, zion, tennessee), got %q", got)
	}
	for _, key := range []string{"[hours:yosemite]", "[hours:zion]", "[hours:tennessee]"} {
		if !strings.Contains(lower, key) {
			t.Fatalf("missing %s in %q", key, got)
		}
	}
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("hour unique-entity path must not label N distinct items: %q", got)
	}
	if strings.Contains(lower, "the answer is") {
		t.Fatalf("must not invent a numeric gold: %q", got)
	}
	if !strings.Contains(got, "\n- ") {
		t.Fatalf("hour unique-entity path must stay unnumbered dashes: %q", got)
	}
}

func TestAssembleCountEvidence_HoursRepeatedDestOneCluster(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("h1", "s1", "I drove six hours to the Outer Banks."),
		factEntry("h2", "s2", "I drove four hours to the Outer Banks last summer."),
	}
	q := "How many hours in total did I spend driving to destinations?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Count(lower, "[hours:") != 1 {
		t.Fatalf("repeated Outer Banks must be one cluster, got %q", got)
	}
	if !strings.Contains(lower, "[hours:outer-banks]") {
		t.Fatalf("missing outer-banks cluster: %q", got)
	}
}

func TestAssembleCountEvidence_KitsUnseenMustangAndRepeatedB29(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("k1", "s1", "I bought a P-51 Mustang kit at the hobby shop."),
		factEntry("k2", "s2", "I'm thinking of working on a 1/72 scale B-29 bomber next."),
		factEntry("k3", "s3", "I just got this 1/72 scale B-29 bomber kit."),
	}
	q := "How many model kits have I worked on or bought?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Count(lower, "[kit:") != 2 {
		t.Fatalf("want 2 kit clusters (mustang + b-29), got %q", got)
	}
	if strings.Count(lower, "[kit:b-29]") != 1 {
		t.Fatalf("repeated B-29 must be one cluster, got %q", got)
	}
	if !strings.Contains(lower, "[kit:p-51-mustang]") {
		t.Fatalf("missing P-51 Mustang kit cluster: %q", got)
	}
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("kit unique-entity path must not label N distinct items: %q", got)
	}
	if !strings.Contains(got, "\n- ") {
		t.Fatalf("kit unique-entity path must stay unnumbered dashes: %q", got)
	}
	if strings.Contains(got, "\n1. ") || strings.Contains(got, "\n2. ") {
		t.Fatalf("kit unique-entity path must not number bullets: %q", got)
	}
}

func TestAssembleCountEvidence_KitsTwoNamesOneTurn(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("k1", "s1", "I bought a P-51 Mustang kit and a B-29 Superfortress kit."),
	}
	q := "How many model kits have I worked on or bought?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Count(lower, "[kit:") != 2 {
		t.Fatalf("two kit names in one turn must be two clusters, got %q", got)
	}
	if !strings.Contains(lower, "[kit:p-51-mustang]") || !strings.Contains(lower, "[kit:b-29]") {
		t.Fatalf("want mustang + b-29, got %q", got)
	}
}

func TestAssembleCountEvidence_PlantsCatalogNotPowerPlant(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("p1", "s1", "I'm trying to care for my peace lily and a succulent I got from the nursery."),
		factEntry("p2", "s2", "I toured the power plant down the river."),
		factEntry("p3", "s3", "I got a monstera plant from the market."),
	}
	q := "How many plants did I acquire in the last month?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Count(lower, "[plant:") != 3 {
		t.Fatalf("want 3 plant clusters (peace-lily, succulent, monstera), got %q", got)
	}
	for _, key := range []string{"[plant:peace-lily]", "[plant:succulent]", "[plant:monstera]"} {
		if !strings.Contains(lower, key) {
			t.Fatalf("missing %s in %q", key, got)
		}
	}
	if strings.Contains(lower, "[plant:power]") {
		t.Fatalf("power plant must not cluster: %q", got)
	}
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("plant unique-entity path must not label N distinct items: %q", got)
	}
}

func TestAssembleCountEvidence_ClothingQueryNotUniqueEntity(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("c1", "s1", "I need to return some boots to Zara."),
		factEntry("c2", "s2", "I still need to pick up my dry cleaning for the navy blue blazer."),
		factEntry("k1", "s3", "I bought a P-51 Mustang kit at the hobby shop."),
		factEntry("r1", "s4", "I tried Banchan Korean restaurant."),
		factEntry("r2", "s5", "I've tried four different ones so far."),
	}
	q := "How many items of clothing do I need to pick up or return from a store?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Contains(lower, "[kit:") || strings.Contains(lower, "[plant:") || strings.Contains(lower, "[hours:") || strings.Contains(lower, "[restaurant:") {
		t.Fatalf("clothing query must not take unique-entity path: %q", got)
	}
	if strings.Contains(lower, "[time:") || strings.Contains(lower, "four different") {
		t.Fatalf("clothing query must not emit restaurant tried-count: %q", got)
	}
	if !strings.Contains(lower, "distinct items") {
		t.Fatalf("clothing path should keep N distinct items: %q", got)
	}
	if !strings.Contains(got, "1. ") {
		t.Fatalf("clothing path should number bullets: %q", got)
	}
}

func TestAssembleCountEvidence_RestaurantsThreeNamesDedupeBanchan(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("r1", "s1", "I tried Banchan Korean restaurant"),
		factEntry("r2", "s2", "I tried Noodle House Korean restaurant"),
		factEntry("r3", "s3", "I tried Seoul Kitchen"),
		factEntry("r4", "s4", "I went back to Banchan"),
	}
	q := "How many Korean restaurants have I tried in my city?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Count(lower, "[restaurant:") != 3 {
		t.Fatalf("want 3 restaurant clusters (Banchan once), got %q", got)
	}
	if strings.Count(lower, "[restaurant:banchan]") != 1 {
		t.Fatalf("repeated Banchan must be one cluster, got %q", got)
	}
	for _, key := range []string{"[restaurant:banchan]", "[restaurant:noodle-house]", "[restaurant:seoul-kitchen]"} {
		if !strings.Contains(lower, key) {
			t.Fatalf("missing %s in %q", key, got)
		}
	}
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("restaurant unique-entity path must not label N distinct items: %q", got)
	}
	if !strings.Contains(got, "\n- ") {
		t.Fatalf("restaurant unique-entity path must stay unnumbered dashes: %q", got)
	}
	if strings.Contains(got, "\n1. ") || strings.Contains(got, "\n2. ") || strings.Contains(got, "\n3. ") {
		t.Fatalf("restaurant unique-entity path must not number bullets: %q", got)
	}
	if strings.Contains(lower, "the answer is") {
		t.Fatalf("must not invent a numeric gold: %q", got)
	}
}

func TestAssembleCountEvidence_RestaurantsQuotedNameAndTwoInOneTurn(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("r1", "s1", `I tried Banchan Korean restaurant and "Harbor Bibimbap"`),
		factEntry("r2", "s2", `I went back to Banchan`),
	}
	q := "How many restaurants have I tried?"
	got := AssembleCountEvidence(q, facts)
	lower := strings.ToLower(got)
	if strings.Count(lower, "[restaurant:") != 2 {
		t.Fatalf("want 2 restaurant clusters (Banchan + Harbor Bibimbap), got %q", got)
	}
	if !strings.Contains(lower, "[restaurant:banchan]") || !strings.Contains(lower, "[restaurant:harbor-bibimbap]") {
		t.Fatalf("want banchan + harbor-bibimbap, got %q", got)
	}
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("restaurant unique-entity path must not label N distinct items: %q", got)
	}
	if strings.Contains(lower, "the answer is") {
		t.Fatalf("must not invent a numeric gold: %q", got)
	}
}

func TestNormalizeRestaurantName_DishNotVenue(t *testing.T) {
	if got := normalizeRestaurantName("Korean-style BBQ"); got != "" {
		t.Fatalf("Korean-style BBQ is a dish, got %q", got)
	}
	if got := normalizeRestaurantName("Korean BBQ"); got != "" {
		t.Fatalf("Korean BBQ is a dish, got %q", got)
	}
	if got := normalizeRestaurantName("If"); got != "" {
		t.Fatalf("if is a stop token, got %q", got)
	}
	if got := normalizeRestaurantName("Seoul Kitchen"); got != "Seoul Kitchen" {
		t.Fatalf("Seoul Kitchen is a venue, got %q", got)
	}
	if got := extractRestaurantPhrases("I'm making Korean-style BBQ at home this week."); len(got) != 0 {
		t.Fatalf("cooking Korean-style BBQ must not extract a venue, got %q", got)
	}
	if got := extractRestaurantPhrases("If restaurants in my city have bibimbap, let me know."); len(got) != 0 {
		t.Fatalf("If restaurants must not cluster, got %q", got)
	}
}

func TestAssembleCountEvidence_RestaurantsTriedCountLatestFirst(t *testing.T) {
	aug := time.Date(2023, 8, 11, 5, 59, 0, 0, time.UTC)
	nov := time.Date(2023, 11, 30, 12, 13, 0, 0, time.UTC)
	facts := []MemoryEntry{
		{
			ID: "aug", Type: "turn_fact", Timestamp: aug, SessionID: "s-aug",
			Content: MemoryContent{Summary: "I've tried three different ones recently, and each has its own unique flavor and style."},
		},
		{
			ID: "nov", Type: "turn_fact", Timestamp: nov, SessionID: "s-nov",
			Content: MemoryContent{Summary: "I've tried four different ones so far, and I'm always looking for new recommendations."},
		},
		{
			ID: "cook", Type: "turn_fact", Timestamp: aug, SessionID: "s-cook",
			Content: MemoryContent{Summary: "I'm making Korean-style BBQ at home this week."},
		},
		{
			ID: "if", Type: "turn_fact", Timestamp: nov, SessionID: "s-if",
			Content: MemoryContent{Summary: "If restaurants in my city have bibimbap, let me know."},
		},
	}
	q := "How many Korean restaurants have I tried in my city?"
	got := AssembleCountEvidence(q, facts)
	if got == "" {
		t.Fatal("expected restaurant count evidence")
	}
	lower := strings.ToLower(got)
	fourAt := strings.Index(lower, "four")
	threeAt := strings.Index(lower, "three")
	if fourAt < 0 || threeAt < 0 {
		t.Fatalf("must list both four and three tried-count mentions: %q", got)
	}
	if fourAt > threeAt {
		t.Fatalf("four must list before three (latest first): %q", got)
	}
	if !strings.Contains(got, "2023-11-30") {
		t.Fatalf("must label Nov time on latest tried-count: %q", got)
	}
	if !strings.Contains(got, "[time: "+nov.UTC().Format(time.RFC3339)+"]") {
		t.Fatalf("must label latest tried-count like latest-value: %q", got)
	}
	if strings.Contains(lower, "[restaurant:if]") {
		t.Fatalf("must not cluster stop-token if: %q", got)
	}
	if strings.Contains(lower, "korean-style-bbq") || strings.Contains(lower, "[restaurant:korean") {
		t.Fatalf("must not treat Korean-style BBQ as a venue: %q", got)
	}
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("restaurant unique-entity path must not label N distinct items: %q", got)
	}
	if strings.Contains(got, "\n1. ") || strings.Contains(got, "\n2. ") {
		t.Fatalf("restaurant unique-entity path must not number bullets: %q", got)
	}
	if strings.Contains(lower, "the answer is") || strings.Contains(lower, "the answer is 4") {
		t.Fatalf("must not invent a numeric gold: %q", got)
	}
}

func TestAssembleCountEvidence_DatedSpanIsNotCountEvidence(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("d1", "s1", "I attended the Sunday mass at St. Mary's Church on January 2nd."),
		factEntry("d2", "s2", "I just came from the Ash Wednesday service at the cathedral on February 1st."),
	}
	q := "How many days had passed between the Sunday mass at St. Mary's Church and the Ash Wednesday service at the cathedral?"
	if got := AssembleCountEvidence(q, facts); got != "" {
		t.Fatalf("dated-span must not assemble count evidence, got %q", got)
	}
}

func TestSearchMemoryWithOptions_CountQueryAssemblesFiveKits(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	conv := "kit-count"
	turns := []struct{ id, sess, full string }{
		{"turn-f15", "hay-1", "I recently finished a simple Revell F-15 Eagle kit that I picked up at the hobby store."},
		{"turn-b29a", "hay-1", "I'm thinking of working on a 1/72 scale B-29 bomber next."},
		{"turn-spit", "hay-2", "I recently finished a Tamiya 1/48 scale Spitfire Mk.V."},
		{"turn-tiger", "hay-3", "I started working on a diorama featuring a 1/16 scale German Tiger I tank."},
		{"turn-camaro", "hay-4", "I just got this 1/72 scale B-29 bomber kit and a 1/24 scale '69 Camaro at a model show."},
	}
	for _, tr := range turns {
		if err := store.IngestTurn(MemoryEntry{
			ID: tr.id, SessionID: tr.sess,
			Content: MemoryContent{Full: tr.full, Tags: []string{ConvTag(conv)}},
		}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 16; i++ {
		if err := store.IngestTurn(MemoryEntry{
			ID: "turn-noise-" + strconv.Itoa(i), SessionID: "hay-noise",
			Content: MemoryContent{
				Full: "The elbow method is an excellent choice for clustering how many model kits appear in dashboard metrics " + strconv.Itoa(i) + ".",
				Tags: []string{ConvTag(conv)},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	q := "How many model kits have I worked on or bought?"
	hits := store.SearchMemoryWithOptions(q, SearchMemoryOptions{SessionID: conv, Limit: 8})
	evidence := AssembleCountEvidence(q, hits)
	lower := strings.ToLower(evidence)
	if strings.Count(lower, "[kit:") != 5 {
		t.Fatalf("search assembly want 5 kit clusters; ids=%v evidence=%q", idsOf(hits), evidence)
	}
	if strings.Count(lower, "[kit:b-29]") != 1 {
		t.Fatalf("B-29 must dedupe across sessions: %q", evidence)
	}
	if strings.Contains(lower, "distinct items") {
		t.Fatalf("kit unique-entity path must not label N distinct items: %q", evidence)
	}
	if !strings.Contains(evidence, "\n- ") {
		t.Fatalf("kit unique-entity path must stay unnumbered dashes: %q", evidence)
	}
	if strings.Contains(evidence, "\n1. ") || strings.Contains(evidence, "\n2. ") || strings.Contains(evidence, "\n3. ") {
		t.Fatalf("kit unique-entity path must not number bullets: %q", evidence)
	}
}

func TestAssembleTemporalEvidence_WebinarBeforeWorkshopSameTimestamp(t *testing.T) {
	ts := time.Date(2023, 5, 28, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "workshop", Type: "turn_fact", Timestamp: ts, SessionID: "s-work",
			Content: MemoryContent{Summary: "I attended the workshop on Effective Time Management last Saturday."},
		},
		{
			ID: "webinar", Type: "turn_fact", Timestamp: ts, SessionID: "s-web",
			Content: MemoryContent{Summary: "I participated in a webinar on Data Analysis using Python two months ago."},
		},
	}
	q := "Which event did I attend first, the 'Effective Time Management' workshop or the 'Data Analysis using Python' webinar?"
	got := AssembleTemporalEvidence(q, entries)
	if got == "" {
		t.Fatal("expected temporal evidence")
	}
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "two months ago") {
		t.Fatalf("missing webinar text date: %q", got)
	}
	if !strings.Contains(lower, "last saturday") {
		t.Fatalf("missing workshop text date: %q", got)
	}
	if !strings.Contains(lower, "ingest:") {
		t.Fatalf("must label ingest Timestamp separately: %q", got)
	}
	web := strings.Index(lower, "webinar")
	work := strings.Index(lower, "workshop")
	if web < 0 || work < 0 || web > work {
		t.Fatalf("webinar (two months ago) must list before workshop, got %q", got)
	}
	if strings.Contains(lower, "days apart") || strings.Contains(lower, "weeks apart") ||
		strings.Contains(lower, "months apart") || strings.Contains(lower, "calendar months") {
		t.Fatalf("which-first must not append a dated-span delta: %q", got)
	}
}

func TestAssembleTemporalEvidence_DatedSpanTwelveDaysApart(t *testing.T) {
	ts := time.Date(2023, 5, 20, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "start", Type: "turn_fact", Timestamp: ts, SessionID: "s-start",
			Content: MemoryContent{Summary: "I started training on May 3."},
		},
		{
			ID: "end", Type: "turn_fact", Timestamp: ts, SessionID: "s-end",
			Content: MemoryContent{Summary: "I finished training on May 15."},
		},
	}
	q := "How many days passed between starting training and finishing training?"
	got := AssembleTemporalEvidence(q, entries)
	if !strings.Contains(got, "May 3") {
		t.Fatalf("missing May 3 bullet: %q", got)
	}
	if !strings.Contains(got, "May 15") {
		t.Fatalf("missing May 15 bullet: %q", got)
	}
	wantDelta := "text dates 12 days apart (May 3 → May 15)"
	if !strings.Contains(got, wantDelta) {
		t.Fatalf("missing text-date delta %q in %q", wantDelta, got)
	}
	if !strings.Contains(got, "2 distinct events") {
		t.Fatalf("must keep distinct-events prefix: %q", got)
	}
	if strings.Contains(strings.ToLower(got), "the answer is") {
		t.Fatalf("must not invent a gold answer: %q", got)
	}
}

func TestAssembleTemporalEvidence_DatedSpanUsesTextDatesNotIngest(t *testing.T) {
	ts := time.Date(2023, 2, 20, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "mass", Type: "turn_fact", Timestamp: ts, SessionID: "s-mass",
			Content: MemoryContent{Summary: "I attended the Sunday mass at St. Mary's Church on January 2nd."},
		},
		{
			ID: "ash", Type: "turn_fact", Timestamp: ts, SessionID: "s-ash",
			Content: MemoryContent{Summary: "I just came from the Ash Wednesday service at the cathedral on February 1st."},
		},
	}
	q := "How many days had passed between the Sunday mass at St. Mary's Church and the Ash Wednesday service at the cathedral?"
	got := AssembleTemporalEvidence(q, entries)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "january 2nd") {
		t.Fatalf("missing January 2nd bullet: %q", got)
	}
	if !strings.Contains(lower, "february 1st") {
		t.Fatalf("missing February 1st bullet: %q", got)
	}
	if !strings.Contains(got, "text dates 30 days apart (January 2nd → February 1st)") {
		t.Fatalf("delta must use parsed text dates, got %q", got)
	}
	if strings.Contains(got, "text dates 0 days apart") {
		t.Fatalf("delta must not use shared ingest Timestamp: %q", got)
	}
	if strings.Contains(lower, "the answer is") {
		t.Fatalf("must not invent a gold answer: %q", got)
	}
}

func TestAssembleTemporalEvidence_SingleDatedBulletNoDelta(t *testing.T) {
	ts := time.Date(2023, 5, 20, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "concert", Type: "turn_fact", Timestamp: ts, SessionID: "s-concert",
			Content: MemoryContent{Summary: "I attended the concert on May 3."},
		},
	}
	q := "How many days have passed since I attended the concert?"
	got := AssembleTemporalEvidence(q, entries)
	if !strings.Contains(got, "May 3") {
		t.Fatalf("missing dated bullet: %q", got)
	}
	if strings.Contains(strings.ToLower(got), "days apart") || strings.Contains(strings.ToLower(got), "weeks apart") ||
		strings.Contains(strings.ToLower(got), "months apart") || strings.Contains(strings.ToLower(got), "calendar months") {
		t.Fatalf("one dated bullet must not append a delta: %q", got)
	}
}

func TestAssembleTemporalEvidence_SlashDatesHouseAndRachel(t *testing.T) {
	ts := time.Date(2022, 3, 2, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "start", Type: "turn_fact", Timestamp: ts, SessionID: "s1",
			Content: MemoryContent{Summary: "Since I started working with Rachel on 2/15, I'm hoping she can help."},
		},
		{
			ID: "house", Type: "turn_fact", Timestamp: ts, SessionID: "s2",
			Content: MemoryContent{Summary: "I recently saw a house that I really love on 3/1."},
		},
	}
	q := "How many days did it take for me to find a house I loved after starting to work with Rachel?"
	got := AssembleTemporalEvidence(q, entries)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "2/15") {
		t.Fatalf("missing 2/15 bullet: %q", got)
	}
	if !strings.Contains(lower, "3/1") {
		t.Fatalf("missing 3/1 bullet: %q", got)
	}
	if !strings.Contains(got, "text dates 14 days apart (2/15 → 3/1)") {
		t.Fatalf("slash-date delta must use parsed text dates, got %q", got)
	}
}

func TestAssembleTemporalEvidence_DaysOnlyOmitsWeekMonthDelta(t *testing.T) {
	ts := time.Date(2023, 5, 20, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "start", Type: "turn_fact", Timestamp: ts, SessionID: "s-start",
			Content: MemoryContent{Summary: "I started training on May 3."},
		},
		{
			ID: "end", Type: "turn_fact", Timestamp: ts, SessionID: "s-end",
			Content: MemoryContent{Summary: "I finished training on May 15."},
		},
	}
	q := "How many days passed between starting training and finishing training?"
	got := AssembleTemporalEvidence(q, entries)
	if !strings.Contains(got, "text dates 12 days apart (May 3 → May 15)") {
		t.Fatalf("days-only must keep the days line, got %q", got)
	}
	lower := strings.ToLower(got)
	if strings.Contains(lower, "weeks apart") || strings.Contains(lower, "calendar months") {
		t.Fatalf("days-only query must not append week/month deltas: %q", got)
	}
}

func TestAssembleTemporalEvidence_DatedSpanWeeksOmitsRemainder(t *testing.T) {
	ts := time.Date(2023, 5, 28, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "start", Type: "turn_fact", Timestamp: ts, SessionID: "s-start",
			Content: MemoryContent{Summary: "I started training on May 3."},
		},
		{
			ID: "end", Type: "turn_fact", Timestamp: ts, SessionID: "s-end",
			Content: MemoryContent{Summary: "I finished training on May 24."},
		},
	}
	q := "How many weeks between May 3 and May 24"
	got := AssembleTemporalEvidence(q, entries)
	wantDays := "text dates 21 days apart (May 3 → May 24)"
	if !strings.Contains(got, wantDays) {
		t.Fatalf("missing days line %q in %q", wantDays, got)
	}
	wantWeeks := "text dates 3 weeks apart (floor days/7)"
	if !strings.Contains(got, wantWeeks) {
		t.Fatalf("missing weeks line %q in %q", wantWeeks, got)
	}
	if strings.Contains(got, "remainder") {
		t.Fatalf("exact weeks must omit remainder: %q", got)
	}
	if strings.Contains(got, "calendar months") {
		t.Fatalf("weeks query must not append calendar-month delta: %q", got)
	}
	if strings.Contains(strings.ToLower(got), "the answer is") {
		t.Fatalf("must not invent a gold answer: %q", got)
	}
}

func TestAssembleTemporalEvidence_DatedSpanWeeksWithRemainder(t *testing.T) {
	ts := time.Date(2023, 5, 28, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "start", Type: "turn_fact", Timestamp: ts, SessionID: "s-start",
			Content: MemoryContent{Summary: "I started training on May 3."},
		},
		{
			ID: "end", Type: "turn_fact", Timestamp: ts, SessionID: "s-end",
			Content: MemoryContent{Summary: "I finished training on May 25."},
		},
	}
	q := "How many weeks between May 3 and May 25"
	got := AssembleTemporalEvidence(q, entries)
	if !strings.Contains(got, "text dates 22 days apart (May 3 → May 25)") {
		t.Fatalf("missing days line in %q", got)
	}
	wantWeeks := "text dates 3 weeks apart (floor days/7; remainder 1 days)"
	if !strings.Contains(got, wantWeeks) {
		t.Fatalf("missing remainder weeks line %q in %q", wantWeeks, got)
	}
}

func TestAssembleTemporalEvidence_DatedSpanCalendarMonths(t *testing.T) {
	ts := time.Date(2023, 5, 15, 12, 0, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "start", Type: "turn_fact", Timestamp: ts, SessionID: "s-start",
			Content: MemoryContent{Summary: "I started the course on January 2nd."},
		},
		{
			ID: "end", Type: "turn_fact", Timestamp: ts, SessionID: "s-end",
			Content: MemoryContent{Summary: "I finished the course on April 2nd."},
		},
	}
	q := "How many months between January 2nd and April 2nd"
	got := AssembleTemporalEvidence(q, entries)
	if !strings.Contains(got, "text dates 90 days apart (January 2nd → April 2nd)") {
		t.Fatalf("missing days line in %q", got)
	}
	wantMonths := "text dates 3 calendar months apart (January 2nd → April 2nd)"
	if !strings.Contains(got, wantMonths) {
		t.Fatalf("missing calendar-month line %q in %q", wantMonths, got)
	}
	if strings.Contains(strings.ToLower(got), "weeks apart") {
		t.Fatalf("months query must not append week delta: %q", got)
	}
	if strings.Contains(strings.ToLower(got), "the answer is") {
		t.Fatalf("must not invent a gold answer: %q", got)
	}
}

func TestIsDatedSpanQuery_WeeksMonthsCues(t *testing.T) {
	cases := []struct {
		q    string
		want bool
	}{
		{"How many weeks between May 3 and May 24", true},
		{"How many months between January 2nd and April 2nd", true},
		{"How many weeks have passed since I started training?", true},
		{"How many months have I been taking the medication?", true},
		{"How many weeks have I been taking the medication?", true},
		{"How many months had passed until the follow-up?", true},
		{"How many days passed between starting and finishing?", true},
		{"How many kits have I been building?", false},
		{"How many weeks of vacation did I book?", false},
		{"Which event did I attend first?", false},
	}
	for _, tc := range cases {
		if got := isDatedSpanQuery(tc.q); got != tc.want {
			t.Errorf("isDatedSpanQuery(%q) = %v, want %v", tc.q, got, tc.want)
		}
	}
}

func TestAssembleCountEvidence_WeeksMonthsNotCountEvidence(t *testing.T) {
	facts := []MemoryEntry{
		factEntry("d1", "s1", "I started training on May 3."),
		factEntry("d2", "s2", "I finished training on May 24."),
		factEntry("d3", "s3", "I started the course on January 2nd."),
		factEntry("d4", "s4", "I finished the course on April 2nd."),
	}
	for _, q := range []string{
		"How many weeks between May 3 and May 24",
		"How many months between January 2nd and April 2nd",
		"How many months have I been taking the medication?",
	} {
		if got := AssembleCountEvidence(q, facts); got != "" {
			t.Fatalf("dated-span %q must not assemble count evidence, got %q", q, got)
		}
	}
}

func TestSearchMemoryWithOptions_TemporalOrderPromotesBothEvents(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	conv := "event-order"
	ts := time.Date(2023, 5, 28, 12, 0, 0, 0, time.UTC)
	if err := store.IngestTurn(MemoryEntry{
		ID: "turn-web", SessionID: "hay-web", Timestamp: ts,
		Content: MemoryContent{
			Full: "I participated in a webinar on Data Analysis using Python two months ago.",
			Tags: []string{ConvTag(conv)},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.IngestTurn(MemoryEntry{
		ID: "turn-work", SessionID: "hay-work", Timestamp: ts,
		Content: MemoryContent{
			Full: "I attended the workshop on Effective Time Management last Saturday.",
			Tags: []string{ConvTag(conv)},
		},
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 16; i++ {
		if err := store.IngestTurn(MemoryEntry{
			ID: "turn-noise-" + strconv.Itoa(i), SessionID: "hay-noise", Timestamp: ts,
			Content: MemoryContent{
				Full: "The elbow method is an excellent choice for clustering which event metrics appear first " + strconv.Itoa(i) + ".",
				Tags: []string{ConvTag(conv)},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	q := "Which event did I attend first, the 'Effective Time Management' workshop or the 'Data Analysis using Python' webinar?"
	hits := store.SearchMemoryWithOptions(q, SearchMemoryOptions{SessionID: conv, Limit: 6})
	if !searchHayContains(hits, "webinar") {
		t.Fatalf("missing webinar under small Limit; ids=%v summaries=%v", idsOf(hits), summariesOf(hits))
	}
	if !searchHayContains(hits, "workshop") {
		t.Fatalf("missing workshop under small Limit; ids=%v summaries=%v", idsOf(hits), summariesOf(hits))
	}
	evidence := AssembleTemporalEvidence(q, hits)
	lower := strings.ToLower(evidence)
	web := strings.Index(lower, "webinar")
	work := strings.Index(lower, "workshop")
	if web < 0 || work < 0 || web > work {
		t.Fatalf("webinar must list first in evidence: %q", evidence)
	}
}

func factEntry(id, session, summary string) MemoryEntry {
	return MemoryEntry{
		ID: id, Type: "turn_fact", SessionID: session,
		Content: MemoryContent{Summary: summary, Tags: []string{"fact_augmented"}},
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
