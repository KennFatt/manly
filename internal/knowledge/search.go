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

// Match records why a concept matched a query. It is the search evidence
// exposed in rendered JSON as matched_fields, matched_terms, and matched_rank.
type Match struct {
	MatchedFields []string
	MatchedTerms  []string
	Rank          ScoreRank
}

type SearchResult struct {
	Concept    *Concept
	Score      float64
	Match      Match
	BundleName string
}

// stopWords are common English function words that carry no retrieval signal.
// They are removed from query and index tokens so that phrases like
// "how do i decide" do not pollute scoring with noise matches.
var stopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
	"be": true, "been": true, "being": true, "but": true, "by": true,
	"can": true, "could": true, "did": true, "do": true, "does": true,
	"for": true, "from": true, "had": true, "has": true, "have": true,
	"he": true, "her": true, "here": true, "hers": true, "him": true,
	"his": true, "how": true, "i": true, "if": true, "in": true, "into": true,
	"is": true, "it": true, "its": true, "may": true, "me": true,
	"might": true, "must": true, "my": true, "no": true, "not": true,
	"of": true, "off": true, "on": true, "or": true, "our": true,
	"ours": true, "she": true, "should": true, "so": true, "than": true,
	"that": true, "the": true, "their": true, "theirs": true, "them": true,
	"then": true, "there": true, "these": true, "they": true, "this": true,
	"those": true, "to": true, "too": true, "up": true, "us": true,
	"was": true, "we": true, "were": true, "what": true, "when": true,
	"where": true, "which": true, "who": true, "whom": true, "why": true,
	"will": true, "with": true, "would": true, "you": true, "your": true,
	"yours": true, "about": true, "above": true, "after": true,
	"again": true, "against": true, "all": true, "am": true, "any": true,
	"because": true, "before": true, "below": true, "between": true,
	"both": true, "down": true, "during": true, "each": true, "few": true,
	"further": true, "more": true, "most": true, "other": true, "over": true,
	"same": true, "some": true, "such": true, "under": true, "until": true,
	"very": true, "while": true, "yet": true,
}

// ScoreRank is the semantic class of a match, ordered from strongest to
// weakest. The enum value is the rank; Weight returns its fixed score
// contribution. Ranks are ordinal, so comparisons like "stronger than a
// description match" are expressed as rank comparisons instead of magic
// numbers.
type ScoreRank int

const (
	// RankExactID is a direct concept-ID lookup, the highest-confidence path.
	RankExactID ScoreRank = iota
	// RankTitlePhrase is the normalized query appearing verbatim in the title.
	RankTitlePhrase
	// RankDescriptionPhrase is the normalized query appearing verbatim in the
	// description.
	RankDescriptionPhrase
	// RankTitle is a token match in the title.
	RankTitle
	// RankTag is a token match in the tags.
	RankTag
	// RankDescription is a token match in the description.
	RankDescription
	// RankID is a token match in the concept ID path.
	RankID
	// RankBody is a token match in the body.
	RankBody
)

// rankWeights is the single source of truth for score contributions. Every
// rank has exactly one weight, and weights strictly decrease with rank so
// that a stronger signal always outranks a weaker one.
var rankWeights = [...]float64{
	RankExactID:           1000,
	RankTitlePhrase:       12,
	RankDescriptionPhrase: 8,
	RankTitle:             6,
	RankTag:               5,
	RankDescription:       3,
	RankID:                2,
	RankBody:              1,
}

// Weight returns the fixed score contribution of the rank.
func (r ScoreRank) Weight() float64 {
	if int(r) < 0 || int(r) >= len(rankWeights) {
		return 0
	}
	return rankWeights[r]
}

// String returns the stable machine-readable name of the rank.
func (r ScoreRank) String() string {
	switch r {
	case RankExactID:
		return "exact_id"
	case RankTitlePhrase:
		return "title_phrase"
	case RankDescriptionPhrase:
		return "description_phrase"
	case RankTitle:
		return "title"
	case RankTag:
		return "tags"
	case RankDescription:
		return "description"
	case RankID:
		return "id"
	case RankBody:
		return "body"
	}
	return "unknown"
}

// Confidence is the semantic confidence tier of a match, ordered from
// strongest to weakest. The enum value is the tier; String returns its stable
// wire name for rendered JSON.
type Confidence int

const (
	// ConfidenceHigh covers exact and phrase matches.
	ConfidenceHigh Confidence = iota
	// ConfidenceMedium covers structured metadata matches.
	ConfidenceMedium
	// ConfidenceLow covers weak lexical matches.
	ConfidenceLow
)

// String returns the stable machine-readable wire name of the tier.
func (c Confidence) String() string {
	switch c {
	case ConfidenceHigh:
		return "high"
	case ConfidenceMedium:
		return "medium"
	case ConfidenceLow:
		return "low"
	}
	return "unknown"
}

// Confidence returns the semantic confidence tier of the rank: high for exact
// and phrase matches, medium for structured metadata, low for weak lexical
// matches. The CLI uses the tier to flag results that should not be treated
// as authoritative context.
func (r ScoreRank) Confidence() Confidence {
	switch r {
	case RankExactID, RankTitlePhrase, RankDescriptionPhrase:
		return ConfidenceHigh
	case RankTitle, RankTag:
		return ConfidenceMedium
	case RankDescription, RankID, RankBody:
		return ConfidenceLow
	}
	return ConfidenceLow
}

// stronger returns the semantically stronger of two ranks. Enum values are
// ordered strongest to weakest, so the smaller value wins.
func stronger(a, b ScoreRank) ScoreRank {
	if a < b {
		return a
	}
	return b
}

// Search ranks concepts by lexical relevance. An exact concept-ID query
// short-circuits with a perfect match; otherwise scoring uses normalized
// tokens with boundary matching, a phrase bonus, and per-field weights.
func Search(bundle *Bundle, query string, options SearchOptions) []SearchResult {
	if concept, ok := exactConceptID(bundle, query); ok {
		return []SearchResult{{
			Concept: concept,
			Score:   RankExactID.Weight(),
			Match: Match{
				MatchedFields: []string{"id"},
				MatchedTerms:  []string{concept.ID},
				Rank:          RankExactID,
			},
		}}
	}
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil
	}
	if options.Limit <= 0 {
		options.Limit = 10
	}
	pathFilter := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(options.Path), ".md"))
	pathFilter = strings.TrimPrefix(pathFilter, "/")
	phrase := normalize(query)
	var results []SearchResult
	for _, concept := range bundle.Concepts {
		if !matchesFilters(concept, options, pathFilter) {
			continue
		}
		score, match := scoreConcept(concept, tokens, phrase)
		if score > 0 {
			results = append(results, SearchResult{Concept: concept, Score: score, Match: match})
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

// exactConceptID returns the concept when the query is a valid concept ID.
// Bare words are accepted as IDs too, but only match when the concept exists.
func exactConceptID(bundle *Bundle, query string) (*Concept, bool) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, false
	}
	canonical, err := CanonicalID(trimmed)
	if err != nil {
		return nil, false
	}
	concept, exists := bundle.ByID[canonical]
	return concept, exists
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

// fieldTokens are the normalized searchable surfaces of one concept.
type fieldTokens struct {
	title       string
	titleJoined string
	description string
	descJoined  string
	titleSet    map[string]bool
	tagSet      map[string]bool
	descSet     map[string]bool
	idSet       map[string]bool
	bodySet     map[string]bool
}

func buildFieldTokens(concept *Concept) fieldTokens {
	title := normalize(displayTitle(concept))
	description := normalize(displayDescription(concept))
	return fieldTokens{
		title:       title,
		titleJoined: strings.ReplaceAll(title, " ", ""),
		description: description,
		descJoined:  strings.ReplaceAll(description, " ", ""),
		titleSet:    tokenSet(displayTitle(concept)),
		tagSet:      tokenSet(strings.Join(concept.Tags, " ")),
		descSet:     tokenSet(displayDescription(concept)),
		idSet:       tokenSet(concept.ID),
		bodySet:     tokenSet(concept.Body),
	}
}

func tokenSet(value string) map[string]bool {
	set := make(map[string]bool)
	for _, token := range tokenize(value) {
		set[token] = true
	}
	return set
}

// scoreConcept weighs token matches by field and adds a phrase bonus when the
// normalized query appears verbatim in the title or description. The phrase
// bonus only applies to multi-token queries: a single word is a token match,
// not a phrase, and joined-form phrase checks would otherwise match arbitrary
// substrings such as "ask" inside "taskmanagement".
func scoreConcept(concept *Concept, tokens []string, phrase string) (float64, Match) {
	fields := buildFieldTokens(concept)
	matched := make(map[string]bool)
	var score float64
	top := RankBody
	if len(tokens) >= 2 && phrase != "" {
		if strings.Contains(fields.title, phrase) || strings.Contains(fields.titleJoined, phrase) {
			score += RankTitlePhrase.Weight()
			matched["title"] = true
			top = stronger(top, RankTitlePhrase)
		}
		if strings.Contains(fields.description, phrase) || strings.Contains(fields.descJoined, phrase) {
			score += RankDescriptionPhrase.Weight()
			matched["description"] = true
			top = stronger(top, RankDescriptionPhrase)
		}
	}
	var matchedTerms []string
	for _, token := range tokens {
		hit := false
		if tokenMatches(fields.titleSet, token) {
			score += RankTitle.Weight()
			matched["title"] = true
			top = stronger(top, RankTitle)
			hit = true
		}
		if tokenMatches(fields.tagSet, token) {
			score += RankTag.Weight()
			matched["tags"] = true
			top = stronger(top, RankTag)
			hit = true
		}
		if tokenMatches(fields.descSet, token) {
			score += RankDescription.Weight()
			matched["description"] = true
			top = stronger(top, RankDescription)
			hit = true
		}
		if tokenMatches(fields.idSet, token) {
			score += RankID.Weight()
			matched["id"] = true
			top = stronger(top, RankID)
			hit = true
		}
		if tokenMatches(fields.bodySet, token) {
			score += RankBody.Weight()
			matched["body"] = true
			top = stronger(top, RankBody)
			hit = true
		}
		if hit {
			matchedTerms = append(matchedTerms, token)
		}
	}
	order := []string{"id", "title", "tags", "description", "body"}
	var matchedFields []string
	for _, field := range order {
		if matched[field] {
			matchedFields = append(matchedFields, field)
		}
	}
	return score, Match{MatchedFields: matchedFields, MatchedTerms: matchedTerms, Rank: top}
}

// tokenMatches reports whether query and index tokens refer to the same word.
// Exact matches always count; otherwise one token must extend the other by a
// known inflection suffix (s, es, ies, ed, d). This covers plurals
// ("project"/"projects", "layer"/"layers") while rejecting accidental
// substrings ("ask"/"task", "pasta"/"past") and loose stems
// ("make"/"making", "react"/"reactive").
func tokenMatches(set map[string]bool, token string) bool {
	if set[token] {
		return true
	}
	for candidate := range set {
		if len(candidate) > len(token) && strings.HasPrefix(candidate, token) {
			if isInflection(candidate[len(token):]) {
				return true
			}
		}
		if len(token) > len(candidate) && strings.HasPrefix(token, candidate) {
			if isInflection(token[len(candidate):]) {
				return true
			}
		}
	}
	return false
}

// isInflection reports whether a suffix is a common plural or verb inflection.
func isInflection(suffix string) bool {
	switch suffix {
	case "s", "es", "ies", "ed", "d":
		return true
	}
	return false
}

// normalize lowercases text and folds punctuation into separators: hyphens and
// underscores become spaces, apostrophes are removed, and every other symbol
// becomes a space. "over-engineering", "over_engineering", and "Over
// Engineering" all normalize to "over engineering"; "don't" becomes "dont".
func normalize(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "'", "")
	value = strings.ReplaceAll(value, "’", "")
	value = strings.ReplaceAll(value, "-", " ")
	value = strings.ReplaceAll(value, "_", " ")
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			builder.WriteRune(r)
		} else {
			builder.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

// tokenize splits normalized text into search tokens, dropping single
// characters and stop words.
func tokenize(value string) []string {
	fields := strings.FieldsFunc(normalize(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) < 2 || stopWords[field] {
			continue
		}
		tokens = append(tokens, field)
	}
	return tokens
}

func containsFold(values []string, wanted string) bool {
	for _, value := range values {
		if strings.EqualFold(value, wanted) {
			return true
		}
	}
	return false
}
