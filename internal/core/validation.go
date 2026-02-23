package core

import (
	"errors"
	"fmt"
)

var (
	ErrValidation           = errors.New("validation error")
	ErrNotFound             = errors.New("not found")
	ErrConflict             = errors.New("conflict")
	ErrInvalidStatus        = errors.New("invalid status")
	ErrEmptyTitle           = errors.New("empty title")
	ErrEmptyName            = errors.New("empty name")
	ErrInvalidTaskStatus    = errors.New("invalid task status")
	ErrInvalidProjectStatus = errors.New("invalid project status")
)

type Policy struct {
	RequireProjectWhenLeavingInbox bool
	RequireContextWhenLeavingInbox bool
	AutoNextOnMoveFromInbox        bool
	AutoNextOnInboxProcess         bool
}

func DefaultPolicy() Policy {
	return Policy{
		RequireProjectWhenLeavingInbox: false,
		RequireContextWhenLeavingInbox: false,
		AutoNextOnMoveFromInbox:        true,
		AutoNextOnInboxProcess:         true,
	}
}

func ValidateTask(task *Task) error {
	if task.Title == "" {
		return ErrEmptyTitle
	}
	if !isValidTaskStatus(task.Status) {
		return fmt.Errorf("%w: %s", ErrInvalidTaskStatus, task.Status)
	}
	return nil
}

func isValidTaskStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusInbox, TaskStatusNext, TaskStatusWaiting, TaskStatusSomeday, TaskStatusTickler, TaskStatusReference, TaskStatusDone:
		return true
	default:
		return false
	}
}

func ValidateProject(project *Project) error {
	if project.Name == "" {
		return ErrEmptyName
	}
	if !isValidProjectStatus(project.Status) {
		return fmt.Errorf("%w: %s", ErrInvalidProjectStatus, project.Status)
	}
	return nil
}

func isValidProjectStatus(status ProjectStatus) bool {
	switch status {
	case ProjectStatusActive, ProjectStatusSomeday, ProjectStatusDone, ProjectStatusArchived:
		return true
	default:
		return false
	}
}

func ValidateContext(context *Context) error {
	if context.Name == "" {
		return ErrEmptyName
	}
	return nil
}

func ValidateArea(area *Area) error {
	if area.Name == "" {
		return ErrEmptyName
	}
	return nil
}

type ProcessInboxParams struct {
	Task       *Task
	ToProject  *string
	AsStatus   TaskStatus
	ContextIDs []string
	WaitingFor *string
	TickleAt   *string
}

func (p *Policy) ProcessInbox(params ProcessInboxParams) error {
	if params.Task.Status != TaskStatusInbox {
		return fmt.Errorf("task is not in inbox: status=%s", params.Task.Status)
	}

	if p.RequireProjectWhenLeavingInbox && params.ToProject == nil {
		return fmt.Errorf("%w: project required when leaving inbox", ErrValidation)
	}

	if p.RequireContextWhenLeavingInbox && len(params.ContextIDs) == 0 {
		return fmt.Errorf("%w: context required when leaving inbox", ErrValidation)
	}

	if params.AsStatus != "" && !isValidTaskStatus(params.AsStatus) {
		return fmt.Errorf("%w: %s", ErrInvalidTaskStatus, params.AsStatus)
	}

	if params.AsStatus == TaskStatusWaiting && params.WaitingFor == nil {
		return fmt.Errorf("%w: waiting_for required for waiting status", ErrValidation)
	}

	if params.AsStatus == TaskStatusTickler && params.TickleAt == nil {
		return fmt.Errorf("%w: tickle_at required for tickler status", ErrValidation)
	}

	return nil
}

func (p *Policy) DetermineStatusAfterInbox(asStatus TaskStatus) TaskStatus {
	if asStatus != "" {
		return asStatus
	}
	if p.AutoNextOnInboxProcess {
		return TaskStatusNext
	}
	return TaskStatusInbox
}

func (p *Policy) DetermineStatusAfterMove(currentStatus TaskStatus) TaskStatus {
	if currentStatus != TaskStatusInbox {
		return currentStatus
	}
	if p.AutoNextOnMoveFromInbox {
		return TaskStatusNext
	}
	return TaskStatusInbox
}
