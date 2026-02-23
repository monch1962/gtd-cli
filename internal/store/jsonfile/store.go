package jsonfile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gofrs/flock"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/store"
	"github.com/anomalyco/gtd-cli/internal/util"
)

type Store struct {
	dataDir string
	idGen   util.IDGenerator
	clock   util.Clock
	lock    *flock.Flock

	tasks    *taskRepo
	projects *projectRepo
	contexts *contextRepo
	areas    *areaRepo
	reviews  *reviewRepo
}

func New(dataDir string, idGen util.IDGenerator, clock util.Clock) (*Store, error) {
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "tasks"),
		filepath.Join(dataDir, "projects"),
		filepath.Join(dataDir, "contexts"),
		filepath.Join(dataDir, "areas"),
		filepath.Join(dataDir, "reviews"),
		filepath.Join(dataDir, "meta"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	lock := flock.New(filepath.Join(dataDir, ".lock"))
	if err := lock.Lock(); err != nil {
		return nil, fmt.Errorf("acquire lock: %w", err)
	}

	s := &Store{
		dataDir: dataDir,
		idGen:   idGen,
		clock:   clock,
		lock:    lock,
	}

	s.tasks = &taskRepo{store: s}
	s.projects = &projectRepo{store: s}
	s.contexts = &contextRepo{store: s}
	s.areas = &areaRepo{store: s}
	s.reviews = &reviewRepo{store: s}

	return s, nil
}

func (s *Store) Tasks() store.TaskRepository       { return s.tasks }
func (s *Store) Projects() store.ProjectRepository { return s.projects }
func (s *Store) Contexts() store.ContextRepository { return s.contexts }
func (s *Store) Areas() store.AreaRepository       { return s.areas }
func (s *Store) Reviews() store.ReviewRepository   { return s.reviews }

func (s *Store) Close() error {
	return s.lock.Unlock()
}

func (s *Store) Migrate(ctx context.Context) error {
	return nil
}

func (s *Store) Health(ctx context.Context) error {
	return nil
}

func (s *Store) writeFile(path string, data any) error {
	tmpPath := path + ".tmp"

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename file: %w", err)
	}

	return nil
}

func (s *Store) readFile(path string, dest any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if err := json.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}

	return nil
}

func (s *Store) listFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

type taskRepo struct {
	store *Store
}

func (r *taskRepo) Create(ctx context.Context, task *core.Task) error {
	path := filepath.Join(r.store.dataDir, "tasks", task.ID+".json")
	return r.store.writeFile(path, task)
}

func (r *taskRepo) Get(ctx context.Context, id string) (*core.Task, error) {
	path := filepath.Join(r.store.dataDir, "tasks", id+".json")
	var task core.Task
	if err := r.store.readFile(path, &task); err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	return &task, nil
}

func (r *taskRepo) List(ctx context.Context, filter core.TaskFilter) (*core.TaskListResult, error) {
	files, err := r.store.listFiles(filepath.Join(r.store.dataDir, "tasks"))
	if err != nil {
		return nil, fmt.Errorf("list task files: %w", err)
	}

	var tasks []core.Task
	for _, f := range files {
		var task core.Task
		if err := r.store.readFile(f, &task); err != nil {
			continue
		}

		if filter.ProjectID != nil && (task.ProjectID == nil || *task.ProjectID != *filter.ProjectID) {
			continue
		}
		if filter.Status != nil && task.Status != *filter.Status {
			continue
		}
		if filter.ContextID != nil && !contains(task.ContextIDs, *filter.ContextID) {
			continue
		}

		tasks = append(tasks, task)
	}

	sortTasks(tasks)

	limit := filter.Limit
	if limit <= 0 {
		limit = core.DefaultLimit
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	if offset > len(tasks) {
		offset = len(tasks)
	}
	end := offset + limit
	if end > len(tasks) {
		end = len(tasks)
	}

	return &core.TaskListResult{
		Items:  tasks[offset:end],
		Count:  len(tasks[offset:end]),
		Limit:  limit,
		Offset: offset,
	}, nil
}

func sortTasks(tasks []core.Task) {
	slices.SortFunc(tasks, func(a, b core.Task) int {
		if a.DueAt == nil && b.DueAt != nil {
			return 1
		}
		if a.DueAt != nil && b.DueAt == nil {
			return -1
		}
		if a.DueAt != nil && b.DueAt != nil {
			if a.DueAt.Before(*b.DueAt) {
				return -1
			}
			if a.DueAt.After(*b.DueAt) {
				return 1
			}
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			if a.CreatedAt.After(b.CreatedAt) {
				return 1
			}
			return 0
		}
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		}
		if a.CreatedAt.After(b.CreatedAt) {
			return 1
		}
		return 0
	})
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (r *taskRepo) Update(ctx context.Context, task *core.Task) error {
	task.UpdatedAt = r.store.clock.Now()
	path := filepath.Join(r.store.dataDir, "tasks", task.ID+".json")
	return r.store.writeFile(path, task)
}

func (r *taskRepo) Delete(ctx context.Context, id string) error {
	path := filepath.Join(r.store.dataDir, "tasks", id+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func (r *taskRepo) MoveToProject(ctx context.Context, taskID, projectID string) error {
	task, err := r.Get(ctx, taskID)
	if err != nil {
		return err
	}
	task.ProjectID = &projectID
	task.UpdatedAt = r.store.clock.Now()
	return r.Update(ctx, task)
}

func (r *taskRepo) Complete(ctx context.Context, taskID string, completedAt string) error {
	task, err := r.Get(ctx, taskID)
	if err != nil {
		return err
	}

	var t time.Time
	if completedAt != "" {
		var err error
		t, err = time.Parse(time.RFC3339, completedAt)
		if err != nil {
			return fmt.Errorf("parse completed_at: %w", err)
		}
	} else {
		t = r.store.clock.Now()
	}

	task.Status = core.TaskStatusDone
	task.CompletedAt = &t
	task.UpdatedAt = r.store.clock.Now()
	return r.Update(ctx, task)
}

func (r *taskRepo) Reopen(ctx context.Context, taskID string) error {
	task, err := r.Get(ctx, taskID)
	if err != nil {
		return err
	}
	task.Status = core.TaskStatusNext
	task.CompletedAt = nil
	task.UpdatedAt = r.store.clock.Now()
	return r.Update(ctx, task)
}

func (r *taskRepo) GetNext(ctx context.Context, projectID, contextID string, limit int) ([]core.Task, error) {
	filter := core.TaskFilter{
		Status:    ptrStatus(core.TaskStatusNext),
		Limit:     limit,
		ProjectID: ptrString(projectID),
		ContextID: ptrString(contextID),
	}
	if projectID == "" {
		filter.ProjectID = nil
	}
	if contextID == "" {
		filter.ContextID = nil
	}

	result, err := r.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func ptrStatus(s core.TaskStatus) *core.TaskStatus {
	return &s
}

func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type projectRepo struct {
	store *Store
}

func (r *projectRepo) Create(ctx context.Context, project *core.Project) error {
	path := filepath.Join(r.store.dataDir, "projects", project.ID+".json")
	return r.store.writeFile(path, project)
}

func (r *projectRepo) Get(ctx context.Context, id string) (*core.Project, error) {
	path := filepath.Join(r.store.dataDir, "projects", id+".json")
	var project core.Project
	if err := r.store.readFile(path, &project); err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	return &project, nil
}

func (r *projectRepo) List(ctx context.Context, filter core.ProjectFilter) ([]core.Project, error) {
	files, err := r.store.listFiles(filepath.Join(r.store.dataDir, "projects"))
	if err != nil {
		return nil, fmt.Errorf("list project files: %w", err)
	}

	var projects []core.Project
	for _, f := range files {
		var p core.Project
		if err := r.store.readFile(f, &p); err != nil {
			continue
		}

		if filter.Status != nil && p.Status != *filter.Status {
			continue
		}

		projects = append(projects, p)
	}

	if filter.Limit > 0 && len(projects) > filter.Limit {
		projects = projects[:filter.Limit]
	}

	return projects, nil
}

func (r *projectRepo) Update(ctx context.Context, project *core.Project) error {
	project.UpdatedAt = r.store.clock.Now()
	path := filepath.Join(r.store.dataDir, "projects", project.ID+".json")
	return r.store.writeFile(path, project)
}

func (r *projectRepo) Delete(ctx context.Context, id string) error {
	path := filepath.Join(r.store.dataDir, "projects", id+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}

func (r *projectRepo) Archive(ctx context.Context, id string) error {
	project, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	project.Status = core.ProjectStatusArchived
	project.UpdatedAt = r.store.clock.Now()
	return r.Update(ctx, project)
}

type contextRepo struct {
	store *Store
}

func (r *contextRepo) Create(ctx context.Context, context_ *core.Context) error {
	path := filepath.Join(r.store.dataDir, "contexts", context_.ID+".json")
	return r.store.writeFile(path, context_)
}

func (r *contextRepo) Get(ctx context.Context, id string) (*core.Context, error) {
	path := filepath.Join(r.store.dataDir, "contexts", id+".json")
	var context_ core.Context
	if err := r.store.readFile(path, &context_); err != nil {
		return nil, fmt.Errorf("get context: %w", err)
	}
	return &context_, nil
}

func (r *contextRepo) List(ctx context.Context) ([]core.Context, error) {
	files, err := r.store.listFiles(filepath.Join(r.store.dataDir, "contexts"))
	if err != nil {
		return nil, fmt.Errorf("list context files: %w", err)
	}

	var contexts []core.Context
	for _, f := range files {
		var c core.Context
		if err := r.store.readFile(f, &c); err != nil {
			continue
		}
		contexts = append(contexts, c)
	}

	return contexts, nil
}

func (r *contextRepo) Update(ctx context.Context, context_ *core.Context) error {
	context_.UpdatedAt = r.store.clock.Now()
	path := filepath.Join(r.store.dataDir, "contexts", context_.ID+".json")
	return r.store.writeFile(path, context_)
}

func (r *contextRepo) Delete(ctx context.Context, id string) error {
	path := filepath.Join(r.store.dataDir, "contexts", id+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete context: %w", err)
	}
	return nil
}

type areaRepo struct {
	store *Store
}

func (r *areaRepo) Create(ctx context.Context, area *core.Area) error {
	path := filepath.Join(r.store.dataDir, "areas", area.ID+".json")
	return r.store.writeFile(path, area)
}

func (r *areaRepo) Get(ctx context.Context, id string) (*core.Area, error) {
	path := filepath.Join(r.store.dataDir, "areas", id+".json")
	var area core.Area
	if err := r.store.readFile(path, &area); err != nil {
		return nil, fmt.Errorf("get area: %w", err)
	}
	return &area, nil
}

func (r *areaRepo) List(ctx context.Context) ([]core.Area, error) {
	files, err := r.store.listFiles(filepath.Join(r.store.dataDir, "areas"))
	if err != nil {
		return nil, fmt.Errorf("list area files: %w", err)
	}

	var areas []core.Area
	for _, f := range files {
		var a core.Area
		if err := r.store.readFile(f, &a); err != nil {
			continue
		}
		areas = append(areas, a)
	}

	return areas, nil
}

func (r *areaRepo) Update(ctx context.Context, area *core.Area) error {
	area.UpdatedAt = r.store.clock.Now()
	path := filepath.Join(r.store.dataDir, "areas", area.ID+".json")
	return r.store.writeFile(path, area)
}

func (r *areaRepo) Delete(ctx context.Context, id string) error {
	path := filepath.Join(r.store.dataDir, "areas", id+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete area: %w", err)
	}
	return nil
}

type reviewRepo struct {
	store *Store
}

func (r *reviewRepo) Create(ctx context.Context, review *core.ReviewSession) error {
	path := filepath.Join(r.store.dataDir, "reviews", review.ID+".json")
	return r.store.writeFile(path, review)
}

func (r *reviewRepo) Get(ctx context.Context, id string) (*core.ReviewSession, error) {
	path := filepath.Join(r.store.dataDir, "reviews", id+".json")
	var review core.ReviewSession
	if err := r.store.readFile(path, &review); err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}
	return &review, nil
}

func (r *reviewRepo) List(ctx context.Context, filter core.ReviewFilter) ([]core.ReviewSession, error) {
	files, err := r.store.listFiles(filepath.Join(r.store.dataDir, "reviews"))
	if err != nil {
		return nil, fmt.Errorf("list review files: %w", err)
	}

	var reviews []core.ReviewSession
	for _, f := range files {
		var rev core.ReviewSession
		if err := r.store.readFile(f, &rev); err != nil {
			continue
		}

		if filter.Type != nil && rev.Type != *filter.Type {
			continue
		}

		reviews = append(reviews, rev)
	}

	if filter.Limit > 0 && len(reviews) > filter.Limit {
		reviews = reviews[:filter.Limit]
	}

	return reviews, nil
}

func (r *reviewRepo) Update(ctx context.Context, review *core.ReviewSession) error {
	review.UpdatedAt = r.store.clock.Now()
	path := filepath.Join(r.store.dataDir, "reviews", review.ID+".json")
	return r.store.writeFile(path, review)
}

func (r *reviewRepo) Complete(ctx context.Context, id string, endedAt string, note string) error {
	review, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	t, err := time.Parse(time.RFC3339, endedAt)
	if err != nil {
		return fmt.Errorf("parse ended_at: %w", err)
	}

	review.EndedAt = &t
	review.Note = note
	review.UpdatedAt = r.store.clock.Now()
	return r.Update(ctx, review)
}
