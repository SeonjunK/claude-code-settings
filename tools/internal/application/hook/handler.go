// Package hook provides hook handlers for Claude Code hooks.
package hook

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/domain/session"
	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/infrastructure/storage"
)

// StopHandler handles stop hooks.
type StopHandler struct {
	sessionsDir string
	fileClient  *storage.SessionFileClient
}

// NewStopHandler creates a new stop handler.
func NewStopHandler(sessionsDir string) *StopHandler {
	return &StopHandler{
		sessionsDir: sessionsDir,
		fileClient:  storage.NewSessionFileClient(),
	}
}

// Handle processes stop hook input.
func (h *StopHandler) Handle(input *StopInput) (*StopOutput, error) {
	// Find session file
	sessionPath := filepath.Join(h.sessionsDir, input.SessionID+".local.md")

	// Check if session exists
	if !h.fileClient.Exists(sessionPath) {
		// Session doesn't exist - allow stop
		return &StopOutput{Action: "allow"}, nil
	}

	// Load session
	sess, err := session.LoadSession(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Check if session is active
	if !sess.Active {
		// Session already stopped - allow stop
		return &StopOutput{Action: "allow"}, nil
	}

	// Increment iteration for next round
	sess.IncrementIteration()

	// Check if we should continue or stop
	if sess.IsComplete() {
		// Session is complete - allow stop
		sess.Active = false
		if err := session.SaveSession(sessionPath, sess); err != nil {
			return nil, fmt.Errorf("failed to save session: %w", err)
		}
		return &StopOutput{Action: "allow"}, nil
	}

	// Block stop and continue session
	if err := session.SaveSession(sessionPath, sess); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return &StopOutput{
		Action:      "block",
		Message:     fmt.Sprintf("team-loops active · 반복: %d/%d", sess.Iteration, sess.MaxIterations),
		SessionID:   sess.SessionID,
		SessionPath: sessionPath,
	}, nil
}

// UserPromptSubmitHandler handles user-prompt-submit hooks.
type UserPromptSubmitHandler struct {
	sessionsDir string
	fileClient  *storage.SessionFileClient
}

// NewUserPromptSubmitHandler creates a new user-prompt-submit handler.
func NewUserPromptSubmitHandler(sessionsDir string) *UserPromptSubmitHandler {
	return &UserPromptSubmitHandler{
		sessionsDir: sessionsDir,
		fileClient:  storage.NewSessionFileClient(),
	}
}

// Handle processes user-prompt-submit hook input.
func (h *UserPromptSubmitHandler) Handle(input *UserPromptSubmitInput, generateSessionID func() string) (*UserPromptSubmitOutput, error) {
	// Check if this is a /team-loops command
	if !isTeamLoopsCommand(input.Prompt) {
		// Not a team-loops command - allow
		return &UserPromptSubmitOutput{Action: "allow"}, nil
	}

	// Extract the goal from the prompt
	goal := extractTeamLoopsGoal(input.Prompt)

	// Create or update session
	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	sessionPath := filepath.Join(h.sessionsDir, sessionID+".local.md")

	// Ensure directory exists
	if err := h.fileClient.EnsureDir(h.sessionsDir); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	// Check if session already exists
	sess, err := session.LoadSession(sessionPath)
	if err != nil {
		// Create new session
		sess = session.CreateSession(goal, sessionID)
	} else {
		// Update existing session
		sess.Goal.Description = goal
	}

	if err := session.SaveSession(sessionPath, sess); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Allow the command to proceed
	return &UserPromptSubmitOutput{Action: "allow"}, nil
}

// isTeamLoopsCommand checks if the prompt is a /team-loops command.
func isTeamLoopsCommand(prompt string) bool {
	return strings.HasPrefix(prompt, "/team-loops") &&
		(len(prompt) == len("/team-loops") || prompt[len("/team-loops")] == ' ')
}

// extractTeamLoopsGoal extracts the goal from a /team-loops prompt.
func extractTeamLoopsGoal(prompt string) string {
	if len(prompt) <= len("/team-loops") {
		return ""
	}
	goal := prompt[len("/team-loops"):]
	return strings.TrimSpace(goal)
}
