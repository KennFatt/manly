package renderer

import (
	"io"
)

type jsonRenderer struct{}

var _ Renderer = jsonRenderer{}

func (jsonRenderer) Format() Format {
	return FormatJSON
}

func (jsonRenderer) Render(w io.Writer, view View) error {
	switch value := view.(type) {
	case ListView:
		return renderJSONList(w, value)
	case ShowView:
		return writeJSON(w, value)
	case SearchView:
		return writeJSON(w, value)
	case ContextView:
		return writeJSON(w, value)
	case LinksView:
		return writeJSON(w, value)
	case BacklinksView:
		return writeJSON(w, value)
	case GraphView:
		return writeJSON(w, value)
	case CheckView:
		return writeJSON(w, value)
	default:
		return unsupportedView(view)
	}
}

func renderJSONList(w io.Writer, view ListView) error {
	directories := make([]string, 0, len(view.Directories))
	for _, directory := range view.Directories {
		directories = append(directories, directory.Path)
	}
	return writeJSON(w, map[string]any{
		"root":        view.Root,
		"path":        view.Path,
		"recursive":   view.Recursive,
		"directories": directories,
		"entries":     view.Entries,
	})
}
