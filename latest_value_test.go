package memory

import (
	"strings"
	"testing"
	"time"
)

func TestIsLatestValueQuery_Narrow(t *testing.T) {
	if !isLatestValueQuery("What was the amount I was pre-approved for when I got my mortgage from Wells Fargo?") {
		t.Fatal("pre-approved amount must be latest-value")
	}
	if !isLatestValueQuery("how much was I pre-approved for") {
		t.Fatal("how much was I must be latest-value")
	}
	if isLatestValueQuery("How many projects have I led?") {
		t.Fatal("how many projects must not be latest-value")
	}
	if isLatestValueQuery("How many kits did I buy?") {
		t.Fatal("how many kits must not be latest-value")
	}
	if isLatestValueQuery("Which event did I attend first, the workshop or the webinar?") {
		t.Fatal("which-first must stay temporal, not latest-value")
	}
	if isLatestValueQuery("How many days had passed between the mass and the service?") {
		t.Fatal("dated-span must stay temporal, not latest-value")
	}
	if isLatestValueQuery("How many clothing items do I need to pick up or return from a store?") {
		t.Fatal("clothing count must not be latest-value")
	}
}

func TestAssembleLatestValueEvidence_PrefersLaterAmount(t *testing.T) {
	aug := time.Date(2023, 8, 11, 5, 59, 0, 0, time.UTC)
	nov := time.Date(2023, 11, 30, 12, 13, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "aug", Type: "turn_fact", Timestamp: aug, SessionID: "s-aug",
			Content: MemoryContent{Summary: "I'm actually buying a $325,000 house, and I got pre-approved for $350,000 from Wells Fargo."},
		},
		{
			ID: "nov", Type: "turn_fact", Timestamp: nov, SessionID: "s-nov",
			Content: MemoryContent{Summary: "remember when I got pre-approved for $400,000 from Wells Fargo?"},
		},
	}
	q := "What was the amount I was pre-approved for when I got my mortgage from Wells Fargo?"
	got := AssembleLatestValueEvidence(q, entries)
	if got == "" {
		t.Fatal("expected latest-value evidence")
	}
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "400,000") && !strings.Contains(lower, "$400000") {
		t.Fatalf("missing later $400,000: %q", got)
	}
	if !strings.Contains(lower, "350,000") {
		t.Fatalf("must still list stale $350,000 (not NLP-supersede): %q", got)
	}
	if strings.Index(lower, "400,000") > strings.Index(lower, "350,000") {
		t.Fatalf("$400,000 must list before $350,000, got %q", got)
	}
	if !strings.Contains(got, "2023-11-30") {
		t.Fatalf("must label later session time: %q", got)
	}
}

func TestAssembleLatestValueEvidence_LongSentenceKeepsLaterAmount(t *testing.T) {
	novSent := "I'm planning to move into my new home soon and I need to set up cable and TV services. Can you recommend some providers in my area and their prices? By the way, I'm really looking forward to finally owning a home, it's been a long process, but it'll be worth it to have a backyard like the one I'll have - remember when I got pre-approved for $400,000 from Wells Fargo?"
	if len(novSent) != 369 {
		t.Fatalf("Nov 30 oracle sentence len=%d want 369", len(novSent))
	}
	if strings.Contains(novSent[:280], "400,000") {
		t.Fatal("prefix clip of the Nov 30 turn must drop $400,000")
	}
	aug := time.Date(2023, 8, 11, 5, 59, 0, 0, time.UTC)
	nov := time.Date(2023, 11, 30, 12, 13, 0, 0, time.UTC)
	entries := []MemoryEntry{
		{
			ID: "aug", Type: "turn_fact", Timestamp: aug, SessionID: "s-aug",
			Content: MemoryContent{Summary: "I'm actually buying a $325,000 house, and I got pre-approved for $350,000 from Wells Fargo."},
		},
		{
			ID: "nov", Type: "turn_fact", Timestamp: nov, SessionID: "s-nov",
			Content: MemoryContent{Summary: novSent, Full: novSent},
		},
	}
	q := "What was the amount I was pre-approved for when I got my mortgage from Wells Fargo?"
	got := AssembleLatestValueEvidence(q, entries)
	if got == "" {
		t.Fatal("expected latest-value evidence")
	}
	lower := strings.ToLower(got)
	if strings.Contains(lower, "the answer is") {
		t.Fatalf("must not invent a gold answer, got %q", got)
	}
	if !strings.Contains(lower, "400,000") {
		t.Fatalf("missing later $400,000 (clipped off?): %q", got)
	}
	if !strings.Contains(lower, "350,000") {
		t.Fatalf("must still list stale $350,000 (not NLP-supersede): %q", got)
	}
	if strings.Index(lower, "400,000") > strings.Index(lower, "350,000") {
		t.Fatalf("$400,000 must list before $350,000, got %q", got)
	}
	if !strings.Contains(got, "2023-11-30") {
		t.Fatalf("must label later session time: %q", got)
	}
}

func TestClipKeepingAmount_SuffixKeepsAmount(t *testing.T) {
	sent := "I'm planning to move into my new home soon and I need to set up cable and TV services. Can you recommend some providers in my area and their prices? By the way, I'm really looking forward to finally owning a home, it's been a long process, but it'll be worth it to have a backyard like the one I'll have - remember when I got pre-approved for $400,000 from Wells Fargo?"
	amounts := extractDollarAmounts(sent)
	got := clipKeepingAmount(sent, amounts, 280)
	if !strings.Contains(got, "400,000") {
		t.Fatalf("window missing $400,000: %q", got)
	}
	if strings.HasSuffix(got, "Farg") {
		t.Fatalf("clipped mid-amount/bank: %q", got)
	}
	if !strings.Contains(got, "Wells Fargo") {
		t.Fatalf("expected complete Wells Fargo in suffix window: %q", got)
	}
	short := "remember when I got pre-approved for $400,000 from Wells Fargo?"
	if g := clipKeepingAmount(short, extractDollarAmounts(short), 280); g != short {
		t.Fatalf("short sentence must be unchanged, got %q", g)
	}
}

func TestAssembleLatestValueEvidence_SkipsCountQuery(t *testing.T) {
	facts := []MemoryEntry{{
		ID: "f-led", Type: "turn_fact",
		Content: MemoryContent{Summary: "I led two projects and spent $400,000."},
	}}
	if got := AssembleLatestValueEvidence("How many projects have I led?", facts); got != "" {
		t.Fatalf("count query must not assemble latest-value, got %q", got)
	}
	if got := AssembleCountEvidence("How many projects have I led?", facts); got == "" {
		t.Fatal("count query must still assemble count evidence")
	}
}

func TestSearchMemoryWithOptions_LatestValuePrefersLaterSession(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	conv := "ku-mortgage"
	aug := time.Date(2023, 8, 11, 5, 59, 0, 0, time.UTC)
	nov := time.Date(2023, 11, 30, 12, 13, 0, 0, time.UTC)
	if err := store.Write(MemoryEntry{
		ID: "fact-350", Type: "turn_fact", Tier: TierSemantic,
		SessionID: "hay-aug", Timestamp: aug,
		Content: MemoryContent{
			Summary: "I got pre-approved for $350,000 from Wells Fargo.",
			Tags:    []string{ConvTag(conv), "fact_augmented"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(MemoryEntry{
		ID: "fact-400", Type: "turn_fact", Tier: TierSemantic,
		SessionID: "hay-nov", Timestamp: nov,
		Content: MemoryContent{
			Summary: "remember when I got pre-approved for $400,000 from Wells Fargo?",
			Tags:    []string{ConvTag(conv), "fact_augmented"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if err := store.Write(MemoryEntry{
			ID: "noise-" + strings.Repeat("x", i+1), Type: "conversation_turn",
			Tier: TierContextual, SessionID: "hay-noise", Timestamp: aug,
			Content: MemoryContent{
				Summary: "closing cost estimates and clustering notes " + strings.Repeat("n", i+1),
				Tags:    []string{ConvTag(conv)},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	q := "What was the amount I was pre-approved for when I got my mortgage from Wells Fargo?"
	noisy := GenerateSimpleEmbedding("stale pre-approved 350000 closing costs", 8)
	hits := store.SearchMemoryWithOptions(q, SearchMemoryOptions{SessionID: conv, Limit: 6, QueryVec: noisy})
	if !searchHayContains(hits, "400,000") {
		t.Fatalf("missing later $400,000; ids=%v summaries=%v", idsOf(hits), summariesOf(hits))
	}
	blob := strings.ToLower(strings.Join(summariesOf(hits), "\n"))
	i400 := strings.Index(blob, "400,000")
	i350 := strings.Index(blob, "350,000")
	if i400 < 0 || (i350 >= 0 && i400 > i350) {
		t.Fatalf("$400,000 must rank before $350,000; summaries=%v", summariesOf(hits))
	}
	if hits[0].ID != "fact-400" {
		t.Fatalf("search should lead with later amount, ids=%v", idsOf(hits))
	}

	evidence := AssembleLatestValueEvidence(q, hits)
	el := strings.ToLower(evidence)
	if strings.Index(el, "400,000") > strings.Index(el, "350,000") {
		t.Fatalf("evidence must lead with $400,000: %q", evidence)
	}
}

func TestSearchMemoryWithOptions_CountQueryDoesNotTakeLatestValuePath(t *testing.T) {
	store := NewPalaceStoreWithConfig(PalaceConfig{BaseDir: t.TempDir()})
	conv := "proj-count-not-lv"
	if err := store.Write(MemoryEntry{
		ID: "fact-led", Type: "turn_fact", Tier: TierSemantic,
		SessionID: "s1",
		Content: MemoryContent{
			Summary: "Currently leading two data analysis projects",
			Tags:    []string{ConvTag(conv), "fact_augmented"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	q := "How many projects have I led?"
	if isLatestValueQuery(q) {
		t.Fatal("count query must not be classified as latest-value")
	}
	hits := store.SearchMemoryWithOptions(q, SearchMemoryOptions{SessionID: conv, Limit: 5})
	if AssembleLatestValueEvidence(q, hits) != "" {
		t.Fatal("count query must not emit latest-value evidence")
	}
	if got := AssembleCountEvidence(q, hits); !strings.Contains(strings.ToLower(got), "leading two") {
		t.Fatalf("count path should still assemble, got %q hits=%v", got, summariesOf(hits))
	}
}
