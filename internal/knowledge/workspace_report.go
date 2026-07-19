package knowledge

import "path/filepath"

func mergeValidationReport(destination *ValidationReport, source ValidationReport, bundleName string) {
	for _, issue := range source.Errors {
		issue.Path = filepath.ToSlash(filepath.Join(bundleName, issue.Path))
		destination.Errors = append(destination.Errors, issue)
	}
	for _, issue := range source.Warnings {
		issue.Path = filepath.ToSlash(filepath.Join(bundleName, issue.Path))
		destination.Warnings = append(destination.Warnings, issue)
	}
	destination.Stats.MarkdownFiles += source.Stats.MarkdownFiles
	destination.Stats.ReservedFiles += source.Stats.ReservedFiles
	destination.Stats.ConceptFiles += source.Stats.ConceptFiles
	destination.Stats.LoadedConcepts += source.Stats.LoadedConcepts
	destination.Stats.InvalidConceptFiles += source.Stats.InvalidConceptFiles
	destination.Stats.LinksChecked += source.Stats.LinksChecked
	destination.Stats.BrokenLinks += source.Stats.BrokenLinks
	destination.Stats.StaleGeneratedIndexes += source.Stats.StaleGeneratedIndexes
	if len(source.Bundles) > 0 {
		bundle := source.Bundles[0]
		bundle.Name = bundleName
		destination.Bundles = append(destination.Bundles, bundle)
	}
}

func finalizeWorkspaceReport(report *ValidationReport, root string) {
	report.Stats.Root = root
	report.Stats.Mode = "workspace"
	report.Stats.Bundles = len(report.Bundles)
	report.Sort()
}
