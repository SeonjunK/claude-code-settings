package telegram

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
)

// FormatEvent formats a notification event as an HTML message for Telegram.
func FormatEvent(event notification.Event) string {
	var b strings.Builder

	fmt.Fprintf(&b, "<b>%s %s</b>\n", event.Type.Emoji(), html.EscapeString(event.Type.Label()))

	if event.Summary != "" {
		b.WriteString(html.EscapeString(event.Summary))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if event.SessionID != "" {
		fmt.Fprintf(&b, "<b>Session:</b> <code>%s</code>\n", html.EscapeString(shortID(event.SessionID)))
	}

	if event.Goal != "" {
		fmt.Fprintf(&b, "<b>Goal:</b> %s\n", html.EscapeString(event.Goal))
	}

	if event.TeamName != "" {
		fmt.Fprintf(&b, "<b>Team:</b> %s\n", html.EscapeString(event.TeamName))
	}

	if event.MaxIter > 0 {
		fmt.Fprintf(&b, "<b>Iteration:</b> %d/%d\n", event.Iteration, event.MaxIter)
	}

	if event.Metrics != nil {
		m := event.Metrics
		fmt.Fprintf(&b, "\n<b>Tasks:</b> %d done, %d active, %d pending\n",
			m.TasksCompleted, m.TasksInProgress, m.TasksPending)
		if m.TotalToolCalls > 0 {
			fmt.Fprintf(&b, "<b>Tool calls:</b> %d\n", m.TotalToolCalls)
		}
	}

	if len(event.Details) > 0 {
		b.WriteString("\n")
		for k, v := range event.Details {
			fmt.Fprintf(&b, "<b>%s:</b> %s\n", html.EscapeString(k), html.EscapeString(v))
		}
	}

	fmt.Fprintf(&b, "\n<i>%s</i>", event.Timestamp.Format(time.RFC3339))

	return b.String()
}

// FormatSessionStatus formats a session status as an HTML message.
func FormatSessionStatus(
	sessionID, teamName, goal string,
	active bool,
	iteration, maxIter int,
	metrics *notification.MetricsSnapshot,
	teammates []TeammateInfo,
) string {
	var b strings.Builder

	status := "\u2705 Active"
	if !active {
		status = "\u23F9 Stopped"
	}

	fmt.Fprintf(&b, "<b>\U0001F4CA Session Status</b> %s\n\n", status)
	fmt.Fprintf(&b, "<b>ID:</b> <code>%s</code>\n", html.EscapeString(shortID(sessionID)))
	fmt.Fprintf(&b, "<b>Team:</b> %s\n", html.EscapeString(teamName))
	fmt.Fprintf(&b, "<b>Goal:</b> %s\n", html.EscapeString(goal))
	fmt.Fprintf(&b, "<b>Iteration:</b> %d/%d\n", iteration, maxIter)

	if metrics != nil {
		fmt.Fprintf(&b, "\n<b>Tasks:</b> %d done, %d active, %d pending\n",
			metrics.TasksCompleted, metrics.TasksInProgress, metrics.TasksPending)
		fmt.Fprintf(&b, "<b>Tool calls:</b> %d\n", metrics.TotalToolCalls)
	}

	if len(teammates) > 0 {
		fmt.Fprintf(&b, "\n<b>Teammates (%d):</b>\n", len(teammates))
		for _, t := range teammates {
			statusIcon := "\u23F3" // hourglass
			switch t.Status {
			case "in_progress":
				statusIcon = "\U0001F7E2" // green circle
			case "completed":
				statusIcon = "\u2705"
			case "pending":
				statusIcon = "\u23F3"
			}
			fmt.Fprintf(&b, "  %s %s (%s)\n", statusIcon, html.EscapeString(t.Name), html.EscapeString(t.Status))
		}
	}

	return b.String()
}

// FormatNoActiveSession returns a message for when no active session is found.
func FormatNoActiveSession() string {
	return "\U0001F4AD No active session found."
}

// FormatHelp returns the bot help message.
func FormatHelp() string {
	return `<b>🤖 Claude Code Hooks Bot</b>

<b>Commands:</b>
/status — Current session status
/tasks — Task list summary
/stop — Stop active session
/help — Show this help

<i>Notifications are sent automatically on hook events.</i>`
}

// TeammateInfo holds teammate display information.
type TeammateInfo struct {
	Name   string
	Status string
}

// shortID truncates a UUID to the first 8 characters.
func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
