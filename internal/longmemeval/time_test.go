package longmemeval

import (
	"testing"
	"time"
)

func TestParseTime_OfficialAndRFC3339(t *testing.T) {
	t.Parallel()
	official, ok := ParseTime("2023/04/10 (Mon) 17:50")
	if !ok {
		t.Fatal("official cleaned date should parse")
	}
	want := time.Date(2023, 4, 10, 17, 50, 0, 0, time.UTC)
	if !official.Equal(want) {
		t.Fatalf("official = %s, want %s", official, want)
	}

	rfc, ok := ParseTime("2024-03-15T10:00:00Z")
	if !ok {
		t.Fatal("RFC3339 should still parse")
	}
	if got := rfc.UTC(); !got.Equal(time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("rfc = %s", got)
	}

	if _, ok := ParseTime(""); ok {
		t.Fatal("empty should fail")
	}
	if _, ok := ParseTime("not-a-date"); ok {
		t.Fatal("garbage should fail")
	}
}
