package knowledge

import (
	"sort"
	"strings"
	"unicode"
)

type SearchOptions struct {
	Tag   string
	Type  string
	Path  string
	Limit int
}

type SearchResult struct {
	Concept *Concept
	Score   float64
}

func Search(bundle *Bundle, query string, options SearchOptions) []SearchResult {
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil
	}
	if options.Limit <= 0 {
		options.Limit = 10
	}
	pathFilter := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(options.Path), ".md"))
	pathFilter = strings.TrimPrefix(pathFilter, "/")
	var results []SearchResult
	for _, concept := range bundle.Concepts {
		if !matchesFilters(concept, options, pathFilter) {
			continue
		}
		score := scoreConcept(concept, query, tokens)
		if score > 0 {
			results = append(results, SearchResult{Concept: concept, Score: score})
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Concept.ID < results[j].Concept.ID
	})
	if len(results) > options.Limit {
		results = results[:options.Limit]
	}
	return results
}

func matchesFilters(concept *Concept, options SearchOptions, pathFilter string) bool {
	if options.Tag != "" && !containsFold(concept.Tags, options.Tag) {
		return false
	}
	if options.Type != "" && !strings.Contains(strings.ToLower(concept.Type), strings.ToLower(options.Type)) {
		return false
	}
	if pathFilter != "" && !strings.HasPrefix(strings.ToLower(strings.TrimPrefix(concept.ID, "/")), pathFilter) {
		return false
	}
	return true
}

func scoreConcept(concept *Concept, query string, tokens []string) float64 {
	title := strings.ToLower(displayTitle(concept))
	description := strings.ToLower(displayDescription(concept))
	tags := strings.ToLower(strings.Join(concept.Tags, " "))
	id := strings.ToLower(concept.ID)
	body := strings.ToLower(concept.Body)
	phrase := strings.ToLower(strings.TrimSpace(query))
	var score float64
	if phrase != "" && strings.Contains(title, phrase) {
		score += 12
	}
	if phrase != "" && strings.Contains(description, phrase) {
		score += 8
	}
	for _, token := range tokens {
		switch {
		case strings.Contains(title, token):
			score += 6
		case strings.Contains(tags, token):
			score += 5
		case strings.Contains(description, token):
			score += 3
		case strings.Contains(id, token):
			score += 2
		case strings.Contains(body, token):
			score += 1
		}
	}
	return score
}

func tokenize(value string) []string {
	return strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func containsFold(values []string, wanted string) bool {
	for _, value := range values {
		if strings.EqualFold(value, wanted) {
			return true
		}
	}
	return false
}
