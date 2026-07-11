package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Issue struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type ValidationReport struct {
	Errors   []Issue
	Warnings []Issue
}

func (r *ValidationReport) AddError(path, message string) {
	r.Errors = append(r.Errors, Issue{Path: path, Message: message})
}

func (r *ValidationReport) AddWarning(path, message string) {
	r.Warnings = append(r.Warnings, Issue{Path: path, Message: message})
}

func (r ValidationReport) Sort() {
	sort.Slice(r.Errors, func(i, j int) bool { return issueLess(r.Errors[i], r.Errors[j]) })
	sort.Slice(r.Warnings, func(i, j int) bool { return issueLess(r.Warnings[i], r.Warnings[j]) })
}

func (r ValidationReport) Valid() bool {
	return len(r.Errors) == 0
}

func Validate(root string, strict bool) (ValidationReport, error) {
	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return ValidationReport{}, fmt.Errorf("resolve root: %w", err)
	}
	info, err := os.Stat(resolvedRoot)
	if err != nil {
		return ValidationReport{}, fmt.Errorf("access root %q: %w", root, err)
	}
	if !info.IsDir() {
		return ValidationReport{}, fmt.Errorf("root is not a directory: %s", resolvedRoot)
	}

	report := ValidationReport{}
	markdown := make(map[string]string)
	concepts := make(map[string]*Concept)
	var paths []string
	err = filepath.WalkDir(resolvedRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			report.AddError(relativePath(resolvedRoot, path), walkErr.Error())
			return nil
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		rel := relativePath(resolvedRoot, path)
		paths = append(paths, rel)
		markdown[rel] = path
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			report.AddError(rel, fmt.Sprintf("cannot read file: %v", readErr))
			return nil
		}
		switch filepath.Base(rel) {
		case "index.md":
			validateIndex(&report, rel, string(data), markdown)
		case "log.md":
			validateLog(&report, rel, string(data))
		default:
			metadata, body, parseErr := parseFrontmatter(string(data))
			if parseErr != nil {
				report.AddError(rel, parseErr.Error())
				return nil
			}
			typeValue, typeIsString := metadata["type"].(string)
			if !typeIsString || strings.TrimSpace(typeValue) == "" {
				report.AddError(rel, "frontmatter requires a non-empty string type")
			}
			validateTimestamp(&report, rel, metadata)
			if strings.TrimSpace(body) == "" {
				report.AddWarning(rel, "concept body is empty")
			}
			concept, conceptErr := parseConcept(path, resolvedRoot, rel)
			if conceptErr == nil {
				concepts[concept.ID] = concept
			}
		}
		return nil
	})
	if err != nil {
		return report, fmt.Errorf("scan bundle: %w", err)
	}

	for _, rel := range paths {
		if filepath.Base(rel) == "log.md" {
			continue
		}
		data, readErr := os.ReadFile(markdown[rel])
		if readErr != nil {
			continue
		}
		for _, link := range resolveDocumentLinksForValidation(resolvedRoot, rel, string(data), markdown, concepts) {
			if link.Broken && !link.External {
				report.AddWarning(rel, fmt.Sprintf("link target not found: %s", link.RawTarget))
			}
		}
	}

	if strict {
		checkIndexCoverage(&report, resolvedRoot, markdown, concepts)
	}
	report.Sort()
	return report, nil
}

func validateIndex(report *ValidationReport, rel, content string, markdown map[string]string) {
	if rel == "index.md" && strings.HasPrefix(strings.ReplaceAll(content, "\r\n", "\n"), "---\n") {
		metadata, _, err := parseFrontmatter(content)
		if err != nil {
			report.AddError(rel, err.Error())
		} else if value, exists := metadata["okf_version"]; exists && strings.TrimSpace(fmt.Sprint(value)) == "" {
			report.AddError(rel, "okf_version must not be empty when present")
		}
	} else if rel != "index.md" && strings.HasPrefix(strings.ReplaceAll(content, "\r\n", "\n"), "---\n") {
		report.AddError(rel, "non-root index.md must not contain frontmatter")
	}
	if !strings.Contains(content, "](") {
		report.AddWarning(rel, "index has no directory entries")
	}
	_ = markdown
}

var logDatePattern = regexp.MustCompile(`^##\s+(.+?)\s*$`)

func validateLog(report *ValidationReport, rel, content string) {
	if strings.HasPrefix(strings.ReplaceAll(content, "\r\n", "\n"), "---\n") {
		report.AddError(rel, "log.md must not contain frontmatter")
	}
	found := false
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		match := logDatePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		found = true
		if _, err := time.Parse("2006-01-02", match[1]); err != nil {
			report.AddError(rel, fmt.Sprintf("invalid log date heading %q; expected YYYY-MM-DD", match[1]))
		}
	}
	if !found {
		report.AddError(rel, "log.md requires at least one ## YYYY-MM-DD heading")
	}
}

func validateTimestamp(report *ValidationReport, rel string, metadata map[string]any) {
	value, exists := metadata["timestamp"]
	if !exists {
		return
	}
	switch timestamp := value.(type) {
	case time.Time:
		return
	case string:
		if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
			report.AddError(rel, fmt.Sprintf("timestamp is not ISO 8601: %q", timestamp))
		}
	default:
		report.AddError(rel, "timestamp must be an ISO 8601 string")
	}
}

func resolveDocumentLinksForValidation(root, sourceRel, content string, markdown map[string]string, concepts map[string]*Concept) []Link {
	bundle := &Bundle{Root: root, Markdown: markdown, ByID: concepts}
	return resolveDocumentLinks(bundle, sourceRel, content)
}

func checkIndexCoverage(report *ValidationReport, root string, markdown map[string]string, concepts map[string]*Concept) {
	conceptList := make([]*Concept, 0, len(concepts))
	for _, concept := range concepts {
		conceptList = append(conceptList, concept)
	}
	bundle := &Bundle{Root: root, Concepts: conceptList, Markdown: markdown, ByID: concepts}
	for rel, path := range markdown {
		if filepath.Base(rel) != "index.md" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)
		start := strings.Index(content, "<!-- manly:generated:start -->")
		end := strings.Index(content, "<!-- manly:generated:end -->")
		if start < 0 || end < start {
			continue
		}
		directory := filepath.ToSlash(filepath.Dir(rel))
		actual := content[start : end+len("<!-- manly:generated:end -->")]
		expected := "<!-- manly:generated:start -->\n" + generatedIndexEntries(bundle, directory) + "<!-- manly:generated:end -->"
		if actual != expected {
			report.AddWarning(rel, "generated index section is stale")
		}
	}
}

func issueLess(left, right Issue) bool {
	if left.Path != right.Path {
		return left.Path < right.Path
	}
	return left.Message < right.Message
}

func relativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
