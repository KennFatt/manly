# manly

A fast, local CLI for turning a folder of Markdown files into a searchable, linked knowledge base.

`manly` brings a Unix-man-page experience to personal knowledge:

```text
man    -> read documentation for a command
manly  -> read and navigate a knowledge concept
```

Markdown remains the source of truth. There is no hosted service, required database, or embedded AI model.

## Why manly?

Important knowledge is often scattered across notes, project folders, chat history, and temporary documents. `manly` gives those notes:

- stable concept IDs based on file paths
- full-text and metadata search
- navigable links between concepts
- backlinks and graph traversal
- structured context for LLM agents
- validation for Markdown and YAML frontmatter
- safe commands for adding, editing, and moving concepts

Useful for:

- university course notes and revision material
- programming explanations and study notes
- software engineering standards and playbooks
- architecture decisions and project knowledge
- debugging procedures and lessons learned
- local knowledge supplied to AI coding agents

## Features

- Plain Markdown files on ordinary filesystem storage
- OKF v0.1-compatible bundles
- Human-readable and JSON output
- Actionable CLI commands in navigation output
- Relative and bundle-root Markdown link resolution
- Outgoing links, backlinks, and bounded graph traversal
- Atomic writes for generated concepts and link updates
- No network access required at runtime

## Installation

Requirements:

- Go 1.25 or newer
- `make` for the convenience targets

Clone and install to `~/.local/bin`:

```bash
git clone https://github.com/KennFatt/manly.git
cd manly
make install
```

Ensure the installation directory is on `PATH`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Install somewhere else:

```bash
make install PREFIX="$HOME/bin"
```

Build a project-local binary:

```bash
make build
./manly --help
```

## Quick start

Initialize a knowledge bundle:

```bash
manly init
```

The default bundle location is `~/.okf`. Use another location with `MANLY_ROOT` or `--root`:

```bash
export MANLY_ROOT="$HOME/knowledge"
manly init
```

Create a concept:

```bash
manly add /programming/learning-go \
  --type Note \
  --title "Learning Go" \
  --description "Working notes and examples from learning Go." \
  --tag programming,go
```

Open it in the configured editor:

```bash
export EDITOR="vim"
manly edit /programming/learning-go
```

Find it later:

```bash
manly search "learning Go"
```

Validate the bundle:

```bash
manly check
```

## Knowledge bundle

A bundle is a directory of Markdown files with YAML frontmatter:

```text
~/.okf/
├── index.md
├── university/
│   ├── algorithms.md
│   └── databases.md
├── programming/
│   ├── debugging.md
│   └── type-safety.md
└── engineering/
    └── architecture-decisions.md
```

A concept ID is the root-relative file path without `.md`:

```text
/programming/type-safety
/engineering/architecture-decisions
```

A concept can link to another concept with standard Markdown:

```markdown
See the [Type Safety](/programming/type-safety.md) concept.
```

The file is the source of truth. `manly` derives search indexes and graph relationships while reading the bundle.

## Specification

`manly` uses the [Open Knowledge Format (OKF) v0.1 specification](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md) as its bundle-format reference. The draft specification defines Markdown documents with YAML frontmatter, concept IDs, reserved `index.md` and `log.md` files, Markdown links, and permissive validation. Consult it when changing bundle structure or compatibility behavior.

## Commands

### `init`

Create the root directory and its initial `index.md`:

```bash
manly init
```

Use `--force` only when intentionally replacing an existing root index.

### `list`

Browse what exists without reading full documents:

```bash
manly list
manly list /programming
manly list /programming --recursive
```

Human output includes concept descriptions and commands for opening each concept:

```text
Programming

  /programming/debugging
      A repeatable process for investigating failures.
      Open: manly show /programming/debugging
```

Use JSON for scripts and agents:

```bash
manly list /programming --format json
```

### `show`

Read one complete concept:

```bash
manly show /programming/debugging
```

The output includes links, backlinks, and next actions:

```text
Actions:
  Open:     manly show /programming/debugging
  Context:  manly context /programming/debugging
  Edit:     manly edit /programming/debugging
  Backlinks: manly backlinks /programming/debugging
```

Use structured output when another tool will consume the result:

```bash
manly show /programming/debugging --format json
```

### `search`

Search titles, descriptions, tags, paths, and document bodies:

```bash
manly search "handling external data"
```

Filter results by metadata or path:

```bash
manly search "testing" --tag react
manly search "deployment" --type Procedure
manly search "error" --path /engineering
manly search "database" --limit 5 --format json
```

### `context`

Retrieve bounded context for an LLM agent or a focused reading session:

```bash
manly context "How should data from an external API be handled?" --limit 5
```

Retrieve one known concept as structured JSON:

```bash
manly context /programming/type-safety --format json
```

Results include content, metadata, scores, links, and navigation actions. Agents can continue navigating with the returned concept IDs:

```bash
manly show /concept/id
manly context /concept/id
manly links /concept/id
```

### `links`

Show outgoing links from a concept:

```bash
manly links /engineering/architecture-decisions
```

Internal links are rendered with their next CLI action, such as:

```text
manly show /programming/type-safety
```

### `backlinks`

Show concepts that link to a target concept:

```bash
manly backlinks /programming/type-safety
```

Backlinks help identify related notes and reveal which documents may be affected by a change.

### `graph`

Traverse connected concepts to a bounded depth:

```bash
manly graph /programming/type-safety --depth 2
```

The graph is calculated from Markdown links and handles cycles safely. It does not introduce a second relationship database.

### `add`

Create a valid concept with YAML frontmatter:

```bash
manly add /university/databases/normalization \
  --type "Study Note" \
  --title "Database Normalization" \
  --description "Definitions and examples from database systems." \
  --tag university,databases
```

Existing concepts are not overwritten unless `--force` is supplied.

### `edit`

Open a concept using `$EDITOR`:

```bash
manly edit /university/databases/normalization
```

### `move`

Move a concept and update known internal Markdown links:

```bash
manly move /programming/go-notes /programming/go/language-notes
```

Concept paths act as IDs, so moves should be deliberate and reviewed.

### `index`

Indexes are optional human navigation documents. Search and listing do not depend on them.

Update sections marked for generated content:

```bash
manly index
```

Check marked sections without writing files:

```bash
manly index --check
```

Manually authored index content is preserved. Index generation currently operates on explicit generated sections.

### `check`

Validate an OKF bundle:

```bash
manly check
```

Checks include:

- YAML frontmatter
- non-empty concept types
- UTF-8 files
- reserved OKF files
- timestamps
- Markdown links
- missing local targets

Broken links are reported as warnings, consistent with OKF v0.1. Use strict advisory checks when reviewing generated index sections:

```bash
manly check --strict
```

Use JSON for automation:

```bash
manly check --format json
```

## Output formats

Commands that return concepts support these formats:

```text
human    Rich output with navigation actions
compact  Minimal line-oriented output
json     Structured concepts, links, and actions
markdown Markdown suitable for another document
```

Human output is optimized for terminal reading. JSON output is designed for shell scripts, editor integrations, and LLM agents.

## Common workflows

### Study notes

```bash
manly add /courses/databases/normalization \
  --type "Study Note" \
  --title "Database Normalization" \
  --description "Definitions, examples, and exam questions." \
  --tag course,databases

manly edit /courses/databases/normalization
manly search "normal forms" --path /courses
manly graph /courses/databases/normalization --depth 2
manly check
```

### Engineering knowledge

```bash
manly add /engineering/decisions/queue-choice \
  --type "Architecture Decision" \
  --title "Queue Choice" \
  --description "Why the service uses this queue implementation." \
  --tag architecture,backend

manly context "How are partial API responses handled?" --format json
manly backlinks /engineering/decisions/queue-choice
manly links /engineering/decisions/queue-choice
manly check
```

### Agent retrieval

```text
Agent task
    |
    v
manly search "relevant concept"
    |
    v
manly context "specific question" --format json
    |
    v
Agent uses linked concepts and actions
```

Knowledge remains readable and reviewable by humans while agents receive bounded, structured context.

## Development

Run tests:

```bash
make test
```

Build and test:

```bash
make check
```

Build a local binary:

```bash
make build
```

Clean the local binary:

```bash
make clean
```

The v1 scope intentionally excludes interactive browsing, terminal hyperlinks, embedded LLM calls, network access, vector databases, and semantic embeddings.
