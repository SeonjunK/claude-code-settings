// Package notifier provides fire-and-forget notification dispatch.
package notifier

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/provider"
)

// DefaultTimeout is the maximum time to wait for notifications to send.
const DefaultTimeout = 3 * time.Second

// Notifier dispatches notification events to matching providers.
type Notifier struct {
	registry *provider.Registry
	filters  map[string][]string // event type -> provider names
	timeout  time.Duration
}

// New creates a Notifier.
// filters maps event type strings to provider name lists.
// If filters is nil, all providers receive all events.
func New(registry *provider.Registry, filters map[string][]string) *Notifier {
	return &Notifier{
		registry: registry,
		filters:  filters,
		timeout:  DefaultTimeout,
	}
}

// matchProviders returns providers that should handle the given event type.
func (n *Notifier) matchProviders(eventType notification.EventType) []provider.Provider {
	if n.filters != nil {
		if names, ok := n.filters[string(eventType)]; ok {
			return n.registry.ByNames(names)
		}
	}
	// No filter or event not in filter: send to all providers.
	return n.registry.Providers()
}

// Dispatch sends an event to all matching providers asynchronously.
// Returns a wait function that blocks until all sends complete or timeout.
// The caller MUST call wait() before os.Exit to ensure delivery.
func (n *Notifier) Dispatch(event notification.Event) (wait func()) {
	providers := n.matchProviders(event.Type)
	if len(providers) == 0 {
		return func() {}
	}

	var wg sync.WaitGroup
	for _, p := range providers {
		wg.Add(1)
		go func(p provider.Provider) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), n.timeout)
			defer cancel()
			if err := p.Send(ctx, event); err != nil {
				fmt.Fprintf(os.Stderr, "[notify] %s error: %v\n", p.Name(), err)
			}
		}(p)
	}

	return func() {
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(n.timeout + 500*time.Millisecond):
			fmt.Fprintln(os.Stderr, "[notify] timeout waiting for providers")
		}
	}
}

// DispatchSync sends an event and waits for completion.
func (n *Notifier) DispatchSync(event notification.Event) {
	wait := n.Dispatch(event)
	wait()
}
