package renderer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Format identifies an output representation.
type Format string

const (
	FormatCompact  Format = "compact"
	FormatFancy    Format = "fancy"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

// Renderer writes one output format for a typed view.
type Renderer interface {
	Format() Format
	Render(io.Writer, View) error
}

// ParseFormat validates a user-provided output format.
func ParseFormat(value string) (Format, error) {
	format := Format(strings.ToLower(strings.TrimSpace(value)))
	switch format {
	case FormatCompact, FormatFancy, FormatJSON, FormatMarkdown:
		return format, nil
	default:
		return "", unsupportedFormat(format)
	}
}

// New creates the renderer for format.
func New(format Format) (Renderer, error) {
	switch format {
	case FormatCompact:
		return compactRenderer{}, nil
	case FormatFancy:
		return fancyRenderer{}, nil
	case FormatJSON:
		return jsonRenderer{}, nil
	case FormatMarkdown:
		return markdownRenderer{}, nil
	default:
		return nil, unsupportedFormat(format)
	}
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func unsupportedFormat(format Format) error {
	return fmt.Errorf("unsupported format %q; available formats: compact, fancy, json, markdown", format)
}

func unsupportedView(view View) error {
	return fmt.Errorf("renderer %q does not support view %T", viewFormat(view), view)
}

func viewFormat(view View) string {
	if view == nil {
		return "unknown"
	}
	return fmt.Sprintf("%T", view)
}

func linkNavigationCommand(link Link) string {
	if link.Target != "" {
		return "manly show " + link.Target
	}
	if strings.HasSuffix(link.TargetPath, "/index.md") {
		return "manly list /" + strings.TrimSuffix(strings.TrimPrefix(link.TargetPath, "/"), "/index.md")
	}
	return ""
}

func compactLinkTarget(link Link) string {
	if link.Target != "" {
		return link.Target
	}
	if link.TargetPath != "" {
		return link.TargetPath
	}
	if link.URL != "" {
		return link.URL
	}
	return "broken"
}
