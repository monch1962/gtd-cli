package core

import "time"

const (
	DefaultLimit       = 100
	DefaultReviewLimit = 50
	StaleProjectDays   = 14
	DateFormat         = "2006-01-02"
)

type TaskStatus string

const (
	TaskStatusInbox     TaskStatus = "inbox"
	TaskStatusNext      TaskStatus = "next"
	TaskStatusWaiting   TaskStatus = "waiting"
	TaskStatusSomeday   TaskStatus = "someday"
	TaskStatusTickler   TaskStatus = "tickler"
	TaskStatusReference TaskStatus = "reference"
	TaskStatusDone      TaskStatus = "done"
)

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Note        string     `json:"note,omitempty"`
	Status      TaskStatus `json:"status"`
	ProjectID   *string    `json:"project_id,omitempty"`
	AreaID      *string    `json:"area_id,omitempty"`
	ContextIDs  []string   `json:"context_ids,omitempty"`
	WaitingFor  *string    `json:"waiting_for,omitempty"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	StartAt     *time.Time `json:"start_at,omitempty"`
	TickleAt    *time.Time `json:"tickle_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Source      *string    `json:"source,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusSomeday  ProjectStatus = "someday"
	ProjectStatusDone     ProjectStatus = "done"
	ProjectStatusArchived ProjectStatus = "archived"
)

type Project struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Note      string        `json:"note,omitempty"`
	Status    ProjectStatus `json:"status"`
	AreaID    *string       `json:"area_id,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type Context struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Area struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReviewType string

const (
	ReviewTypeDaily   ReviewType = "daily"
	ReviewTypeWeekly  ReviewType = "weekly"
	ReviewTypeMonthly ReviewType = "monthly"
)

type ReviewSession struct {
	ID        string     `json:"id"`
	Type      ReviewType `json:"type"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Note      string     `json:"note,omitempty"`
	Stats     any        `json:"stats,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type TaskListResult struct {
	Items  []Task `json:"items"`
	Count  int    `json:"count"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type TaskFilter struct {
	ProjectID *string
	ContextID *string
	Status    *TaskStatus
	Limit     int
	Offset    int
}

type ProjectFilter struct {
	Status *ProjectStatus
	Limit  int
	Offset int
}

type ReviewFilter struct {
	Type  *ReviewType
	Limit int
}
