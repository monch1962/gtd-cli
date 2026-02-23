package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/util"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := dir + "/test.db"

	store, err := New(dbPath, util.NewRealIDGenerator(util.RealClock{}))
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	if err := store.Migrate(context.Background()); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return store
}

func TestStore_Migrate(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	if err := store.Health(context.Background()); err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestProjectRepository_Create(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	project := &core.Project{
		ID:     "prj_01HX0000000000000000000000",
		Name:   "Test Project",
		Status: core.ProjectStatusActive,
	}

	err := store.Projects().Create(ctx, project)
	if err != nil {
		t.Fatalf("Create project failed: %v", err)
	}

	got, err := store.Projects().Get(ctx, project.ID)
	if err != nil {
		t.Fatalf("Get project failed: %v", err)
	}

	if got.Name != project.Name {
		t.Errorf("Name = %s, want %s", got.Name, project.Name)
	}
	if got.Status != project.Status {
		t.Errorf("Status = %s, want %s", got.Status, project.Status)
	}
}

func TestTaskRepository_Create(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	task := &core.Task{
		ID:     "tsk_01HX0000000000000000000000",
		Title:  "Test Task",
		Status: core.TaskStatusInbox,
	}

	err := store.Tasks().Create(ctx, task)
	if err != nil {
		t.Fatalf("Create task failed: %v", err)
	}

	got, err := store.Tasks().Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get task failed: %v", err)
	}

	if got.Title != task.Title {
		t.Errorf("Title = %s, want %s", got.Title, task.Title)
	}
	if got.Status != task.Status {
		t.Errorf("Status = %s, want %s", got.Status, task.Status)
	}
}

func TestTaskRepository_List(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		task := &core.Task{
			ID:     "tsk_01HX000000000000000000000" + string(rune('0'+i)),
			Title:  "Task",
			Status: core.TaskStatusInbox,
		}
		if err := store.Tasks().Create(ctx, task); err != nil {
			t.Fatalf("Create task %d failed: %v", i, err)
		}
	}

	result, err := store.Tasks().List(ctx, core.TaskFilter{Limit: 3})
	if err != nil {
		t.Fatalf("List tasks failed: %v", err)
	}

	if len(result.Items) != 3 {
		t.Errorf("Items count = %d, want 3", len(result.Items))
	}
	if result.Count != 3 {
		t.Errorf("Count = %d, want 3", result.Count)
	}
}

func TestContextRepository_CRUD(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	context_ := &core.Context{
		ID:   "ctx_01HX0000000000000000000000",
		Name: "@calls",
	}

	if err := store.Contexts().Create(ctx, context_); err != nil {
		t.Fatalf("Create context failed: %v", err)
	}

	got, err := store.Contexts().Get(ctx, context_.ID)
	if err != nil {
		t.Fatalf("Get context failed: %v", err)
	}

	if got.Name != context_.Name {
		t.Errorf("Name = %s, want %s", got.Name, context_.Name)
	}

	list, err := store.Contexts().List(ctx)
	if err != nil {
		t.Fatalf("List contexts failed: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("List count = %d, want 1", len(list))
	}
}

func TestAreaRepository_CRUD(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	area := &core.Area{
		ID:   "area_01HX0000000000000000000000",
		Name: "Health",
	}

	if err := store.Areas().Create(ctx, area); err != nil {
		t.Fatalf("Create area failed: %v", err)
	}

	got, err := store.Areas().Get(ctx, area.ID)
	if err != nil {
		t.Fatalf("Get area failed: %v", err)
	}

	if got.Name != area.Name {
		t.Errorf("Name = %s, want %s", got.Name, area.Name)
	}
}

func TestTaskRepository_Complete(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	task := &core.Task{
		ID:     "tsk_01HX0000000000000000000000",
		Title:  "Test Task",
		Status: core.TaskStatusNext,
	}

	if err := store.Tasks().Create(ctx, task); err != nil {
		t.Fatalf("Create task failed: %v", err)
	}

	if err := store.Tasks().Complete(ctx, task.ID, "2024-01-15T10:30:00Z"); err != nil {
		t.Fatalf("Complete task failed: %v", err)
	}

	got, err := store.Tasks().Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get task failed: %v", err)
	}

	if got.Status != core.TaskStatusDone {
		t.Errorf("Status = %s, want done", got.Status)
	}
	if got.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestTaskRepository_MoveToProject(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	project := &core.Project{
		ID:     "prj_01HX0000000000000000000000",
		Name:   "Test Project",
		Status: core.ProjectStatusActive,
	}
	if err := store.Projects().Create(ctx, project); err != nil {
		t.Fatalf("Create project failed: %v", err)
	}

	task := &core.Task{
		ID:     "tsk_01HX0000000000000000000000",
		Title:  "Test Task",
		Status: core.TaskStatusInbox,
	}
	if err := store.Tasks().Create(ctx, task); err != nil {
		t.Fatalf("Create task failed: %v", err)
	}

	if err := store.Tasks().MoveToProject(ctx, task.ID, project.ID); err != nil {
		t.Fatalf("MoveToProject failed: %v", err)
	}

	got, err := store.Tasks().Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get task failed: %v", err)
	}

	if got.ProjectID == nil || *got.ProjectID != project.ID {
		t.Errorf("ProjectID = %v, want %s", got.ProjectID, project.ID)
	}
}

func TestReviewRepository_Create(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	startedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	review := &core.ReviewSession{
		ID:        "rev_01HX0000000000000000000000",
		Type:      core.ReviewTypeWeekly,
		StartedAt: startedAt,
	}

	if err := store.Reviews().Create(ctx, review); err != nil {
		t.Fatalf("Create review failed: %v", err)
	}

	got, err := store.Reviews().Get(ctx, review.ID)
	if err != nil {
		t.Fatalf("Get review failed: %v", err)
	}

	if got.Type != review.Type {
		t.Errorf("Type = %s, want %s", got.Type, review.Type)
	}
}
