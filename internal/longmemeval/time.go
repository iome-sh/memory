package longmemeval

import (
	"strings"
	"time"
)

// OfficialDateLayout is the cleaned LongMemEval V1 haystack_dates / question_date
// layout used by xiaowu0162/longmemeval-cleaned (e.g. "2023/04/10 (Mon) 17:50").
const OfficialDateLayout = "2006/01/02 (Mon) 15:04"

var parseLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	OfficialDateLayout,
	"2006/01/02 (Mon) 15:04:05",
	"2006/01/02 15:04",
	"2006/01/02",
	"2006-01-02",
}

// ParseTime accepts RFC3339 and official LongMemEval cleaned dates.
// Empty or unknown strings return (zero, false) so callers keep their fallback.
func ParseTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range parseLayouts {
		if ts, err := time.Parse(layout, s); err == nil {
			return ts, true
		}
	}
	return time.Time{}, false
}
