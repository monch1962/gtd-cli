package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/jsonout"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  "Commands for viewing and managing tasks.",
}

var taskShowCmd = &cobra.Command{
	Use:   "show <task-id>",
	Short: "Show task details",
	Long: `Show detailed information about a task.

Example:
  gtd-cli task show tsk_01HXYZ`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli task show", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		task, err := app.Store.Tasks().Get(context.Background(), taskID)
		if err != nil {
			writeError(cmd, "gtd-cli task show", jsonout.ErrNotFound, "task not found", map[string]any{"id": taskID})
			return
		}

		writeSuccess(cmd, "gtd-cli task show", task)
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List tasks with optional filters.

Examples:
  gtd-cli task list
  gtd-cli task list --status next
  gtd-cli task list --project prj_01ABC
  gtd-cli task list --context @calls --limit 10`,
	Run: func(cmd *cobra.Command, args []string) {
		projectID, _ := cmd.Flags().GetString("project")
		contextID, _ := cmd.Flags().GetString("context")
		statusStr, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli task list", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		filter := core.TaskFilter{
			Limit:  limit,
			Offset: offset,
		}

		if projectID != "" {
			filter.ProjectID = &projectID
		}
		if contextID != "" {
			filter.ContextID = &contextID
		}
		if statusStr != "" {
			status := core.TaskStatus(statusStr)
			filter.Status = &status
		}

		result, err := app.Store.Tasks().List(context.Background(), filter)
		if err != nil {
			writeError(cmd, "gtd-cli task list", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli task list", result)
	},
}

var taskMoveCmd = &cobra.Command{
	Use:   "move <task-id> --to-project <project-id>",
	Short: "Move a task to a different project",
	Long: `Move a task to a different project.

If the task was in the inbox, it will automatically become "next" status (unless configured otherwise).

Example:
  gtd-cli task move tsk_01HXYZ --to-project prj_01ABC`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		projectID, _ := cmd.Flags().GetString("to-project")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli task move", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		task, err := app.Store.Tasks().Get(context.Background(), taskID)
		if err != nil {
			writeError(cmd, "gtd-cli task move", jsonout.ErrNotFound, "task not found", map[string]any{"id": taskID})
			return
		}

		if err := app.Store.Tasks().MoveToProject(context.Background(), taskID, projectID); err != nil {
			writeError(cmd, "gtd-cli task move", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		policy := app.Policy()
		newStatus := policy.DetermineStatusAfterMove(task.Status)
		if newStatus != task.Status {
			task.Status = newStatus
			task.UpdatedAt = app.Clock.Now()
			if err := app.Store.Tasks().Update(context.Background(), task); err != nil {
				writeError(cmd, "gtd-cli task move", jsonout.ErrInternal, err.Error(), nil)
				return
			}
		}

		updated, _ := app.Store.Tasks().Get(context.Background(), taskID)
		writeSuccess(cmd, "gtd-cli task move", updated)
	},
}

var taskCompleteCmd = &cobra.Command{
	Use:   "complete <task-id>",
	Short: "Mark a task as complete",
	Long: `Mark a task as completed.

Example:
  gtd-cli task complete tsk_01HXYZ
  gtd-cli task complete tsk_01HXYZ --at "2024-01-15T10:30:00Z"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		completedAt, _ := cmd.Flags().GetString("at")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli task complete", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		if err := app.Store.Tasks().Complete(context.Background(), taskID, completedAt); err != nil {
			writeError(cmd, "gtd-cli task complete", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		task, _ := app.Store.Tasks().Get(context.Background(), taskID)
		writeSuccess(cmd, "gtd-cli task complete", task)
	},
}

var taskReopenCmd = &cobra.Command{
	Use:   "reopen <task-id>",
	Short: "Reopen a completed task",
	Long: `Reopen a completed task, setting its status to "next".

Example:
  gtd-cli task reopen tsk_01HXYZ`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli task reopen", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		if err := app.Store.Tasks().Reopen(context.Background(), taskID); err != nil {
			writeError(cmd, "gtd-cli task reopen", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		task, _ := app.Store.Tasks().Get(context.Background(), taskID)
		writeSuccess(cmd, "gtd-cli task reopen", task)
	},
}

var taskNextCmd = &cobra.Command{
	Use:   "next (--project <project-id> | --context <@context>)",
	Short: "Get next actions",
	Long: `Get "what's next" items ordered by priority heuristics.

Ordering:
  1. Tasks with status=next only
  2. Due date (soonest first, nulls last)
  3. Created date (oldest first)

Examples:
  gtd-cli task next --project prj_01ABC
  gtd-cli task next --context @calls --limit 5`,
	Run: func(cmd *cobra.Command, args []string) {
		projectID, _ := cmd.Flags().GetString("project")
		contextID, _ := cmd.Flags().GetString("context")
		limit, _ := cmd.Flags().GetInt("limit")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli task next", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		tasks, err := app.Store.Tasks().GetNext(context.Background(), projectID, contextID, limit)
		if err != nil {
			writeError(cmd, "gtd-cli task next", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli task next", map[string]any{
			"items": tasks,
			"count": len(tasks),
		})
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskMoveCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	taskCmd.AddCommand(taskReopenCmd)
	taskCmd.AddCommand(taskNextCmd)

	taskListCmd.Flags().String("project", "", "filter by project ID")
	taskListCmd.Flags().String("context", "", "filter by context ID")
	taskListCmd.Flags().String("status", "", "filter by status (inbox|next|waiting|someday|tickler|reference|done)")
	taskListCmd.Flags().Int("limit", 100, "maximum number of results")
	taskListCmd.Flags().Int("offset", 0, "offset for pagination")

	taskMoveCmd.Flags().String("to-project", "", "project to move the task to (required)")
	taskMoveCmd.MarkFlagRequired("to-project")

	taskCompleteCmd.Flags().String("at", "", "completion timestamp (RFC3339)")

	taskNextCmd.Flags().String("project", "", "filter by project ID")
	taskNextCmd.Flags().String("context", "", "filter by context ID")
	taskNextCmd.Flags().Int("limit", 10, "maximum number of results")
}
