package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/jsonout"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage contexts",
	Long:  "Commands for managing contexts (e.g., @calls, @computer, @home).",
}

var contextAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new context",
	Long: `Add a new context.

Context names conventionally start with @ (e.g., @calls, @computer).

Example:
  gtd-cli context add "@calls"
  gtd-cli context add "@computer"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli context add", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		now := app.Clock.Now()
		context_ := &core.Context{
			ID:        app.IDGen.NewID("ctx_"),
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := core.ValidateContext(context_); err != nil {
			writeError(cmd, "gtd-cli context add", jsonout.ErrValidation, err.Error(), nil)
			return
		}

		if err := app.Store.Contexts().Create(context.Background(), context_); err != nil {
			writeError(cmd, "gtd-cli context add", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli context add", context_)
	},
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all contexts",
	Long: `List all contexts.

Example:
  gtd-cli context list`,
	Run: func(cmd *cobra.Command, args []string) {
		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli context list", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		contexts, err := app.Store.Contexts().List(context.Background())
		if err != nil {
			writeError(cmd, "gtd-cli context list", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli context list", map[string]any{
			"items": contexts,
			"count": len(contexts),
		})
	},
}

var contextRenameCmd = &cobra.Command{
	Use:   "rename <context-id> --name <new-name>",
	Short: "Rename a context",
	Long: `Rename a context.

Example:
  gtd-cli context rename ctx_01ABC --name "@phone"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		contextID := args[0]
		name, _ := cmd.Flags().GetString("name")

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli context rename", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		context_, err := app.Store.Contexts().Get(context.Background(), contextID)
		if err != nil {
			writeError(cmd, "gtd-cli context rename", jsonout.ErrNotFound, "context not found", map[string]any{"id": contextID})
			return
		}

		context_.Name = name
		context_.UpdatedAt = app.Clock.Now()

		if err := app.Store.Contexts().Update(context.Background(), context_); err != nil {
			writeError(cmd, "gtd-cli context rename", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli context rename", context_)
	},
}

var contextDeleteCmd = &cobra.Command{
	Use:   "delete <context-id>",
	Short: "Delete a context",
	Long: `Delete a context.

This removes the context from all tasks.

Example:
  gtd-cli context delete ctx_01ABC`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		contextID := args[0]

		app, err := newApp(rootCmd.Version)
		if err != nil {
			writeError(cmd, "gtd-cli context delete", jsonout.ErrInternal, err.Error(), nil)
			return
		}
		defer app.Store.Close()

		if err := app.Store.Contexts().Delete(context.Background(), contextID); err != nil {
			writeError(cmd, "gtd-cli context delete", jsonout.ErrInternal, err.Error(), nil)
			return
		}

		writeSuccess(cmd, "gtd-cli context delete", map[string]any{"deleted": true, "id": contextID})
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextAddCmd)
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextRenameCmd)
	contextCmd.AddCommand(contextDeleteCmd)

	contextRenameCmd.Flags().String("name", "", "new name for the context (required)")
	contextRenameCmd.MarkFlagRequired("name")
}
