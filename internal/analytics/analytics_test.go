package analytics

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewBatchDeduplicatesConcepts(t *testing.T) {
	loads, err := NewBatch(EntryPointContext, []string{"/b", "/a", "/b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(loads) != 2 {
		t.Fatalf("loads = %#v, want two events", loads)
	}
	if loads[0].BatchID == "" || len(loads[0].BatchID) != batchIDSize*2 {
		t.Fatalf("batch ID = %q", loads[0].BatchID)
	}
	if loads[0].BatchID != loads[1].BatchID || !loads[0].OccurredAt.Equal(loads[1].OccurredAt) {
		t.Fatalf("loads do not share batch metadata: %#v", loads)
	}
	if loads[0].ConceptID != "/b" || loads[1].ConceptID != "/a" {
		t.Fatalf("concept order = %#v", loads)
	}
}

func TestSQLiteAndCSVReportsMatch(t *testing.T) {
	loads := testLoads()
	options := ReportOptions{}

	sqlite, err := newSQLiteProvider(filepath.Join(t.TempDir(), "sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer sqlite.Close()
	csv, err := newCSVProvider(filepath.Join(t.TempDir(), "csv"))
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlite.RecordConceptLoads(context.Background(), loads); err != nil {
		t.Fatal(err)
	}
	if err := csv.RecordConceptLoads(context.Background(), loads); err != nil {
		t.Fatal(err)
	}

	sqliteReport, err := sqlite.Report(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	csvReport, err := csv.Report(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sqliteReport, csvReport) {
		t.Fatalf("provider reports differ:\nsqlite: %#v\ncsv: %#v", sqliteReport, csvReport)
	}
	if sqliteReport.ConceptLoads != 3 || sqliteReport.RetrievalBatches != 2 {
		t.Fatalf("summary = %#v", sqliteReport)
	}
	if sqliteReport.AverageConceptsPerBatch != 1.5 {
		t.Fatalf("average = %v", sqliteReport.AverageConceptsPerBatch)
	}
	if sqliteReport.TopConcepts[0].ConceptID != "/a" || sqliteReport.TopConcepts[0].LoadCount != 2 {
		t.Fatalf("top concepts = %#v", sqliteReport.TopConcepts)
	}
	if sqliteReport.RecentBatches[0].EntryPoint != EntryPointContext || len(sqliteReport.RecentBatches[0].ConceptIDs) != 1 {
		t.Fatalf("recent batches = %#v", sqliteReport.RecentBatches)
	}

	since := testLoads()[0].OccurredAt.Add(30 * time.Minute)
	filtered, err := sqlite.Report(context.Background(), ReportOptions{Since: &since})
	if err != nil {
		t.Fatal(err)
	}
	if filtered.ConceptLoads != 1 || filtered.RetrievalBatches != 1 || filtered.TopConcepts[0].ConceptID != "/a" {
		t.Fatalf("filtered report = %#v", filtered)
	}
}

func TestSQLiteSinceUsesSubsecondPrecision(t *testing.T) {
	provider, err := newSQLiteProvider(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	base := time.Date(2026, time.August, 2, 9, 0, 0, 0, time.UTC)
	loads := []ConceptLoad{
		{OccurredAt: base.Add(time.Nanosecond), BatchID: "nanosecond", ConceptID: "/nanosecond", EntryPoint: EntryPointShow},
		{OccurredAt: base.Add(time.Millisecond), BatchID: "millisecond", ConceptID: "/millisecond", EntryPoint: EntryPointShow},
		{OccurredAt: base.Add(10 * time.Millisecond), BatchID: "centisecond", ConceptID: "/centisecond", EntryPoint: EntryPointShow},
		{OccurredAt: base.Add(100 * time.Millisecond), BatchID: "decisecond", ConceptID: "/decisecond", EntryPoint: EntryPointShow},
	}
	if err := provider.RecordConceptLoads(context.Background(), loads); err != nil {
		t.Fatal(err)
	}
	cutoff := base.Add(50 * time.Millisecond)
	report, err := provider.Report(context.Background(), ReportOptions{Since: &cutoff})
	if err != nil {
		t.Fatal(err)
	}
	if report.ConceptLoads != 1 || len(report.TopConcepts) != 1 || report.TopConcepts[0].ConceptID != "/decisecond" {
		t.Fatalf("subsecond report = %#v", report)
	}
}

func TestCSVWritesHeaderOnce(t *testing.T) {
	directory := t.TempDir()
	provider, err := newCSVProvider(directory)
	if err != nil {
		t.Fatal(err)
	}
	loads := testLoads()[:1]
	if err := provider.RecordConceptLoads(context.Background(), loads); err != nil {
		t.Fatal(err)
	}
	if err := provider.RecordConceptLoads(context.Background(), loads); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, csvFilename))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(data), "occurred_at,batch_id,concept_id,entry_point"); got != 1 {
		t.Fatalf("header count = %d, data = %q", got, data)
	}
}

func TestParseSinceSupportsDays(t *testing.T) {
	now := time.Date(2026, time.August, 2, 12, 0, 0, 0, time.FixedZone("PDT", -7*60*60))
	cutoff, err := ParseSince("7d", now)
	if err != nil {
		t.Fatal(err)
	}
	want := now.UTC().Add(-7 * 24 * time.Hour)
	if cutoff == nil || !cutoff.Equal(want) || cutoff.Location() != time.UTC {
		t.Fatalf("cutoff = %v, want %v UTC", cutoff, want)
	}
}

func TestDisabledServiceDoesNotCreateStorage(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "analytics")
	service := New(Settings{Provider: ProviderSQLite, Directory: directory})
	if err := service.RecordConceptLoads(context.Background(), testLoads()); err != nil {
		t.Fatal(err)
	}
	report, err := service.Report(context.Background(), ReportOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if report.Enabled {
		t.Fatal("disabled report was enabled")
	}
	if _, err := os.Stat(directory); !os.IsNotExist(err) {
		t.Fatalf("disabled service created storage: %v", err)
	}
}

func testLoads() []ConceptLoad {
	base := time.Date(2026, time.August, 2, 9, 0, 0, 0, time.UTC)
	return []ConceptLoad{
		{OccurredAt: base, BatchID: "batch-show", ConceptID: "/b", EntryPoint: EntryPointShow},
		{OccurredAt: base, BatchID: "batch-show", ConceptID: "/a", EntryPoint: EntryPointShow},
		{OccurredAt: base.Add(time.Hour), BatchID: "batch-context", ConceptID: "/a", EntryPoint: EntryPointContext},
	}
}
