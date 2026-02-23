package cli

import (
	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/jsonout"
)

func writeSuccess(cmd *cobra.Command, command string, data any) {
	rw := jsonout.NewResponseWriter(cmd.OutOrStdout(), nil, backend, profile, rootCmd.Version, pretty)
	rw.WriteSuccess(command, data)
}

func writeError(cmd *cobra.Command, command string, code jsonout.ErrorCode, message string, details map[string]any) {
	rw := jsonout.NewResponseWriter(cmd.OutOrStdout(), nil, backend, profile, rootCmd.Version, pretty)
	rw.WriteError(command, code, message, details)
}
