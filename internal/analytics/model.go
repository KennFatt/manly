// Package analytics records and reports local concept retrieval events.
package analytics

import (
	"context"
	"time"
)

// Provider identifies a local analytics storage format.
type Provider string

const (
	ProviderSQLite Provider = "sqlite"
	ProviderCSV    Provider = "csv"
)

// EntryPoint identifies the command that returned a concept body.
type EntryPoint string

const (
	EntryPointShow    EntryPoint = "show"
	EntryPointContext EntryPoint = "context"
)

// ConceptLoad is one successfully rendered concept retrieval event.
type ConceptLoad struct {
	OccurredAt time.Time
	BatchID    string
	ConceptID  string
	EntryPoint EntryPoint
}

// Recorder persists concept retrieval events.
type Recorder interface {
	RecordConceptLoads(context.Context, []ConceptLoad) error
}

// Reader produces aggregate analytics reports.
type Reader interface {
	Report(context.Context, ReportOptions) (Report, error)
}

// ReportOptions controls analytics report filtering and result limits.
type ReportOptions struct {
	Since *time.Time
	Limit int
}

// Report contains aggregate concept usage and retrieval batches.
type Report struct {
	Enabled                 bool
	Provider                Provider
	Since                   *time.Time
	ConceptLoads            int
	RetrievalBatches        int
	AverageConceptsPerBatch float64
	EntryPoints             map[EntryPoint]int
	TopConcepts             []ConceptUsage
	RecentBatches           []RetrievalBatch
}

// ConceptUsage contains one concept's retrieval count.
type ConceptUsage struct {
	ConceptID string
	LoadCount int
}

// RetrievalBatch groups concepts returned by one retrieval invocation.
type RetrievalBatch struct {
	BatchID    string
	OccurredAt time.Time
	EntryPoint EntryPoint
	ConceptIDs []string
}

func emptyReport(options ReportOptions) Report {
	return Report{
		Since:         normalizedSince(options.Since),
		EntryPoints:   make(map[EntryPoint]int),
		TopConcepts:   make([]ConceptUsage, 0),
		RecentBatches: make([]RetrievalBatch, 0),
	}
}

func normalizedSince(since *time.Time) *time.Time {
	if since == nil {
		return nil
	}
	value := since.UTC()
	return &value
}
