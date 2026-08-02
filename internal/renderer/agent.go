package renderer

import "io"

type agentRenderer struct{}

var _ Renderer = agentRenderer{}

func (agentRenderer) Format() Format {
	return FormatAgent
}

func (agentRenderer) Render(w io.Writer, view View) error {
	list, ok := view.(ListView)
	if !ok {
		return unsupportedView(view)
	}
	return renderAgentList(w, list)
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
