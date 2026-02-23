package jsonfile

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/util"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	clock := util.NewFixedClock(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
	idGen := util.NewDeterministicIDGenerator()

	store, err := New(dir, idGen, clock)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	return store
}

func TestStore_Health(t *testing.T) {
	store := setupTestStore(t)

	if err := store.Health(context.Background()); err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestProjectRepository_Create(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()
	project := &core.Project{
		ID:        "prj_01HX0000000000000000000000",
		Name:      "Test Project",
		Status:    core.ProjectStatusActive,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := store.Projects().Create(ctx, project); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	path := filepath.Join(store.dataDir, "projects", "prj_01HX0000000000000000000000.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Project file not created at %s", path)
	}

	got, err := store.Projects().Get(ctx, project.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Name != project.Name {
		t.Errorf("Name = %s, want %s", got.Name, project.Name)
	}
}

func TestTaskRepository_Create(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()
	task := &core.Task{
		ID:        "tsk_01HX0000000000000000000000",
		Title:     "Test Task",
		Status:    core.TaskStatusInbox,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := store.Tasks().Create(ctx, task); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	path := filepath.Join(store.dataDir, "tasks", "tsk_01HX0000000000000000000000.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Task file not created at %s", path)
	}

	got, err := store.Tasks().Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Title != task.Title {
		t.Errorf("Title = %s, want %s", got.Title, task.Title)
	}
}

func TestTaskRepository_List(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		task := &core.Task{
			ID:        "tsk_01HX000000000000000000000" + string(rune('0'+i)),
			Title:     "Task",
			Status:    core.TaskStatusInbox,
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		}
		if err := store.Tasks().Create(ctx, task); err != nil {
			t.Fatalf("Create task %d failed: %v", i, err)
		}
	}

	result, err := store.Tasks().List(ctx, core.TaskFilter{Limit: 3})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(result.Items) != 3 {
		t.Errorf("Items count = %d, want 3", len(result.Items))
	}
}

func TestContextRepository_CRUD(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()
	context_ := &core.Context{
		ID:        "ctx_01HX0000000000000000000000",
		Name:      "@calls",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := store.Contexts().Create(ctx, context_); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := store.Contexts().Get(ctx, context_.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Name != context_.Name {
		t.Errorf("Name = %s, want %s", got.Name, context_.Name)
	}

	list, err := store.Contexts().List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("List count = %d, want 1", len(list))
	}
}

func TestAreaRepository_CRUD(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()
	area := &core.Area{
		ID:        "area_01HX0000000000000000000000",
		Name:      "Health",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := store.Areas().Create(ctx, area); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := store.Areas().Get(ctx, area.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Name != area.Name {
		t.Errorf("Name = %s, want %s", got.Name, area.Name)
	}
}

func TestTaskRepository_Complete(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()
	task := &core.Task{
		ID:        "tsk_01HX0000000000000000000000",
		Title:     "Test Task",
		Status:    core.TaskStatusNext,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := store.Tasks().Create(ctx, task); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := store.Tasks().Complete(ctx, task.ID, "2024-01-15T12:00:00Z"); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	got, err := store.Tasks().Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Status != core.TaskStatusDone {
		t.Errorf("Status = %s, want done", got.Status)
	}
}

func TestReviewRepository_Create(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()
	review := &core.ReviewSession{
		ID:        "rev_01HX0000000000000000000000",
		Type:      core.ReviewTypeWeekly,
		StartedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := store.Reviews().Create(ctx, review); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := store.Reviews().Get(ctx, review.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Type != review.Type {
		t.Errorf("Type = %s, want %s", got.Type, review.Type)
	}
}
