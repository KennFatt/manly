package knowledge

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

type Concept struct {
	ID          string
	RelPath     string
	AbsPath     string
	Metadata    map[string]any
	Type        string
	Title       string
	Description string
	Tags        []string
	Body        string
	Links       []Link
}

type Link struct {
	Label      string
	RawTarget  string
	TargetPath string
	TargetID   string
	External   bool
	Broken     bool
}

type Bundle struct {
	Root     string
	Title    string
	Concepts []*Concept
	ByID     map[string]*Concept
	Markdown map[string]string
}

func Load(root string) (*Bundle, error) {
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

	bundle := &Bundle{
		Root:     resolvedRoot,
		ByID:     make(map[string]*Concept),
		Markdown: make(map[string]string),
	}
	var loadErr error
	err = filepath.WalkDir(resolvedRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		relPath, err := filepath.Rel(resolvedRoot, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)
		bundle.Markdown[relPath] = path
		if relPath == "index.md" {
			if data, readErr := os.ReadFile(path); readErr == nil {
				if metadata, _, parseErr := parseFrontmatter(string(data)); parseErr == nil {
					bundle.Title = metadataString(metadata, "title")
				}
			}
		}
		if isReservedMarkdown(relPath) {
			return nil
		}

		concept, err := parseConcept(path, resolvedRoot, relPath)
		if err != nil {
			loadErr = fmt.Errorf("%s: %w", relPath, err)
			return filepath.SkipAll
		}
		if _, exists := bundle.ByID[concept.ID]; exists {
			loadErr = fmt.Errorf("duplicate concept ID: %s", concept.ID)
			return filepath.SkipAll
		}
		bundle.Concepts = append(bundle.Concepts, concept)
		bundle.ByID[concept.ID] = concept
		return nil
	})
	if err != nil && !errors.Is(err, filepath.SkipAll) {
		return nil, fmt.Errorf("scan bundle: %w", err)
	}
	if loadErr != nil {
		return nil, loadErr
	}

	sort.Slice(bundle.Concepts, func(i, j int) bool {
		return bundle.Concepts[i].ID < bundle.Concepts[j].ID
	})
	resolveLinks(bundle)
	return bundle, nil
}

func (b *Bundle) Get(id string) (*Concept, error) {
	canonical, err := CanonicalID(id)
	if err != nil {
		return nil, err
	}
	concept, exists := b.ByID[canonical]
	if !exists {
		return nil, fmt.Errorf("concept not found: %s", canonical)
	}
	return concept, nil
}

func CanonicalID(value string) (string, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.TrimPrefix(value, "/")
	value = strings.TrimSuffix(value, "/")
	if strings.HasSuffix(value, ".md") {
		value = strings.TrimSuffix(value, ".md")
	}
	if value == "" || value == "." {
		return "", fmt.Errorf("concept ID must not be empty")
	}
	clean := filepath.ToSlash(filepath.Clean(value))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("invalid concept ID: %q", value)
	}
	if strings.HasSuffix(clean, "/index") || strings.HasSuffix(clean, "/log") || clean == "index" || clean == "log" {
		return "", fmt.Errorf("reserved file is not a concept: %q", value)
	}
	return "/" + clean, nil
}

func (b *Bundle) ConceptPath(id string) (string, string, error) {
	canonical, err := CanonicalID(id)
	if err != nil {
		return "", "", err
	}
	relPath := strings.TrimPrefix(canonical, "/") + ".md"
	absPath := filepath.Join(b.Root, filepath.FromSlash(relPath))
	return canonical, absPath, nil
}

func parseConcept(path, root, relPath string) (*Concept, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	metadata, body, err := parseFrontmatter(string(data))
	if err != nil {
		return nil, err
	}
	id, err := CanonicalID(strings.TrimSuffix(relPath, ".md"))
	if err != nil {
		return nil, err
	}
	return &Concept{
		ID:          id,
		RelPath:     relPath,
		AbsPath:     path,
		Metadata:    metadata,
		Type:        metadataString(metadata, "type"),
		Title:       metadataString(metadata, "title"),
		Description: metadataString(metadata, "description"),
		Tags:        metadataStrings(metadata, "tags"),
		Body:        body,
	}, nil
}

func parseFrontmatter(content string) (map[string]any, string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return nil, "", fmt.Errorf("missing opening YAML frontmatter delimiter")
	}
	closing := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			closing = i
			break
		}
	}
	if closing == -1 {
		return nil, "", fmt.Errorf("missing closing YAML frontmatter delimiter")
	}
	var metadata map[string]any
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:closing], "\n")), &metadata); err != nil {
		return nil, "", fmt.Errorf("frontmatter is not valid YAML: %w", err)
	}
	if metadata == nil {
		metadata = make(map[string]any)
	}
	body := strings.Join(lines[closing+1:], "\n")
	return metadata, body, nil
}

func metadataString(metadata map[string]any, key string) string {
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return fmt.Sprint(value)
	}
	return text
}

func metadataStrings(metadata map[string]any, key string) []string {
	value, ok := metadata[key]
	if !ok {
		return nil
	}
	switch values := value.(type) {
	case []any:
		result := make([]string, 0, len(values))
		for _, value := range values {
			result = append(result, fmt.Sprint(value))
		}
		return result
	case []string:
		return append([]string(nil), values...)
	case string:
		return []string{values}
	default:
		return nil
	}
}

func isReservedMarkdown(relPath string) bool {
	base := filepath.Base(relPath)
	return base == "index.md" || base == "log.md"
}

func displayTitle(concept *Concept) string {
	if concept.Title != "" {
		return concept.Title
	}
	name := strings.TrimSuffix(filepath.Base(concept.RelPath), ".md")
	name = strings.ReplaceAll(name, "-", " ")
	words := strings.Fields(name)
	for index, word := range words {
		runes := []rune(word)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			words[index] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

func displayDescription(concept *Concept) string {
	if concept.Description != "" {
		return concept.Description
	}
	for _, line := range strings.Split(concept.Body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}
