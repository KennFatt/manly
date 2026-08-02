package renderer

import (
	"fmt"
	"io"
	"sort"
	"time"
)

func renderCompactAnalytics(w io.Writer, view AnalyticsView) error {
	if !view.Enabled {
		_, err := fmt.Fprintln(w, "Analytics is disabled.")
		return err
	}
	fmt.Fprintln(w, "Analytics")
	fmt.Fprintf(w, "Provider: %s\n", view.Provider)
	fmt.Fprintf(w, "Period: %s\n\n", analyticsPeriod(view))
	if view.ConceptLoads == 0 {
		_, err := fmt.Fprintln(w, "No concept loads recorded.")
		return err
	}

	fmt.Fprintf(w, "Concept loads: %d\n", view.ConceptLoads)
	fmt.Fprintf(w, "Retrieval batches: %d\n", view.RetrievalBatches)
	fmt.Fprintf(w, "Average concepts per batch: %.1f\n\n", view.AverageConceptsPerBatch)
	renderCompactEntryPoints(w, view.EntryPoints)
	renderCompactTopConcepts(w, view.TopConcepts)
	renderCompactRecentBatches(w, view.RecentBatches)
	return nil
}

func renderCompactEntryPoints(w io.Writer, entryPoints map[string]int) {
	if len(entryPoints) == 0 {
		return
	}
	fmt.Fprintln(w, "Entry points")
	keys := sortedEntryPointKeys(entryPoints)
	for _, key := range keys {
		fmt.Fprintf(w, "%s  %d\n", key, entryPoints[key])
	}
	fmt.Fprintln(w)
}

func renderCompactTopConcepts(w io.Writer, concepts []AnalyticsConcept) {
	if len(concepts) == 0 {
		return
	}
	fmt.Fprintln(w, "Top concepts")
	for _, concept := range concepts {
		fmt.Fprintf(w, "%d  %s\n", concept.LoadCount, concept.ConceptID)
	}
	fmt.Fprintln(w)
}

func renderCompactRecentBatches(w io.Writer, batches []AnalyticsBatch) {
	if len(batches) == 0 {
		return
	}
	fmt.Fprintln(w, "Recent batches")
	for index, batch := range batches {
		fmt.Fprintf(w, "%s  %s  %d concepts\n", formatAnalyticsTime(batch.OccurredAt), batch.EntryPoint, batch.ConceptCount)
		for _, conceptID := range batch.ConceptIDs {
			fmt.Fprintf(w, "  %s\n", conceptID)
		}
		if index < len(batches)-1 {
			fmt.Fprintln(w)
		}
	}
}

func renderFancyAnalytics(w io.Writer, view AnalyticsView) error {
	if !view.Enabled {
		_, err := fmt.Fprintln(w, "Analytics is disabled.")
		return err
	}
	fmt.Fprintln(w, "Analytics")
	fmt.Fprintf(w, "\nProvider: %s\n", view.Provider)
	fmt.Fprintf(w, "Period: %s\n", analyticsPeriod(view))
	if view.ConceptLoads == 0 {
		_, err := fmt.Fprintln(w, "\nNo concept loads recorded.")
		return err
	}
	fmt.Fprintf(w, "\nConcept loads: %d\n", view.ConceptLoads)
	fmt.Fprintf(w, "Retrieval batches: %d\n", view.RetrievalBatches)
	fmt.Fprintf(w, "Average concepts per batch: %.1f\n", view.AverageConceptsPerBatch)
	if len(view.EntryPoints) > 0 {
		fmt.Fprintln(w, "\nEntry points")
		for _, key := range sortedEntryPointKeys(view.EntryPoints) {
			fmt.Fprintf(w, "  %-8s %d\n", key, view.EntryPoints[key])
		}
	}
	if len(view.TopConcepts) > 0 {
		fmt.Fprintln(w, "\nTop concepts")
		for _, concept := range view.TopConcepts {
			fmt.Fprintf(w, "  %d  %s\n", concept.LoadCount, concept.ConceptID)
		}
	}
	if len(view.RecentBatches) > 0 {
		fmt.Fprintln(w, "\nRecent batches")
		for _, batch := range view.RecentBatches {
			fmt.Fprintf(w, "  %s  %s  %d concepts\n", formatAnalyticsTime(batch.OccurredAt), batch.EntryPoint, batch.ConceptCount)
			for _, conceptID := range batch.ConceptIDs {
				fmt.Fprintf(w, "    %s\n", conceptID)
			}
		}
	}
	return nil
}

func renderMarkdownAnalytics(w io.Writer, view AnalyticsView) error {
	if !view.Enabled {
		_, err := fmt.Fprintln(w, "Analytics is disabled.")
		return err
	}
	fmt.Fprint(w, "# Analytics\n\n")
	fmt.Fprintf(w, "- Provider: `%s`\n", view.Provider)
	fmt.Fprintf(w, "- Period: %s\n", analyticsPeriod(view))
	if view.ConceptLoads == 0 {
		_, err := fmt.Fprintln(w, "\nNo concept loads recorded.")
		return err
	}
	fmt.Fprintf(w, "- Concept loads: %d\n", view.ConceptLoads)
	fmt.Fprintf(w, "- Retrieval batches: %d\n", view.RetrievalBatches)
	fmt.Fprintf(w, "- Average concepts per batch: %.1f\n", view.AverageConceptsPerBatch)
	if len(view.EntryPoints) > 0 {
		fmt.Fprintln(w, "\n## Entry points\n\n| Entry point | Loads |\n|---|---:|")
		for _, key := range sortedEntryPointKeys(view.EntryPoints) {
			fmt.Fprintf(w, "| %s | %d |\n", key, view.EntryPoints[key])
		}
	}
	if len(view.TopConcepts) > 0 {
		fmt.Fprintln(w, "\n## Top concepts\n\n| Concept | Loads |\n|---|---:|")
		for _, concept := range view.TopConcepts {
			fmt.Fprintf(w, "| `%s` | %d |\n", concept.ConceptID, concept.LoadCount)
		}
	}
	if len(view.RecentBatches) > 0 {
		fmt.Fprintln(w, "\n## Recent batches")
		for _, batch := range view.RecentBatches {
			fmt.Fprintf(w, "\n### `%s` · %s · %d concepts\n", formatAnalyticsTime(batch.OccurredAt), batch.EntryPoint, batch.ConceptCount)
			for _, conceptID := range batch.ConceptIDs {
				fmt.Fprintf(w, "- `%s`\n", conceptID)
			}
		}
	}
	return nil
}

func analyticsPeriod(view AnalyticsView) string {
	if view.Period.Since == nil {
		return "all time"
	}
	return "since " + formatAnalyticsTime(*view.Period.Since)
}

func formatAnalyticsTime(value time.Time) string {
	return value.UTC().Format("2006-01-02 15:04:05Z07:00")
}

func sortedEntryPointKeys(entryPoints map[string]int) []string {
	keys := make([]string, 0, len(entryPoints))
	for key := range entryPoints {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
