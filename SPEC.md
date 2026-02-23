You are Opencode. Implement a production-quality, single-binary Go CLI application called `gtd-cli` that provides a complete GTD (Getting Things Done) platform with a JSON-first interface, extensive `--help`, pluggable storage backends, and an architecture designed to support multi-user later (but single-user now).

DO NOT ask the human for more clarifications unless you truly cannot proceed. Make reasonable defaults consistent with this prompt. Follow TDD for all feature development: write tests first, then implement. Ensure deterministic tests.

================================================================================
1) PRODUCT GOALS (NON-NEGOTIABLE)
================================================================================
- Single binary CLI: `gtd-cli` (this is the executable name and the root command).
- Primary interface is the CLI; future API server should be feasible without redesign.
- Output is JSON for ALL commands (including errors). No incidental plain text output.
- `--help` must be extensive at every level: root, subcommands, examples, flag descriptions.
- Storage backends are pluggable:
  - Backend 1: SQLite (primary, default) as the source of truth.
  - Backend 2: JSON files backend (implemented now), designed for sync via tools like Syncthing.
- IDs are stable across backends and exports: use ULIDs and string prefixes (tsk_, prj_, ctx_, area_, rev_).
- Heavy scripting expected: stable schemas, automation-safe, NDJSON option.
- Scale target: ~500 tasks (optimize for correctness + ergonomics over extreme performance).
- Enforce GTD “rules” via configurable policies (Option C: configurable). Provide defaults now.
- Review sessions: weekly/daily/monthly start should create persisted review session objects.

================================================================================
2) CORE GTD ARTIFACTS REQUIRED (IMPLEMENT ALL)
================================================================================
Implement support for:
- Inbox capture & processing
- Projects
- Next Actions
- Contexts
- Areas of Focus
- Someday/Maybe
- Waiting For
- Ticklers
- Reference
- Daily review
- Weekly review
- Monthly review
- “What’s next?” by project or context
- Task completion
- Task move between projects (and inbox -> project)

IMPORTANT: Model most of these as Task “status / type” plus optional fields rather than separate entities (except Projects, Contexts, Areas, Reviews). See Data Model below.

================================================================================
3) CLI COMMANDS (MINIMUM REQUIRED SET)
================================================================================
Implement the following commands and flags. Use a git-style subcommand tree.

Global flags (supported by all commands):
- --config <path> (default: ~/.config/gtd-cli/config.yaml)
- --profile <name> (default: "default")
- --backend <sqlite|json> (optional override; otherwise from config)
- --db <path> (sqlite convenience override; otherwise from config)
- --data-dir <path> (json backend root; otherwise from config)
- --format <json|ndjson> (default json)
- --pretty (pretty-print JSON; default false)
- --quiet (suppress non-JSON logs; default true behavior is already JSON-only)
- --help

Command tree (must implement):

A) Inbox / capture
- gtd-cli inbox add "<title>" [--note "..."] [--source <cli|sms|email|other>] [--context @calls ... optional] [--project <id> optional]
  - Default behavior: creates a task with status="inbox" and no project unless --project provided.
- gtd-cli list inbox
  - Lists tasks currently in inbox status.
- gtd-cli inbox process --task <task-id> --to-project <project-id> [--as <next|waiting|someday|tickler|reference>] [--context @x repeatable] [--waiting-for "..."] [--tickle <date>]
  - Moves task out of inbox into specified project and sets status based on --as (default "next").
  - This command is non-interactive by default. (Interactive mode may come later, do not implement now.)

B) Tasks
- gtd-cli task show <task-id>
- gtd-cli task list [--project <project-id>] [--context <@context>] [--status <inbox|next|waiting|someday|tickler|reference|done>] [--limit N] [--offset N]
- gtd-cli task move <task-id> --to-project <project-id>
  - If task was inbox, it becomes status="next" unless configured otherwise.
- gtd-cli task complete <task-id> [--at <timestamp optional>]
- gtd-cli task reopen <task-id> (optional but recommended; implement if easy)
- gtd-cli task next (--project <project-id> | --context <@context>) [--limit N]
  - Returns “what’s next” items ordered by priority heuristics (see ordering below).

C) Projects
- gtd-cli project create "<name>" [--note "..."] [--area <area-id optional>]
- gtd-cli project list [--status <active|someday|done|archived>]
- gtd-cli project show <project-id>
- gtd-cli project archive <project-id>

D) Contexts
- gtd-cli context add "@calls"
- gtd-cli context list
- gtd-cli context rename <context-id> --name "@new"
- gtd-cli context delete <context-id>

E) Areas of focus
- gtd-cli area add "Health"
- gtd-cli area list
- gtd-cli area rename <area-id> --name "..."
- gtd-cli area delete <area-id>

F) Ticklers (as tasks with tickle_at)
- gtd-cli tickler add "<title>" --date <YYYY-MM-DD> [--note "..."] [--project <id optional>]
- gtd-cli tickler list [--from <date>] [--to <date>] [--limit N]

G) Reference (as tasks with status="reference" and optional path/uri)
- gtd-cli reference add --title "<title>" [--path "<path-or-uri>"] [--note "..."]
- gtd-cli reference list [--limit N]

H) Reviews (persisted sessions)
- gtd-cli daily-review start
- gtd-cli weekly-review start
- gtd-cli monthly-review start
Each start command:
  - Creates a review session record (type, started_at).
  - Returns a JSON “review plan” with checklist + computed items needing attention.

Optional but recommended:
- gtd-cli <daily|weekly|monthly>-review complete <review-id> [--note "..."]
- gtd-cli review list [--type daily|weekly|monthly] [--limit N]
If time permits, implement these. If not, at least implement `start`.

================================================================================
4) OUTPUT JSON ENVELOPE (MANDATORY)
================================================================================
All commands return:
{
  "ok": true|false,
  "command": "gtd-cli task list",
  "data": <object|null>,
  "error": <object|null>,
  "meta": {
    "timestamp": "RFC3339 with timezone",
    "backend": "sqlite|json",
    "profile": "default",
    "version": "semver string"
  }
}

Error object:
{
  "code": "NOT_FOUND|VALIDATION|CONFLICT|IO|INTERNAL",
  "message": "human-readable",
  "details": { ... }
}

List responses:
data: {
  "items": [ ... ],
  "count": N,
  "limit": N,
  "offset": N
}

NDJSON mode:
- Print one JSON object per line.
- For list commands, output each item as its own line (just the item object, not envelope) OR output envelopes per line. Choose ONE approach and document it in help. Recommended:
  - NDJSON outputs items only for list commands.
  - Non-list commands output single envelope line.

================================================================================
5) DATA MODEL (DOMAIN)
================================================================================
Entities:

Task:
- id: string (e.g. "tsk_01HX...")
- title: string
- note: string optional
- status: enum: inbox|next|waiting|someday|tickler|reference|done
- project_id: string nullable
- area_id: string nullable
- context_ids: []string (many-to-many)
- waiting_for: string nullable (used when status=waiting)
- due_at: time nullable (RFC3339)
- start_at: time nullable
- tickle_at: date/time nullable (ticklers)
- completed_at: time nullable
- source: string nullable (cli|sms|email|other)
- created_at, updated_at: time

Project:
- id: "prj_..."
- name
- note optional
- status: active|someday|done|archived
- area_id optional
- created_at, updated_at

Context:
- id: "ctx_..."
- name: string (convention: begins with '@', but do not hard enforce)
- created_at, updated_at

Area:
- id: "area_..."
- name
- created_at, updated_at

ReviewSession:
- id: "rev_..."
- type: daily|weekly|monthly
- started_at, ended_at nullable
- note optional
- stats JSON optional (store snapshot counts)
- created_at, updated_at

Ordering heuristics for "task next":
- Primary: status=next only
- Secondary: tasks with due_at soonest first (nulls last)
- Tertiary: created_at oldest first
Document in help.

Configurable GTD enforcement policy (Option C):
- Provide config defaults (see config section).
- Enforcement examples:
  - require_project_when_leaving_inbox: false by default
  - require_context_when_leaving_inbox: false by default
  - auto_next_on_move_from_inbox: true by default
This policy is evaluated by CLI commands that move/process tasks.

================================================================================
6) BACKEND ABSTRACTION (PLUGGABLE)
================================================================================
Implement a `store` package defining interfaces:
- Store interface:
  - Tasks() TaskRepository
  - Projects() ProjectRepository
  - Contexts() ContextRepository
  - Areas() AreaRepository
  - Reviews() ReviewRepository
  - Migrate(ctx) error (where supported)
  - Health(ctx) error
  - Close() error

Repositories should expose methods needed by CLI:
- Create, Get, List (with filters), Update, Delete/Archive, Move, Complete, etc.

Keep domain logic (validation, policy enforcement) in a `core` package, NOT in store.

================================================================================
7) SQLITE BACKEND REQUIREMENTS
================================================================================
- Use SQLite as default backend.
- Implement migrations using a lightweight Go migration approach:
  - Either embed SQL migration files and maintain schema_migrations table.
  - Or use a migration library, but keep dependencies minimal.
- Schema (minimum tables):
  - tasks
  - projects
  - contexts
  - areas
  - task_contexts (join)
  - reviews
  - schema_migrations
- Indexes:
  - tasks(status)
  - tasks(project_id, status)
  - tasks(due_at)
  - task_contexts(context_id)
- Concurrency:
  - Set busy_timeout.
  - Use transactions for writes.
- Tests:
  - Use a temp SQLite file or in-memory, but must support migrations.

================================================================================
8) JSON FILE BACKEND REQUIREMENTS
================================================================================
Goal: human-portable, sync-friendly backend.
Use per-entity files to reduce merge conflicts:
- <data-dir>/
  - tasks/tsk_x.json
  - projects/prj_x.json
  - contexts/ctx_x.json
  - areas/area_x.json
  - reviews/rev_x.json
  - meta/index.json (optional)
Atomic writes:
- Write to temp file then rename.
Locking:
- Use a file lock in data-dir (simple advisory lock) to prevent concurrent corruption.
Query strategy:
- For 500 tasks, scanning entity files is acceptable.
- Optional: maintain lightweight index cache; not required initially.

Tests:
- Use temp dirs; verify atomic write semantics and correctness.

================================================================================
9) CONFIGURATION
================================================================================
Config file: YAML at ~/.config/gtd-cli/config.yaml by default.

Example config (implement parsing + defaults):
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

CLI flags override config.

================================================================================
10) HELP SYSTEM QUALITY
================================================================================
Use Cobra (recommended) or urfave/cli; Cobra is acceptable and common.
For each command:
- Provide:
  - short description
  - long description
  - examples
  - flag documentation
  - output schema note (what `data` contains)

Add `gtd-cli help` that prints:
- Short GTD overview
- Command map
- Example workflows (capture -> process -> do -> review)

IMPORTANT: Even help text should be invoked via `--help` and normal help output can be plain text (help is user-facing). However, for command execution outputs: JSON only.
(If you prefer JSON help, implement `gtd-cli help --json` later; do not do now.)

================================================================================
11) TEST-DRIVEN DEVELOPMENT REQUIREMENTS
================================================================================
TDD RULES:
- For each feature/command:
  1) Write failing tests first.
  2) Implement minimal code to pass.
  3) Refactor.
- Use Go’s testing package.
- Prefer table-driven tests.
- Tests MUST cover:
  - CLI parsing and JSON output envelope
  - Policy enforcement behavior
  - SQLite backend operations
  - JSON backend operations
  - Migration behavior (SQLite)
  - Error codes and messages stable
- Use golden files for JSON outputs where helpful (but keep stable by controlling timestamps in tests; inject clock).

Determinism:
- Use dependency injection for:
  - clock (time provider)
  - ID generator (ULID generator)
  - filesystem interface for JSON backend (optional)
This ensures stable tests.

================================================================================
12) PROJECT STRUCTURE (RECOMMENDED)
================================================================================
/cmd/gtd-cli/               (main)
/internal/
  app/                      (wiring: config, store selection)
  cli/                      (cobra commands)
  core/                     (domain types + validation + policy)
  store/                    (interfaces)
    sqlite/                 (sqlite store)
    json/                   (json files store)
  jsonout/                  (envelope formatting, error mapping)
  config/                   (config load/merge defaults)
  util/                     (clock, ulid, fs helpers, locking)
 /migrations/sqlite/        (embedded migration SQL files)
/testdata/                  (goldens)

No external services required.

================================================================================
13) IMPLEMENTATION PLAN (DO IN THIS ORDER)
================================================================================
1) Skeleton + config + json envelope output + error mapping
   - Tests: config load, envelope render, error render.
2) Store interface + SQLite migration + minimal CRUD for Projects/Tasks
   - Tests: migrations, create/list/get.
3) Implement core domain validation + policy enforcement
   - Tests: policy cases for inbox process/move.
4) Implement required CLI commands for inbox/tasks/projects
   - Tests: CLI command outputs and exit codes.
5) Add contexts/areas and join behavior
6) Implement review start commands creating ReviewSession + generating review plan JSON
7) Implement JSON backend parity for the operations
8) Improve help & examples, add NDJSON output mode
9) Add any optional commands (review complete, list) if time allows

================================================================================
14) REVIEW PLAN CONTENT (FOR daily/weekly/monthly start)
================================================================================
When starting a review, return:
data: {
  "review": {ReviewSession...},
  "checklist": [ { "id": "...", "title": "...", "description": "...", "items_count": N } ... ],
  "attention": {
     "inbox": { "count": N, "items": [<task summaries>] },
     "stale_projects": { "count": N, "items": [<project summaries>] },
     "waiting": { "count": N, "items": [...] },
     "ticklers_due": { "count": N, "items": [...] }
  },
  "stats": { ... snapshot counts ... }
}

Define “stale project” heuristic:
- Active projects with no next actions (status=next) OR last updated > 14 days (choose one, document and test).
For daily review, focus on due soon + inbox count; weekly includes everything; monthly can include someday list summary.

================================================================================
15) LIKELY QUESTIONS FROM YOU (OPENCODE) + ANSWERS
================================================================================
Q1: Should help output be JSON too?
A1: No. Help text can be plain text; actual command execution outputs must be JSON only.

Q2: Should commands ever prompt interactively?
A2: No. Non-interactive by default. Do not implement interactive prompts now.

Q3: How should timestamps be handled in tests?
A3: Inject a clock; in tests, use a fixed time. Ensure JSON outputs stable.

Q4: How do we handle ULIDs deterministically?
A4: Inject an ID generator. In tests, use a deterministic sequence generator.

Q5: Should NDJSON include envelope?
A5: Choose and document a consistent approach. Recommended:
    - list commands in ndjson mode output one item per line (without envelope).
    - non-list commands output a single envelope line.

Q6: How to support multi-user later?
A6: Design schemas and interfaces to allow optional user_id fields later, but do not implement user/auth now. Keep repo interfaces clean so adding user scoping is straightforward.

Q7: What’s the exact behavior when moving inbox task to project?
A7: Default: status becomes "next" unless --as specified, controlled by policy:
    - auto_next_on_move_from_inbox=true
    - auto_next_on_inbox_process=true

Q8: Should contexts require '@' prefix?
A8: Encourage convention but don’t hard enforce. Validate only non-empty.

Q9: What about references and ticklers—separate entities?
A9: Model as tasks with status=reference or tickler plus fields (path optional for reference; tickle_at for tickler).

Q10: What about due dates and start dates?
A10: Include fields in Task model, but CLI flags for them can be added later; implement tickle date for ticklers now.

Q11: Should errors exit non-zero?
A11: Yes. Exit codes:
    - 0 success
    - 2 validation error
    - 3 not found
    - 4 conflict
    - 5 IO
    - 1 internal/unknown
But still emit JSON error envelope to stdout (or stderr). Choose one approach and keep consistent; recommended: JSON to stdout always, logs none.

Q12: How should config file path expansion (~) work?
A12: Expand ~ to home directory. Test this behavior.

Q13: Should we support Windows?
A13: Keep code portable. File paths and locking should be cross-platform where feasible. If locking is OS-specific, choose a well-known Go library or implement a conservative solution.

Q14: Is performance sufficient with scanning JSON files?
A14: Yes for 500 tasks. Ensure correctness and avoid partial writes.

================================================================================
16) ACCEPTANCE CRITERIA
================================================================================
- `gtd-cli --help` shows clear guidance and command map.
- Commands listed above exist and work with both backends.
- All outputs for command execution are valid JSON and match stable schemas.
- SQLite backend uses migrations and passes tests.
- JSON backend uses per-entity files, atomic writes, and locking, passes tests.
- Policy enforcement works and is configurable via config file.
- Reviews create persisted sessions and return review plan JSON.
- Full test suite passes reliably.

Deliver:
- Source code with clean structure.
- README with usage examples + config examples.
- Makefile (optional) with `test`, `lint` (optional), `build`.

Start by creating the repo skeleton and implementing feature set in the order described with TDD.
