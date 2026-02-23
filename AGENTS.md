# AGENTS.md

Guidelines for agentic coding tools working in the gtd-cli repository.

## Project Overview

gtd-cli is a GTD (Getting Things Done) command-line tool written in Go. It outputs JSON for all commands, supports multiple storage backends (SQLite and JSON files), and follows TDD practices.

## Build/Test/Lint Commands

```bash
make build                      # Build the binary
make test                       # Run all tests: go test ./... -v
make lint                       # Run golangci-lint
go test ./internal/core -v -run TestValidateTask  # Run single test
go test ./internal/store/jsonfile/... -v          # Run tests for package
```

## Code Style Guidelines

### Imports

Group in three sections: standard library, external packages, internal packages.

```go
import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/core"
)
```

### Comments

Do NOT add comments unless explicitly requested. Code should be self-documenting.

### Types and Constants

- Use typed strings for enums (TaskStatus, ProjectStatus)
- Define constants for magic values at package level

```go
const (
	DefaultLimit = 100
	DateFormat   = "2006-01-02"
)

type TaskStatus string

const (
	TaskStatusInbox TaskStatus = "inbox"
	TaskStatusNext  TaskStatus = "next"
)
```

### Naming Conventions

- Packages: lowercase, single word (core, store, cli, util)
- Types: PascalCase (Task, Project, TaskFilter)
- Interfaces: noun or noun+er suffix (Store, TaskRepository)
- Functions: PascalCase for exported, camelCase for unexported

### Error Handling

- Define sentinel errors at package level
- Use errors.Is() for checking in tests
- Return errors with fmt.Errorf and %w for wrapping
- Never ignore errors in production code

```go
var ErrEmptyTitle = errors.New("empty title")

if err != nil {
	return fmt.Errorf("create task: %w", err)
}
```

## Testing Conventions

### Structure

- Use table-driven tests with t.Run() for subtests
- Use t.Helper() for test helpers
- Use t.TempDir() for temporary directories

```go
func TestValidateTask(t *testing.T) {
	tests := []struct {
		name    string
		task    *Task
		wantErr error
	}{
		{"valid task", &Task{Title: "Test", Status: TaskStatusInbox}, nil},
		{"empty title", &Task{Title: "", Status: TaskStatusInbox}, ErrEmptyTitle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTask(tt.task)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateTask() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
```

### Determinism

- Inject clock and ID generator for testability
- Use util.FixedClock and util.DeterministicIDGenerator in tests

```go
clock := util.NewFixedClock(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
idGen := util.NewDeterministicIDGenerator()
```

## TDD Workflow

1. Write failing test first
2. Implement minimum code to pass
3. Refactor
4. Run tests to verify

## Project Structure

```
/cmd/gtd-cli/main.go          # Entry point
/internal/
  cli/                        # Cobra commands
  config/                     # Config loading, profiles
  core/                       # Domain types, validation, policy
  jsonout/                    # JSON envelope formatting
  store/                      # Storage interfaces
    interfaces.go             # Repository interfaces
    jsonfile/                 # JSON file backend
    sqlite/                   # SQLite backend
  util/                       # Clock, ID generators, helpers
```

## Key Patterns

### Repository Interface

```go
type TaskRepository interface {
	Create(ctx context.Context, task *core.Task) error
	Get(ctx context.Context, id string) (*core.Task, error)
	List(ctx context.Context, filter core.TaskFilter) (*core.TaskListResult, error)
	Update(ctx context.Context, task *core.Task) error
	Delete(ctx context.Context, id string) error
}
```

### CLI Command Handler

```go
var taskShowCmd = &cobra.Command{
	Use:   "show <task-id>",
	Short: "Show task details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli task show", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		task, err := app.Store.Tasks().Get(context.Background(), args[0])
		if err != nil {
			writeError(cmd, "gtd-cli task show", jsonout.ErrNotFound, "task not found", nil)
			return
		}
		writeSuccess(cmd, "gtd-cli task show", task)
	},
}
```

## Common Gotchas

- Default backend is `json` (defined in config.DefaultBackend)
- IDs use ULIDs with prefixes: tsk_, prj_, ctx_, area_, rev_
- JSON backend uses file locking (gofrs/flock)
- SQLite uses modernc.org/sqlite (pure Go, no CGO)
- Exit codes: 0=success, 2=validation, 3=not found, 4=conflict, 5=IO, 1=internal
