package telegram

import (
	"strings"
	"testing"
	"time"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
)

func TestFormatEvent(t *testing.T) {
	fixedTime := time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		event    notification.Event
		contains []string
	}{
		{
			name: "basic event with summary",
			event: notification.Event{
				Type:      notification.EventSessionStart,
				SessionID: "abcd1234-5678-9abc-def0-123456789abc",
				Timestamp: fixedTime,
				Summary:   "New session started",
			},
			contains: []string{
				"<b>",
				"Session Started",
				"New session started",
				"<code>abcd1234</code>",
				fixedTime.Format(time.RFC3339),
			},
		},
		{
			name: "event with goal",
			event: notification.Event{
				Type:      notification.EventTaskComplete,
				SessionID: "sess-1234-long-id",
				Timestamp: fixedTime,
				Summary:   "task done",
				Goal:      "Implement feature X",
			},
			contains: []string{
				"Task Completed",
				"<b>Goal:</b> Implement feature X",
			},
		},
		{
			name: "event with iteration",
			event: notification.Event{
				Type:      notification.EventIteration,
				SessionID: "sess-1234",
				Timestamp: fixedTime,
				Iteration: 3,
				MaxIter:   10,
			},
			contains: []string{
				"Iteration",
				"<b>Iteration:</b> 3/10",
			},
		},
		{
			name: "event with team name",
			event: notification.Event{
				Type:      notification.EventSessionStart,
				SessionID: "sess-1234",
				Timestamp: fixedTime,
				TeamName:  "backend-team",
			},
			contains: []string{
				"<b>Team:</b> backend-team",
			},
		},
		{
			name: "event with metrics",
			event: notification.Event{
				Type:      notification.EventSessionStop,
				SessionID: "sess-1234",
				Timestamp: fixedTime,
				Metrics: &notification.MetricsSnapshot{
					TasksCompleted:  5,
					TasksInProgress: 2,
					TasksPending:    3,
					TotalToolCalls:  42,
				},
			},
			contains: []string{
				"<b>Tasks:</b> 5 done, 2 active, 3 pending",
				"<b>Tool calls:</b> 42",
			},
		},
		{
			name: "event with details",
			event: notification.Event{
				Type:      notification.EventGuardDeny,
				SessionID: "sess-1234",
				Timestamp: fixedTime,
				Details: map[string]string{
					"reason": "blocked by policy",
				},
			},
			contains: []string{
				"Guard Denied",
				"<b>reason:</b> blocked by policy",
			},
		},
		{
			name: "event with HTML special characters",
			event: notification.Event{
				Type:      notification.EventTaskComplete,
				SessionID: "sess-1234",
				Timestamp: fixedTime,
				Summary:   "Fix <script>alert('xss')</script>",
				Goal:      "Handle & escape < > chars",
			},
			contains: []string{
				"&lt;script&gt;",
				"&amp;",
			},
		},
		{
			name: "minimal event",
			event: notification.Event{
				Type:      notification.EventSessionStart,
				Timestamp: fixedTime,
			},
			contains: []string{
				"Session Started",
				fixedTime.Format(time.RFC3339),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatEvent(tt.event)

			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("FormatEvent() output missing %q\ngot:\n%s", substr, got)
				}
			}
		})
	}
}

func TestFormatEvent_ShortSessionID(t *testing.T) {
	event := notification.Event{
		Type:      notification.EventSessionStart,
		SessionID: "abc",
		Timestamp: time.Now(),
	}

	got := FormatEvent(event)
	if !strings.Contains(got, "<code>abc</code>") {
		t.Errorf("FormatEvent() should show short ID as-is; got:\n%s", got)
	}
}

func TestFormatEvent_MetricsZeroToolCalls(t *testing.T) {
	event := notification.Event{
		Type:      notification.EventSessionStop,
		SessionID: "sess-1234",
		Timestamp: time.Now(),
		Metrics: &notification.MetricsSnapshot{
			TasksCompleted: 1,
			TotalToolCalls: 0,
		},
	}

	got := FormatEvent(event)
	if strings.Contains(got, "Tool calls:") {
		t.Error("FormatEvent() should not show 'Tool calls' when TotalToolCalls is 0")
	}
}

func TestFormatSessionStatus(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		teamName  string
		goal      string
		active    bool
		iteration int
		maxIter   int
		metrics   *notification.MetricsSnapshot
		teammates []TeammateInfo
		contains  []string
	}{
		{
			name:      "active session",
			sessionID: "session-12345678-abcd",
			teamName:  "my-team",
			goal:      "Build something",
			active:    true,
			iteration: 3,
			maxIter:   10,
			contains: []string{
				"Session Status",
				"\u2705 Active",
				"<code>session-</code>",
				"my-team",
				"Build something",
				"3/10",
			},
		},
		{
			name:      "stopped session",
			sessionID: "session-xyz",
			teamName:  "team-b",
			goal:      "Done",
			active:    false,
			iteration: 5,
			maxIter:   5,
			contains: []string{
				"\u23F9 Stopped",
			},
		},
		{
			name:      "with metrics",
			sessionID: "sess-1234",
			teamName:  "t",
			goal:      "g",
			active:    true,
			iteration: 1,
			maxIter:   1,
			metrics: &notification.MetricsSnapshot{
				TasksCompleted:  2,
				TasksInProgress: 1,
				TasksPending:    0,
				TotalToolCalls:  15,
			},
			contains: []string{
				"2 done, 1 active, 0 pending",
				"Tool calls:</b> 15",
			},
		},
		{
			name:      "with teammates",
			sessionID: "sess-1234",
			teamName:  "t",
			goal:      "g",
			active:    true,
			iteration: 1,
			maxIter:   1,
			teammates: []TeammateInfo{
				{Name: "worker-1", Status: "in_progress"},
				{Name: "worker-2", Status: "completed"},
				{Name: "worker-3", Status: "pending"},
			},
			contains: []string{
				"Teammates (3)",
				"worker-1",
				"worker-2",
				"worker-3",
				"\U0001F7E2", // green circle for in_progress
				"\u2705",     // check for completed
				"\u23F3",     // hourglass for pending
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSessionStatus(
				tt.sessionID, tt.teamName, tt.goal,
				tt.active, tt.iteration, tt.maxIter,
				tt.metrics, tt.teammates,
			)

			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("FormatSessionStatus() missing %q\ngot:\n%s", substr, got)
				}
			}
		})
	}
}

func TestFormatNoActiveSession(t *testing.T) {
	got := FormatNoActiveSession()

	if !strings.Contains(got, "No active session found") {
		t.Errorf("FormatNoActiveSession() = %q, want to contain 'No active session found'", got)
	}
}

func TestFormatHelp(t *testing.T) {
	got := FormatHelp()

	expectedSubstrings := []string{
		"Claude Code Hooks Bot",
		"/status",
		"/tasks",
		"/stop",
		"/help",
		"Notifications are sent automatically",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(got, substr) {
			t.Errorf("FormatHelp() missing %q\ngot:\n%s", substr, got)
		}
	}
}
