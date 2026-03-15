// Package notification provides notification event types.
package notification

import "time"

// EventType represents the type of notification event.
type EventType string

const (
	EventSessionStart EventType = "session_start"
	EventSessionStop  EventType = "session_stop"
	EventIteration    EventType = "iteration"
	EventVerifyPass   EventType = "verify_pass"
	EventVerifyFail   EventType = "verify_fail"
	EventTaskComplete EventType = "task_complete"
	EventTaskBlocked  EventType = "task_blocked"
	EventGuardDeny    EventType = "guard_deny"
)

// AllEventTypes returns all known event types.
func AllEventTypes() []EventType {
	return []EventType{
		EventSessionStart, EventSessionStop, EventIteration,
		EventVerifyPass, EventVerifyFail,
		EventTaskComplete, EventTaskBlocked,
		EventGuardDeny,
	}
}

// Event is the canonical notification payload passed to all providers.
type Event struct {
	Type      EventType         `json:"type"`
	SessionID string            `json:"session_id"`
	Timestamp time.Time         `json:"timestamp"`
	Summary   string            `json:"summary"`
	Details   map[string]string `json:"details,omitempty"`
	Goal      string            `json:"goal,omitempty"`
	Iteration int               `json:"iteration,omitempty"`
	MaxIter   int               `json:"max_iterations,omitempty"`
	TeamName  string            `json:"team_name,omitempty"`
	Metrics   *MetricsSnapshot  `json:"metrics,omitempty"`
}

// MetricsSnapshot is a point-in-time copy of session metrics.
type MetricsSnapshot struct {
	TasksCompleted  int `json:"tasks_completed"`
	TasksPending    int `json:"tasks_pending"`
	TasksInProgress int `json:"tasks_in_progress"`
	TotalToolCalls  int `json:"total_tool_calls"`
}

// Emoji returns an emoji prefix for the event type.
func (t EventType) Emoji() string {
	switch t {
	case EventSessionStart:
		return "\U0001F4CB" // clipboard
	case EventSessionStop:
		return "\U0001F3C1" // checkered flag
	case EventIteration:
		return "\U0001F504" // counterclockwise arrows
	case EventVerifyPass:
		return "\u2705" // check mark
	case EventVerifyFail:
		return "\u274C" // cross mark
	case EventTaskComplete:
		return "\u2714\uFE0F" // heavy check
	case EventTaskBlocked:
		return "\U0001F6A7" // construction
	case EventGuardDeny:
		return "\U0001F6E1\uFE0F" // shield
	default:
		return "\U0001F514" // bell
	}
}

// Label returns a human-readable label for the event type.
func (t EventType) Label() string {
	switch t {
	case EventSessionStart:
		return "Session Started"
	case EventSessionStop:
		return "Session Stopped"
	case EventIteration:
		return "Iteration"
	case EventVerifyPass:
		return "Verify Passed"
	case EventVerifyFail:
		return "Verify Failed"
	case EventTaskComplete:
		return "Task Completed"
	case EventTaskBlocked:
		return "Task Blocked"
	case EventGuardDeny:
		return "Guard Denied"
	default:
		return string(t)
	}
}
