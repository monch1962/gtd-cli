package store

import (
	"context"

	"github.com/anomalyco/gtd-cli/internal/core"
)

type Store interface {
	Tasks() TaskRepository
	Projects() ProjectRepository
	Contexts() ContextRepository
	Areas() AreaRepository
	Reviews() ReviewRepository
	Migrate(ctx context.Context) error
	Health(ctx context.Context) error
	Close() error
}

type TaskRepository interface {
	Create(ctx context.Context, task *core.Task) error
	Get(ctx context.Context, id string) (*core.Task, error)
	List(ctx context.Context, filter core.TaskFilter) (*core.TaskListResult, error)
	Update(ctx context.Context, task *core.Task) error
	Delete(ctx context.Context, id string) error
	MoveToProject(ctx context.Context, taskID, projectID string) error
	Complete(ctx context.Context, taskID string, completedAt string) error
	Reopen(ctx context.Context, taskID string) error
	GetNext(ctx context.Context, projectID, contextID string, limit int) ([]core.Task, error)
}

type ProjectRepository interface {
	Create(ctx context.Context, project *core.Project) error
	Get(ctx context.Context, id string) (*core.Project, error)
	List(ctx context.Context, filter core.ProjectFilter) ([]core.Project, error)
	Update(ctx context.Context, project *core.Project) error
	Delete(ctx context.Context, id string) error
	Archive(ctx context.Context, id string) error
}

type ContextRepository interface {
	Create(ctx context.Context, context_ *core.Context) error
	Get(ctx context.Context, id string) (*core.Context, error)
	List(ctx context.Context) ([]core.Context, error)
	Update(ctx context.Context, context_ *core.Context) error
	Delete(ctx context.Context, id string) error
}

type AreaRepository interface {
	Create(ctx context.Context, area *core.Area) error
	Get(ctx context.Context, id string) (*core.Area, error)
	List(ctx context.Context) ([]core.Area, error)
	Update(ctx context.Context, area *core.Area) error
	Delete(ctx context.Context, id string) error
}

type ReviewRepository interface {
	Create(ctx context.Context, review *core.ReviewSession) error
	Get(ctx context.Context, id string) (*core.ReviewSession, error)
	List(ctx context.Context, filter core.ReviewFilter) ([]core.ReviewSession, error)
	Update(ctx context.Context, review *core.ReviewSession) error
	Complete(ctx context.Context, id string, endedAt string, note string) error
}
