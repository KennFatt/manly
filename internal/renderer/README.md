# Renderer

`internal/renderer` owns all presentation logic for `manly` command output.

It converts typed presentation views into one of four output formats:

```text
compact   Minimal terminal output; default
fancy     Rich terminal output
json      Structured machine-readable output
markdown  Markdown output
```

The package deliberately does not import `internal/knowledge`. It does not load bundles, search concepts, resolve links, or parse CLI arguments.

## Architecture

```text
cmd/manly
   |
   |  query and transform knowledge data
   v
renderer.View
   |
   |  renderer.New(format)
   v
renderer.Render(writer, view)
   |
   +--> compact
   +--> fancy
   +--> json
   +--> markdown
```

The dependency direction is:

```text
cmd/manly -> internal/knowledge
cmd/manly -> internal/renderer
internal/renderer -> Go standard library
```

## Public API

### Formats

`Format` identifies an output format:

```go
format, err := renderer.ParseFormat(value)
if err != nil {
    return err
}
```

Supported values are defined by:

```go
renderer.FormatCompact
renderer.FormatFancy
renderer.FormatJSON
renderer.FormatMarkdown
```

Invalid values produce an error containing all available formats:

```text
unsupported format "human"; available formats: compact, fancy, json, markdown
```

### Renderer interface

Every format renderer implements:

```go
type Renderer interface {
    Format() Format
    Render(io.Writer, View) error
}
```

Use the factory rather than constructing a concrete renderer directly:

```go
outputRenderer, err := renderer.New(format)
if err != nil {
    return err
}

return outputRenderer.Render(os.Stdout, view)
```

`io.Writer` keeps rendering independent from stdout. It also allows callers to render into a buffer, file, pipe, or test fixture.

### Views

`View` is a sealed interface. The package defines the supported view types in `model.go`:

- `ListView`
- `ShowView`
- `SearchView`
- `ContextView`
- `LinksView`
- `BacklinksView`
- `GraphView`
- `CheckView`

The command layer constructs these views from knowledge-layer objects. Views should contain presentation-ready values such as titles, descriptions, IDs, links, actions, scores, and content.

## Renderer files

| File | Responsibility |
|---|---|
| `renderer.go` | Format parsing, factory, interface, shared helpers |
| `model.go` | Typed presentation views |
| `compact.go` | Minimal line-oriented and table-like output |
| `fancy.go` | Rich terminal output and navigation actions |
| `json.go` | JSON serialization |
| `markdown.go` | Markdown serialization |

Each renderer uses a type switch over the supported `View` types. Unsupported view types return an error instead of producing partial output.

## Current output rules

### Compact

Compact output is the default terminal format.

Recursive lists use an aligned table:

```text
ID                                      TITLE
/general/comments                       Comments
/go/modern-features                     Modern Features

Details: manly show <ID>
```

Non-recursive lists label their mixed directory/concept columns accurately:

```text
PATH                                    TITLE / CONCEPTS
/general/                               12 concepts
/general/comments                       Comments

Details: manly show <ID>
```

Other compact outputs remain line-oriented:

```text
search:  score  ID  title
links:   label  target
graph:   depth  ID  title
```

### Fancy

Fancy output keeps headings, descriptions, numbered links, backlinks, and navigation actions such as:

```text
Open: manly show /concept
Context: manly context /concept
```

### JSON

JSON output is the machine-readable contract. Changes to JSON field names or nesting should be treated as compatibility changes.

### Markdown

Markdown output is intended for embedding in documents. It should remain valid, readable Markdown rather than terminal-oriented output.

## How to adjust existing output

1. Find the format file for the behavior:
   - compact: `compact.go`
   - fancy: `fancy.go`
   - JSON: `json.go`
   - Markdown: `markdown.go`
2. Update only that renderer unless the output contract intentionally changes across formats.
3. Preserve the view model as the source of data. Do not load bundles or query knowledge from a renderer.
4. Keep output directed to the supplied `io.Writer`.
5. Preserve existing machine-readable JSON fields unless the change is deliberate and documented.
6. Update the relevant CLI or process-level tests when output behavior changes.

For example, changing compact list spacing belongs in `compact.go`, not in `cmd/manly/command_list_show.go`.

## How to add a new command output

Suppose a new command needs a `StatsView`.

### 1. Add a typed view

Add the presentation model to `model.go`:

```go
type StatsView struct {
    Concepts int
    Links    int
}

func (StatsView) view() {}
```

The unexported `view` method keeps the `View` interface sealed to this package.

### 2. Add support to every renderer

Add a `StatsView` case to `Render` in:

- `compact.go`
- `fancy.go`
- `json.go`
- `markdown.go`

Each case should call a focused format-specific function:

```go
case StatsView:
    return renderCompactStats(w, value)
```

Do not silently fall back to another format. Every supported command view should have an intentional representation in every format.

### 3. Build the view in `cmd/manly`

The command layer converts knowledge data into `renderer.StatsView` and invokes the existing factory flow:

```go
view := renderer.StatsView{
    Concepts: len(bundle.Concepts),
    Links:    linkCount,
}

outputRenderer, err := renderer.New(format)
if err != nil {
    return err
}
return outputRenderer.Render(os.Stdout, view)
```

### 4. Update documentation and tests

Update:

- CLI usage or README examples when the user-facing contract changes.
- Renderer or CLI tests for each format.
- `docs/v1.1.0-technical-review.md` when the architecture or output contract changes.

## Design rules

- Keep domain logic in `internal/knowledge`.
- Keep presentation logic in this package.
- Do not import `internal/knowledge` here.
- Do not parse flags here.
- Do not write directly to `os.Stdout`; use the supplied writer.
- Do not use a generic untyped payload instead of a typed view.
- Keep JSON output stable.
- Add compile-time interface assertions for renderer implementations:

```go
var _ Renderer = compactRenderer{}
```

- When adding a view, update all four format renderers and their tests.

## Verification

From the repository root:

```bash
make test
go build ./...
```

The current repository has CLI and process-level coverage for the major commands and formats. Dedicated unit test files inside `internal/renderer` should be added when renderer behavior requires exact output or writer-error coverage.
