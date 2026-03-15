package hook

import (
	"testing"
	"time"
)

func TestParseStopInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid input",
			input:   `{"session_id":"test-session","timestamp":"2026-03-14T00:00:00Z"}`,
			wantErr: false,
		},
		{
			name:    "missing session_id",
			input:   `{"timestamp":"2026-03-14T00:00:00Z"}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{not valid json`,
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   `{}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseStopInput([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("result should not be nil")
			}
		})
	}
}

func TestParseUserPromptSubmitInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid input with prompt",
			input:   `{"session_id":"test","prompt":"/team-loops test goal","timestamp":"2026-03-14T00:00:00Z"}`,
			wantErr: false,
		},
		{
			name:    "prompt only",
			input:   `{"prompt":"hello world"}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{broken`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseUserPromptSubmitInput([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("result should not be nil")
			}
		})
	}
}

func TestOutputFormats(t *testing.T) {
	t.Run("AllowOutput", func(t *testing.T) {
		output := AllowOutput()
		if output != `{"action":"allow"}` {
			t.Errorf("AllowOutput() = %q, want %q", output, `{"action":"allow"}`)
		}
	})

	t.Run("BlockOutput", func(t *testing.T) {
		message := "session active"
		output := BlockOutput(message)
		expected := `{"action":"block","message":"session active"}`
		if output != expected {
			t.Errorf("BlockOutput(%q) = %q, want %q", message, output, expected)
		}
	})

	t.Run("ErrorOutput", func(t *testing.T) {
		message := "something went wrong"
		output := ErrorOutput(message)
		expected := `{"action":"error","message":"something went wrong"}`
		if output != expected {
			t.Errorf("ErrorOutput(%q) = %q, want %q", message, output, expected)
		}
	})
}

func TestStopOutputSchema(t *testing.T) {
	output := &StopOutput{
		Action:      "block",
		Message:     "session active",
		SessionID:   "test-session",
		SessionPath: "/path/to/session.md",
	}

	jsonBytes, err := output.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Verify JSON contains expected fields
	if output.Action != "block" {
		t.Errorf("Action = %q, want %q", output.Action, "block")
	}
	if output.SessionID != "test-session" {
		t.Errorf("SessionID = %q, want %q", output.SessionID, "test-session")
	}

	t.Logf("StopOutput JSON: %s", string(jsonBytes))
}

func TestUserPromptSubmitInputSchema(t *testing.T) {
	input := &UserPromptSubmitInput{
		SessionID:      "session-123",
		TranscriptPath: "/path/to/transcript",
		Prompt:         "/team-loops test goal",
		Timestamp:      time.Now(),
	}

	if input.SessionID != "session-123" {
		t.Errorf("SessionID = %q, want %q", input.SessionID, "session-123")
	}
	if input.Prompt != "/team-loops test goal" {
		t.Errorf("Prompt = %q, want %q", input.Prompt, "/team-loops test goal")
	}
}
