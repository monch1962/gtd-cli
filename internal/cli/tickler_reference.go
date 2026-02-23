package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/jsonout"
)

var ticklerCmd = &cobra.Command{
	Use:   "tickler",
	Short: "Manage tickler items",
	Long:  "Commands for managing tickler items (tasks that resurface at a future date).",
}

var ticklerAddCmd = &cobra.Command{
	Use:   "add <title> --date <YYYY-MM-DD>",
	Short: "Add a tickler item",
	Long: `Add a tickler item that will resurface on a specific date.

Creates a task with status="tickler" and tickle_at set to the specified date.

Example:
  gtd-cli tickler add "Follow up with John" --date 2024-02-15
  gtd-cli tickler add "Tax deadline reminder" --date 2024-04-01 --note "File taxes"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		dateStr, _ := cmd.Flags().GetString("date")
		note, _ := cmd.Flags().GetString("note")
		projectID, _ := cmd.Flags().GetString("project")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli tickler add", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		tickleAt, err := time.Parse(core.DateFormat, dateStr)
		if err != nil {
			writeError(cmd, "gtd-cli tickler add", jsonout.ErrValidation, "invalid date format (use YYYY-MM-DD)", nil)
			return
		}

		now := app.Clock.Now()
		task := &core.Task{
			ID:        app.IDGen.NewID("tsk_"),
			Title:     title,
			Note:      note,
			Status:    core.TaskStatusTickler,
			TickleAt:  &tickleAt,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if projectID != "" {
			task.ProjectID = &projectID
		}

		if err := core.ValidateTask(task); err != nil {
			writeError(cmd, "gtd-cli tickler add", jsonout.ErrValidation, err.Error(), nil)
			return
		}

		if err := app.Store.Tasks().Create(context.Background(), task); err != nil {
			writeError(cmd, "gtd-cli tickler add", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli tickler add", task)
	},
}

var ticklerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tickler items",
	Long: `List tickler items with optional date range filters.

Examples:
  gtd-cli tickler list
  gtd-cli tickler list --from 2024-01-01 --to 2024-02-01
  gtd-cli tickler list --limit 10`,
	Run: func(cmd *cobra.Command, args []string) {
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		limit, _ := cmd.Flags().GetInt("limit")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli tickler list", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		status := core.TaskStatusTickler
		filter := core.TaskFilter{
			Status: &status,
			Limit:  limit,
		}

		if fromStr != "" {
			_, err := time.Parse(core.DateFormat, fromStr)
			if err != nil {
				writeError(cmd, "gtd-cli tickler list", jsonout.ErrValidation, "invalid from date format (use YYYY-MM-DD)", nil)
				return
			}
		}

		if toStr != "" {
			_, err := time.Parse(core.DateFormat, toStr)
			if err != nil {
				writeError(cmd, "gtd-cli tickler list", jsonout.ErrValidation, "invalid to date format (use YYYY-MM-DD)", nil)
				return
			}
		}

		result, err := app.Store.Tasks().List(context.Background(), filter)
		if err != nil {
			writeError(cmd, "gtd-cli tickler list", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		var items []core.Task
		for _, t := range result.Items {
			if fromStr != "" {
				from, _ := time.Parse(core.DateFormat, fromStr)
				if t.TickleAt != nil && t.TickleAt.Before(from) {
					continue
				}
			}
			if toStr != "" {
				to, _ := time.Parse(core.DateFormat, toStr)
				if t.TickleAt != nil && t.TickleAt.After(to) {
					continue
				}
			}
			items = append(items, t)
		}

		writeSuccess(cmd, "gtd-cli tickler list", map[string]any{
			"items": items,
			"count": len(items),
		})
	},
}

var referenceCmd = &cobra.Command{
	Use:   "reference",
	Short: "Manage reference items",
	Long:  "Commands for managing reference items (tasks that store information).",
}

var referenceAddCmd = &cobra.Command{
	Use:   "add --title <title>",
	Short: "Add a reference item",
	Long: `Add a reference item.

Creates a task with status="reference" for storing information.

Example:
  gtd-cli reference add --title "WiFi Password"
  gtd-cli reference add --title "API Docs" --path "https://docs.example.com" --note "v2 API"`,
	Run: func(cmd *cobra.Command, args []string) {
		title, _ := cmd.Flags().GetString("title")
		path, _ := cmd.Flags().GetString("path")
		note, _ := cmd.Flags().GetString("note")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli reference add", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		now := app.Clock.Now()
		task := &core.Task{
			ID:        app.IDGen.NewID("tsk_"),
			Title:     title,
			Note:      note,
			Status:    core.TaskStatusReference,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if path != "" {
			task.Note = path
			if note != "" {
				task.Note = path + "\n" + note
			}
		}

		if err := core.ValidateTask(task); err != nil {
			writeError(cmd, "gtd-cli reference add", jsonout.ErrValidation, err.Error(), nil)
			return
		}

		if err := app.Store.Tasks().Create(context.Background(), task); err != nil {
			writeError(cmd, "gtd-cli reference add", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli reference add", task)
	},
}

var referenceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List reference items",
	Long: `List reference items.

Example:
  gtd-cli reference list --limit 20`,
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli reference list", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		status := core.TaskStatusReference
		result, err := app.Store.Tasks().List(context.Background(), core.TaskFilter{
			Status: &status,
			Limit:  limit,
		})
		if err != nil {
			writeError(cmd, "gtd-cli reference list", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli reference list", result)
	},
}

func init() {
	rootCmd.AddCommand(ticklerCmd)
	rootCmd.AddCommand(referenceCmd)

	ticklerCmd.AddCommand(ticklerAddCmd)
	ticklerCmd.AddCommand(ticklerListCmd)

	ticklerAddCmd.Flags().String("date", "", "date when the tickler should resurface (YYYY-MM-DD) (required)")
	ticklerAddCmd.Flags().String("note", "", "additional note")
	ticklerAddCmd.Flags().String("project", "", "project to assign the tickler to")
	ticklerAddCmd.MarkFlagRequired("date")

	ticklerListCmd.Flags().String("from", "", "filter from date (YYYY-MM-DD)")
	ticklerListCmd.Flags().String("to", "", "filter to date (YYYY-MM-DD)")
	ticklerListCmd.Flags().Int("limit", core.DefaultLimit, "maximum number of results")

	referenceCmd.AddCommand(referenceAddCmd)
	referenceCmd.AddCommand(referenceListCmd)

	referenceAddCmd.Flags().String("title", "", "title for the reference (required)")
	referenceAddCmd.Flags().String("path", "", "path or URI for the reference")
	referenceAddCmd.Flags().String("note", "", "additional note")
	referenceAddCmd.MarkFlagRequired("title")

	referenceListCmd.Flags().Int("limit", core.DefaultLimit, "maximum number of results")
}
