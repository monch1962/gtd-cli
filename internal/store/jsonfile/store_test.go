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

func TestStore_Close(t *testing.T) {
	store := setupTestStore(t)

	if err := store.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestStore_Migrate(t *testing.T) {
	store := setupTestStore(t)

	if err := store.Migrate(context.Background()); err != nil {
		t.Errorf("Migrate failed: %v", err)
	}
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

func TestTaskRepository_Update(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()
	task := &core.Task{
		ID:        "tsk_01HX0000000000000000000000",
		Title:     "Original Title",
		Status:    core.TaskStatusInbox,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := store.Tasks().Create(ctx, task); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	task.Title = "Updated Title"
	task.Status = core.TaskStatusNext

	if err := store.Tasks().Update(ctx, task); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := store.Tasks().Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Title != "Updated Title" {
		t.Errorf("Title = %s, want Updated Title", got.Title)
	}
}

func TestTaskRepository_Delete(t *testing.T) {
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

	if err := store.Tasks().Delete(ctx, task.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := store.Tasks().Get(ctx, task.ID)
	if err == nil {
		t.Error("Get should fail for deleted task")
	}
}

func TestTaskRepository_MoveToProject(t *testing.T) {
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
		t.Fatalf("Create project failed: %v", err)
	}

	task := &core.Task{
		ID:        "tsk_01HX0000000000000000000000",
		Title:     "Test Task",
		Status:    core.TaskStatusInbox,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	if err := store.Tasks().Create(ctx, task); err != nil {
		t.Fatalf("Create task failed: %v", err)
	}

	if err := store.Tasks().MoveToProject(ctx, task.ID, project.ID); err != nil {
		t.Fatalf("MoveToProject failed: %v", err)
	}

	got, err := store.Tasks().Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.ProjectID == nil || *got.ProjectID != project.ID {
		t.Errorf("ProjectID = %v, want %s", got.ProjectID, project.ID)
	}
}

func TestTaskRepository_Reopen(t *testing.T) {
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

	if err := store.Tasks().Complete(ctx, task.ID, ""); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	if err := store.Tasks().Reopen(ctx, task.ID); err != nil {
		t.Fatalf("Reopen failed: %v", err)
	}

	got, err := store.Tasks().Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Status != core.TaskStatusNext {
		t.Errorf("Status = %s, want next", got.Status)
	}
	if got.CompletedAt != nil {
		t.Error("CompletedAt should be nil after reopen")
	}
}

func TestTaskRepository_GetNext(t *testing.T) {
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
		t.Fatalf("Create project failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		task := &core.Task{
			ID:        "tsk_01HX000000000000000000000" + string(rune('0'+i)),
			Title:     "Next Task",
			Status:    core.TaskStatusNext,
			ProjectID: &project.ID,
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		}
		if err := store.Tasks().Create(ctx, task); err != nil {
			t.Fatalf("Create task %d failed: %v", i, err)
		}
	}

	inboxTask := &core.Task{
		ID:        "tsk_01HX0000000000000000000003",
		Title:     "Inbox Task",
		Status:    core.TaskStatusInbox,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	if err := store.Tasks().Create(ctx, inboxTask); err != nil {
		t.Fatalf("Create inbox task failed: %v", err)
	}

	tasks, err := store.Tasks().GetNext(ctx, project.ID, "", 10)
	if err != nil {
		t.Fatalf("GetNext failed: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("GetNext count = %d, want 3", len(tasks))
	}

	for _, task := range tasks {
		if task.Status != core.TaskStatusNext {
			t.Errorf("GetNext returned non-next task: %s", task.Status)
		}
	}
}

func TestProjectRepository_Update(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()
	project := &core.Project{
		ID:        "prj_01HX0000000000000000000000",
		Name:      "Original Name",
		Status:    core.ProjectStatusActive,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := store.Projects().Create(ctx, project); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	project.Name = "Updated Name"
	project.Note = "Added note"

	if err := store.Projects().Update(ctx, project); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := store.Projects().Get(ctx, project.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Name != "Updated Name" {
		t.Errorf("Name = %s, want Updated Name", got.Name)
	}
}

func TestProjectRepository_Delete(t *testing.T) {
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

	if err := store.Projects().Delete(ctx, project.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := store.Projects().Get(ctx, project.ID)
	if err == nil {
		t.Error("Get should fail for deleted project")
	}
}

func TestProjectRepository_Archive(t *testing.T) {
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

	if err := store.Projects().Archive(ctx, project.ID); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	got, err := store.Projects().Get(ctx, project.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Status != core.ProjectStatusArchived {
		t.Errorf("Status = %s, want archived", got.Status)
	}
}

func TestContextRepository_Update(t *testing.T) {
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

	context_.Name = "@phone"

	if err := store.Contexts().Update(ctx, context_); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := store.Contexts().Get(ctx, context_.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Name != "@phone" {
		t.Errorf("Name = %s, want @phone", got.Name)
	}
}

func TestContextRepository_Delete(t *testing.T) {
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

	if err := store.Contexts().Delete(ctx, context_.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := store.Contexts().Get(ctx, context_.ID)
	if err == nil {
		t.Error("Get should fail for deleted context")
	}
}

func TestAreaRepository_Update(t *testing.T) {
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

	area.Name = "Wellness"

	if err := store.Areas().Update(ctx, area); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := store.Areas().Get(ctx, area.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Name != "Wellness" {
		t.Errorf("Name = %s, want Wellness", got.Name)
	}
}

func TestAreaRepository_Delete(t *testing.T) {
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

	if err := store.Areas().Delete(ctx, area.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := store.Areas().Get(ctx, area.ID)
	if err == nil {
		t.Error("Get should fail for deleted area")
	}
}

func TestReviewRepository_List(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()

	for i := 0; i < 2; i++ {
		review := &core.ReviewSession{
			ID:        "rev_01HX000000000000000000000" + string(rune('0'+i)),
			Type:      core.ReviewTypeWeekly,
			StartedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		}
		if err := store.Reviews().Create(ctx, review); err != nil {
			t.Fatalf("Create review %d failed: %v", i, err)
		}
	}

	daily := &core.ReviewSession{
		ID:        "rev_01HX0000000000000000000002",
		Type:      core.ReviewTypeDaily,
		StartedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	if err := store.Reviews().Create(ctx, daily); err != nil {
		t.Fatalf("Create daily review failed: %v", err)
	}

	weekly := core.ReviewTypeWeekly
	reviews, err := store.Reviews().List(ctx, core.ReviewFilter{Type: &weekly})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(reviews) != 2 {
		t.Errorf("Weekly reviews count = %d, want 2", len(reviews))
	}
}

func TestReviewRepository_Update(t *testing.T) {
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

	review.Note = "Updated note"

	if err := store.Reviews().Update(ctx, review); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := store.Reviews().Get(ctx, review.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Note != "Updated note" {
		t.Errorf("Note = %s, want Updated note", got.Note)
	}
}

func TestReviewRepository_Complete(t *testing.T) {
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

	if err := store.Reviews().Complete(ctx, review.ID, "2024-01-15T11:00:00Z", "All done"); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	got, err := store.Reviews().Get(ctx, review.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.EndedAt == nil {
		t.Error("EndedAt should not be nil")
	}
	if got.Note != "All done" {
		t.Errorf("Note = %s, want All done", got.Note)
	}
}

func TestTaskRepository_ListWithFilters(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()

	project := &core.Project{
		ID:        "prj_01HX0000000000000000000000",
		Name:      "Project A",
		Status:    core.ProjectStatusActive,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	if err := store.Projects().Create(ctx, project); err != nil {
		t.Fatalf("Create project failed: %v", err)
	}

	context_ := &core.Context{
		ID:        "ctx_01HX0000000000000000000000",
		Name:      "@calls",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	if err := store.Contexts().Create(ctx, context_); err != nil {
		t.Fatalf("Create context failed: %v", err)
	}

	task := &core.Task{
		ID:         "tsk_01HX0000000000000000000000",
		Title:      "Task with project and context",
		Status:     core.TaskStatusNext,
		ProjectID:  &project.ID,
		ContextIDs: []string{context_.ID},
		CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	if err := store.Tasks().Create(ctx, task); err != nil {
		t.Fatalf("Create task failed: %v", err)
	}

	inboxTask := &core.Task{
		ID:        "tsk_01HX0000000000000000000001",
		Title:     "Inbox task",
		Status:    core.TaskStatusInbox,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	if err := store.Tasks().Create(ctx, inboxTask); err != nil {
		t.Fatalf("Create inbox task failed: %v", err)
	}

	t.Run("filter by project", func(t *testing.T) {
		result, err := store.Tasks().List(ctx, core.TaskFilter{ProjectID: &project.ID})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(result.Items) != 1 {
			t.Errorf("Expected 1 task, got %d", len(result.Items))
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		status := core.TaskStatusInbox
		result, err := store.Tasks().List(ctx, core.TaskFilter{Status: &status})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(result.Items) != 1 {
			t.Errorf("Expected 1 inbox task, got %d", len(result.Items))
		}
	})

	t.Run("filter by context", func(t *testing.T) {
		result, err := store.Tasks().List(ctx, core.TaskFilter{ContextID: &context_.ID})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(result.Items) != 1 {
			t.Errorf("Expected 1 task with context, got %d", len(result.Items))
		}
	})
}

func TestProjectRepository_ListWithFilter(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()

	active := &core.Project{
		ID:        "prj_01HX0000000000000000000000",
		Name:      "Active Project",
		Status:    core.ProjectStatusActive,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	if err := store.Projects().Create(ctx, active); err != nil {
		t.Fatalf("Create active project failed: %v", err)
	}

	someday := &core.Project{
		ID:        "prj_01HX0000000000000000000001",
		Name:      "Someday Project",
		Status:    core.ProjectStatusSomeday,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	if err := store.Projects().Create(ctx, someday); err != nil {
		t.Fatalf("Create someday project failed: %v", err)
	}

	t.Run("filter by active status", func(t *testing.T) {
		status := core.ProjectStatusActive
		projects, err := store.Projects().List(ctx, core.ProjectFilter{Status: &status})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(projects) != 1 {
			t.Errorf("Expected 1 active project, got %d", len(projects))
		}
	})

	t.Run("filter by someday status", func(t *testing.T) {
		status := core.ProjectStatusSomeday
		projects, err := store.Projects().List(ctx, core.ProjectFilter{Status: &status})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(projects) != 1 {
			t.Errorf("Expected 1 someday project, got %d", len(projects))
		}
	})
}

func TestContextRepository_List(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()

	for i := 0; i < 3; i++ {
		context_ := &core.Context{
			ID:        "ctx_01HX000000000000000000000" + string(rune('0'+i)),
			Name:      "@context" + string(rune('0'+i)),
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		}
		if err := store.Contexts().Create(ctx, context_); err != nil {
			t.Fatalf("Create context %d failed: %v", i, err)
		}
	}

	contexts, err := store.Contexts().List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(contexts) != 3 {
		t.Errorf("Expected 3 contexts, got %d", len(contexts))
	}
}

func TestAreaRepository_List(t *testing.T) {
	store := setupTestStore(t)

	ctx := context.Background()

	for i := 0; i < 2; i++ {
		area := &core.Area{
			ID:        "area_01HX000000000000000000000" + string(rune('0'+i)),
			Name:      "Area " + string(rune('0'+i)),
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		}
		if err := store.Areas().Create(ctx, area); err != nil {
			t.Fatalf("Create area %d failed: %v", i, err)
		}
	}

	areas, err := store.Areas().List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(areas) != 2 {
		t.Errorf("Expected 2 areas, got %d", len(areas))
	}
}
