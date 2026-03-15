package notification

import (
	"slices"
	"testing"
)

func TestEventType_Emoji(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		want      string
	}{
		{name: "session_start", eventType: EventSessionStart, want: "\U0001F4CB"},
		{name: "session_stop", eventType: EventSessionStop, want: "\U0001F3C1"},
		{name: "iteration", eventType: EventIteration, want: "\U0001F504"},
		{name: "verify_pass", eventType: EventVerifyPass, want: "\u2705"},
		{name: "verify_fail", eventType: EventVerifyFail, want: "\u274C"},
		{name: "task_complete", eventType: EventTaskComplete, want: "\u2714\uFE0F"},
		{name: "task_blocked", eventType: EventTaskBlocked, want: "\U0001F6A7"},
		{name: "guard_deny", eventType: EventGuardDeny, want: "\U0001F6E1\uFE0F"},
		{name: "unknown returns bell", eventType: EventType("unknown"), want: "\U0001F514"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.eventType.Emoji()
			if got != tt.want {
				t.Errorf("EventType(%q).Emoji() = %q, want %q", tt.eventType, got, tt.want)
			}
		})
	}
}

func TestEventType_Label(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		want      string
	}{
		{name: "session_start", eventType: EventSessionStart, want: "Session Started"},
		{name: "session_stop", eventType: EventSessionStop, want: "Session Stopped"},
		{name: "iteration", eventType: EventIteration, want: "Iteration"},
		{name: "verify_pass", eventType: EventVerifyPass, want: "Verify Passed"},
		{name: "verify_fail", eventType: EventVerifyFail, want: "Verify Failed"},
		{name: "task_complete", eventType: EventTaskComplete, want: "Task Completed"},
		{name: "task_blocked", eventType: EventTaskBlocked, want: "Task Blocked"},
		{name: "guard_deny", eventType: EventGuardDeny, want: "Guard Denied"},
		{name: "unknown returns raw string", eventType: EventType("custom_event"), want: "custom_event"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.eventType.Label()
			if got != tt.want {
				t.Errorf("EventType(%q).Label() = %q, want %q", tt.eventType, got, tt.want)
			}
		})
	}
}

func TestAllEventTypes_Completeness(t *testing.T) {
	all := AllEventTypes()

	expected := []EventType{
		EventSessionStart, EventSessionStop, EventIteration,
		EventVerifyPass, EventVerifyFail,
		EventTaskComplete, EventTaskBlocked,
		EventGuardDeny,
	}

	if len(all) != len(expected) {
		t.Fatalf("AllEventTypes() returned %d types, want %d", len(all), len(expected))
	}

	for _, et := range expected {
		if !slices.Contains(all, et) {
			t.Errorf("AllEventTypes() missing %q", et)
		}
	}
}

func TestAllEventTypes_EmojiAndLabelCoverage(t *testing.T) {
	// Every event type from AllEventTypes should have a non-default emoji and label.
	bellEmoji := "\U0001F514"

	for _, et := range AllEventTypes() {
		t.Run(string(et), func(t *testing.T) {
			emoji := et.Emoji()
			if emoji == bellEmoji {
				t.Errorf("EventType(%q).Emoji() returned default bell; expected a specific emoji", et)
			}
			if emoji == "" {
				t.Errorf("EventType(%q).Emoji() returned empty string", et)
			}

			label := et.Label()
			if label == string(et) {
				t.Errorf("EventType(%q).Label() returned raw string; expected a human-readable label", et)
			}
			if label == "" {
				t.Errorf("EventType(%q).Label() returned empty string", et)
			}
		})
	}
}
