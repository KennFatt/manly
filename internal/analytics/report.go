package analytics

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

func aggregateEvents(events []ConceptLoad, options ReportOptions) (Report, error) {
	report := emptyReport(options)
	conceptCounts := make(map[string]int)
	batches := make(map[string]*batchReport)
	for _, load := range events {
		if err := validateLoad(load); err != nil {
			return Report{}, err
		}
		occurredAt := load.OccurredAt.UTC()
		if report.Since != nil && occurredAt.Before(*report.Since) {
			continue
		}

		report.ConceptLoads++
		report.EntryPoints[load.EntryPoint]++
		conceptCounts[load.ConceptID]++

		batch, exists := batches[load.BatchID]
		if !exists {
			batch = &batchReport{
				batchID:    load.BatchID,
				occurredAt: occurredAt,
				entryPoint: load.EntryPoint,
				conceptIDs: make(map[string]struct{}),
			}
			batches[load.BatchID] = batch
		}
		if batch.entryPoint != load.EntryPoint {
			return Report{}, fmt.Errorf("analytics: batch %q contains multiple entry points", load.BatchID)
		}
		if occurredAt.After(batch.occurredAt) {
			batch.occurredAt = occurredAt
		}
		batch.conceptIDs[load.ConceptID] = struct{}{}
	}

	report.RetrievalBatches = len(batches)
	if report.RetrievalBatches > 0 {
		report.AverageConceptsPerBatch = float64(report.ConceptLoads) / float64(report.RetrievalBatches)
	}

	report.TopConcepts = make([]ConceptUsage, 0, len(conceptCounts))
	for conceptID, loadCount := range conceptCounts {
		report.TopConcepts = append(report.TopConcepts, ConceptUsage{ConceptID: conceptID, LoadCount: loadCount})
	}
	sort.Slice(report.TopConcepts, func(i, j int) bool {
		if report.TopConcepts[i].LoadCount != report.TopConcepts[j].LoadCount {
			return report.TopConcepts[i].LoadCount > report.TopConcepts[j].LoadCount
		}
		return report.TopConcepts[i].ConceptID < report.TopConcepts[j].ConceptID
	})
	if options.Limit > 0 && len(report.TopConcepts) > options.Limit {
		report.TopConcepts = report.TopConcepts[:options.Limit]
	}

	report.RecentBatches = make([]RetrievalBatch, 0, len(batches))
	for _, batch := range batches {
		conceptIDs := make([]string, 0, len(batch.conceptIDs))
		for conceptID := range batch.conceptIDs {
			conceptIDs = append(conceptIDs, conceptID)
		}
		sort.Strings(conceptIDs)
		report.RecentBatches = append(report.RecentBatches, RetrievalBatch{
			BatchID:    batch.batchID,
			OccurredAt: batch.occurredAt,
			EntryPoint: batch.entryPoint,
			ConceptIDs: conceptIDs,
		})
	}
	sort.Slice(report.RecentBatches, func(i, j int) bool {
		if !report.RecentBatches[i].OccurredAt.Equal(report.RecentBatches[j].OccurredAt) {
			return report.RecentBatches[i].OccurredAt.After(report.RecentBatches[j].OccurredAt)
		}
		return report.RecentBatches[i].BatchID < report.RecentBatches[j].BatchID
	})
	if options.Limit > 0 && len(report.RecentBatches) > options.Limit {
		report.RecentBatches = report.RecentBatches[:options.Limit]
	}
	return report, nil
}

type batchReport struct {
	batchID    string
	occurredAt time.Time
	entryPoint EntryPoint
	conceptIDs map[string]struct{}
}

func validateLoad(load ConceptLoad) error {
	if load.OccurredAt.IsZero() {
		return fmt.Errorf("analytics: event %q has no timestamp", load.BatchID)
	}
	if load.BatchID == "" {
		return errors.New("analytics: event batch ID must not be empty")
	}
	if load.ConceptID == "" {
		return errors.New("analytics: event concept ID must not be empty")
	}
	if !validEntryPoint(load.EntryPoint) {
		return fmt.Errorf("analytics: unsupported entry point %q", load.EntryPoint)
	}
	return nil
}
