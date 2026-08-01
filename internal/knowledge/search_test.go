package knowledge

import (
	"strings"
	"testing"
)

// searchFixture builds a bundle that mirrors the recommended memory layout:
// preferences, decisions, and references, plus a decoy concept whose title
// contains "task" to guard against substring matching regressions.
func searchFixture(t *testing.T) *Bundle {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, "index.md", "# Test Bundle\n")
	concepts := []struct {
		path        string
		kind        string
		title       string
		description string
		tags        string
		body        string
	}{
		{
			path:        "preferences/ask-dont-assume.md",
			kind:        "Guideline",
			title:       "Ask, Don't Assume",
			description: "Ask focused questions instead of assuming intent.",
			tags:        "[ask, assume, clarification]",
			body:        "Clarify intent by asking instead of assuming. When the user's intent is unclear, ask a focused question before making changes.",
		},
		{
			path:        "preferences/simplicity-before-abstraction.md",
			kind:        "Guideline",
			title:       "Simplicity Before Abstraction",
			description: "Prefer the simplest solution that satisfies demonstrated requirements.",
			tags:        "[simplicity, occams-razor]",
			body:        "Prefer the simplest solution that fully satisfies demonstrated requirements. Before adding an abstraction or architectural boundary, decide what demonstrated requirement it solves and what ongoing cost it introduces.",
		},
		{
			path:        "preferences/avoid-over-engineering.md",
			kind:        "Guideline",
			title:       "Avoid Over-Engineering",
			description: "Do not add complexity for hypothetical future needs.",
			tags:        "[simplicity, overengineering, scope-control]",
			body:        "Do not add abstractions, layers, or dependencies for hypothetical future needs. Choose the solution with fewer moving parts.",
		},
		{
			path:        "decisions/use-stdlib.md",
			kind:        "Decision",
			title:       "Use Standard Library",
			description: "Prefer the standard library over external dependencies.",
			tags:        "[decisions, project, dependencies]",
			body:        "Past project decisions: prefer the standard library over external dependencies.",
		},
		{
			path:        "decisions/go-version.md",
			kind:        "Decision",
			title:       "Go 1.25 Baseline",
			description: "Target Go 1.25 as the minimum supported version.",
			tags:        "[decisions, project, go]",
			body:        "The project targets Go 1.25 as the minimum supported version.",
		},
		{
			path:        "references/mermaid-flowcharts.md",
			kind:        "Reference",
			title:       "Mermaid Flowcharts",
			description: "Model processes, decisions, and dependencies with Mermaid flowcharts.",
			tags:        "[mermaid, diagrams]",
			body:        "Flowcharts model processes, decisions, and dependencies.",
		},
		{
			path:        "references/task-management.md",
			kind:        "Reference",
			title:       "Task Management",
			description: "Track pending tasks and owners.",
			tags:        "[tasks]",
			body:        "Track pending tasks and their owners.",
		},
	}
	for _, concept := range concepts {
		writeTestFile(t, root, concept.path, "---\n"+
			"type: "+concept.kind+"\n"+
			"title: \""+concept.title+"\"\n"+
			"description: \""+concept.description+"\"\n"+
			"tags: "+concept.tags+"\n"+
			"---\n\n"+concept.body+"\n")
	}
	bundle, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return bundle
}

func searchTopID(t *testing.T, bundle *Bundle, query string, options SearchOptions) string {
	t.Helper()
	results := Search(bundle, query, options)
	if len(results) == 0 {
		return ""
	}
	return results[0].Concept.ID
}

func TestSearchNaturalLanguageTask(t *testing.T) {
	bundle := searchFixture(t)
	query := "Please help me decide whether to add an abstraction layer to the CLI"
	results := Search(bundle, query, SearchOptions{Limit: 5})
	if len(results) == 0 {
		t.Fatalf("Search(%q) returned no results", query)
	}
	if got := results[0].Concept.ID; got != "/preferences/simplicity-before-abstraction" {
		t.Fatalf("Search(%q) top result = %s, want simplicity-before-abstraction", query, got)
	}
	fields := results[0].Match.MatchedFields
	if !containsString(fields, "title") {
		t.Fatalf("Search(%q) matched fields = %v, want title", query, fields)
	}
}

func TestSearchAskDontAssume(t *testing.T) {
	bundle := searchFixture(t)
	query := "Ask, don't assume."
	results := Search(bundle, query, SearchOptions{Limit: 5})
	if len(results) == 0 {
		t.Fatalf("Search(%q) returned no results", query)
	}
	if got := results[0].Concept.ID; got != "/preferences/ask-dont-assume" {
		t.Fatalf("Search(%q) top result = %s, want ask-dont-assume", query, got)
	}
	if results[0].Score < 30 {
		t.Fatalf("Search(%q) score = %v, want phrase + title token match >= 30", query, results[0].Score)
	}
	if results[0].Match.Rank != RankTitlePhrase {
		t.Fatalf("Search(%q) rank = %v, want RankTitlePhrase", query, results[0].Match.Rank)
	}
	match := results[0].Match
	if !containsString(match.MatchedFields, "title") {
		t.Fatalf("Search(%q) matched fields = %v, want title", query, match.MatchedFields)
	}
	for _, term := range []string{"ask", "dont", "assume"} {
		if !containsString(match.MatchedTerms, term) {
			t.Fatalf("Search(%q) matched terms = %v, want %q", query, match.MatchedTerms, term)
		}
	}
}

func TestSearchAvoidOverEngineeringNormalization(t *testing.T) {
	bundle := searchFixture(t)
	for _, query := range []string{
		"avoid over-engineering",
		"avoid overengineering",
		"avoid over_engineering",
	} {
		results := Search(bundle, query, SearchOptions{Limit: 5})
		if len(results) == 0 {
			t.Fatalf("Search(%q) returned no results", query)
		}
		if got := results[0].Concept.ID; got != "/preferences/avoid-over-engineering" {
			t.Fatalf("Search(%q) top result = %s, want avoid-over-engineering", query, got)
		}
	}
}

func TestSearchPastProjectDecisions(t *testing.T) {
	bundle := searchFixture(t)
	query := "what decisions did we make in past projects"
	results := Search(bundle, query, SearchOptions{Limit: 5})
	if len(results) < 2 {
		t.Fatalf("Search(%q) returned %d results, want at least 2", query, len(results))
	}
	if got := results[0].Concept.ID; got != "/decisions/use-stdlib" {
		t.Fatalf("Search(%q) top result = %s, want /decisions/use-stdlib", query, got)
	}
	if got := results[1].Concept.ID; got != "/decisions/go-version" {
		t.Fatalf("Search(%q) second result = %s, want /decisions/go-version", query, got)
	}
	for _, result := range results[:2] {
		if result.Concept.Type != "Decision" {
			t.Fatalf("Search(%q) top results must be decisions, got %s (%s)", query, result.Concept.ID, result.Concept.Type)
		}
	}
}

func TestSearchUnrelatedQueryReturnsNothing(t *testing.T) {
	bundle := searchFixture(t)
	for _, query := range []string{
		"how do i make pasta",
		"best pasta recipe cooking",
		"quantum physics equations",
	} {
		if results := Search(bundle, query, SearchOptions{Limit: 5}); len(results) != 0 {
			t.Fatalf("Search(%q) returned %d results, want none", query, len(results))
		}
	}
}

func TestSearchStopWordsOnlyReturnsNothing(t *testing.T) {
	bundle := searchFixture(t)
	if results := Search(bundle, "the and of to", SearchOptions{Limit: 5}); len(results) != 0 {
		t.Fatalf("Search(stop words only) returned %d results, want none", len(results))
	}
}

func TestSearchMatchesTokenBoundaries(t *testing.T) {
	bundle := searchFixture(t)
	results := Search(bundle, "ask", SearchOptions{Limit: 5})
	if got := searchTopID(t, bundle, "ask", SearchOptions{Limit: 5}); got != "/preferences/ask-dont-assume" {
		t.Fatalf("Search(ask) top result = %s, want ask-dont-assume", got)
	}
	for _, result := range results {
		if result.Concept.ID == "/references/task-management" {
			t.Fatalf("Search(ask) matched task-management: substring matching regression")
		}
	}
	if got := searchTopID(t, bundle, "task", SearchOptions{Limit: 5}); got != "/references/task-management" {
		t.Fatalf("Search(task) top result = %s, want task-management", got)
	}
}

func TestSearchExactConceptIDShortCircuits(t *testing.T) {
	bundle := searchFixture(t)
	for _, query := range []string{
		"/preferences/ask-dont-assume",
		"preferences/ask-dont-assume",
		"/preferences/ask-dont-assume.md",
	} {
		results := Search(bundle, query, SearchOptions{Limit: 5})
		if len(results) != 1 {
			t.Fatalf("Search(%q) returned %d results, want exactly 1", query, len(results))
		}
		result := results[0]
		if result.Concept.ID != "/preferences/ask-dont-assume" {
			t.Fatalf("Search(%q) result = %s, want /preferences/ask-dont-assume", query, result.Concept.ID)
		}
		if result.Score != RankExactID.Weight() {
			t.Fatalf("Search(%q) score = %v, want %v", query, result.Score, RankExactID.Weight())
		}
		if result.Match.Rank != RankExactID {
			t.Fatalf("Search(%q) rank = %v, want RankExactID", query, result.Match.Rank)
		}
		if len(result.Match.MatchedFields) != 1 || result.Match.MatchedFields[0] != "id" {
			t.Fatalf("Search(%q) matched fields = %v, want [id]", query, result.Match.MatchedFields)
		}
		if len(result.Match.MatchedTerms) != 1 || result.Match.MatchedTerms[0] != "/preferences/ask-dont-assume" {
			t.Fatalf("Search(%q) matched terms = %v, want canonical ID", query, result.Match.MatchedTerms)
		}
	}
}

func TestSearchReportsMatchedFieldsAndTerms(t *testing.T) {
	bundle := searchFixture(t)

	results := Search(bundle, "mermaid diagram", SearchOptions{Limit: 5})
	if got := searchTopID(t, bundle, "mermaid diagram", SearchOptions{Limit: 5}); got != "/references/mermaid-flowcharts" {
		t.Fatalf("Search(mermaid diagram) top result = %s, want mermaid-flowcharts", got)
	}
	match := results[0].Match
	for _, field := range []string{"title", "tags"} {
		if !containsString(match.MatchedFields, field) {
			t.Fatalf("Search(mermaid diagram) matched fields = %v, want %q", match.MatchedFields, field)
		}
	}
	for _, term := range []string{"mermaid", "diagram"} {
		if !containsString(match.MatchedTerms, term) {
			t.Fatalf("Search(mermaid diagram) matched terms = %v, want %q", match.MatchedTerms, term)
		}
	}

	results = Search(bundle, "owners", SearchOptions{Limit: 5})
	if len(results) == 0 || results[0].Concept.ID != "/references/task-management" {
		t.Fatalf("Search(owners) top result = %v, want task-management", results)
	}
	match = results[0].Match
	if match.Rank != RankDescription {
		t.Fatalf("Search(owners) rank = %v, want RankDescription", match.Rank)
	}
	if len(match.MatchedFields) != 2 || match.MatchedFields[0] != "description" || match.MatchedFields[1] != "body" {
		t.Fatalf("Search(owners) matched fields = %v, want [description body]", match.MatchedFields)
	}
	if len(match.MatchedTerms) != 1 || match.MatchedTerms[0] != "owners" {
		t.Fatalf("Search(owners) matched terms = %v, want [owners]", match.MatchedTerms)
	}
}

func TestSearchResultsSortedByScore(t *testing.T) {
	bundle := searchFixture(t)
	results := Search(bundle, "decisions", SearchOptions{Limit: 5})
	if len(results) < 3 {
		t.Fatalf("Search(decisions) returned %d results, want at least 3", len(results))
	}
	for i := 1; i < len(results); i++ {
		if results[i-1].Score < results[i].Score {
			t.Fatalf("results not sorted by score: %v before %v", results[i-1].Score, results[i].Score)
		}
	}
	if got := results[0].Concept.ID; got != "/decisions/use-stdlib" {
		t.Fatalf("Search(decisions) top result = %s, want /decisions/use-stdlib", got)
	}
}

func TestSearchLimitTruncatesResults(t *testing.T) {
	bundle := searchFixture(t)
	results := Search(bundle, "decisions", SearchOptions{Limit: 2})
	if len(results) != 2 {
		t.Fatalf("Search(decisions, limit 2) returned %d results, want 2", len(results))
	}
}

func TestSearchTokenizeNormalization(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"Ask, don't assume.", []string{"ask", "dont", "assume"}},
		{"avoid over-engineering", []string{"avoid", "engineering"}},
		{"avoid over_engineering", []string{"avoid", "engineering"}},
		{"the and of to", nil},
		{"go 1.25", []string{"go", "25"}},
	}
	for _, test := range tests {
		if got := tokenize(test.input); !equalStrings(got, test.want) {
			t.Fatalf("tokenize(%q) = %v, want %v", test.input, got, test.want)
		}
	}
}

func TestScoreRankSemantics(t *testing.T) {
	ranks := []ScoreRank{
		RankExactID, RankTitlePhrase, RankDescriptionPhrase,
		RankTitle, RankTag, RankDescription, RankID, RankBody,
	}
	// Enum order is the semantic order: weights strictly decrease with rank.
	for i := 1; i < len(ranks); i++ {
		previous := ranks[i-1].Weight()
		current := ranks[i].Weight()
		if previous <= current {
			t.Fatalf("rank %v weight %v must be greater than rank %v weight %v", ranks[i-1], previous, ranks[i], current)
		}
	}
	// Every rank has a stable machine-readable name.
	names := map[ScoreRank]string{
		RankExactID:           "exact_id",
		RankTitlePhrase:       "title_phrase",
		RankDescriptionPhrase: "description_phrase",
		RankTitle:             "title",
		RankTag:               "tags",
		RankDescription:       "description",
		RankID:                "id",
		RankBody:              "body",
	}
	for rank, want := range names {
		if got := rank.String(); got != want {
			t.Fatalf("rank %v String() = %q, want %q", rank, got, want)
		}
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSearchNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Ask, Don't Assume.", "ask dont assume"},
		{"over-engineering", "over engineering"},
		{"over_engineering", "over engineering"},
		{"GO 1.25!", "go 1 25"},
	}
	for _, test := range tests {
		if got := normalize(test.input); got != test.want {
			t.Fatalf("normalize(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestWorkspaceSearchCarriesMatchEvidence(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "prefs/index.md", "---\nokf_version: \"0.1\"\ntype: Bundle\n---\n\n# Prefs\n")
	writeTestConcept(t, root, "prefs/ask-dont-assume.md", "Ask, Don't Assume", "Ask instead of assuming.")
	writeTestFile(t, root, "docs/index.md", "---\nokf_version: \"0.1\"\ntype: Bundle\n---\n\n# Docs\n")
	writeTestConcept(t, root, "docs/flowcharts.md", "Mermaid Flowcharts", "Flowcharts model processes.")

	workspace, err := LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace() error = %v", err)
	}
	results, err := workspace.Search("ask assume", SearchOptions{Limit: 5})
	if err != nil {
		t.Fatalf("workspace.Search() error = %v", err)
	}
	if len(results) == 0 {
		t.Fatal("workspace.Search() returned no results")
	}
	if results[0].Concept.ID != "/ask-dont-assume" {
		t.Fatalf("workspace.Search() top result = %s, want /ask-dont-assume", results[0].Concept.ID)
	}
	if !containsString(results[0].Match.MatchedFields, "title") {
		t.Fatalf("workspace.Search() matched fields = %v, want title", results[0].Match.MatchedFields)
	}
	if !strings.Contains(results[0].Match.MatchedTerms[0], "ask") {
		t.Fatalf("workspace.Search() matched terms = %v, want ask", results[0].Match.MatchedTerms)
	}
}
