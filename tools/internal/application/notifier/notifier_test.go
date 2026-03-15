package notifier

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/provider"
)

// mockProvider implements provider.Provider for testing.
type mockProvider struct {
	name    string
	sendFn  func(ctx context.Context, event notification.Event) error
	mu      sync.Mutex
	called  int
	lastEvt notification.Event
}

func newMockProvider(name string) *mockProvider {
	return &mockProvider{name: name}
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Send(ctx context.Context, event notification.Event) error {
	m.mu.Lock()
	m.called++
	m.lastEvt = event
	m.mu.Unlock()

	if m.sendFn != nil {
		return m.sendFn(ctx, event)
	}
	return nil
}

func (m *mockProvider) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}

func (m *mockProvider) lastEvent() notification.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastEvt
}

func makeEvent(eventType notification.EventType) notification.Event {
	return notification.Event{
		Type:      eventType,
		SessionID: "test-session-id",
		Timestamp: time.Now(),
		Summary:   "test event",
	}
}

func TestDispatch_AllProviders_NoFilters(t *testing.T) {
	p1 := newMockProvider("telegram")
	p2 := newMockProvider("slack")
	reg := provider.NewRegistry(p1, p2)

	n := New(reg, nil)
	event := makeEvent(notification.EventSessionStart)

	wait := n.Dispatch(event)
	wait()

	if p1.callCount() != 1 {
		t.Errorf("telegram provider called %d times, want 1", p1.callCount())
	}
	if p2.callCount() != 1 {
		t.Errorf("slack provider called %d times, want 1", p2.callCount())
	}
}

func TestDispatch_WithFilters(t *testing.T) {
	p1 := newMockProvider("telegram")
	p2 := newMockProvider("slack")
	reg := provider.NewRegistry(p1, p2)

	filters := map[string][]string{
		"session_start": {"telegram"},
	}
	n := New(reg, filters)

	// session_start should only go to telegram.
	event := makeEvent(notification.EventSessionStart)
	wait := n.Dispatch(event)
	wait()

	if p1.callCount() != 1 {
		t.Errorf("telegram provider called %d times, want 1", p1.callCount())
	}
	if p2.callCount() != 0 {
		t.Errorf("slack provider called %d times, want 0", p2.callCount())
	}
}

func TestDispatch_FilteredEvent_FallsBackToAll(t *testing.T) {
	p1 := newMockProvider("telegram")
	p2 := newMockProvider("slack")
	reg := provider.NewRegistry(p1, p2)

	filters := map[string][]string{
		"session_start": {"telegram"},
	}
	n := New(reg, filters)

	// guard_deny is not in filters, so all providers should receive it.
	event := makeEvent(notification.EventGuardDeny)
	wait := n.Dispatch(event)
	wait()

	if p1.callCount() != 1 {
		t.Errorf("telegram provider called %d times, want 1", p1.callCount())
	}
	if p2.callCount() != 1 {
		t.Errorf("slack provider called %d times, want 1", p2.callCount())
	}
}

func TestDispatch_NoProviders_ReturnsImmediately(t *testing.T) {
	reg := provider.NewRegistry()
	n := New(reg, nil)

	event := makeEvent(notification.EventSessionStart)
	wait := n.Dispatch(event)

	// Should not hang.
	done := make(chan struct{})
	go func() {
		wait()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(1 * time.Second):
		t.Fatal("wait() did not return for empty provider list")
	}
}

func TestDispatch_ProviderError_DoesNotBlock(t *testing.T) {
	p1 := newMockProvider("failing")
	p1.sendFn = func(ctx context.Context, event notification.Event) error {
		return fmt.Errorf("send failed")
	}
	reg := provider.NewRegistry(p1)

	n := New(reg, nil)
	event := makeEvent(notification.EventSessionStart)

	wait := n.Dispatch(event)
	wait()

	if p1.callCount() != 1 {
		t.Errorf("failing provider called %d times, want 1", p1.callCount())
	}
}

func TestDispatch_EventPassedToProvider(t *testing.T) {
	p1 := newMockProvider("telegram")
	reg := provider.NewRegistry(p1)

	n := New(reg, nil)
	event := notification.Event{
		Type:      notification.EventTaskComplete,
		SessionID: "session-abc",
		Timestamp: time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC),
		Summary:   "task done",
		Goal:      "build feature",
	}

	wait := n.Dispatch(event)
	wait()

	got := p1.lastEvent()
	if got.Type != notification.EventTaskComplete {
		t.Errorf("event type = %q, want %q", got.Type, notification.EventTaskComplete)
	}
	if got.SessionID != "session-abc" {
		t.Errorf("session_id = %q, want %q", got.SessionID, "session-abc")
	}
	if got.Summary != "task done" {
		t.Errorf("summary = %q, want %q", got.Summary, "task done")
	}
	if got.Goal != "build feature" {
		t.Errorf("goal = %q, want %q", got.Goal, "build feature")
	}
}

func TestDispatchSync(t *testing.T) {
	var called atomic.Int32
	p1 := newMockProvider("telegram")
	p1.sendFn = func(ctx context.Context, event notification.Event) error {
		called.Add(1)
		return nil
	}
	reg := provider.NewRegistry(p1)

	n := New(reg, nil)
	event := makeEvent(notification.EventSessionStart)
	n.DispatchSync(event)

	if called.Load() != 1 {
		t.Errorf("DispatchSync did not call provider; called = %d", called.Load())
	}
}

func TestDispatch_Timeout_SlowProvider(t *testing.T) {
	p1 := newMockProvider("slow")
	p1.sendFn = func(ctx context.Context, event notification.Event) error {
		// Simulate a provider that blocks until context is cancelled.
		<-ctx.Done()
		return ctx.Err()
	}
	reg := provider.NewRegistry(p1)

	n := &Notifier{
		registry: reg,
		filters:  nil,
		timeout:  50 * time.Millisecond, // very short timeout for test
	}

	event := makeEvent(notification.EventSessionStart)
	wait := n.Dispatch(event)

	done := make(chan struct{})
	go func() {
		wait()
		close(done)
	}()

	select {
	case <-done:
		// OK - wait() returned (either providers finished or outer timeout fired)
	case <-time.After(2 * time.Second):
		t.Fatal("wait() did not return within expected timeout window")
	}
}

func TestDispatch_MultipleEvents(t *testing.T) {
	p1 := newMockProvider("telegram")
	reg := provider.NewRegistry(p1)
	n := New(reg, nil)

	events := []notification.EventType{
		notification.EventSessionStart,
		notification.EventIteration,
		notification.EventVerifyPass,
		notification.EventSessionStop,
	}

	for _, et := range events {
		wait := n.Dispatch(makeEvent(et))
		wait()
	}

	if p1.callCount() != len(events) {
		t.Errorf("provider called %d times, want %d", p1.callCount(), len(events))
	}
}
