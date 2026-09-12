package memory

import "testing"

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
