package renderer

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestJSONSearchViewIncludesMatchEvidence(t *testing.T) {
	view := SearchView{
		Query: "ask",
		Results: []SearchResult{
			{
				Concept:       Concept{ID: "/preferences/ask-dont-assume", Title: "Ask, Don't Assume"},
				Score:         32,
				MatchedFields: []string{"title", "body"},
				MatchedTerms:  []string{"ask", "dont", "assume"},
				MatchedRank:   "title_phrase",
				Actions:       []Action{{Name: "show", Command: "manly show /preferences/ask-dont-assume"}},
			},
		},
	}
	var buf bytes.Buffer
	renderer, err := New(FormatJSON)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := renderer.Render(&buf, view); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	results, ok := decoded["results"].([]any)
	if !ok || len(results) != 1 {
		t.Fatalf("results = %v, want one result", decoded["results"])
	}
	result, ok := results[0].(map[string]any)
	if !ok {
		t.Fatalf("result is not an object: %v", results[0])
	}
	if score := result["score"]; score != float64(32) {
		t.Fatalf("score = %v, want 32", score)
	}
	if fields := result["matched_fields"]; !reflect.DeepEqual(fields, []any{"title", "body"}) {
		t.Fatalf("matched_fields = %v, want [title body]", fields)
	}
	if terms := result["matched_terms"]; !reflect.DeepEqual(terms, []any{"ask", "dont", "assume"}) {
		t.Fatalf("matched_terms = %v, want [ask dont assume]", terms)
	}
	if rank := result["matched_rank"]; rank != "title_phrase" {
		t.Fatalf("matched_rank = %v, want title_phrase", rank)
	}
}

func TestJSONSearchViewOmitsEmptyMatchEvidence(t *testing.T) {
	view := SearchView{
		Query: "x",
		Results: []SearchResult{
			{Concept: Concept{ID: "/legacy", Title: "Legacy"}, Score: 1},
		},
	}
	var buf bytes.Buffer
	renderer, err := New(FormatJSON)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := renderer.Render(&buf, view); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	result := decoded["results"].([]any)[0].(map[string]any)
	for _, key := range []string{"matched_fields", "matched_terms", "matched_rank"} {
		if _, present := result[key]; present {
			t.Fatalf("JSON output should omit %q: %v", key, result)
		}
	}
}
