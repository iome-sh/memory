package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type recallGateCase struct {
	name            string
	convID          string
	turns           []ingestTurn
	query           string
	expectedKeyword string
}

type ingestTurn struct {
	Role      string
	Content   string
	Timestamp time.Time
	Cycle     int
}

func TestLongMemEval_RecallGate_MultiCase(t *testing.T) {
	setupTestHarness(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", handleIngest)
	mux.HandleFunc("/retrieve", handleRetrieve)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cases := []recallGateCase{
		{
			name:   "pet-adoption",
			convID: "recall-gate-pet",
			turns: []ingestTurn{
				{
					Role:      "user",
					Content:   "I adopted a golden retriever named Max in March 2024.",
					Timestamp: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
					Cycle:     1,
				},
			},
			// Semantic: no "Max" or "retriever" in the query.
			query:           "Which furry family member did I bring home last spring?",
			expectedKeyword: "max",
		},
		{
			name:   "career",
			convID: "recall-gate-career",
			turns: []ingestTurn{
				{
					Role:      "user",
					Content:   "I started working as a marine biologist at the Monterey Bay Aquarium in June 2023.",
					Timestamp: time.Date(2023, 6, 1, 9, 0, 0, 0, time.UTC),
					Cycle:     1,
				},
			},
			query:           "Where do I work studying ocean life?",
			expectedKeyword: "monterey",
		},
		{
			name:   "allergy",
			convID: "recall-gate-allergy",
			turns: []ingestTurn{
				{
					Role:      "user",
					Content:   "I'm severely allergic to shellfish and always carry an epinephrine pen.",
					Timestamp: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
					Cycle:     1,
				},
			},
			query:           "What foods should I avoid at a seafood restaurant?",
			expectedKeyword: "shellfish",
		},
		{
			name:   "hobby",
			convID: "recall-gate-hobby",
			turns: []ingestTurn{
				{
					Role:      "user",
					Content:   "I've been learning to play the cello since my grandmother gave me hers in 2022.",
					Timestamp: time.Date(2022, 11, 5, 18, 0, 0, 0, time.UTC),
					Cycle:     1,
				},
			},
			query:           "What string instrument am I practicing?",
			expectedKeyword: "cello",
		},
	}

	const minRecall = 3
	hits := 0

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if err := postIngest(srv.URL, tc.convID, tc.turns); err != nil {
				t.Fatal(err)
			}

			out, err := postRetrieve(srv.URL, tc.query, 5)
			if err != nil {
				t.Fatal(err)
			}
			if len(out.Memories) == 0 {
				t.Fatal("expected at least one retrieved memory")
			}

			top := strings.ToLower(out.Memories[0].Full + " " + out.Memories[0].Summary)
			if strings.Contains(top, tc.expectedKeyword) {
				hits++
				return
			}
			t.Errorf("top memory should contain %q; got summary=%q full=%q",
				tc.expectedKeyword, out.Memories[0].Summary, out.Memories[0].Full)
		})
	}

	if hits < minRecall {
		t.Fatalf("recall gate: %d/%d cases passed, need at least %d", hits, len(cases), minRecall)
	}
}

func postIngest(baseURL, convID string, turns []ingestTurn) error {
	payloadTurns := make([]struct {
		Role      string `json:"role"`
		Content   string `json:"content"`
		Timestamp string `json:"timestamp"`
		Cycle     int    `json:"cycle"`
	}, len(turns))
	for i, turn := range turns {
		payloadTurns[i].Role = turn.Role
		payloadTurns[i].Content = turn.Content
		payloadTurns[i].Timestamp = turn.Timestamp.Format(time.RFC3339)
		payloadTurns[i].Cycle = turn.Cycle
	}

	body, err := json.Marshal(IngestRequest{ConvID: convID, Turns: payloadTurns})
	if err != nil {
		return err
	}
	resp, err := http.Post(baseURL+"/ingest", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errIngestStatus(resp.StatusCode)
	}
	return nil
}

func postRetrieve(baseURL, query string, limit int) (RetrieveResponse, error) {
	body, err := json.Marshal(RetrieveRequest{Query: query, Limit: limit})
	if err != nil {
		return RetrieveResponse{}, err
	}
	resp, err := http.Post(baseURL+"/retrieve", "application/json", bytes.NewReader(body))
	if err != nil {
		return RetrieveResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return RetrieveResponse{}, errRetrieveStatus(resp.StatusCode)
	}
	var out RetrieveResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return RetrieveResponse{}, err
	}
	return out, nil
}

type statusError string

func (e statusError) Error() string { return string(e) }

func errIngestStatus(code int) error {
	return statusError(fmt.Sprintf("ingest status = %d", code))
}

func errRetrieveStatus(code int) error {
	return statusError(fmt.Sprintf("retrieve status = %d", code))
}