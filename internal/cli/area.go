package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/jsonout"
)

var areaCmd = &cobra.Command{
	Use:   "area",
	Short: "Manage areas of focus",
	Long:  "Commands for managing areas of focus (e.g., Health, Career, Relationships).",
}

var areaAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new area of focus",
	Long: `Add a new area of focus.

Areas represent broad categories of responsibility in your life.

Example:
  gtd-cli area add "Health"
  gtd-cli area add "Career"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli area add", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		now := app.Clock.Now()
		area := &core.Area{
			ID:        app.IDGen.NewID("area_"),
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := app.Store.Areas().Create(context.Background(), area); err != nil {
			writeError(cmd, "gtd-cli area add", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli area add", area)
	},
}

var areaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all areas of focus",
	Long: `List all areas of focus.

Example:
  gtd-cli area list`,
	Run: func(cmd *cobra.Command, args []string) {
		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli area list", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		areas, err := app.Store.Areas().List(context.Background())
		if err != nil {
			writeError(cmd, "gtd-cli area list", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli area list", map[string]any{
			"items": areas,
			"count": len(areas),
		})
	},
}

var areaRenameCmd = &cobra.Command{
	Use:   "rename <area-id> --name <new-name>",
	Short: "Rename an area of focus",
	Long: `Rename an area of focus.

Example:
  gtd-cli area rename area_01ABC --name "Personal Development"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		areaID := args[0]
		name, _ := cmd.Flags().GetString("name")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli area rename", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		area, err := app.Store.Areas().Get(context.Background(), areaID)
		if err != nil {
			writeError(cmd, "gtd-cli area rename", jsonout.ErrNotFound, "area not found", map[string]any{"id": areaID})
			return
		}

		area.Name = name
		area.UpdatedAt = app.Clock.Now()

		if err := app.Store.Areas().Update(context.Background(), area); err != nil {
			writeError(cmd, "gtd-cli area rename", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli area rename", area)
	},
}

var areaDeleteCmd = &cobra.Command{
	Use:   "delete <area-id>",
	Short: "Delete an area of focus",
	Long: `Delete an area of focus.

This does not delete projects or tasks in this area.

Example:
  gtd-cli area delete area_01ABC`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		areaID := args[0]

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli area delete", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		if err := app.Store.Areas().Delete(context.Background(), areaID); err != nil {
			writeError(cmd, "gtd-cli area delete", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli area delete", map[string]any{"deleted": true, "id": areaID})
	},
}

func init() {
	rootCmd.AddCommand(areaCmd)
	areaCmd.AddCommand(areaAddCmd)
	areaCmd.AddCommand(areaListCmd)
	areaCmd.AddCommand(areaRenameCmd)
	areaCmd.AddCommand(areaDeleteCmd)

	areaRenameCmd.Flags().String("name", "", "new name for the area (required)")
	areaRenameCmd.MarkFlagRequired("name")
}
