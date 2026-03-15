// Package session provides session management for team-loops.
package session

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Goal represents the session goal.
type Goal struct {
	Description      string   `yaml:"description"`
	SourceDoc        string   `yaml:"source_doc"`
	DefinitionOfDone []string `yaml:"definition_of_done,omitempty"`
}

// Teammate represents a teammate in the session.
type Teammate struct {
	Name         string    `yaml:"name"`
	SubagentType string    `yaml:"subagent_type"`
	Status       string    `yaml:"status"`
	TaskIDs      []string  `yaml:"task_ids,omitempty"`
	SpawnedAt    time.Time `yaml:"spawned_at"`
}

// Metrics represents session metrics.
type Metrics struct {
	TasksCompleted         int        `yaml:"tasks_completed"`
	TasksPending           int        `yaml:"tasks_pending"`
	TasksInProgress        int        `yaml:"tasks_in_progress"`
	TotalToolCalls         int        `yaml:"total_tool_calls"`
	AvgTaskDurationSeconds *float64   `yaml:"avg_task_duration_seconds,omitempty"`
	LastActivityAt         *time.Time `yaml:"last_activity_at,omitempty"`
}

// Control represents session control settings.
type Control struct {
	AutoSpawnTeammates   bool `yaml:"auto_spawn_teammates"`
	MaxParallelTeammates int  `yaml:"max_parallel_teammates"`
}

// Session represents a team-loops session.
type Session struct {
	Active            bool       `yaml:"active"`
	Iteration         int        `yaml:"iteration"`
	MaxIterations     int        `yaml:"max_iterations"`
	CompletionPromise string     `yaml:"completion_promise"`
	StartedAt         time.Time  `yaml:"started_at"`
	SessionID         string     `yaml:"session_id"`
	TeamName          string     `yaml:"team_name"`
	Goal              Goal       `yaml:"goal"`
	Teammates         []Teammate `yaml:"teammates,omitempty"`
	Metrics           Metrics    `yaml:"metrics"`
	Control           Control    `yaml:"control"`
	Body              string     `yaml:"-"`
}

// Default values for session.
const (
	DefaultMaxIterations        = 20
	DefaultCompletionPromise    = "QUEUE_EMPTY_ALL_TASKS_COMPLETED"
	DefaultMaxParallelTeammates = 3
)

// Option is a functional option for creating sessions.
type Option func(*Session)

// WithMaxIterations sets the maximum iterations.
func WithMaxIterations(max int) Option {
	return func(s *Session) {
		s.MaxIterations = max
	}
}

// WithTeamName sets the team name.
func WithTeamName(name string) Option {
	return func(s *Session) {
		s.TeamName = name
	}
}

// WithSourceDoc sets the source document.
func WithSourceDoc(doc string) Option {
	return func(s *Session) {
		s.Goal.SourceDoc = doc
	}
}

// WithControl sets the control settings.
func WithControl(control Control) Option {
	return func(s *Session) {
		s.Control = control
	}
}

// WithMaxParallelTeammates sets the maximum parallel teammates.
func WithMaxParallelTeammates(max int) Option {
	return func(s *Session) {
		s.Control.MaxParallelTeammates = max
	}
}

// safeSlice safely slices a string to the given length.
func safeSlice(s string, maxLen int) string {
	if len(s) < maxLen {
		return s
	}
	return s[:maxLen]
}

// CreateSession creates a new session with the given goal and options.
func CreateSession(goal, sessionID string, opts ...Option) *Session {
	now := time.Now()

	s := &Session{
		Active:            true,
		Iteration:         1,
		MaxIterations:     DefaultMaxIterations,
		CompletionPromise: DefaultCompletionPromise,
		StartedAt:         now,
		SessionID:         sessionID,
		TeamName:          "team-loops-" + safeSlice(sessionID, 8),
		Goal: Goal{
			Description: goal,
		},
		Teammates: []Teammate{},
		Metrics: Metrics{
			TasksCompleted:  0,
			TasksPending:    0,
			TasksInProgress: 0,
			TotalToolCalls:  0,
		},
		Control: Control{
			AutoSpawnTeammates:   true,
			MaxParallelTeammates: DefaultMaxParallelTeammates,
		},
		Body: generateDefaultBody(goal),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// generateDefaultBody generates the default markdown body for a session.
func generateDefaultBody(goal string) string {
	return fmt.Sprintf(`목표: %s

## 규칙
- claim한 task만 수행
- 막히면 blockedBy 작성
- 끝나면 completed 전환

<promise>%s</promise>`, goal, DefaultCompletionPromise)
}

// frontmatterRegex matches YAML frontmatter in a markdown file.
var frontmatterRegex = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)$`)

// LoadSession loads a session from a file.
func LoadSession(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	return ParseSession(data)
}

// ParseSession parses session data from bytes.
func ParseSession(data []byte) (*Session, error) {
	matches := frontmatterRegex.FindSubmatch(data)
	if matches == nil {
		return nil, fmt.Errorf("invalid session file format: missing frontmatter")
	}

	var s Session
	if err := yaml.Unmarshal(matches[1], &s); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	s.Body = strings.TrimSpace(string(matches[2]))

	return &s, nil
}

// SaveSession saves a session to a file.
func SaveSession(path string, s *Session) error {
	data, err := SerializeSession(s)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// SerializeSession serializes a session to bytes.
func SerializeSession(s *Session) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("---\n")

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(s); err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}
	encoder.Close()

	buf.WriteString("---\n")
	buf.WriteString(s.Body)
	buf.WriteString("\n")

	return buf.Bytes(), nil
}

// AddTeammate adds a new teammate to the session.
func (s *Session) AddTeammate(name, subagentType string, taskIDs []string) {
	s.Teammates = append(s.Teammates, Teammate{
		Name:         name,
		SubagentType: subagentType,
		Status:       "pending",
		TaskIDs:      taskIDs,
		SpawnedAt:    time.Now(),
	})
}

// UpdateTeammateStatus updates the status of a teammate.
func (s *Session) UpdateTeammateStatus(name, status string) {
	for i := range s.Teammates {
		if s.Teammates[i].Name == name {
			s.Teammates[i].Status = status
			break
		}
	}
}

// GetActiveTeammates returns all active teammates.
func (s *Session) GetActiveTeammates() []Teammate {
	var active []Teammate
	for _, t := range s.Teammates {
		if t.Status == "in_progress" || t.Status == "pending" {
			active = append(active, t)
		}
	}
	return active
}

// CanSpawnTeammate checks if a new teammate can be spawned.
func (s *Session) CanSpawnTeammate() bool {
	return len(s.GetActiveTeammates()) < s.Control.MaxParallelTeammates
}

// IncrementIteration increments the iteration counter.
func (s *Session) IncrementIteration() {
	s.Iteration++
}

// IsComplete checks if the session is complete.
func (s *Session) IsComplete() bool {
	return !s.Active || s.Iteration > s.MaxIterations
}

// UpdateMetrics updates session metrics.
func (s *Session) UpdateMetrics(completed, pending, inProgress int) {
	s.Metrics.TasksCompleted = completed
	s.Metrics.TasksPending = pending
	s.Metrics.TasksInProgress = inProgress
	now := time.Now()
	s.Metrics.LastActivityAt = &now
}

// IncrementToolCalls increments the tool call counter.
func (s *Session) IncrementToolCalls() {
	s.Metrics.TotalToolCalls++
}
