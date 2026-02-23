package core

import (
	"errors"
	"testing"
)

func TestValidateTask(t *testing.T) {
	tests := []struct {
		name    string
		task    *Task
		wantErr error
	}{
		{
			name:    "valid task",
			task:    &Task{Title: "Test", Status: TaskStatusInbox},
			wantErr: nil,
		},
		{
			name:    "empty title",
			task:    &Task{Title: "", Status: TaskStatusInbox},
			wantErr: ErrEmptyTitle,
		},
		{
			name:    "invalid status",
			task:    &Task{Title: "Test", Status: TaskStatus("invalid")},
			wantErr: ErrInvalidTaskStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTask(tt.task)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateTask() = %v, want nil", err)
				}
			} else {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateTask() = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestValidateProject(t *testing.T) {
	tests := []struct {
		name    string
		project *Project
		wantErr error
	}{
		{
			name:    "valid project",
			project: &Project{Name: "Test", Status: ProjectStatusActive},
			wantErr: nil,
		},
		{
			name:    "empty name",
			project: &Project{Name: "", Status: ProjectStatusActive},
			wantErr: ErrEmptyName,
		},
		{
			name:    "invalid status",
			project: &Project{Name: "Test", Status: ProjectStatus("invalid")},
			wantErr: ErrInvalidProjectStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProject(tt.project)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateProject() = %v, want nil", err)
				}
			} else {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateProject() = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestPolicy_ProcessInbox(t *testing.T) {
	projectID := "prj_123"

	tests := []struct {
		name    string
		policy  Policy
		params  ProcessInboxParams
		wantErr bool
	}{
		{
			name:   "basic process",
			policy: DefaultPolicy(),
			params: ProcessInboxParams{
				Task:      &Task{Status: TaskStatusInbox},
				ToProject: &projectID,
				AsStatus:  TaskStatusNext,
			},
			wantErr: false,
		},
		{
			name:   "require project enabled, no project",
			policy: Policy{RequireProjectWhenLeavingInbox: true},
			params: ProcessInboxParams{
				Task:     &Task{Status: TaskStatusInbox},
				AsStatus: TaskStatusNext,
			},
			wantErr: true,
		},
		{
			name:   "require context enabled, no context",
			policy: Policy{RequireContextWhenLeavingInbox: true},
			params: ProcessInboxParams{
				Task:     &Task{Status: TaskStatusInbox},
				AsStatus: TaskStatusNext,
			},
			wantErr: true,
		},
		{
			name:   "task not in inbox",
			policy: DefaultPolicy(),
			params: ProcessInboxParams{
				Task:     &Task{Status: TaskStatusNext},
				AsStatus: TaskStatusNext,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.ProcessInbox(tt.params)
			if tt.wantErr && err == nil {
				t.Error("ProcessInbox() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ProcessInbox() = %v, want nil", err)
			}
		})
	}
}

func TestPolicy_DetermineStatusAfterInbox(t *testing.T) {
	policy := DefaultPolicy()

	if got := policy.DetermineStatusAfterInbox(TaskStatusSomeday); got != TaskStatusSomeday {
		t.Errorf("DetermineStatusAfterInbox(someday) = %v, want someday", got)
	}

	if got := policy.DetermineStatusAfterInbox(""); got != TaskStatusNext {
		t.Errorf("DetermineStatusAfterInbox(empty) = %v, want next", got)
	}

	policyNoAuto := Policy{AutoNextOnInboxProcess: false}
	if got := policyNoAuto.DetermineStatusAfterInbox(""); got != TaskStatusInbox {
		t.Errorf("DetermineStatusAfterInbox(empty, no auto) = %v, want inbox", got)
	}
}

func TestPolicy_DetermineStatusAfterMove(t *testing.T) {
	policy := DefaultPolicy()

	if got := policy.DetermineStatusAfterMove(TaskStatusNext); got != TaskStatusNext {
		t.Errorf("DetermineStatusAfterMove(next) = %v, want next", got)
	}

	if got := policy.DetermineStatusAfterMove(TaskStatusInbox); got != TaskStatusNext {
		t.Errorf("DetermineStatusAfterMove(inbox) = %v, want next", got)
	}

	policyNoAuto := Policy{AutoNextOnMoveFromInbox: false}
	if got := policyNoAuto.DetermineStatusAfterMove(TaskStatusInbox); got != TaskStatusInbox {
		t.Errorf("DetermineStatusAfterMove(inbox, no auto) = %v, want inbox", got)
	}
}
