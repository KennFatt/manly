package renderer

import "io"

type agentRenderer struct{}

var _ Renderer = agentRenderer{}

func (agentRenderer) Format() Format {
	return FormatAgent
}

func (agentRenderer) Render(w io.Writer, view View) error {
	switch value := view.(type) {
	case ListView:
		return renderAgentList(w, value)
	case TreeView:
		return renderAgentTree(w, value)
	default:
		return unsupportedView(view)
	}
}

type agentConcept struct {
	ID          string   `json:"id"`
	Type        string   `json:"type,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type agentList struct {
	Path        string         `json:"path"`
	Recursive   bool           `json:"recursive"`
	Directories []string       `json:"directories"`
	Concepts    []agentConcept `json:"concepts"`
	Truncated   bool           `json:"truncated"`
}

type agentTree struct {
	Path         string     `json:"path"`
	Heading      string     `json:"heading,omitempty"`
	Description  string     `json:"description,omitempty"`
	Depth        int        `json:"depth"`
	ConceptCount int        `json:"concept_count"`
	Nodes        []TreeNode `json:"nodes"`
}

func renderAgentTree(w io.Writer, view TreeView) error {
	nodes := view.Nodes
	if nodes == nil {
		nodes = []TreeNode{}
	}
	return writeJSON(w, agentTree{
		Path:         view.Path,
		Heading:      view.Heading,
		Description:  view.Description,
		Depth:        view.Depth,
		ConceptCount: view.ConceptCount,
		Nodes:        nodes,
	})
}

func renderAgentList(w io.Writer, view ListView) error {
	directories := make([]string, 0, len(view.Directories))
	for _, directory := range view.Directories {
		directories = append(directories, directory.Path)
	}

	concepts := make([]agentConcept, 0, len(view.Entries))
	for _, entry := range view.Entries {
		concept := entry.Concept
		concepts = append(concepts, agentConcept{
			ID:          concept.ID,
			Type:        concept.Type,
			Title:       concept.Title,
			Description: concept.Description,
			Tags:        append([]string(nil), concept.Tags...),
		})
	}

	return writeJSON(w, agentList{
		Path:        view.Path,
		Recursive:   view.Recursive,
		Directories: directories,
		Concepts:    concepts,
		Truncated:   false,
	})
}
