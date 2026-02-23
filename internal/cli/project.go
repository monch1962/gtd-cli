package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/jsonout"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  "Commands for creating and managing projects.",
}

var projectCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project",
	Long: `Create a new project.

Example:
  gtd-cli project create "Website Redesign"
  gtd-cli project create "Q1 Planning" --area area_01XYZ`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		note, _ := cmd.Flags().GetString("note")
		areaID, _ := cmd.Flags().GetString("area")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli project create", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		now := app.Clock.Now()
		project := &core.Project{
			ID:        app.IDGen.NewID("prj_"),
			Name:      name,
			Note:      note,
			Status:    core.ProjectStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if areaID != "" {
			project.AreaID = &areaID
		}

		if err := core.ValidateProject(project); err != nil {
			writeError(cmd, "gtd-cli project create", jsonout.ErrValidation, err.Error(), nil)
			return
		}

		if err := app.Store.Projects().Create(context.Background(), project); err != nil {
			writeError(cmd, "gtd-cli project create", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli project create", project)
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Long: `List projects with optional filters.

Examples:
  gtd-cli project list
  gtd-cli project list --status active`,
	Run: func(cmd *cobra.Command, args []string) {
		statusStr, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli project list", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		filter := core.ProjectFilter{
			Limit: limit,
		}

		if statusStr != "" {
			status := core.ProjectStatus(statusStr)
			filter.Status = &status
		}

		projects, err := app.Store.Projects().List(context.Background(), filter)
		if err != nil {
			writeError(cmd, "gtd-cli project list", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli project list", map[string]any{
			"items": projects,
			"count": len(projects),
		})
	},
}

var projectShowCmd = &cobra.Command{
	Use:   "show <project-id>",
	Short: "Show project details",
	Long: `Show detailed information about a project.

Example:
  gtd-cli project show prj_01ABC`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectID := args[0]

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli project show", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		project, err := app.Store.Projects().Get(context.Background(), projectID)
		if err != nil {
			writeError(cmd, "gtd-cli project show", jsonout.ErrNotFound, "project not found", map[string]any{"id": projectID})
			return
		}

		writeSuccess(cmd, "gtd-cli project show", project)
	},
}

var projectArchiveCmd = &cobra.Command{
	Use:   "archive <project-id>",
	Short: "Archive a project",
	Long: `Archive a project.

Marks the project as archived. The project and its tasks are retained but hidden from active lists.

Example:
  gtd-cli project archive prj_01ABC`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectID := args[0]

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli project archive", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		if err := app.Store.Projects().Archive(context.Background(), projectID); err != nil {
			writeError(cmd, "gtd-cli project archive", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		project, err := app.Store.Projects().Get(context.Background(), projectID)
		if err != nil {
			writeError(cmd, "gtd-cli project archive", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		writeSuccess(cmd, "gtd-cli project archive", project)
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectShowCmd)
	projectCmd.AddCommand(projectArchiveCmd)

	projectCreateCmd.Flags().String("note", "", "additional note for the project")
	projectCreateCmd.Flags().String("area", "", "area of focus to assign the project to")

	projectListCmd.Flags().String("status", "", "filter by status (active|someday|done|archived)")
	projectListCmd.Flags().Int("limit", core.DefaultLimit, "maximum number of results")
}
