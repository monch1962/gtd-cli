package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/jsonout"
)

var inboxCmd = &cobra.Command{
	Use:   "inbox",
	Short: "Manage inbox items",
	Long:  "Commands for capturing and processing inbox items in GTD.",
}

var inboxAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new item to the inbox",
	Long: `Add a new item to the inbox.

Creates a task with status="inbox". Use optional flags to add notes, source, or assign to a project.

Example:
  gtd-cli inbox add "Call dentist"
  gtd-cli inbox add "Review proposal" --note "From client" --source email`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		note, _ := cmd.Flags().GetString("note")
		source, _ := cmd.Flags().GetString("source")
		projectID, _ := cmd.Flags().GetString("project")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli inbox add", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		now := app.Clock.Now()
		task := &core.Task{
			ID:        app.IDGen.NewID("tsk_"),
			Title:     title,
			Note:      note,
			Status:    core.TaskStatusInbox,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if source != "" {
			task.Source = &source
		}
		if projectID != "" {
			task.ProjectID = &projectID
		}

		if err := app.Store.Tasks().Create(context.Background(), task); err != nil {
			writeError(cmd, "gtd-cli inbox add", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli inbox add", task)
	},
}

var inboxListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all inbox items",
	Long: `List all items currently in the inbox.

Example:
  gtd-cli inbox list`,
	Run: func(cmd *cobra.Command, args []string) {
		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli inbox list", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		status := core.TaskStatusInbox
		result, err := app.Store.Tasks().List(context.Background(), core.TaskFilter{Status: &status})
		if err != nil {
			writeError(cmd, "gtd-cli inbox list", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli inbox list", result)
	},
}

var inboxProcessCmd = &cobra.Command{
	Use:   "process --task <task-id> --to-project <project-id>",
	Short: "Process an inbox item",
	Long: `Process an item from the inbox into a project.

Moves the task out of the inbox and into the specified project, setting the status based on --as flag (default "next").

Example:
  gtd-cli inbox process --task tsk_01HXYZ --to-project prj_01ABC
  gtd-cli inbox process --task tsk_01HXYZ --to-project prj_01ABC --as waiting --waiting-for "John"`,
	Run: func(cmd *cobra.Command, args []string) {
		taskID, _ := cmd.Flags().GetString("task")
		projectID, _ := cmd.Flags().GetString("to-project")
		asStatus, _ := cmd.Flags().GetString("as")
		contextIDs, _ := cmd.Flags().GetStringSlice("context")
		waitingFor, _ := cmd.Flags().GetString("waiting-for")
		tickleAt, _ := cmd.Flags().GetString("tickle")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli inbox process", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		task, err := app.Store.Tasks().Get(context.Background(), taskID)
		if err != nil {
			writeError(cmd, "gtd-cli inbox process", jsonout.ErrNotFound, "task not found", map[string]any{"id": taskID})
			return
		}

		policy := app.Policy()
		as := core.TaskStatus(asStatus)
		if as == "" {
			as = policy.DetermineStatusAfterInbox("")
		}

		params := core.ProcessInboxParams{
			Task:       task,
			ToProject:  &projectID,
			AsStatus:   as,
			ContextIDs: contextIDs,
			WaitingFor: &waitingFor,
			TickleAt:   &tickleAt,
		}
		if waitingFor == "" {
			params.WaitingFor = nil
		}
		if tickleAt == "" {
			params.TickleAt = nil
		}

		if err := policy.ProcessInbox(params); err != nil {
			writeError(cmd, "gtd-cli inbox process", jsonout.ErrValidation, err.Error(), nil)
			return
		}

		task.ProjectID = &projectID
		task.Status = as
		task.ContextIDs = contextIDs
		task.UpdatedAt = app.Clock.Now()

		if waitingFor != "" && as == core.TaskStatusWaiting {
			task.WaitingFor = &waitingFor
		}

		if tickleAt != "" && as == core.TaskStatusTickler {
			t, err := time.Parse("2006-01-02", tickleAt)
			if err != nil {
				writeError(cmd, "gtd-cli inbox process", jsonout.ErrValidation, "invalid tickle date format (use YYYY-MM-DD)", nil)
				return
			}
			task.TickleAt = &t
		}

		if err := app.Store.Tasks().Update(context.Background(), task); err != nil {
			writeError(cmd, "gtd-cli inbox process", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli inbox process", task)
	},
}

func init() {
	rootCmd.AddCommand(inboxCmd)
	inboxCmd.AddCommand(inboxAddCmd)
	inboxCmd.AddCommand(inboxListCmd)
	inboxCmd.AddCommand(inboxProcessCmd)

	inboxAddCmd.Flags().String("note", "", "additional note for the task")
	inboxAddCmd.Flags().String("source", "", "source of the task (cli|sms|email|other)")
	inboxAddCmd.Flags().String("project", "", "project to assign the task to")

	inboxProcessCmd.Flags().String("task", "", "task ID to process (required)")
	inboxProcessCmd.Flags().String("to-project", "", "project to move the task to (required)")
	inboxProcessCmd.Flags().String("as", "", "status after processing (next|waiting|someday|tickler|reference)")
	inboxProcessCmd.Flags().StringSlice("context", nil, "context(s) to assign (e.g., @calls)")
	inboxProcessCmd.Flags().String("waiting-for", "", "who/what we're waiting for (for 'waiting' status)")
	inboxProcessCmd.Flags().String("tickle", "", "tickle date YYYY-MM-DD (for 'tickler' status)")
	inboxProcessCmd.MarkFlagRequired("task")
	inboxProcessCmd.MarkFlagRequired("to-project")
}
