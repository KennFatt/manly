package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Workspace coordinates one bundle or a directory containing direct child bundles.
type Workspace struct {
	Root       string
	SingleRoot bool
	Bundles    []*Bundle
	ByName     map[string]*Bundle
}

// WorkspaceMode describes how a workspace root is interpreted: as one bundle
// or as a directory of child bundles. The enum value is the mode; String
// returns its stable wire name for rendered JSON.
type WorkspaceMode int

const (
	// ModeWorkspace treats the root as a directory of child bundles.
	ModeWorkspace WorkspaceMode = iota
	// ModeSingle treats the root as one bundle.
	ModeSingle
)

// String returns the stable machine-readable wire name of the mode.
func (m WorkspaceMode) String() string {
	switch m {
	case ModeWorkspace:
		return "workspace"
	case ModeSingle:
		return "single"
	}
	return "unknown"
}

// Mode reports how this workspace interprets its root.
func (w *Workspace) Mode() WorkspaceMode {
	if w.SingleRoot {
		return ModeSingle
	}
	return ModeWorkspace
}

// ConceptRef identifies a concept together with the bundle that owns it.
type ConceptRef struct {
	BundleName string
	Bundle     *Bundle
	Concept    *Concept
}

// SearchResultRef is a workspace-qualified search result.
type SearchResultRef struct {
	BundleName string
	Concept    *Concept
	Score      float64
	Match      Match
}

// LoadWorkspace loads root as one bundle or discovers its direct child bundles.
func LoadWorkspace(root string) (*Workspace, error) {
	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}
	info, err := os.Stat(resolvedRoot)
	if err != nil {
		return nil, fmt.Errorf("access root %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root is not a directory: %s", resolvedRoot)
	}

	if rootLooksLikeBundle(resolvedRoot) {
		bundle, err := Load(resolvedRoot)
		if err != nil {
			return nil, err
		}
		return &Workspace{
			Root:       resolvedRoot,
			SingleRoot: true,
			Bundles:    []*Bundle{bundle},
			ByName:     map[string]*Bundle{"": bundle},
		}, nil
	}

	entries, err := os.ReadDir(resolvedRoot)
	if err != nil {
		return nil, fmt.Errorf("scan workspace: %w", err)
	}
	workspace := &Workspace{Root: resolvedRoot, ByName: make(map[string]*Bundle)}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidateRoot := filepath.Join(resolvedRoot, entry.Name())
		indexPath := filepath.Join(candidateRoot, "index.md")
		if _, statErr := os.Stat(indexPath); os.IsNotExist(statErr) {
			continue
		} else if statErr != nil {
			return nil, fmt.Errorf("inspect bundle candidate %q: %w", entry.Name(), statErr)
		}
		if err := validateBundleMetadata(indexPath); err != nil {
			return nil, fmt.Errorf("bundle candidate %q: %w", entry.Name(), err)
		}
		bundle, err := Load(candidateRoot)
		if err != nil {
			return nil, fmt.Errorf("load bundle %q: %w", entry.Name(), err)
		}
		workspace.Bundles = append(workspace.Bundles, bundle)
		workspace.ByName[entry.Name()] = bundle
	}
	if len(workspace.Bundles) == 0 {
		return nil, fmt.Errorf("workspace contains no bundles")
	}
	sort.Slice(workspace.Bundles, func(i, j int) bool {
		return workspace.bundleName(workspace.Bundles[i]) < workspace.bundleName(workspace.Bundles[j])
	})
	return workspace, nil
}

func rootLooksLikeBundle(root string) bool {
	indexPath := filepath.Join(root, "index.md")
	data, err := os.ReadFile(indexPath)
	if err == nil {
		metadata, _, parseErr := parseFrontmatter(string(data))
		if parseErr == nil && strings.TrimSpace(metadataString(metadata, "okf_version")) != "" {
			// Accept old manly roots that predate type: Bundle.
			typeValue := strings.TrimSpace(metadataString(metadata, "type"))
			return typeValue == "" || typeValue == "Bundle"
		}
		// A legacy root index without bundle metadata is still a single bundle
		// unless it has direct child bundle candidates.
		if !hasDirectChildIndex(root) {
			return true
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" && !isReservedMarkdown(entry.Name()) {
			return true
		}
	}
	return false
}

func hasDirectChildIndex(root string) bool {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if _, err := os.Stat(filepath.Join(root, entry.Name(), "index.md")); err == nil {
				return true
			}
		}
	}
	return false
}

func validateBundleMetadata(indexPath string) error {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("read index.md: %w", err)
	}
	metadata, _, err := parseFrontmatter(string(data))
	if err != nil {
		return fmt.Errorf("index.md frontmatter: %w", err)
	}
	if strings.TrimSpace(metadataString(metadata, "okf_version")) == "" {
		return fmt.Errorf("index.md frontmatter requires a non-empty okf_version")
	}
	if strings.TrimSpace(metadataString(metadata, "type")) != "Bundle" {
		return fmt.Errorf("index.md frontmatter requires type: Bundle")
	}
	return nil
}

func (w *Workspace) bundleName(bundle *Bundle) string {
	if w.SingleRoot {
		return ""
	}
	for name, candidate := range w.ByName {
		if candidate == bundle {
			return name
		}
	}
	return ""
}

// QualifyID converts a bundle-local ID to the CLI ID for this workspace.
func (w *Workspace) QualifyID(bundleName, localID string) string {
	if w.SingleRoot {
		return localID
	}
	if localID == "" {
		return ""
	}
	return "/" + bundleName + "/" + strings.TrimPrefix(localID, "/")
}

// ResolveConcept resolves a local ID in a single bundle or a qualified ID in a workspace.
func (w *Workspace) ResolveConcept(value string) (*ConceptRef, error) {
	if w.SingleRoot {
		concept, err := w.Bundles[0].Get(value)
		if err != nil {
			return nil, err
		}
		return &ConceptRef{Bundle: w.Bundles[0], Concept: concept}, nil
	}
	name, local, err := w.splitQualifiedPath(value, false)
	if err != nil {
		return nil, err
	}
	bundle := w.ByName[name]
	if bundle == nil {
		return nil, fmt.Errorf("bundle not found: %s", name)
	}
	concept, err := bundle.Get(local)
	if err != nil {
		return nil, err
	}
	return &ConceptRef{BundleName: name, Bundle: bundle, Concept: concept}, nil
}

// ResolveDirectory resolves a workspace directory path to its owning bundle and local prefix.
// An empty prefix in workspace mode refers to the workspace itself and returns a nil bundle.
func (w *Workspace) ResolveDirectory(value string) (*Bundle, string, string, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.Trim(value, "/")
	if value == "" {
		if w.SingleRoot {
			return w.Bundles[0], "", "", nil
		}
		return nil, "", "", nil
	}
	clean := filepath.ToSlash(filepath.Clean(value))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return nil, "", "", fmt.Errorf("invalid directory path: %q", value)
	}
	if w.SingleRoot {
		return w.Bundles[0], clean, "", nil
	}
	name, local, err := w.splitQualifiedPath(clean, true)
	if err != nil {
		return nil, "", "", err
	}
	return w.ByName[name], local, name, nil
}

func (w *Workspace) splitQualifiedPath(value string, directory bool) (string, string, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.Trim(value, "/")
	if strings.HasSuffix(value, ".md") {
		value = strings.TrimSuffix(value, ".md")
	}
	clean := filepath.ToSlash(filepath.Clean(value))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", "", fmt.Errorf("invalid workspace path: %q", value)
	}
	parts := strings.Split(clean, "/")
	if len(parts) < 2 && !directory {
		return "", "", fmt.Errorf("workspace concept path must include a bundle: %q", value)
	}
	name := parts[0]
	if name == "" || w.ByName[name] == nil {
		return "", "", fmt.Errorf("bundle not found: %s", name)
	}
	local := strings.Join(parts[1:], "/")
	if local == "" {
		if directory {
			return name, "", nil
		}
		return "", "", fmt.Errorf("concept path must not name a bundle: %q", value)
	}
	if directory {
		return name, local, nil
	}
	return name, "/" + local, nil
}

// Search searches all bundles, applying a qualified path filter when supplied.
func (w *Workspace) Search(query string, options SearchOptions) ([]SearchResultRef, error) {
	if w.SingleRoot {
		results := Search(w.Bundles[0], query, options)
		return searchRefs("", results), nil
	}
	var results []SearchResultRef
	if strings.TrimSpace(options.Path) != "" && strings.Trim(options.Path, "/") != "" {
		bundle, local, name, err := w.ResolveDirectory(options.Path)
		if err != nil {
			return nil, err
		}
		if bundle == nil {
			return nil, fmt.Errorf("invalid workspace search path: %s", options.Path)
		}
		options.Path = local
		return searchRefs(name, Search(bundle, query, options)), nil
	}
	options.Path = ""
	for _, bundle := range w.Bundles {
		name := w.bundleName(bundle)
		results = append(results, searchRefs(name, Search(bundle, query, options))...)
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return w.QualifyID(results[i].BundleName, results[i].Concept.ID) < w.QualifyID(results[j].BundleName, results[j].Concept.ID)
	})
	limit := options.Limit
	if limit <= 0 {
		limit = 10
	}
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func searchRefs(name string, results []SearchResult) []SearchResultRef {
	refs := make([]SearchResultRef, 0, len(results))
	for _, result := range results {
		refs = append(refs, SearchResultRef{BundleName: name, Concept: result.Concept, Score: result.Score, Match: result.Match})
	}
	return refs
}

// ValidateWorkspace validates each bundle and prefixes issues with its workspace path.
func (w *Workspace) Validate(strict bool) (ValidationReport, error) {
	if w.SingleRoot {
		return Validate(w.Root, strict)
	}
	report := ValidationReport{}
	for _, bundle := range w.Bundles {
		name := w.bundleName(bundle)
		bundleReport, err := Validate(bundle.Root, strict)
		if err != nil {
			return report, err
		}
		mergeValidationReport(&report, bundleReport, name)
	}
	finalizeWorkspaceReport(&report, w.Root)
	return report, nil
}
