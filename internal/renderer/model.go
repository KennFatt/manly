package renderer

import "time"

// View is a typed presentation model accepted by a Renderer.
type View interface {
	view()
}

// Concept contains presentation-ready concept metadata and content.
type Concept struct {
	ID          string   `json:"id"`
	Path        string   `json:"path"`
	Type        string   `json:"type,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Content     string   `json:"content,omitempty"`
}

func (Concept) view() {}

// Link contains presentation-ready link information.
type Link struct {
	Label      string `json:"label"`
	Title      string `json:"-"`
	Target     string `json:"target,omitempty"`
	TargetPath string `json:"target_path,omitempty"`
	URL        string `json:"url,omitempty"`
	Broken     bool   `json:"broken,omitempty"`
	External   bool   `json:"external,omitempty"`
}

// Action contains a navigational CLI action.
type Action struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// Directory contains a directory path and its concept count.
type Directory struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// ListEntry contains one concept and its available actions.
type ListEntry struct {
	Concept Concept  `json:"concept"`
	Actions []Action `json:"actions,omitempty"`
}

// ListView contains directory listing data.
type ListView struct {
	Root        string      `json:"root"`
	Path        string      `json:"path"`
	Heading     string      `json:"heading,omitempty"`
	Recursive   bool        `json:"recursive"`
	Directories []Directory `json:"directories"`
	Entries     []ListEntry `json:"entries"`
	Count       int         `json:"count,omitempty"`
	HideActions bool        `json:"-"`
	HideUsage   bool        `json:"-"`
}

func (ListView) view() {}

// ShowView contains one complete concept and its relationships.
type ShowView struct {
	Concept     Concept  `json:"concept"`
	Links       []Link   `json:"links"`
	Backlinks   []Link   `json:"backlinks"`
	Actions     []Action `json:"actions,omitempty"`
	HideActions bool     `json:"-"`
	HideUsage   bool     `json:"-"`
}

func (ShowView) view() {}

// ShowResult contains one concept and its relationships in a collection.
type ShowResult struct {
	Concept   Concept  `json:"concept"`
	Links     []Link   `json:"links"`
	Backlinks []Link   `json:"backlinks"`
	Actions   []Action `json:"actions,omitempty"`
	HideUsage bool     `json:"-"`
}

// ShowCollectionView contains multiple complete concepts and their relationships.
type ShowCollectionView struct {
	Results []ShowResult `json:"results"`
}

func (ShowCollectionView) view() {}

// SourceInfo identifies where results came from.
type SourceInfo struct {
	Root string `json:"root"`
	Mode string `json:"mode"`
}

// SearchResult contains one scored search result.
type SearchResult struct {
	Concept       Concept  `json:"concept"`
	Score         float64  `json:"score"`
	MatchedFields []string `json:"matched_fields,omitempty"`
	MatchedTerms  []string `json:"matched_terms,omitempty"`
	MatchedRank   string   `json:"matched_rank,omitempty"`
	Confidence    string   `json:"confidence,omitempty"`
	Bundle        string   `json:"bundle,omitempty"`
	Actions       []Action `json:"actions"`
}

// SearchView contains search results for a query.
type SearchView struct {
	Query     string         `json:"query"`
	Source    SourceInfo     `json:"source"`
	Confident bool           `json:"confident"`
	Results   []SearchResult `json:"results"`
}

func (SearchView) view() {}

// ContextResult contains one context concept and its links.
type ContextResult struct {
	Concept       Concept  `json:"concept"`
	Score         float64  `json:"score"`
	MatchedFields []string `json:"matched_fields,omitempty"`
	MatchedTerms  []string `json:"matched_terms,omitempty"`
	MatchedRank   string   `json:"matched_rank,omitempty"`
	Confidence    string   `json:"confidence,omitempty"`
	Bundle        string   `json:"bundle,omitempty"`
	Links         []Link   `json:"links"`
	Actions       []Action `json:"actions"`
}

// ContextView contains bounded context results for a query.
type ContextView struct {
	Query     string          `json:"query"`
	Source    SourceInfo      `json:"source"`
	Confident bool            `json:"confident"`
	Results   []ContextResult `json:"results"`
}

func (ContextView) view() {}

// LinksView contains outgoing links for one concept.
type LinksView struct {
	Source string `json:"source"`
	Links  []Link `json:"links"`
}

func (LinksView) view() {}

// BacklinksView contains incoming links for one concept.
type BacklinksView struct {
	Target    string `json:"target"`
	Backlinks []Link `json:"backlinks"`
}

func (BacklinksView) view() {}

// GraphNode contains one concept and its traversal depth.
type GraphNode struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Depth int    `json:"depth"`
}

// GraphView contains graph traversal nodes.
type GraphView struct {
	Nodes []GraphNode `json:"nodes"`
}

func (GraphView) view() {}

// AnalyticsView contains local concept-usage analytics.
type AnalyticsView struct {
	Enabled                 bool               `json:"enabled"`
	Provider                string             `json:"provider,omitempty"`
	Period                  AnalyticsPeriod    `json:"period"`
	ConceptLoads            int                `json:"concept_loads"`
	RetrievalBatches        int                `json:"retrieval_batches"`
	AverageConceptsPerBatch float64            `json:"average_concepts_per_batch"`
	EntryPoints             map[string]int     `json:"entry_points"`
	TopConcepts             []AnalyticsConcept `json:"top_concepts"`
	RecentBatches           []AnalyticsBatch   `json:"recent_batches"`
}

// AnalyticsPeriod identifies the lower bound used for a report.
type AnalyticsPeriod struct {
	Since *time.Time `json:"since"`
}

// AnalyticsConcept contains one concept's usage count.
type AnalyticsConcept struct {
	ConceptID string `json:"concept_id"`
	LoadCount int    `json:"load_count"`
}

// AnalyticsBatch contains one recent retrieval group.
type AnalyticsBatch struct {
	BatchID      string    `json:"batch_id"`
	OccurredAt   time.Time `json:"occurred_at"`
	EntryPoint   string    `json:"entry_point"`
	ConceptCount int       `json:"concept_count"`
	ConceptIDs   []string  `json:"concept_ids"`
}

func (AnalyticsView) view() {}

// Issue contains one validation issue.
type Issue struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// CheckStats contains aggregate validation and scan statistics.
type CheckStats struct {
	Bundles               int `json:"bundles"`
	MarkdownFiles         int `json:"markdown_files"`
	ReservedFiles         int `json:"reserved_files"`
	ConceptFiles          int `json:"concept_files"`
	LoadedConcepts        int `json:"loaded_concepts"`
	InvalidConceptFiles   int `json:"invalid_concept_files"`
	LinksChecked          int `json:"links_checked"`
	BrokenLinks           int `json:"broken_links"`
	StaleGeneratedIndexes int `json:"stale_generated_indexes"`
	Errors                int `json:"errors"`
	Warnings              int `json:"warnings"`
}

// CheckBundle contains per-bundle validation statistics.
type CheckBundle struct {
	Name                  string `json:"name"`
	Root                  string `json:"root"`
	MarkdownFiles         int    `json:"markdown_files"`
	ReservedFiles         int    `json:"reserved_files"`
	ConceptFiles          int    `json:"concept_files"`
	LoadedConcepts        int    `json:"loaded_concepts"`
	InvalidConceptFiles   int    `json:"invalid_concept_files"`
	LinksChecked          int    `json:"links_checked"`
	BrokenLinks           int    `json:"broken_links"`
	StaleGeneratedIndexes int    `json:"stale_generated_indexes"`
}

// CheckView contains bundle validation results.
type CheckView struct {
	Root     string        `json:"root"`
	Mode     string        `json:"mode"`
	Stats    CheckStats    `json:"stats"`
	Bundles  []CheckBundle `json:"bundles,omitempty"`
	Errors   []Issue       `json:"Errors"`
	Warnings []Issue       `json:"Warnings"`
	Valid    bool          `json:"valid"`
}

func (CheckView) view() {}
