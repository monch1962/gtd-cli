package jsonout

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/anomalyco/gtd-cli/internal/util"
)

type ErrorCode string

const (
	ErrValidation ErrorCode = "VALIDATION"
	ErrNotFound   ErrorCode = "NOT_FOUND"
	ErrConflict   ErrorCode = "CONFLICT"
	ErrIO         ErrorCode = "IO"
	ErrInternal   ErrorCode = "INTERNAL"
)

type Meta struct {
	Timestamp string `json:"timestamp"`
	Backend   string `json:"backend"`
	Profile   string `json:"profile"`
	Version   string `json:"version"`
}

type ErrorBody struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type Envelope struct {
	OK      bool       `json:"ok"`
	Command string     `json:"command"`
	Data    any        `json:"data"`
	Error   *ErrorBody `json:"error"`
	Meta    Meta       `json:"meta"`
}

func ExitCode(code ErrorCode) int {
	switch code {
	case ErrValidation:
		return 2
	case ErrNotFound:
		return 3
	case ErrConflict:
		return 4
	case ErrIO:
		return 5
	default:
		return 1
	}
}

func Success(command string, data any, backend, profile, version string) Envelope {
	return Envelope{
		OK:      true,
		Command: command,
		Data:    data,
		Error:   nil,
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Backend:   backend,
			Profile:   profile,
			Version:   version,
		},
	}
}

func Failure(command string, code ErrorCode, message string, details map[string]any, backend, profile, version string) Envelope {
	return Envelope{
		OK:      false,
		Command: command,
		Data:    nil,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Backend:   backend,
			Profile:   profile,
			Version:   version,
		},
	}
}

type ResponseWriter struct {
	w       io.Writer
	clock   util.Clock
	backend string
	profile string
	version string
	pretty  bool
}

func NewResponseWriter(w io.Writer, clock util.Clock, backend, profile, version string, pretty bool) *ResponseWriter {
	if clock == nil {
		clock = util.RealClock{}
	}
	return &ResponseWriter{
		w:       w,
		clock:   clock,
		backend: backend,
		profile: profile,
		version: version,
		pretty:  pretty,
	}
}

func (rw *ResponseWriter) WriteSuccess(command string, data any) error {
	env := Envelope{
		OK:      true,
		Command: command,
		Data:    data,
		Error:   nil,
		Meta: Meta{
			Timestamp: rw.clock.Now().UTC().Format(time.RFC3339),
			Backend:   rw.backend,
			Profile:   rw.profile,
			Version:   rw.version,
		},
	}
	return rw.writeJSON(env)
}

func (rw *ResponseWriter) WriteError(command string, code ErrorCode, message string, details map[string]any) error {
	env := Envelope{
		OK:      false,
		Command: command,
		Data:    nil,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: Meta{
			Timestamp: rw.clock.Now().UTC().Format(time.RFC3339),
			Backend:   rw.backend,
			Profile:   rw.profile,
			Version:   rw.version,
		},
	}
	return rw.writeJSON(env)
}

func (rw *ResponseWriter) writeJSON(v any) error {
	if rw.pretty {
		enc := json.NewEncoder(rw.w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	_, err = rw.w.Write(append(data, '\n'))
	return err
}
