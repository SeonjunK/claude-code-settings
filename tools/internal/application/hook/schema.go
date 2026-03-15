// Package hook provides hook schema definitions and handlers.
package hook

import (
	"encoding/json"
	"fmt"
	"time"
)

// Type represents the type of hook.
type Type string

const (
	// TypeStop is triggered when Claude Code session stops.
	TypeStop Type = "stop"
	// TypeUserPromptSubmit is triggered when user submits a prompt.
	TypeUserPromptSubmit Type = "user-prompt-submit"
	// TypeSessionStart is triggered when a session starts.
	TypeSessionStart Type = "session-start"
	// TypePreToolUse is triggered before a tool is used.
	TypePreToolUse Type = "pre-tool-use"
	// TypePostToolUse is triggered after a tool is used.
	TypePostToolUse Type = "post-tool-use"
)

// --- Stop Hook ---

// StopInput represents the input for stop hook.
type StopInput struct {
	SessionID      string    `json:"session_id"`
	TranscriptPath string    `json:"transcript_path"`
	StopReason     string    `json:"stop_reason,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// StopOutput represents the output for stop hook.
type StopOutput struct {
	Action      string `json:"action"` // "block", "allow", "error"
	Message     string `json:"message,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
	SessionPath string `json:"session_path,omitempty"`
}

// ParseStopInput parses JSON data into StopInput.
func ParseStopInput(data []byte) (*StopInput, error) {
	var input StopInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse stop input: %w", err)
	}
	return &input, nil
}

// ToJSON serializes StopOutput to JSON.
func (o *StopOutput) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// --- User-Prompt-Submit Hook ---

// UserPromptSubmitInput represents the input for user-prompt-submit hook.
type UserPromptSubmitInput struct {
	SessionID      string    `json:"session_id"`
	TranscriptPath string    `json:"transcript_path"`
	Prompt         string    `json:"prompt"`
	Timestamp      time.Time `json:"timestamp"`
}

// UserPromptSubmitOutput represents the output for user-prompt-submit hook.
type UserPromptSubmitOutput struct {
	Action    string `json:"action"` // "block", "allow", "transform", "error"
	Message   string `json:"message,omitempty"`
	Transform string `json:"transform,omitempty"` // transformed prompt if action is "transform"
}

// ParseUserPromptSubmitInput parses JSON data into UserPromptSubmitInput.
func ParseUserPromptSubmitInput(data []byte) (*UserPromptSubmitInput, error) {
	var input UserPromptSubmitInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse user-prompt-submit input: %w", err)
	}
	return &input, nil
}

// ToJSON serializes UserPromptSubmitOutput to JSON.
func (o *UserPromptSubmitOutput) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// --- Session-Start Hook ---

// SessionStartInput represents the input for session-start hook.
type SessionStartInput struct {
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
}

// SessionStartOutput represents the output for session-start hook.
type SessionStartOutput struct {
	Action  string `json:"action"` // "block", "allow", "error"
	Message string `json:"message,omitempty"`
}

// --- Pre-Tool-Use Hook ---

// PreToolUseInput represents the input for pre-tool-use hook.
type PreToolUseInput struct {
	SessionID      string                 `json:"session_id"`
	ToolName       string                 `json:"tool_name"`
	ToolInput      map[string]interface{} `json:"tool_input"`
	TranscriptPath string                 `json:"transcript_path"`
	Timestamp      time.Time              `json:"timestamp"`
}

// PreToolUseOutput represents the output for pre-tool-use hook.
type PreToolUseOutput struct {
	Action  string `json:"action"` // "block", "allow", "error"
	Message string `json:"message,omitempty"`
}

// --- Post-Tool-Use Hook ---

// PostToolUseInput represents the input for post-tool-use hook.
type PostToolUseInput struct {
	SessionID      string                 `json:"session_id"`
	ToolName       string                 `json:"tool_name"`
	ToolInput      map[string]interface{} `json:"tool_input"`
	ToolOutput     interface{}            `json:"tool_output"`
	ToolError      string                 `json:"tool_error,omitempty"`
	TranscriptPath string                 `json:"transcript_path"`
	Timestamp      time.Time              `json:"timestamp"`
}

// PostToolUseOutput represents the output for post-tool-use hook.
type PostToolUseOutput struct {
	Action  string `json:"action"` // "block", "allow", "error"
	Message string `json:"message,omitempty"`
}

// --- Common Output Helpers ---

// AllowOutput creates an allow action output.
func AllowOutput() string {
	return `{"action":"allow"}`
}

// BlockOutput creates a block action output.
func BlockOutput(message string) string {
	return fmt.Sprintf(`{"action":"block","message":%q}`, message)
}

// ErrorOutput creates an error action output.
func ErrorOutput(message string) string {
	return fmt.Sprintf(`{"action":"error","message":%q}`, message)
}

// ToJSONBytes serializes any value to JSON bytes.
func ToJSONBytes(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize to JSON: %w", err)
	}
	return data, nil
}

// ToJSONString serializes any value to JSON string.
func ToJSONString(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to serialize to JSON: %w", err)
	}
	return string(data), nil
}

// --- Claude Code Official Schema Output Helpers ---

// StopBlockOutput creates a Stop/SubagentStop block output conforming to Claude Code hook schema.
func StopBlockOutput(reason, message string) map[string]interface{} {
	return map[string]interface{}{
		"decision":      "block",
		"reason":        reason,
		"systemMessage": message,
	}
}

// SystemMessageOutput creates a systemMessage-only output (usable with any hook event).
func SystemMessageOutput(message string) map[string]interface{} {
	return map[string]interface{}{
		"systemMessage": message,
	}
}

// SessionContextOutput creates a SessionStart additionalContext output.
func SessionContextOutput(context string) map[string]interface{} {
	return map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":     "SessionStart",
			"additionalContext": context,
		},
	}
}
