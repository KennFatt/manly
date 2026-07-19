# `manly` Recipe Book

Complete usage patterns for every command. Start at the top or jump to the command you need.

---

## init

Initialize an OKF bundle so `manly` has a home for your concepts. A configured root may also be a workspace containing multiple bundles.

```
manly init [--force]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Overwrite an existing `index.md` without error |

The default bundle location is `~/.okf`. You can change it by setting `MANLY_ROOT` or passing `--root` before any subcommand:

```
manly --root ~/my-knowledge init
export MANLY_ROOT=~/my-knowledge
manly init
```

### Persistent configuration

`manly` creates `$HOME/.config/manly/config.yml` on first startup. Its built-in contents are:

```yaml
root: ~/.okf

defaults:
  format: compact
  list:
    recursive: false

display:
  actions: true
  usage: true
```

The precedence for the bundle root is `--root > MANLY_ROOT > config.root > ~/.okf`. The configured format applies to `list`, `show`, `search`, `context`, `links`, `backlinks`, `graph`, and `check`; each command's `--format` flag overrides it. `--recursive` and `--no-recursive` override `defaults.list.recursive`. `display.actions` controls list/show action data and human-readable list/show action commands, while `display.usage` controls list usage hints and human-readable show action commands. Existing configuration files are read without rewriting and invalid configuration fails startup.

**Examples**

```bash
# First time setup with the default location
manly init

# Set up a custom bundle in your project
manly --root ./team-docs init

# Re-initialize an existing root index (destroys current index.md)
manly init --force
```

**What it does**

Creates the root directory as needed, with a minimal `index.md` containing YAML frontmatter (`okf_version: "0.1"` and `type: Bundle`) and a `# Knowledge Bundle` heading. The root must be writable.

---

## bundle and workspace roots

`manly` accepts either a single bundle root or a workspace root. A workspace contains independent bundles as direct child directories:

```text
knowledge/
├── engineering-preferences/
│   ├── index.md              # okf_version and type: Bundle
│   └── react/handler-naming.md
└── personal-notes/
    ├── index.md              # okf_version and type: Bundle
    └── meetings.md
```

A single-bundle root uses local concept IDs:

```text
/programming/type-safety
```

Workspace concept commands require the direct child bundle name:

```text
/engineering-preferences/react/handler-naming
```

`list`, `search`, and `check` aggregate across all discovered bundles. Concept-specific commands such as `show`, `context`, `links`, `backlinks`, `graph`, `add`, `edit`, and `move` use the explicit workspace-qualified path. Markdown links remain bundle-local, so a link inside `engineering-preferences` should use `/react/handler-naming.md`, not `/engineering-preferences/react/handler-naming.md`.

A discovered workspace bundle must have parseable frontmatter in its root `index.md` with a non-empty `okf_version` and `type: Bundle`. Discovery is limited to direct child directories.

---

## list

List concept-containing directories or concepts in a bundle or workspace. This is your primary browsing command.

```
manly list [path] [--recursive|--no-recursive] [--format FORMAT]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--recursive` / `--no-recursive` | bool | false (configurable) | Include or exclude all nested concepts; overrides the configured recursive default |
| `--format` | string | compact (configurable) | Output format: `compact`, `fancy`, `json`, or `markdown` |

If no path is given, `/` (the root) is used.

**Examples**

```bash
# List all top-level concept-containing directories or workspace bundles
manly list

# List concepts inside /programming in a single bundle
manly list /programming

# List concepts inside one workspace bundle
manly list /engineering-preferences/react

# List every concept in the entire bundle or workspace
manly list --recursive

# List concepts in a directory and all subdirectories
manly list /programming --recursive

# JSON output for scripts or agents
manly list --format json
manly list /programming --recursive --format json

# Fancy output with descriptions and open commands
manly list --format fancy

# Markdown output for embedding in another document
manly list --format markdown
```

**What each format shows**

- **compact**: Table with `PATH` and `TITLE / CONCEPTS` columns for non-recursive lists, or `ID` and `TITLE` columns for recursive lists. It also includes a footer hint (`List: manly list <PATH>` for non-recursive, `Details: manly show <ID>` for recursive) and the resolved root path.
- **fancy**: Directory listing with concept counts, or per-concept entries with a description when present and an `Open:` action. Suitable for interactive browsing.
- **json**: Structured object with `root`, `path`, `recursive`, `directories`, and `entries`. Each entry includes the concept metadata exposed by `manly`.
- **markdown**: Bullet list with Markdown links. Suitable for pasting into a note.

---

## show

Read one or more concepts. A directory argument loads all concepts beneath that directory recursively. In workspace mode, concept and directory arguments must include the bundle name. Each result includes its ID, title, content, links, backlinks, and available actions where supported.

```
manly show <concept-id-or-directory>... [--format FORMAT]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | compact | Output format: `compact`, `fancy`, `json`, or `markdown` |

**Examples**

```bash
# Read a concept in compact format (ID + body)
manly show /programming/go/type-safety

# Read a workspace-qualified concept
manly show /engineering-preferences/react/handler-naming

# Load every concept under a directory, including nested directories
manly show /programming --format compact

# Load multiple concepts in one command
manly show /programming/go/type-safety /programming/go/concurrency

# Fancy format includes links, backlinks, and navigation actions
manly show /programming/go/type-safety --format fancy

# JSON for programmatic consumption
manly show /programming/go/type-safety --format json

# Markdown rendering
manly show /programming/go/type-safety --format markdown
```

**What each format shows**

- **compact**: For one concept, the ID is followed by its trimmed body. Collections repeat this block for each concept.
- **fancy**: Each concept includes full content, numbered links with navigation commands, backlinks, and an `Actions:` section with `Open`, `Context`, `Edit`, and `Backlinks` commands.
- **json**: A single concept returns the existing concept object. Collections return a `results` array with each concept's metadata fields, content, links, backlinks, and actions. Arbitrary unrecognized frontmatter fields are not included.
- **markdown**: A single concept is rendered as one heading and body; collections contain one heading and body section per concept.

---

## search

Search concepts by title, description, tags, path, and body text. In workspace mode, searches aggregate across bundles; use a bundle-qualified `--path` to restrict results to one bundle.

```
manly search <query> [--tag TAG] [--type TYPE] [--path PATH] [--limit N] [--format FORMAT]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag` | string | (none) | Filter results to concepts with this tag |
| `--type` | string | (none) | Filter results to concepts of this type |
| `--path` | string | (none) | Restrict search to concept IDs matching this path prefix |
| `--limit` | int | 10 | Maximum number of results |
| `--format` | string | compact | Output format: `compact`, `fancy`, `json`, or `markdown` |

**Examples**

```bash
# Basic full-text search
manly search "type safety"

# Search with a tag filter
manly search "testing" --tag react

# Search within a specific path
manly search "deployment" --path /engineering

# Search within one workspace bundle
manly search "handlers" --path /engineering-preferences/react

# Search by concept type
manly search "architecture" --type "Architecture Decision"

# Combine filters
manly search "error handling" --tag typescript --path /programming --limit 5

# JSON output for agent consumption
manly search "database" --limit 3 --format json

# Fancy output with descriptions and open commands
manly search "memoization" --format fancy
```

**Output formats**

- **compact**: Tab-separated `<score>\t<id>\t<title>` lines.
- **fancy**: Numbered results with title, ID, a description when present, and `Open:`/`Context:` commands.
- **json**: Structured object with `query` and `results` array (each result has the concept metadata exposed by `manly`, its score, and actions).
- **markdown**: Bullet list with Markdown links and descriptions.

---

## context

Retrieve bounded context for an LLM agent or a focused reading session. The JSON format includes concept bodies, links, scores, and navigation actions.

```
manly context <query-or-concept-id> [--limit N] [--format FORMAT]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | 5 | Maximum number of concepts to return |
| `--format` | string | compact | Output format: `compact`, `fancy`, `json`, or `markdown` |

If the argument resolves to an existing concept ID, that concept is returned directly without searching. Otherwise, a full-text search runs with the given limit. In workspace mode, direct concept retrieval requires a bundle-qualified ID; search and context results are qualified before rendering.

**Examples**

```bash
# Search-based context for an agent query
manly context "How should data from an external API be handled?" --limit 5

# Direct concept retrieval (exact ID match, no search)
manly context /programming/type-safety

# JSON context for LLM tool consumption
manly context "error handling patterns" --format json

# Narrow the results
manly context "testing" --limit 3 --format fancy
```

**Output formats**

- **compact**: Blank-line-separated concept ID + body blocks.
- **fancy**: Rich output with title, body, and `Open:` action per result.
- **json**: Structured object with `query` and `results` array. Each result includes concept metadata, score, links, and actions.
- **markdown**: `## Title` sections with body content.

**Agent workflow pattern**

```
manly search "relevant concept"           # find candidates
manly context "specific question" --format json  # get bounded context
manly show /matched/concept-id --format json     # dive deeper
```

---

## links

Show outgoing links from a concept. Internal concept links include navigation commands in fancy output.

```
manly links <concept-id> [--format FORMAT]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | compact | Output format: `compact`, `fancy`, `json`, or `markdown` |

**Examples**

```bash
# Compact tab-separated output
manly links /engineering/architecture-decisions

# Fancy output with navigation commands
manly links /engineering/architecture-decisions --format fancy

# JSON for programmatic link analysis
manly links /engineering/architecture-decisions --format json
```

**Output formats**

- **compact**: Tab-separated `<label>\t<target>` lines.
- **fancy**: Numbered links with label, target, and navigation command (`manly show <id>` for internal, raw URL for external).
- **json**: Object with `source` and `links` array. Link objects may include `label`, `target`, `target_path`, `url`, `broken`, and `external`; empty fields are omitted.
- **markdown**: Bullet list with Markdown links.

---

## backlinks

Show concepts that link to a target concept. Useful for finding related notes and assessing the impact of a change.

```
manly backlinks <concept-id> [--format FORMAT]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | compact | Output format: `compact`, `fancy`, `json`, or `markdown` |

**Examples**

```bash
# Who links to this concept?
manly backlinks /programming/type-safety

# Fancy with titles and open commands
manly backlinks /programming/type-safety --format fancy

# JSON for dependency analysis
manly backlinks /programming/type-safety --format json
```

**Output formats**

- **compact**: Tab-separated `<target>\t<label>` lines.
- **fancy**: Numbered backlinks with title, target, label, and `manly show <id>` command.
- **json**: Object with `target` and `backlinks` array.
- **markdown**: Bullet list with Markdown links.

---

## graph

Traverse linked concepts to a bounded depth. Handles cycles safely.

```
manly graph <concept-id> [--depth N] [--format FORMAT]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--depth` | int | 1 | Maximum traversal depth from the starting concept |
| `--format` | string | compact | Output format: `compact`, `fancy`, `json`, or `markdown` |

**Examples**

```bash
# Immediate neighbors of a concept
manly graph /programming/type-safety

# Traverse two levels deep
manly graph /programming/type-safety --depth 2

# Deep graph exploration
manly graph /engineering/architecture-decisions --depth 3 --format fancy

# JSON for visualization tools
manly graph /programming/type-safety --depth 2 --format json
```

**Output formats**

- **compact**: Tab-separated `<depth>\t<id>\t<title>` lines.
- **fancy**: Indented tree with depth-based spacing (`<id>  <title>`).
- **json**: Object with `nodes` array (each node has `id`, `title`, `depth`).
- **markdown**: Bullet list with depth annotation.

---

## add

Create a new concept with YAML frontmatter. The concept ID determines both its path and identity.

```
manly add <concept-id> --type TYPE [--title TITLE] [--description TEXT] [--tag TAG,...] [--force]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type` | string | (required) | Concept type (e.g. "Note", "Guide", "Architecture Decision") |
| `--title` | string | derived | Concept title; defaults to the last path segment with hyphens replaced by spaces |
| `--description` | string | (none) | Description text; one sentence is recommended but not enforced |
| `--tag` | string | (none) | Comma-separated tags |
| `--force` | bool | false | Overwrite an existing concept |

**Examples**

```bash
# Minimal concept (only --type is required)
manly add /notes/quick-idea --type Note

# Full concept with all metadata
manly add /programming/go/concurrency \
  --type "Study Note" \
  --title "Go Concurrency Patterns" \
  --description "Goroutines, channels, and sync primitives with examples." \
  --tag go,concurrency,programming

# Overwrite an existing concept (replaces the file)
manly add /notes/meeting-notes --type Note --force

# Multiple tags separated by commas
manly add /reference/books --type Reference --tag books,reading,reference
```

**What it does**

Creates a Markdown file at `<root>/<concept-id>.md` with YAML frontmatter containing `type` and `title`, plus optional `description` and `tags`. The generated body contains a `# <title>` heading and the description when provided. Prints the created concept ID on success.

Existing concepts return an error unless `--force` is used.

---

## edit

Open a concept in your configured text editor.

```
manly edit <concept-id>
```

No flags. The editor is read from the `EDITOR` environment variable.

**Examples**

```bash
# Set your editor
export EDITOR="vim"

# Open a concept for editing
manly edit /programming/go/concurrency

# Use a different editor for one session
EDITOR="code --wait" manly edit /engineering/architecture-decisions
```

**Troubleshooting**

If you see `EDITOR is not set`, export the variable or prefix the command:
```
EDITOR="nano" manly edit /notes/quick-idea
```

---

## move

Move a concept and rewrite recognized internal Markdown links that point to it.

```
manly move <old-concept-id> <new-concept-id>
```

No flags.

**Examples**

```bash
# Rename a concept
manly move /programming/go-notes /programming/go/language-notes

# Reorganize into a different directory
manly move /notes/misc /engineering/decisions/misc

# Fix a typo in a concept path
manly move /programing/type-safety /programming/type-safety
```

**What it does**

1. Moves the `.md` file from the old path to the new path.
2. Scans Markdown files in the bundle for links matching the old target and rewrites those links to the new target.
3. Prints the old ID, new ID, and the number of links updated.

Concept IDs are file paths, so moves affect navigation and search. Recognized matching links are updated automatically; review the link update count to confirm the expected scope.

---

## index

Update marked generated sections in `index.md` files. Indexes are optional human-navigation documents — search and concept discovery do not depend on them.

```
manly index [--check]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--check` | bool | false | Report stale generated sections without writing files |

**Examples**

```bash
# Regenerate all marked index sections
manly index

# Check which indexes are stale (dry run)
manly index --check
```

**What it does**

Scans `index.md` files for marked generated-content sections and regenerates their contents based on the current bundle state. Manually authored content outside those markers is preserved. With `--check`, it prints each stale path and exits with a non-zero code if any are out of date — useful in CI or pre-commit hooks.

---

## check

Validate a single OKF bundle or all bundles in a workspace for structural and semantic issues.

```
manly check [--strict] [--format FORMAT]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--strict` | bool | false | Enable advisory checks, including stale generated indexes |
| `--format` | string | compact | Output format: `compact`, `fancy`, `json`, or `markdown` |

**Examples**

```bash
# Basic validation
manly check

# Strict mode with advisory checks
manly check --strict

# JSON output for CI or automation
manly check --format json

# Fancy output with readable issue descriptions and bundle statistics
manly check --strict --format fancy
```

**What it checks**

- YAML frontmatter validity
- Non-empty concept types
- Reserved OKF file constraints (`index.md`, `log.md`)
- Timestamp formats
- Missing local link targets, reported as warnings
- Stale generated indexes when `--strict` is enabled

The command also reports the resolved root, whether it is a `single-bundle` or `workspace` root, discovered bundle count, scanned Markdown files, loaded and invalid concepts, checked and broken links, errors, and warnings.

**Output formats**

- **compact**: Issue lines followed by an aggregate summary containing `Root`, `Mode`, file/concept/link counts, and validation totals.
- **fancy**: Human-readable issues, a validation summary, and per-bundle statistics.
- **json**: Structured `root`, `mode`, `stats`, `bundles`, `Errors`, `Warnings`, and `valid` fields. Existing `Errors` and `Warnings` fields are preserved for compatibility.
- **markdown**: Markdown issue sections, aggregate statistics, and a per-bundle table.

Example compact summary:

```text
Root: /Users/you/knowledge
Mode: workspace
Bundles: 2
Markdown files: 53
Concept files: 45
Loaded concepts: 45
Invalid concept files: 0
Links checked: 99
Broken links: 0
Errors: 0
Warnings: 0
OKF validation passed	0 warning(s)
```

Warnings do not produce a non-zero exit status. Errors do.

---

## version

Print the build-stamped version of the `manly` executable.

```
manly version
```

No flags or arguments.

**Examples**

```bash
# Check which version is installed
manly version
```

**What it shows**

Prints `manly <version>` and exits. The version value depends on how the binary was built:

- **`make build` / `make install`**: Includes the Git-derived version (e.g. `manly v1.2.0` on an exact release tag, `manly 1.1.0-1-g<hash>` on a post-release commit, or `manly dev` if Git metadata is unavailable).
- **Direct `go build`**: Always reports `manly dev` because no linker stamp is applied.
- **Overridden builds**: `VERSION=v1.2.0 make build` uses the supplied value verbatim.

---

## Output Formats

All read commands (`list`, `show`, `search`, `context`, `links`, `backlinks`, `graph`, `check`) accept `--format`.

| Format | Best for |
|--------|---------|
| `compact` | Terminal reading, piping to other CLI tools |
| `fancy` | Interactive browsing, reading concepts with navigation hints |
| `json` | Scripts, editor integrations, LLM agent consumption |
| `markdown` | Embedding output in another Markdown document |

The default is always `compact`.

---

## Agent / LLM Workflows

### Retrieval pipeline

```
manly search "relevant topic" --limit 5 --format json    # 1. Find candidates
manly context "specific question" --limit 3 --format json # 2. Get bounded context
manly show /matched/id --format json                     # 3. Dive deeper
manly links /matched/id --format json                    # 4. Follow relationships
```

### One-shot context retrieval

```
manly context "How should React components handle external data?" --format json
```

By default, returns up to 5 relevant concepts. JSON includes full bodies, links, scores, and navigation actions; other formats render their documented subsets. The agent can continue by calling `manly show` or `manly links` on any returned concept ID.

### Validation guard

```
manly check --format json
```

Run before committing or deploying. A non-zero exit indicates validation errors or an operational command error; warnings alone do not produce a non-zero exit.

---

## Troubleshooting

### `EDITOR is not set`

`manly edit` reads the `EDITOR` environment variable. Set it in your shell profile:

```bash
export EDITOR="vim"
```

Or prefix the command:

```bash
EDITOR="nano" manly edit /concept/id
```

### Bundle or workspace not found / wrong root

`manly` resolves the root as: `--root` flag > `MANLY_ROOT` env var > `~/.okf`.

```bash
# Check which root and mode are active
manly check

# Inspect root and workspace statistics as JSON
manly check --format json | jq '{root, mode, stats}'

# Override for one command
manly --root ./my-bundle list

# Set persistently
export MANLY_ROOT=~/Documents/knowledge
```

If the root is a workspace, concept-specific commands require the direct child bundle name, for example `/engineering-preferences/react/handler-naming`. Each direct child bundle must contain an `index.md` with `okf_version` and `type: Bundle`.

### Concept already exists

Use `--force` with `add` to overwrite, or use `move` to rename:

```bash
manly add /notes/meeting --type Note --force
manly move /notes/meeting /notes/meeting-2024
```

### Broken links

Run `check` to find broken recognized internal links in bundle Markdown files (except `log.md`):

```bash
manly check
# WARNING\tengineering-preferences/react/file.md\tlink target not found: /missing/concept
```

In a workspace, Markdown links remain bundle-local. For a concept inside `engineering-preferences`, use `/react/handler-naming.md`, not `/engineering-preferences/react/handler-naming.md`. Broken links are reported as warnings. OKF v0.1 requires consumers to tolerate them. Use `move` to fix recognized path mismatches or `edit` to correct the Markdown.

### Stale generated indexes

Check without writing:

```bash
manly index --check
```

Regenerate if stale:

```bash
manly index
```

---

## FAQ

### How do I retrieve a specific concept?

For a single bundle:

```bash
manly show /exact/concept-id
```

For a workspace, include the bundle directory name:

```bash
manly show /engineering-preferences/react/handler-naming
```

If you want the full context (links, backlinks, actions):

```bash
manly show /exact/concept-id --format fancy
```

For programmatic use:

```bash
manly show /exact/concept-id --format json
```

### How do I list all concepts?

```bash
manly list --recursive
```

Add `--format json` for structured output.

### How do I list all concepts in a specific directory?

```bash
manly list /programming --recursive
```

Omit `--recursive` to see only immediate children.

### How do I find concepts by tag?

```bash
manly search "any query" --tag react
```

The query must contain a keyword; an empty query returns no results, even when a tag filter is provided.

### How do I find concepts by type?

```bash
manly search "any query" --type "Architecture Decision"
```

Combine with `--path` to narrow further:

```bash
manly search "study" --type "Study Note" --path /courses
```

### How do I find concepts that link to a specific concept?

```bash
manly backlinks /target/concept-id
```

### How do I explore a concept and everything it links to?

```bash
manly graph /starting/concept --depth 2 --format fancy
```

Start with depth 1, increase as needed. The graph avoids cycles.

### How do I get machine-readable output?

Add `--format json` to any read command:

```bash
manly list --recursive --format json | jq '.entries[].concept.id'
manly search "error handling" --format json | jq '.results'
manly check --format json | jq '{root, mode, stats, bundles, errors: .Errors, warnings: .Warnings, valid}'
```

### How do I change a concept's path?

```bash
manly move /old/path /new/path
```

Recognized matching internal links are updated automatically. The command prints how many links were changed.

### How do I create a concept without opening an editor?

```bash
manly add /path/to/concept --type Note --title "My Title" --description "What this is about."
```

The file is created immediately with valid frontmatter. Use `manly edit /path/to/concept` later to add body content.

### How do I validate my bundle or workspace before sharing it?

```bash
manly check --strict
```

Non-zero exit means validation errors occurred; warnings alone do not produce a non-zero exit. Use `--format json` for CI pipelines.

### How do I check if generated indexes need updating?

```bash
manly index --check
```

Exits non-zero if any index section is stale.

### How do I use manly with AI coding agents?

Feed agents structured context:

```bash
manly context "How should I handle errors in TypeScript?" --format json --limit 3
```

Agents get concept bodies, links, and navigation actions. The agent can follow up with `manly show` or `manly links` on any returned ID.

### How do I know which root and mode are active?

```bash
manly check
manly check --format json | jq '{root, mode, stats}'
```

### How do I use a different bundle or workspace for a single command?

```bash
manly --root /path/to/bundle list
```

The `--root` flag must come before the subcommand.
