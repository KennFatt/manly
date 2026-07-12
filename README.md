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

`manly` uses the [Open Knowledge Format (OKF) v0.1](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md). Concepts are Markdown files with YAML frontmatter, identified by root-relative path without `.md`. Links use standard Markdown syntax.

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize an OKF bundle |
| `list` | List directories or concepts |
| `show` | Show concepts or recursively load a concept directory |
| `search` | Search concepts |
| `context` | Retrieve bounded agent context |
| `links` | Show outgoing links |
| `backlinks` | Show incoming links |
| `graph` | Traverse linked concepts |
| `add` | Create a concept |
| `edit` | Open a concept in $EDITOR |
| `move` | Move a concept and update links |
| `index` | Update marked generated index sections |
| `check` | Validate the bundle |

All read commands support `--format compact|fancy|json|markdown` (default: `compact`).

See **[docs/recipe.md](docs/recipe.md)** for complete flag tables, examples, agent workflows, and FAQ.

## Agent retrieval

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
