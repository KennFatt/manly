package knowledge

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var markdownLinkPattern = regexp.MustCompile(`\[([^\]\n]+)\]\((<[^>\n]+>|[^)\s]+)(?:\s+[^)]*)?\)`)

func resolveLinks(bundle *Bundle) {
	for _, concept := range bundle.Concepts {
		concept.Links = resolveDocumentLinks(bundle, concept.RelPath, concept.Body)
	}
}

func resolveDocumentLinks(bundle *Bundle, sourceRelPath, content string) []Link {
	links := extractMarkdownLinks(content)
	result := make([]Link, 0, len(links))
	for _, link := range links {
		resolved := link
		resolved.RawTarget = strings.Trim(resolved.RawTarget, "<>")
		if isExternalTarget(resolved.RawTarget) {
			resolved.External = true
			result = append(result, resolved)
			continue
		}
		resolved.TargetPath, resolved.TargetID, resolved.Broken = resolveLocalTarget(bundle, sourceRelPath, resolved.RawTarget)
		result = append(result, resolved)
	}
	return result
}

func extractMarkdownLinks(content string) []Link {
	var result []Link
	inFence := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		matches := markdownLinkPattern.FindAllStringSubmatchIndex(line, -1)
		for _, match := range matches {
			label := line[match[2]:match[3]]
			rawTarget := line[match[4]:match[5]]
			if match[0] > 0 && line[match[0]-1] == '!' {
				continue
			}
			result = append(result, Link{Label: label, RawTarget: rawTarget})
		}
	}
	return result
}

func isExternalTarget(target string) bool {
	if target == "" || strings.HasPrefix(target, "#") {
		return true
	}
	parsed, err := url.Parse(target)
	if err == nil && parsed.Scheme != "" {
		return true
	}
	return strings.HasPrefix(target, "//")
}

func resolveLocalTarget(bundle *Bundle, sourceRelPath, rawTarget string) (string, string, bool) {
	target := strings.SplitN(rawTarget, "#", 2)[0]
	if target == "" {
		return "", "", false
	}
	target = strings.Trim(target, "<>")
	target = strings.ReplaceAll(target, "\\", "/")
	var relative string
	if strings.HasPrefix(target, "/") {
		relative = strings.TrimPrefix(target, "/")
	} else {
		sourceDir := filepath.ToSlash(filepath.Dir(sourceRelPath))
		relative = filepath.ToSlash(filepath.Join(sourceDir, target))
	}
	relative = filepath.ToSlash(filepath.Clean(relative))
	if relative == "." || relative == ".." || strings.HasPrefix(relative, "../") {
		return relative, "", true
	}

	candidates := []string{relative}
	if filepath.Ext(relative) == "" {
		candidates = append(candidates, relative+".md", filepath.Join(relative, "index.md"))
	}
	for _, candidate := range candidates {
		candidate = filepath.ToSlash(candidate)
		if _, exists := bundle.Markdown[candidate]; !exists {
			continue
		}
		if isReservedMarkdown(candidate) {
			return candidate, "", false
		}
		id, err := CanonicalID(strings.TrimSuffix(candidate, ".md"))
		if err != nil {
			return candidate, "", false
		}
		if _, exists := bundle.ByID[id]; exists {
			return candidate, id, false
		}
		return candidate, "", false
	}
	return relative, "", true
}

func (b *Bundle) Outgoing(id string) ([]Link, error) {
	concept, err := b.Get(id)
	if err != nil {
		return nil, err
	}
	links := append([]Link(nil), concept.Links...)
	return links, nil
}

func (b *Bundle) Backlinks(id string) ([]Backlink, error) {
	canonical, err := CanonicalID(id)
	if err != nil {
		return nil, err
	}
	if _, exists := b.ByID[canonical]; !exists {
		return nil, fmt.Errorf("concept not found: %s", canonical)
	}
	var result []Backlink
	for _, concept := range b.Concepts {
		for _, link := range concept.Links {
			if link.TargetID == canonical {
				result = append(result, Backlink{Concept: concept, Link: link})
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Concept.ID < result[j].Concept.ID
	})
	return result, nil
}

type Backlink struct {
	Concept *Concept
	Link    Link
}

type GraphNode struct {
	Concept *Concept
	Depth   int
}

func (b *Bundle) Graph(id string, maxDepth int) ([]GraphNode, error) {
	start, err := b.Get(id)
	if err != nil {
		return nil, err
	}
	if maxDepth < 0 {
		return nil, fmt.Errorf("depth must not be negative")
	}
	result := []GraphNode{{Concept: start, Depth: 0}}
	visited := map[string]bool{start.ID: true}
	queue := []GraphNode{{Concept: start, Depth: 0}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.Depth >= maxDepth {
			continue
		}
		for _, link := range current.Concept.Links {
			if link.TargetID == "" || visited[link.TargetID] {
				continue
			}
			target := b.ByID[link.TargetID]
			if target == nil {
				continue
			}
			visited[target.ID] = true
			node := GraphNode{Concept: target, Depth: current.Depth + 1}
			result = append(result, node)
			queue = append(queue, node)
		}
	}
	return result, nil
}
