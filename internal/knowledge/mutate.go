package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type NewConcept struct {
	Type        string
	Title       string
	Description string
	Tags        []string
}

func Add(root, id string, input NewConcept, force bool) (string, error) {
	bundle, err := Load(root)
	if err != nil {
		return "", err
	}
	canonical, path, err := bundle.ConceptPath(id)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil && !force {
		return "", fmt.Errorf("concept already exists: %s", canonical)
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("check concept path: %w", err)
	}
	if strings.TrimSpace(input.Type) == "" {
		return "", fmt.Errorf("--type is required")
	}
	if input.Title == "" {
		input.Title = titleFromID(canonical)
	}
	metadata := map[string]any{"type": input.Type, "title": input.Title}
	if input.Description != "" {
		metadata["description"] = input.Description
	}
	if len(input.Tags) > 0 {
		metadata["tags"] = input.Tags
	}
	frontmatter, err := yaml.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("encode frontmatter: %w", err)
	}
	body := fmt.Sprintf("---\n%s---\n\n# %s\n\n%s\n", string(frontmatter), input.Title, input.Description)
	if err := writeAtomic(path, []byte(body)); err != nil {
		return "", fmt.Errorf("write concept: %w", err)
	}
	return canonical, nil
}

func Move(root, oldID, newID string) (int, error) {
	bundle, err := Load(root)
	if err != nil {
		return 0, err
	}
	oldConcept, err := bundle.Get(oldID)
	if err != nil {
		return 0, err
	}
	newCanonical, newPath, err := bundle.ConceptPath(newID)
	if err != nil {
		return 0, err
	}
	if _, err := os.Stat(newPath); err == nil {
		return 0, fmt.Errorf("destination already exists: %s", newCanonical)
	} else if err != nil && !os.IsNotExist(err) {
		return 0, fmt.Errorf("check destination: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return 0, fmt.Errorf("create destination directory: %w", err)
	}
	if err := os.Rename(oldConcept.AbsPath, newPath); err != nil {
		return 0, fmt.Errorf("move concept: %w", err)
	}

	oldRel := filepath.ToSlash(oldConcept.RelPath)
	newRel, err := filepath.Rel(bundle.Root, newPath)
	if err != nil {
		return 0, fmt.Errorf("resolve destination: %w", err)
	}
	changed := 0
	for relPath, path := range bundle.Markdown {
		if relPath == oldRel {
			continue
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return changed, fmt.Errorf("read %s after move: %w", relPath, readErr)
		}
		updated, replacements := rewriteInternalLinks(string(data), relPath, oldRel, filepath.ToSlash(newRel))
		if replacements == 0 {
			continue
		}
		if writeErr := writeAtomic(path, []byte(updated)); writeErr != nil {
			return changed, fmt.Errorf("update links in %s: %w", relPath, writeErr)
		}
		changed += replacements
	}
	return changed, nil
}

func UpdateIndexes(root string, checkOnly bool) ([]string, error) {
	bundle, err := Load(root)
	if err != nil {
		return nil, err
	}
	indexPaths := make([]string, 0)
	for relPath, path := range bundle.Markdown {
		if filepath.Base(relPath) == "index.md" {
			indexPaths = append(indexPaths, path)
		}
	}
	sort.Strings(indexPaths)
	var changed []string
	for _, indexPath := range indexPaths {
		data, err := os.ReadFile(indexPath)
		if err != nil {
			return changed, err
		}
		content := string(data)
		start := strings.Index(content, "<!-- manly:generated:start -->")
		end := strings.Index(content, "<!-- manly:generated:end -->")
		if start < 0 || end < start {
			continue
		}
		relDir := filepath.Dir(relativePath(bundle.Root, indexPath))
		generated := generatedIndexEntries(bundle, relDir)
		replacement := "<!-- manly:generated:start -->\n" + generated + "<!-- manly:generated:end -->"
		end += len("<!-- manly:generated:end -->")
		updated := content[:start] + replacement + content[end:]
		if updated == content {
			continue
		}
		if checkOnly {
			changed = append(changed, relativePath(bundle.Root, indexPath))
			continue
		}
		if err := writeAtomic(indexPath, []byte(updated)); err != nil {
			return changed, err
		}
		changed = append(changed, relativePath(bundle.Root, indexPath))
	}
	return changed, nil
}

func generatedIndexEntries(bundle *Bundle, directory string) string {
	prefix := filepath.ToSlash(directory)
	if prefix == "." {
		prefix = ""
	}
	var concepts []*Concept
	for _, concept := range bundle.Concepts {
		conceptDirectory := filepath.ToSlash(filepath.Dir(concept.RelPath))
		if conceptDirectory == "." {
			conceptDirectory = ""
		}
		if conceptDirectory == prefix {
			concepts = append(concepts, concept)
		}
	}
	sort.Slice(concepts, func(i, j int) bool { return concepts[i].ID < concepts[j].ID })
	var lines strings.Builder
	for _, concept := range concepts {
		relative, err := filepath.Rel(directory, concept.RelPath)
		if err != nil {
			continue
		}
		lines.WriteString(fmt.Sprintf("* [%s](%s) - %s\n", displayTitle(concept), filepath.ToSlash(relative), displayDescription(concept)))
	}
	return lines.String()
}

func rewriteInternalLinks(content, sourceRel, oldRel, newRel string) (string, int) {
	var builder strings.Builder
	last := 0
	replacements := 0
	for _, match := range markdownLinkPattern.FindAllStringSubmatchIndex(content, -1) {
		if match[0] > 0 && content[match[0]-1] == '!' {
			continue
		}
		rawTarget := strings.Trim(content[match[4]:match[5]], "<>")
		if isExternalTarget(rawTarget) {
			continue
		}
		target := strings.SplitN(rawTarget, "#", 2)[0]
		resolved, _, broken := resolveLinkPath(sourceRel, target)
		if broken || resolved != oldRel {
			continue
		}
		replacement := linkTargetForSource(sourceRel, newRel, rawTarget)
		builder.WriteString(content[last:match[4]])
		builder.WriteString(replacement)
		last = match[5]
		replacements++
	}
	if replacements == 0 {
		return content, 0
	}
	builder.WriteString(content[last:])
	return builder.String(), replacements
}

func resolveLinkPath(sourceRel, target string) (string, string, bool) {
	if target == "" {
		return "", "", false
	}
	target = strings.ReplaceAll(target, "\\", "/")
	if strings.HasPrefix(target, "/") {
		target = strings.TrimPrefix(target, "/")
	} else {
		target = filepath.ToSlash(filepath.Join(filepath.Dir(sourceRel), target))
	}
	target = filepath.ToSlash(filepath.Clean(target))
	if filepath.Ext(target) == "" {
		target += ".md"
	}
	return target, "", false
}

func linkTargetForSource(sourceRel, newRel, oldTarget string) string {
	fragment := ""
	if index := strings.Index(oldTarget, "#"); index >= 0 {
		fragment = oldTarget[index:]
	}
	newTarget := ""
	if strings.HasPrefix(oldTarget, "/") {
		newTarget = "/" + newRel
	} else {
		relative, err := filepath.Rel(filepath.Dir(sourceRel), filepath.FromSlash(newRel))
		if err == nil {
			newTarget = filepath.ToSlash(relative)
		} else {
			newTarget = newRel
		}
	}
	if !strings.HasSuffix(oldTarget, ".md") {
		newTarget = strings.TrimSuffix(newTarget, ".md")
	}
	return newTarget + fragment
}

func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".manly-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, path)
}

func titleFromID(id string) string {
	base := filepath.Base(strings.TrimSuffix(id, ".md"))
	base = strings.ReplaceAll(base, "-", " ")
	if base == "" {
		return "Concept"
	}
	return strings.ToUpper(base[:1]) + base[1:]
}
