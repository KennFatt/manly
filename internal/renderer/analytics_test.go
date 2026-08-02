package renderer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCompactAnalyticsReport(t *testing.T) {
	view := AnalyticsView{
		Enabled:                 true,
		Provider:                "sqlite",
		ConceptLoads:            3,
		RetrievalBatches:        2,
		AverageConceptsPerBatch: 1.5,
		EntryPoints:             map[string]int{"show": 2, "context": 1},
		TopConcepts:             []AnalyticsConcept{{ConceptID: "/a", LoadCount: 2}},
		RecentBatches: []AnalyticsBatch{{
			BatchID:      "batch",
			OccurredAt:   time.Date(2026, time.August, 2, 9, 0, 0, 0, time.UTC),
			EntryPoint:   "show",
			ConceptCount: 1,
			ConceptIDs:   []string{"/a"},
		}},
	}
	var output strings.Builder
	if err := (compactRenderer{}).Render(&output, view); err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		"Analytics\n",
		"Provider: sqlite",
		"Concept loads: 3",
		"Average concepts per batch: 1.5",
		"Top concepts",
		"Recent batches",
		"2026-08-02 09:00:00Z  show  1 concepts",
	} {
		if !strings.Contains(output.String(), fragment) {
			t.Fatalf("output %q does not contain %q", output.String(), fragment)
		}
	}
}

func TestDisabledAnalyticsReport(t *testing.T) {
	var output strings.Builder
	if err := (compactRenderer{}).Render(&output, AnalyticsView{}); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "Analytics is disabled.\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestJSONAnalyticsReport(t *testing.T) {
	view := AnalyticsView{
		Enabled:       true,
		Provider:      "csv",
		EntryPoints:   map[string]int{},
		TopConcepts:   []AnalyticsConcept{},
		RecentBatches: []AnalyticsBatch{},
	}
	var output bytes.Buffer
	if err := (jsonRenderer{}).Render(&output, view); err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["provider"] != "csv" || decoded["concept_loads"] != float64(0) {
		t.Fatalf("decoded = %#v", decoded)
	}
	if _, ok := decoded["recent_batches"].([]any); !ok {
		t.Fatalf("recent_batches = %#v", decoded["recent_batches"])
	}
}
