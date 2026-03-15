// Package provider defines the notification provider interface and registry.
package provider

import (
	"context"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
)

// Provider sends a notification event to an external channel.
type Provider interface {
	// Name returns the provider identifier (e.g. "telegram", "slack", "webhook").
	Name() string
	// Send delivers the event. Implementations must respect ctx deadline.
	Send(ctx context.Context, event notification.Event) error
}

// Registry holds enabled providers.
type Registry struct {
	providers []Provider
	byName    map[string]Provider
}

// NewRegistry creates a registry with the given providers.
func NewRegistry(providers ...Provider) *Registry {
	r := &Registry{
		byName: make(map[string]Provider),
	}
	for _, p := range providers {
		r.Add(p)
	}
	return r
}

// Add registers a provider.
func (r *Registry) Add(p Provider) {
	r.providers = append(r.providers, p)
	r.byName[p.Name()] = p
}

// Get returns a provider by name.
func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.byName[name]
	return p, ok
}

// Providers returns all registered providers.
func (r *Registry) Providers() []Provider {
	return r.providers
}

// ByNames returns providers matching the given names.
func (r *Registry) ByNames(names []string) []Provider {
	var result []Provider
	for _, name := range names {
		if p, ok := r.byName[name]; ok {
			result = append(result, p)
		}
	}
	return result
}
