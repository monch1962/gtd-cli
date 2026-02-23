package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/jsonout"
)

func TestWriteSuccess(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	backend = "sqlite"
	profile = "default"
	rootCmd.Version = "1.0.0"
	pretty = false

	writeSuccess(cmd, "test command", map[string]string{"key": "value"})

	var env testEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if !env.OK {
		t.Error("OK should be true")
	}
	if env.Command != "test command" {
		t.Errorf("Command = %s, want 'test command'", env.Command)
	}
	if env.Error != nil {
		t.Error("Error should be nil for success")
	}
}

func TestWriteError(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	backend = "sqlite"
	profile = "default"
	rootCmd.Version = "1.0.0"
	pretty = false

	writeError(cmd, "test command", jsonout.ErrNotFound, "item not found", map[string]any{"id": "123"})

	var env testEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if env.OK {
		t.Error("OK should be false for error")
	}
	if env.Command != "test command" {
		t.Errorf("Command = %s, want 'test command'", env.Command)
	}
	if env.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if env.Error.Code != string(jsonout.ErrNotFound) {
		t.Errorf("Error.Code = %s, want NOT_FOUND", env.Error.Code)
	}
	if env.Error.Message != "item not found" {
		t.Errorf("Error.Message = %s, want 'item not found'", env.Error.Message)
	}
}

func TestWriteSuccess_Pretty(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	backend = "sqlite"
	profile = "default"
	rootCmd.Version = "1.0.0"
	pretty = true

	writeSuccess(cmd, "test", map[string]string{"key": "value"})

	if !strings.Contains(buf.String(), "\n") {
		t.Error("Pretty output should contain newlines")
	}
	if !strings.Contains(buf.String(), "  ") {
		t.Error("Pretty output should contain indentation")
	}
}

func TestGetChecklist(t *testing.T) {
	tests := []struct {
		reviewType    core.ReviewType
		expectedCount int
	}{
		{core.ReviewTypeDaily, 4},
		{core.ReviewTypeWeekly, 6},
		{core.ReviewTypeMonthly, 5},
	}

	for _, tt := range tests {
		t.Run(string(tt.reviewType), func(t *testing.T) {
			checklist := getChecklist(tt.reviewType)
			if len(checklist) != tt.expectedCount {
				t.Errorf("getChecklist(%s) returned %d items, want %d", tt.reviewType, len(checklist), tt.expectedCount)
			}
			for i, item := range checklist {
				if item["id"] == "" {
					t.Errorf("Checklist item %d missing id", i)
				}
				if item["title"] == "" {
					t.Errorf("Checklist item %d missing title", i)
				}
				if item["description"] == "" {
					t.Errorf("Checklist item %d missing description", i)
				}
			}
		})
	}
}

func TestGetChecklist_UnknownType(t *testing.T) {
	checklist := getChecklist(core.ReviewType("unknown"))
	if checklist != nil {
		t.Errorf("getChecklist(unknown) = %v, want nil", checklist)
	}
}

type testEnvelope struct {
	OK      bool   `json:"ok"`
	Command string `json:"command"`
	Data    any    `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details"`
	} `json:"error"`
	Meta struct {
		Timestamp string `json:"timestamp"`
		Backend   string `json:"backend"`
		Profile   string `json:"profile"`
		Version   string `json:"version"`
	} `json:"meta"`
}
