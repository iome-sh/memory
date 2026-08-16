package longmemeval

import (
	"fmt"
	"strings"
	"time"

	"github.com/iome-sh/memory"
)

// ContextItem matches official V2 Memory.query items: type text|image, value string.
type ContextItem struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// PalaceMemory is an Insert/Query adapter over PalaceStore for the official V2 harness.
// Text steps only. Images are ignored. Hash embeddings are the no-dep default
// when the store was constructed without ONNX. dual_write stays off. Not Memory GA.
type PalaceMemory struct {
	Store *memory.PalaceStore
	TopK  int
}

// Insert indexes one V2 trajectory's text steps via IngestTurn.
func (m *PalaceMemory) Insert(trajectory V2Trajectory) error {
	if m == nil || m.Store == nil {
		return fmt.Errorf("longmemeval-v2: palace adapter missing store")
	}
	id := strings.TrimSpace(trajectory.ID)
	if id == "" {
		return fmt.Errorf("longmemeval-v2: trajectory missing id")
	}
	steps := TextSteps(trajectory)
	if len(steps) == 0 {
		return fmt.Errorf("longmemeval-v2: trajectory %s has no text steps", id)
	}
	now := time.Now().UTC()
	for i, text := range steps {
		entry := memory.MemoryEntry{
			ID:           memory.GenerateMemoryID(),
			Type:         "conversation_turn",
			Tier:         memory.TierWorking,
			Content:      memory.MemoryContent{Full: text, Summary: truncate(text, 280)},
			Cycle:        i + 1,
			CreatedAt:    now,
			UpdatedAt:    now,
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			TurnID:       memory.GenerateMemoryID(),
			SessionID:    id,
			OriginalText: text,
		}
		if err := m.Store.IngestTurn(entry); err != nil {
			return fmt.Errorf("longmemeval-v2: ingest %s step %d: %w", id, i, err)
		}
	}
	return nil
}

// Query returns official-shaped context items from SearchMemory (text only).
func (m *PalaceMemory) Query(query string, _ string) []ContextItem {
	if m == nil || m.Store == nil || strings.TrimSpace(query) == "" {
		return nil
	}
	topK := m.TopK
	if topK <= 0 {
		topK = 5
	}
	hits := m.Store.SearchMemory(query, nil, topK, nil)
	out := make([]ContextItem, 0, len(hits))
	for _, hit := range hits {
		val := strings.TrimSpace(hit.Content.Full)
		if val == "" {
			val = strings.TrimSpace(hit.Content.Summary)
		}
		if val == "" {
			continue
		}
		out = append(out, ContextItem{Type: "text", Value: val})
	}
	return out
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
