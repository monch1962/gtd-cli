package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/store"
	"github.com/anomalyco/gtd-cli/internal/util"
)

type Store struct {
	db       *sql.DB
	idGen    util.IDGenerator
	clock    util.Clock
	tasks    *taskRepo
	projects *projectRepo
	contexts *contextRepo
	areas    *areaRepo
	reviews  *reviewRepo
}

func New(path string, idGen util.IDGenerator) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	s := &Store{
		db:    db,
		idGen: idGen,
		clock: util.RealClock{},
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
	return s.db.Close()
}

func (s *Store) Health(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

//go:embed migrations
var migrationsFS embed.FS

func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	var version string
	err := s.db.QueryRow(`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("check migrations: %w", err)
	}

	files, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}
		if f.Name() <= version {
			continue
		}

		data, err := migrationsFS.ReadFile("migrations/" + f.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f.Name(), err)
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		if _, err := tx.Exec(string(data)); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", f.Name(), err)
		}

		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, f.Name()); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", f.Name(), err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", f.Name(), err)
		}
	}

	return nil
}

type taskRepo struct {
	store *Store
}

func (r *taskRepo) Create(ctx context.Context, task *core.Task) error {
	_, err := r.store.db.ExecContext(ctx, `
		INSERT INTO tasks (id, title, note, status, project_id, area_id, waiting_for, due_at, start_at, tickle_at, completed_at, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ID, task.Title, task.Note, task.Status, task.ProjectID, task.AreaID, task.WaitingFor, task.DueAt, task.StartAt, task.TickleAt, task.CompletedAt, task.Source, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	for _, ctxID := range task.ContextIDs {
		_, err := r.store.db.ExecContext(ctx, `INSERT INTO task_contexts (task_id, context_id) VALUES (?, ?)`, task.ID, ctxID)
		if err != nil {
			return fmt.Errorf("insert task_context: %w", err)
		}
	}

	return nil
}

func (r *taskRepo) Get(ctx context.Context, id string) (*core.Task, error) {
	task := &core.Task{}
	err := r.store.db.QueryRowContext(ctx, `
		SELECT id, title, COALESCE(note, ''), status, project_id, area_id, waiting_for, due_at, start_at, tickle_at, completed_at, source, created_at, updated_at
		FROM tasks WHERE id = ?
	`, id).Scan(&task.ID, &task.Title, &task.Note, &task.Status, &task.ProjectID, &task.AreaID, &task.WaitingFor, &task.DueAt, &task.StartAt, &task.TickleAt, &task.CompletedAt, &task.Source, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	rows, err := r.store.db.QueryContext(ctx, `SELECT context_id FROM task_contexts WHERE task_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("get task contexts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ctxID string
		if err := rows.Scan(&ctxID); err != nil {
			return nil, fmt.Errorf("scan context: %w", err)
		}
		task.ContextIDs = append(task.ContextIDs, ctxID)
	}

	return task, nil
}

func (r *taskRepo) List(ctx context.Context, filter core.TaskFilter) (*core.TaskListResult, error) {
	query := `SELECT id, title, COALESCE(note, ''), status, project_id, area_id, waiting_for, due_at, start_at, tickle_at, completed_at, source, created_at, updated_at FROM tasks WHERE 1=1`
	args := []any{}

	if filter.ProjectID != nil {
		query += ` AND project_id = ?`
		args = append(args, *filter.ProjectID)
	}
	if filter.Status != nil {
		query += ` AND status = ?`
		args = append(args, *filter.Status)
	}

	if filter.ContextID != nil {
		query += ` AND id IN (SELECT task_id FROM task_contexts WHERE context_id = ?)`
		args = append(args, *filter.ContextID)
	}

	query += ` ORDER BY due_at ASC NULLS LAST, created_at ASC`

	limit := filter.Limit
	if limit <= 0 {
		limit = core.DefaultLimit
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query += ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []core.Task
	for rows.Next() {
		var task core.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Note, &task.Status, &task.ProjectID, &task.AreaID, &task.WaitingFor, &task.DueAt, &task.StartAt, &task.TickleAt, &task.CompletedAt, &task.Source, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return &core.TaskListResult{
		Items:  tasks,
		Count:  len(tasks),
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (r *taskRepo) Update(ctx context.Context, task *core.Task) error {
	task.UpdatedAt = r.store.clock.Now()
	_, err := r.store.db.ExecContext(ctx, `
		UPDATE tasks SET title = ?, note = ?, status = ?, project_id = ?, area_id = ?, waiting_for = ?, due_at = ?, start_at = ?, tickle_at = ?, completed_at = ?, source = ?, updated_at = ?
		WHERE id = ?
	`, task.Title, task.Note, task.Status, task.ProjectID, task.AreaID, task.WaitingFor, task.DueAt, task.StartAt, task.TickleAt, task.CompletedAt, task.Source, task.UpdatedAt, task.ID)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	_, err = r.store.db.ExecContext(ctx, `DELETE FROM task_contexts WHERE task_id = ?`, task.ID)
	if err != nil {
		return fmt.Errorf("delete task contexts: %w", err)
	}

	for _, ctxID := range task.ContextIDs {
		_, err := r.store.db.ExecContext(ctx, `INSERT INTO task_contexts (task_id, context_id) VALUES (?, ?)`, task.ID, ctxID)
		if err != nil {
			return fmt.Errorf("insert task_context: %w", err)
		}
	}

	return nil
}

func (r *taskRepo) Delete(ctx context.Context, id string) error {
	_, err := r.store.db.ExecContext(ctx, `DELETE FROM task_contexts WHERE task_id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete task contexts: %w", err)
	}
	_, err = r.store.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func (r *taskRepo) MoveToProject(ctx context.Context, taskID, projectID string) error {
	now := r.store.clock.Now()
	_, err := r.store.db.ExecContext(ctx, `
		UPDATE tasks SET project_id = ?, updated_at = ? WHERE id = ?
	`, projectID, now, taskID)
	if err != nil {
		return fmt.Errorf("move task: %w", err)
	}
	return nil
}

func (r *taskRepo) Complete(ctx context.Context, taskID string, completedAt string) error {
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

	_, err := r.store.db.ExecContext(ctx, `
		UPDATE tasks SET status = ?, completed_at = ?, updated_at = ? WHERE id = ?
	`, core.TaskStatusDone, t, r.store.clock.Now(), taskID)
	if err != nil {
		return fmt.Errorf("complete task: %w", err)
	}
	return nil
}

func (r *taskRepo) Reopen(ctx context.Context, taskID string) error {
	_, err := r.store.db.ExecContext(ctx, `
		UPDATE tasks SET status = ?, completed_at = NULL, updated_at = ? WHERE id = ?
	`, core.TaskStatusNext, r.store.clock.Now(), taskID)
	if err != nil {
		return fmt.Errorf("reopen task: %w", err)
	}
	return nil
}

func (r *taskRepo) GetNext(ctx context.Context, projectID, contextID string, limit int) ([]core.Task, error) {
	query := `SELECT id, title, COALESCE(note, ''), status, project_id, area_id, waiting_for, due_at, start_at, tickle_at, completed_at, source, created_at, updated_at
		FROM tasks WHERE status = ?`
	args := []any{core.TaskStatusNext}

	if projectID != "" {
		query += ` AND project_id = ?`
		args = append(args, projectID)
	}
	if contextID != "" {
		query += ` AND id IN (SELECT task_id FROM task_contexts WHERE context_id = ?)`
		args = append(args, contextID)
	}

	query += ` ORDER BY due_at ASC NULLS LAST, created_at ASC`

	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := r.store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get next tasks: %w", err)
	}
	defer rows.Close()

	var tasks []core.Task
	for rows.Next() {
		var task core.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Note, &task.Status, &task.ProjectID, &task.AreaID, &task.WaitingFor, &task.DueAt, &task.StartAt, &task.TickleAt, &task.CompletedAt, &task.Source, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

type projectRepo struct {
	store *Store
}

func (r *projectRepo) Create(ctx context.Context, project *core.Project) error {
	_, err := r.store.db.ExecContext(ctx, `
		INSERT INTO projects (id, name, note, status, area_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, project.ID, project.Name, project.Note, project.Status, project.AreaID, project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert project: %w", err)
	}
	return nil
}

func (r *projectRepo) Get(ctx context.Context, id string) (*core.Project, error) {
	project := &core.Project{}
	err := r.store.db.QueryRowContext(ctx, `
		SELECT id, name, COALESCE(note, ''), status, area_id, created_at, updated_at
		FROM projects WHERE id = ?
	`, id).Scan(&project.ID, &project.Name, &project.Note, &project.Status, &project.AreaID, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	return project, nil
}

func (r *projectRepo) List(ctx context.Context, filter core.ProjectFilter) ([]core.Project, error) {
	query := `SELECT id, name, COALESCE(note, ''), status, area_id, created_at, updated_at FROM projects WHERE 1=1`
	args := []any{}

	if filter.Status != nil {
		query += ` AND status = ?`
		args = append(args, *filter.Status)
	}

	query += ` ORDER BY name ASC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	rows, err := r.store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []core.Project
	for rows.Next() {
		var p core.Project
		err := rows.Scan(&p.ID, &p.Name, &p.Note, &p.Status, &p.AreaID, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (r *projectRepo) Update(ctx context.Context, project *core.Project) error {
	project.UpdatedAt = r.store.clock.Now()
	_, err := r.store.db.ExecContext(ctx, `
		UPDATE projects SET name = ?, note = ?, status = ?, area_id = ?, updated_at = ? WHERE id = ?
	`, project.Name, project.Note, project.Status, project.AreaID, project.UpdatedAt, project.ID)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	return nil
}

func (r *projectRepo) Delete(ctx context.Context, id string) error {
	_, err := r.store.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}

func (r *projectRepo) Archive(ctx context.Context, id string) error {
	_, err := r.store.db.ExecContext(ctx, `
		UPDATE projects SET status = ?, updated_at = ? WHERE id = ?
	`, core.ProjectStatusArchived, r.store.clock.Now(), id)
	if err != nil {
		return fmt.Errorf("archive project: %w", err)
	}
	return nil
}

type contextRepo struct {
	store *Store
}

func (r *contextRepo) Create(ctx context.Context, context_ *core.Context) error {
	_, err := r.store.db.ExecContext(ctx, `
		INSERT INTO contexts (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, context_.ID, context_.Name, context_.CreatedAt, context_.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert context: %w", err)
	}
	return nil
}

func (r *contextRepo) Get(ctx context.Context, id string) (*core.Context, error) {
	context_ := &core.Context{}
	err := r.store.db.QueryRowContext(ctx, `
		SELECT id, name, created_at, updated_at FROM contexts WHERE id = ?
	`, id).Scan(&context_.ID, &context_.Name, &context_.CreatedAt, &context_.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get context: %w", err)
	}
	return context_, nil
}

func (r *contextRepo) List(ctx context.Context) ([]core.Context, error) {
	rows, err := r.store.db.QueryContext(ctx, `SELECT id, name, created_at, updated_at FROM contexts ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list contexts: %w", err)
	}
	defer rows.Close()

	var contexts []core.Context
	for rows.Next() {
		var c core.Context
		err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan context: %w", err)
		}
		contexts = append(contexts, c)
	}

	return contexts, nil
}

func (r *contextRepo) Update(ctx context.Context, context_ *core.Context) error {
	context_.UpdatedAt = r.store.clock.Now()
	_, err := r.store.db.ExecContext(ctx, `
		UPDATE contexts SET name = ?, updated_at = ? WHERE id = ?
	`, context_.Name, context_.UpdatedAt, context_.ID)
	if err != nil {
		return fmt.Errorf("update context: %w", err)
	}
	return nil
}

func (r *contextRepo) Delete(ctx context.Context, id string) error {
	_, err := r.store.db.ExecContext(ctx, `DELETE FROM task_contexts WHERE context_id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete task_contexts: %w", err)
	}
	_, err = r.store.db.ExecContext(ctx, `DELETE FROM contexts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete context: %w", err)
	}
	return nil
}

type areaRepo struct {
	store *Store
}

func (r *areaRepo) Create(ctx context.Context, area *core.Area) error {
	_, err := r.store.db.ExecContext(ctx, `
		INSERT INTO areas (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, area.ID, area.Name, area.CreatedAt, area.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert area: %w", err)
	}
	return nil
}

func (r *areaRepo) Get(ctx context.Context, id string) (*core.Area, error) {
	area := &core.Area{}
	err := r.store.db.QueryRowContext(ctx, `
		SELECT id, name, created_at, updated_at FROM areas WHERE id = ?
	`, id).Scan(&area.ID, &area.Name, &area.CreatedAt, &area.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get area: %w", err)
	}
	return area, nil
}

func (r *areaRepo) List(ctx context.Context) ([]core.Area, error) {
	rows, err := r.store.db.QueryContext(ctx, `SELECT id, name, created_at, updated_at FROM areas ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list areas: %w", err)
	}
	defer rows.Close()

	var areas []core.Area
	for rows.Next() {
		var a core.Area
		err := rows.Scan(&a.ID, &a.Name, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan area: %w", err)
		}
		areas = append(areas, a)
	}

	return areas, nil
}

func (r *areaRepo) Update(ctx context.Context, area *core.Area) error {
	area.UpdatedAt = r.store.clock.Now()
	_, err := r.store.db.ExecContext(ctx, `
		UPDATE areas SET name = ?, updated_at = ? WHERE id = ?
	`, area.Name, area.UpdatedAt, area.ID)
	if err != nil {
		return fmt.Errorf("update area: %w", err)
	}
	return nil
}

func (r *areaRepo) Delete(ctx context.Context, id string) error {
	_, err := r.store.db.ExecContext(ctx, `DELETE FROM areas WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete area: %w", err)
	}
	return nil
}

type reviewRepo struct {
	store *Store
}

func (r *reviewRepo) Create(ctx context.Context, review *core.ReviewSession) error {
	_, err := r.store.db.ExecContext(ctx, `
		INSERT INTO reviews (id, type, started_at, ended_at, note, stats, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, review.ID, review.Type, review.StartedAt, review.EndedAt, review.Note, review.Stats, review.CreatedAt, review.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert review: %w", err)
	}
	return nil
}

func (r *reviewRepo) Get(ctx context.Context, id string) (*core.ReviewSession, error) {
	review := &core.ReviewSession{}
	err := r.store.db.QueryRowContext(ctx, `
		SELECT id, type, started_at, ended_at, COALESCE(note, ''), stats, created_at, updated_at
		FROM reviews WHERE id = ?
	`, id).Scan(&review.ID, &review.Type, &review.StartedAt, &review.EndedAt, &review.Note, &review.Stats, &review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}
	return review, nil
}

func (r *reviewRepo) List(ctx context.Context, filter core.ReviewFilter) ([]core.ReviewSession, error) {
	query := `SELECT id, type, started_at, ended_at, COALESCE(note, ''), stats, created_at, updated_at FROM reviews WHERE 1=1`
	args := []any{}

	if filter.Type != nil {
		query += ` AND type = ?`
		args = append(args, *filter.Type)
	}

	query += ` ORDER BY started_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	rows, err := r.store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	var reviews []core.ReviewSession
	for rows.Next() {
		var r core.ReviewSession
		err := rows.Scan(&r.ID, &r.Type, &r.StartedAt, &r.EndedAt, &r.Note, &r.Stats, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

func (r *reviewRepo) Update(ctx context.Context, review *core.ReviewSession) error {
	review.UpdatedAt = r.store.clock.Now()
	_, err := r.store.db.ExecContext(ctx, `
		UPDATE reviews SET type = ?, started_at = ?, ended_at = ?, note = ?, stats = ?, updated_at = ? WHERE id = ?
	`, review.Type, review.StartedAt, review.EndedAt, review.Note, review.Stats, review.UpdatedAt, review.ID)
	if err != nil {
		return fmt.Errorf("update review: %w", err)
	}
	return nil
}

func (r *reviewRepo) Complete(ctx context.Context, id string, endedAt string, note string) error {
	_, err := r.store.db.ExecContext(ctx, `
		UPDATE reviews SET ended_at = ?, note = ?, updated_at = ? WHERE id = ?
	`, endedAt, note, r.store.clock.Now(), id)
	if err != nil {
		return fmt.Errorf("complete review: %w", err)
	}
	return nil
}
