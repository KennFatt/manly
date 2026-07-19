package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
)

// IsSingleBundleRoot reports whether root uses the single-bundle mode.
func IsSingleBundleRoot(root string) (bool, error) {
	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return false, fmt.Errorf("resolve root: %w", err)
	}
	info, err := os.Stat(resolvedRoot)
	if err != nil {
		return false, fmt.Errorf("access root %q: %w", root, err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("root is not a directory: %s", resolvedRoot)
	}
	return rootLooksLikeBundle(resolvedRoot), nil
}

// ValidateWorkspaceRoot validates a single bundle or all direct child bundles
// without loading concepts into the in-memory Bundle model first.
func ValidateWorkspaceRoot(root string, strict bool) (ValidationReport, error) {
	single, err := IsSingleBundleRoot(root)
	if err != nil {
		return ValidationReport{}, err
	}
	if single {
		return Validate(root, strict)
	}
	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return ValidationReport{}, fmt.Errorf("resolve root: %w", err)
	}
	entries, err := os.ReadDir(resolvedRoot)
	if err != nil {
		return ValidationReport{}, fmt.Errorf("scan workspace: %w", err)
	}
	report := ValidationReport{}
	bundles := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		bundleRoot := filepath.Join(resolvedRoot, entry.Name())
		indexPath := filepath.Join(bundleRoot, "index.md")
		if _, statErr := os.Stat(indexPath); os.IsNotExist(statErr) {
			continue
		} else if statErr != nil {
			return report, fmt.Errorf("inspect bundle candidate %q: %w", entry.Name(), statErr)
		}
		if err := validateBundleMetadata(indexPath); err != nil {
			return report, fmt.Errorf("bundle candidate %q: %w", entry.Name(), err)
		}
		bundleReport, err := Validate(bundleRoot, strict)
		if err != nil {
			return report, err
		}
		bundles++
		mergeValidationReport(&report, bundleReport, entry.Name())
	}
	if bundles == 0 {
		return report, fmt.Errorf("workspace contains no bundles")
	}
	finalizeWorkspaceReport(&report, resolvedRoot)
	return report, nil
}
