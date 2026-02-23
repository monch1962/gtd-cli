package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/jsonout"
)

func TestCLI_InboxAdd(t *testing.T) {
	tmpDir := t.TempDir()

	backend = "json"
	dataDir = tmpDir
	defer func() {
		backend = ""
		dataDir = ""
	}()

	app, err := newApp("test")
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}
	defer app.Store.Close()

	now := app.Clock.Now()
	task := &core.Task{
		ID:        app.IDGen.NewID("tsk_"),
		Title:     "Test Task",
		Status:    core.TaskStatusInbox,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := app.Store.Tasks().Create(context.Background(), task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	got, err := app.Store.Tasks().Get(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if got.Title != task.Title {
		t.Errorf("Title = %s, want %s", got.Title, task.Title)
	}
	if got.Status != core.TaskStatusInbox {
		t.Errorf("Status = %s, want inbox", got.Status)
	}
}

func TestCLI_ProjectCreate(t *testing.T) {
	tmpDir := t.TempDir()
	backend = "json"
	dataDir = tmpDir
	defer func() {
		backend = ""
		dataDir = ""
	}()

	app, err := newApp("test")
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}
	defer app.Store.Close()

	now := app.Clock.Now()
	project := &core.Project{
		ID:        app.IDGen.NewID("prj_"),
		Name:      "Test Project",
		Status:    core.ProjectStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := app.Store.Projects().Create(context.Background(), project); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	got, err := app.Store.Projects().Get(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}

	if got.Name != project.Name {
		t.Errorf("Name = %s, want %s", got.Name, project.Name)
	}
}

func TestCLI_InboxProcess(t *testing.T) {
	tmpDir := t.TempDir()
	backend = "json"
	dataDir = tmpDir
	defer func() {
		backend = ""
		dataDir = ""
	}()

	app, err := newApp("test")
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}
	defer app.Store.Close()

	now := app.Clock.Now()

	project := &core.Project{
		ID:        app.IDGen.NewID("prj_"),
		Name:      "Test Project",
		Status:    core.ProjectStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := app.Store.Projects().Create(context.Background(), project); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	task := &core.Task{
		ID:        app.IDGen.NewID("tsk_"),
		Title:     "Test Task",
		Status:    core.TaskStatusInbox,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := app.Store.Tasks().Create(context.Background(), task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	policy := app.Policy()
	newStatus := policy.DetermineStatusAfterInbox("")
	task.ProjectID = &project.ID
	task.Status = newStatus
	task.UpdatedAt = now

	if err := app.Store.Tasks().Update(context.Background(), task); err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	got, err := app.Store.Tasks().Get(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if got.ProjectID == nil || *got.ProjectID != project.ID {
		t.Errorf("ProjectID = %v, want %s", got.ProjectID, project.ID)
	}
	if got.Status != core.TaskStatusNext {
		t.Errorf("Status = %s, want next", got.Status)
	}
}

func TestCLI_TaskComplete(t *testing.T) {
	tmpDir := t.TempDir()
	backend = "json"
	dataDir = tmpDir
	defer func() {
		backend = ""
		dataDir = ""
	}()

	app, err := newApp("test")
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}
	defer app.Store.Close()

	now := app.Clock.Now()
	task := &core.Task{
		ID:        app.IDGen.NewID("tsk_"),
		Title:     "Test Task",
		Status:    core.TaskStatusNext,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := app.Store.Tasks().Create(context.Background(), task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if err := app.Store.Tasks().Complete(context.Background(), task.ID, ""); err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	got, err := app.Store.Tasks().Get(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if got.Status != core.TaskStatusDone {
		t.Errorf("Status = %s, want done", got.Status)
	}
	if got.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestCLI_ContextCRUD(t *testing.T) {
	tmpDir := t.TempDir()
	backend = "json"
	dataDir = tmpDir
	defer func() {
		backend = ""
		dataDir = ""
	}()

	app, err := newApp("test")
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}
	defer app.Store.Close()

	now := app.Clock.Now()
	context_ := &core.Context{
		ID:        app.IDGen.NewID("ctx_"),
		Name:      "@calls",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := app.Store.Contexts().Create(context.Background(), context_); err != nil {
		t.Fatalf("CreateContext failed: %v", err)
	}

	got, err := app.Store.Contexts().Get(context.Background(), context_.ID)
	if err != nil {
		t.Fatalf("GetContext failed: %v", err)
	}

	if got.Name != context_.Name {
		t.Errorf("Name = %s, want %s", got.Name, context_.Name)
	}
}

func TestCLI_ReviewCreate(t *testing.T) {
	tmpDir := t.TempDir()
	backend = "json"
	dataDir = tmpDir
	defer func() {
		backend = ""
		dataDir = ""
	}()

	app, err := newApp("test")
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}
	defer app.Store.Close()

	now := app.Clock.Now()
	review := &core.ReviewSession{
		ID:        app.IDGen.NewID("rev_"),
		Type:      core.ReviewTypeWeekly,
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := app.Store.Reviews().Create(context.Background(), review); err != nil {
		t.Fatalf("CreateReview failed: %v", err)
	}

	got, err := app.Store.Reviews().Get(context.Background(), review.ID)
	if err != nil {
		t.Fatalf("GetReview failed: %v", err)
	}

	if got.Type != review.Type {
		t.Errorf("Type = %s, want %s", got.Type, review.Type)
	}
}

func TestCLI_JSONBackend(t *testing.T) {
	tmpDir := t.TempDir()

	backend = "json"
	dataDir = tmpDir
	defer func() {
		backend = ""
		dataDir = ""
	}()

	app, err := newApp("test")
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}
	defer app.Store.Close()

	now := app.Clock.Now()
	task := &core.Task{
		ID:        app.IDGen.NewID("tsk_"),
		Title:     "JSON Task",
		Status:    core.TaskStatusInbox,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := app.Store.Tasks().Create(context.Background(), task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	taskPath := filepath.Join(tmpDir, "tasks", task.ID+".json")
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		t.Errorf("Task file not created at %s", taskPath)
	}

	data, err := os.ReadFile(taskPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	var loaded core.Task
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if loaded.Title != task.Title {
		t.Errorf("Title = %s, want %s", loaded.Title, task.Title)
	}
}

func TestEnvelopeOutput(t *testing.T) {
	var buf bytes.Buffer
	rw := jsonout.NewResponseWriter(&buf, nil, "sqlite", "default", "1.0.0", false)

	err := rw.WriteSuccess("test command", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	var env cliTestEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if env.OK != true {
		t.Errorf("OK = %v, want true", env.OK)
	}
	if env.Command != "test command" {
		t.Errorf("Command = %s, want 'test command'", env.Command)
	}
	if env.Meta.Backend != "sqlite" {
		t.Errorf("Backend = %s, want sqlite", env.Meta.Backend)
	}
}

type cliTestEnvelope struct {
	OK      bool   `json:"ok"`
	Command string `json:"command"`
	Data    any    `json:"data"`
	Error   any    `json:"error"`
	Meta    struct {
		Timestamp string `json:"timestamp"`
		Backend   string `json:"backend"`
		Profile   string `json:"profile"`
		Version   string `json:"version"`
	} `json:"meta"`
}
