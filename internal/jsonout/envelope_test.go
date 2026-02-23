package jsonout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/anomalyco/gtd-cli/internal/util"
)

func TestEnvelope_Success(t *testing.T) {
	env := Envelope{
		OK:      true,
		Command: "gtd-cli task list",
		Data:    map[string]any{"items": []string{"a", "b"}, "count": 2},
		Error:   nil,
		Meta: Meta{
			Timestamp: "2024-01-15T10:30:00Z",
			Backend:   "sqlite",
			Profile:   "default",
			Version:   "1.0.0",
		},
	}

	got, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	expected := `{"ok":true,"command":"gtd-cli task list","data":{"count":2,"items":["a","b"]},"error":null,"meta":{"timestamp":"2024-01-15T10:30:00Z","backend":"sqlite","profile":"default","version":"1.0.0"}}`
	if string(got) != expected {
		t.Errorf("Marshal envelope = %s, want %s", got, expected)
	}
}

func TestEnvelope_Error(t *testing.T) {
	env := Envelope{
		OK:      false,
		Command: "gtd-cli task show",
		Data:    nil,
		Error: &ErrorBody{
			Code:    ErrNotFound,
			Message: "task not found",
			Details: map[string]any{"id": "tsk_01HXYZ"},
		},
		Meta: Meta{
			Timestamp: "2024-01-15T10:30:00Z",
			Backend:   "sqlite",
			Profile:   "default",
			Version:   "1.0.0",
		},
	}

	got, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	expected := `{"ok":false,"command":"gtd-cli task show","data":null,"error":{"code":"NOT_FOUND","message":"task not found","details":{"id":"tsk_01HXYZ"}},"meta":{"timestamp":"2024-01-15T10:30:00Z","backend":"sqlite","profile":"default","version":"1.0.0"}}`
	if string(got) != expected {
		t.Errorf("Marshal envelope = %s, want %s", got, expected)
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected int
	}{
		{"validation", ErrValidation, 2},
		{"not found", ErrNotFound, 3},
		{"conflict", ErrConflict, 4},
		{"io", ErrIO, 5},
		{"internal", ErrInternal, 1},
		{"unknown", ErrorCode("UNKNOWN"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExitCode(tt.code)
			if got != tt.expected {
				t.Errorf("ExitCode(%s) = %d, want %d", tt.code, got, tt.expected)
			}
		})
	}
}

func TestResponseWriter_Success(t *testing.T) {
	var buf bytes.Buffer
	clock := util.NewFixedClock(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
	rw := NewResponseWriter(&buf, clock, "sqlite", "default", "1.0.0", false)

	err := rw.WriteSuccess("gtd-cli task list", map[string]any{"count": 0})
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if env.OK != true {
		t.Errorf("env.OK = %v, want true", env.OK)
	}
	if env.Command != "gtd-cli task list" {
		t.Errorf("env.Command = %s, want gtd-cli task list", env.Command)
	}
	if env.Meta.Timestamp != "2024-01-15T10:30:00Z" {
		t.Errorf("env.Meta.Timestamp = %s, want 2024-01-15T10:30:00Z", env.Meta.Timestamp)
	}
}

func TestResponseWriter_Error(t *testing.T) {
	var buf bytes.Buffer
	clock := util.NewFixedClock(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
	rw := NewResponseWriter(&buf, clock, "sqlite", "default", "1.0.0", false)

	err := rw.WriteError("gtd-cli task show", ErrNotFound, "task not found", map[string]any{"id": "tsk_01HXYZ"})
	if err != nil {
		t.Fatalf("WriteError failed: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if env.OK != false {
		t.Errorf("env.OK = %v, want false", env.OK)
	}
	if env.Error == nil {
		t.Fatal("env.Error is nil")
	}
	if env.Error.Code != ErrNotFound {
		t.Errorf("env.Error.Code = %s, want NOT_FOUND", env.Error.Code)
	}
}

func TestResponseWriter_Pretty(t *testing.T) {
	var buf bytes.Buffer
	clock := util.NewFixedClock(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
	rw := NewResponseWriter(&buf, clock, "sqlite", "default", "1.0.0", true)

	err := rw.WriteSuccess("gtd-cli task list", map[string]any{"count": 0})
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\n  ") {
		t.Errorf("Pretty output should contain indentation, got: %s", output)
	}
}

func TestSuccess(t *testing.T) {
	env := Success("gtd-cli task list", map[string]any{"count": 0}, "sqlite", "default", "1.0.0")

	if !env.OK {
		t.Error("Success should return OK=true")
	}
	if env.Command != "gtd-cli task list" {
		t.Errorf("Command = %s, want gtd-cli task list", env.Command)
	}
	if env.Error != nil {
		t.Error("Error should be nil")
	}
	if env.Meta.Backend != "sqlite" {
		t.Errorf("Backend = %s, want sqlite", env.Meta.Backend)
	}
	if env.Meta.Profile != "default" {
		t.Errorf("Profile = %s, want default", env.Meta.Profile)
	}
	if env.Meta.Version != "1.0.0" {
		t.Errorf("Version = %s, want 1.0.0", env.Meta.Version)
	}
	if env.Meta.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
}

func TestFailure(t *testing.T) {
	env := Failure("gtd-cli task show", ErrNotFound, "task not found", map[string]any{"id": "tsk_123"}, "sqlite", "default", "1.0.0")

	if env.OK {
		t.Error("Failure should return OK=false")
	}
	if env.Command != "gtd-cli task show" {
		t.Errorf("Command = %s, want gtd-cli task show", env.Command)
	}
	if env.Data != nil {
		t.Error("Data should be nil")
	}
	if env.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if env.Error.Code != ErrNotFound {
		t.Errorf("Error.Code = %s, want NOT_FOUND", env.Error.Code)
	}
	if env.Error.Message != "task not found" {
		t.Errorf("Error.Message = %s, want 'task not found'", env.Error.Message)
	}
	if env.Error.Details == nil {
		t.Error("Error.Details should not be nil")
	}
}

func TestFailure_NilDetails(t *testing.T) {
	env := Failure("gtd-cli task show", ErrInternal, "internal error", nil, "json", "work", "2.0.0")

	if env.OK {
		t.Error("Failure should return OK=false")
	}
	if env.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if env.Error.Details != nil {
		t.Errorf("Error.Details should be nil, got %v", env.Error.Details)
	}
	if env.Meta.Backend != "json" {
		t.Errorf("Backend = %s, want json", env.Meta.Backend)
	}
}

func TestResponseWriter_NilClock(t *testing.T) {
	var buf bytes.Buffer
	rw := NewResponseWriter(&buf, nil, "sqlite", "default", "1.0.0", false)

	err := rw.WriteSuccess("test", nil)
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if !env.OK {
		t.Error("OK should be true")
	}
	if env.Meta.Timestamp == "" {
		t.Error("Timestamp should not be empty when clock is nil")
	}
}

func TestResponseWriter_WriteJSONError(t *testing.T) {
	rw := NewResponseWriter(&errorWriter{}, nil, "sqlite", "default", "1.0.0", false)

	err := rw.WriteSuccess("test", map[string]any{"key": "value"})
	if err == nil {
		t.Error("WriteSuccess should fail with error writer")
	}
}

type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write error")
}

func TestEnvelope_MarshalJSON_AllErrorCodes(t *testing.T) {
	codes := []ErrorCode{ErrValidation, ErrNotFound, ErrConflict, ErrIO, ErrInternal}

	for _, code := range codes {
		t.Run(string(code), func(t *testing.T) {
			env := Envelope{
				OK:      false,
				Command: "test",
				Error: &ErrorBody{
					Code:    code,
					Message: "error message",
				},
				Meta: Meta{
					Timestamp: "2024-01-15T10:30:00Z",
					Backend:   "sqlite",
					Profile:   "default",
					Version:   "1.0.0",
				},
			}

			data, err := json.Marshal(env)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			if !bytes.Contains(data, []byte(code)) {
				t.Errorf("Marshal should contain error code %s", code)
			}
		})
	}
}
