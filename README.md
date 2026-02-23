# gtd-cli

A command-line tool for managing tasks and projects using the Getting Things Done (GTD) methodology. All commands output JSON for easy scripting and automation.

## Features

- **Pluggable storage backends**: SQLite (default) and JSON files (sync-friendly)
- **Complete GTD workflow**: inbox, projects, tasks, contexts, areas, ticklers, references, reviews
- **JSON output**: All commands return structured JSON envelopes
- **ULID-based IDs**: Stable, sortable IDs with type prefixes (`tsk_`, `prj_`, `ctx_`, `area_`, `rev_`)
- **Configurable policies**: Customize GTD enforcement rules
- **Review sessions**: Daily, weekly, and monthly reviews with checklists

## Installation

```bash
go build -o gtd-cli ./cmd/gtd-cli
```

Or with Make:

```bash
make build
```

## Quick Start

```bash
# Add a task to inbox
gtd-cli inbox add "Call dentist"

# Create a project
gtd-cli project create "Website Redesign"

# Process inbox item into project
gtd-cli inbox process --task tsk_01HXYZ --to-project prj_01ABC

# See what's next
gtd-cli task next --project prj_01ABC

# Start a weekly review
gtd-cli weekly-review start
```

## Commands

### Inbox

```bash
gtd-cli inbox add "<title>" [--note "..."] [--source cli|sms|email|other]
gtd-cli inbox list
gtd-cli inbox process --task <id> --to-project <id> [--as next|waiting|someday|tickler|reference]
```

### Tasks

```bash
gtd-cli task show <id>
gtd-cli task list [--project <id>] [--context @name] [--status <status>] [--limit N]
gtd-cli task move <id> --to-project <id>
gtd-cli task complete <id> [--at <timestamp>]
gtd-cli task reopen <id>
gtd-cli task next (--project <id> | --context @name) [--limit N]
```

### Projects

```bash
gtd-cli project create "<name>" [--note "..."] [--area <id>]
gtd-cli project list [--status active|someday|done|archived]
gtd-cli project show <id>
gtd-cli project archive <id>
```

### Contexts

```bash
gtd-cli context add "@calls"
gtd-cli context list
gtd-cli context rename <id> --name "@new"
gtd-cli context delete <id>
```

### Areas of Focus

```bash
gtd-cli area add "Health"
gtd-cli area list
gtd-cli area rename <id> --name "..."
gtd-cli area delete <id>
```

### Ticklers

```bash
gtd-cli tickler add "<title>" --date YYYY-MM-DD [--note "..."]
gtd-cli tickler list [--from YYYY-MM-DD] [--to YYYY-MM-DD]
```

### Reference

```bash
gtd-cli reference add --title "<title>" [--path "<uri>"] [--note "..."]
gtd-cli reference list
```

### Reviews

```bash
gtd-cli daily-review start
gtd-cli weekly-review start
gtd-cli monthly-review start
```

## Global Flags

```bash
--config <path>      Config file (default: ~/.config/gtd-cli/config.yaml)
--profile <name>     Config profile (default: "default")
--backend <type>     Storage backend: sqlite or json
--db <path>          SQLite database path
--data-dir <path>    JSON backend data directory
--format <type>      Output format: json or ndjson (default: json)
--pretty             Pretty-print JSON output
--quiet              Suppress non-JSON output (default: true)
```

## Configuration

Config file: `~/.config/gtd-cli/config.yaml`

```yaml
profiles:
  default:
    backend: sqlite
    sqlite:
      path: "~/.local/share/gtd-cli/gtd.sqlite"
    json:
      data_dir: "~/.local/share/gtd-cli/jsondb"
    output:
      format: json
      pretty: false
    policy:
      require_project_when_leaving_inbox: false
      require_context_when_leaving_inbox: false
      auto_next_on_move_from_inbox: true
      auto_next_on_inbox_process: true
```

## Output Format

All commands return a JSON envelope:

```json
{
  "ok": true,
  "command": "gtd-cli task list",
  "data": { ... },
  "error": null,
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "backend": "sqlite",
    "profile": "default",
    "version": "1.0.0"
  }
}
```

Error responses:

```json
{
  "ok": false,
  "command": "gtd-cli task show",
  "data": null,
  "error": {
    "code": "NOT_FOUND",
    "message": "task not found",
    "details": {"id": "tsk_01HXYZ"}
  },
  "meta": { ... }
}
```

## Storage Backends

### SQLite (default)

Single database file, ideal for local use. Supports migrations and transactions.

```bash
gtd-cli --backend sqlite --db /path/to/gtd.db inbox add "Task"
```

### JSON Files

Per-entity files, designed for sync via tools like Syncthing or Dropbox.

```bash
gtd-cli --backend json --data-dir /path/to/jsondb inbox add "Task"
```

## Exit Codes

- `0` - Success
- `1` - Internal/unknown error
- `2` - Validation error
- `3` - Not found
- `4` - Conflict
- `5` - I/O error

## Development

```bash
make test       # Run tests
make test-cover # Run tests with coverage
make build      # Build binary
make lint       # Run linter (requires golangci-lint)
```

## License

MIT License - see [LICENSE](LICENSE) for details.
