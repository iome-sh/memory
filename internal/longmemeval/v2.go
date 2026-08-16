package longmemeval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// V2Question is one official LongMemEval-V2 questions.jsonl row.
// Gold answer is optional (some public rows omit it); overlap bench uses it when present.
type V2Question struct {
	ID       string `json:"id"`
	Domain   string `json:"domain"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Image    string `json:"image,omitempty"`
}

// V2Trajectory is one official trajectories.jsonl row. Extra fields are ignored.
type V2Trajectory struct {
	ID     string           `json:"id"`
	Domain string           `json:"domain"`
	States []map[string]any `json:"states"`
	Steps  []map[string]any `json:"steps"`
}

// V2Dataset is a loaded official-shaped V2 tree (questions + trajectories + haystack).
// Do not vendor the 7GB snapshot; tests use testdata/longmemeval_v2_subset.
type V2Dataset struct {
	Questions    []V2Question
	Trajectories map[string]V2Trajectory
	Haystack     map[string][]string
}

// LoadV2 reads questions.jsonl, trajectories.jsonl, and haystacks/lme_v2_<tier>.json.
func LoadV2(dataRoot, tier string) (V2Dataset, error) {
	root := strings.TrimSpace(dataRoot)
	if root == "" {
		return V2Dataset{}, fmt.Errorf("longmemeval-v2: data root required")
	}
	tier = strings.TrimSpace(tier)
	if tier == "" {
		tier = "small"
	}
	questions, err := loadV2Questions(filepath.Join(root, "questions.jsonl"))
	if err != nil {
		return V2Dataset{}, err
	}
	trajs, err := loadV2Trajectories(filepath.Join(root, "trajectories.jsonl"))
	if err != nil {
		return V2Dataset{}, err
	}
	haystack, err := loadV2Haystack(filepath.Join(root, "haystacks", "lme_v2_"+tier+".json"))
	if err != nil {
		return V2Dataset{}, err
	}
	return V2Dataset{Questions: questions, Trajectories: trajs, Haystack: haystack}, nil
}

func loadV2Questions(path string) ([]V2Question, error) {
	rows, err := readJSONL(path)
	if err != nil {
		return nil, err
	}
	out := make([]V2Question, 0, len(rows))
	for i, raw := range rows {
		var q V2Question
		if err := json.Unmarshal(raw, &q); err != nil {
			return nil, fmt.Errorf("longmemeval-v2: question %d: %w", i+1, err)
		}
		if strings.TrimSpace(q.ID) == "" || strings.TrimSpace(q.Question) == "" {
			return nil, fmt.Errorf("longmemeval-v2: question %d missing id/question", i+1)
		}
		if q.Answer == "" {
			q.Answer = stringField(raw, "gold_answer", "gold", "reference")
		}
		out = append(out, q)
	}
	return out, nil
}

func loadV2Trajectories(path string) (map[string]V2Trajectory, error) {
	rows, err := readJSONL(path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]V2Trajectory, len(rows))
	for i, raw := range rows {
		var tr V2Trajectory
		if err := json.Unmarshal(raw, &tr); err != nil {
			return nil, fmt.Errorf("longmemeval-v2: trajectory %d: %w", i+1, err)
		}
		if strings.TrimSpace(tr.ID) == "" {
			return nil, fmt.Errorf("longmemeval-v2: trajectory %d missing id", i+1)
		}
		if _, dup := out[tr.ID]; dup {
			return nil, fmt.Errorf("longmemeval-v2: duplicate trajectory id %q", tr.ID)
		}
		out[tr.ID] = tr
	}
	return out, nil
}

func loadV2Haystack(path string) (map[string][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("longmemeval-v2: read haystack: %w", err)
	}
	var out map[string][]string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("longmemeval-v2: parse haystack: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("longmemeval-v2: haystack is empty")
	}
	return out, nil
}

func readJSONL(path string) ([]json.RawMessage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("longmemeval-v2: open %s: %w", path, err)
	}
	defer f.Close()
	var out []json.RawMessage
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if !json.Valid([]byte(line)) {
			return nil, fmt.Errorf("longmemeval-v2: invalid JSON at %s:%d", path, lineNo)
		}
		out = append(out, json.RawMessage(line))
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("longmemeval-v2: scan %s: %w", path, err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("longmemeval-v2: %s is empty", path)
	}
	return out, nil
}

func stringField(raw json.RawMessage, keys ...string) string {
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return ""
	}
	for _, k := range keys {
		if v, ok := obj[k].(string); ok && strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// TextSteps extracts ingestible text from a V2 trajectory (text first; images skipped).
func TextSteps(tr V2Trajectory) []string {
	var out []string
	for _, state := range tr.States {
		if joined := strings.Join(mapTextChunks(state), "\n"); strings.TrimSpace(joined) != "" {
			out = append(out, joined)
		}
	}
	for _, step := range tr.Steps {
		if joined := strings.Join(mapTextChunks(step), "\n"); strings.TrimSpace(joined) != "" {
			out = append(out, joined)
		}
	}
	return out
}

func mapTextChunks(m map[string]any) []string {
	if m == nil {
		return nil
	}
	keys := []string{
		"action", "thought", "goal", "utterance", "text", "observation",
		"instruction", "note", "content", "axtree", "slice_axtree_text",
		"slice_action_sequence",
	}
	var out []string
	for _, k := range keys {
		s := strings.TrimSpace(asString(m[k]))
		if s == "" {
			continue
		}
		if len(s) > 4000 {
			s = s[:4000]
		}
		out = append(out, s)
	}
	return out
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return ""
	}
}
