package provider

import (
	"context"
	"testing"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
)

// stubProvider is a minimal Provider implementation for registry tests.
type stubProvider struct {
	name string
}

func (s *stubProvider) Name() string { return s.name }
func (s *stubProvider) Send(_ context.Context, _ notification.Event) error {
	return nil
}

func TestNewRegistry(t *testing.T) {
	p1 := &stubProvider{name: "telegram"}
	p2 := &stubProvider{name: "slack"}

	reg := NewRegistry(p1, p2)

	providers := reg.Providers()
	if len(providers) != 2 {
		t.Fatalf("Providers() length = %d, want 2", len(providers))
	}
	if providers[0].Name() != "telegram" {
		t.Errorf("Providers()[0].Name() = %q, want %q", providers[0].Name(), "telegram")
	}
	if providers[1].Name() != "slack" {
		t.Errorf("Providers()[1].Name() = %q, want %q", providers[1].Name(), "slack")
	}
}

func TestNewRegistry_Empty(t *testing.T) {
	reg := NewRegistry()

	providers := reg.Providers()
	if len(providers) != 0 {
		t.Errorf("Providers() length = %d, want 0", len(providers))
	}
}

func TestRegistry_Add(t *testing.T) {
	reg := NewRegistry()

	reg.Add(&stubProvider{name: "webhook"})

	providers := reg.Providers()
	if len(providers) != 1 {
		t.Fatalf("Providers() length = %d, want 1", len(providers))
	}
	if providers[0].Name() != "webhook" {
		t.Errorf("Providers()[0].Name() = %q, want %q", providers[0].Name(), "webhook")
	}
}

func TestRegistry_Get(t *testing.T) {
	tests := []struct {
		name       string
		registered []string
		lookup     string
		wantFound  bool
	}{
		{
			name:       "found",
			registered: []string{"telegram", "slack"},
			lookup:     "telegram",
			wantFound:  true,
		},
		{
			name:       "not found",
			registered: []string{"telegram", "slack"},
			lookup:     "webhook",
			wantFound:  false,
		},
		{
			name:       "empty registry",
			registered: nil,
			lookup:     "telegram",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var providers []Provider
			for _, name := range tt.registered {
				providers = append(providers, &stubProvider{name: name})
			}
			reg := NewRegistry(providers...)

			p, ok := reg.Get(tt.lookup)
			if ok != tt.wantFound {
				t.Errorf("Get(%q) found = %v, want %v", tt.lookup, ok, tt.wantFound)
			}
			if tt.wantFound && p.Name() != tt.lookup {
				t.Errorf("Get(%q).Name() = %q, want %q", tt.lookup, p.Name(), tt.lookup)
			}
			if !tt.wantFound && p != nil {
				t.Errorf("Get(%q) returned non-nil provider when not found", tt.lookup)
			}
		})
	}
}

func TestRegistry_ByNames(t *testing.T) {
	tests := []struct {
		name       string
		registered []string
		lookup     []string
		wantNames  []string
	}{
		{
			name:       "all exist",
			registered: []string{"telegram", "slack", "webhook"},
			lookup:     []string{"telegram", "webhook"},
			wantNames:  []string{"telegram", "webhook"},
		},
		{
			name:       "some missing",
			registered: []string{"telegram", "slack"},
			lookup:     []string{"telegram", "webhook"},
			wantNames:  []string{"telegram"},
		},
		{
			name:       "none exist",
			registered: []string{"telegram"},
			lookup:     []string{"slack", "webhook"},
			wantNames:  nil,
		},
		{
			name:       "empty lookup",
			registered: []string{"telegram", "slack"},
			lookup:     nil,
			wantNames:  nil,
		},
		{
			name:       "empty registry",
			registered: nil,
			lookup:     []string{"telegram"},
			wantNames:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var providers []Provider
			for _, name := range tt.registered {
				providers = append(providers, &stubProvider{name: name})
			}
			reg := NewRegistry(providers...)

			result := reg.ByNames(tt.lookup)

			if len(result) != len(tt.wantNames) {
				t.Fatalf("ByNames(%v) returned %d providers, want %d", tt.lookup, len(result), len(tt.wantNames))
			}
			for i, p := range result {
				if p.Name() != tt.wantNames[i] {
					t.Errorf("ByNames(%v)[%d].Name() = %q, want %q", tt.lookup, i, p.Name(), tt.wantNames[i])
				}
			}
		})
	}
}

func TestRegistry_ByNames_PreservesOrder(t *testing.T) {
	reg := NewRegistry(
		&stubProvider{name: "c"},
		&stubProvider{name: "a"},
		&stubProvider{name: "b"},
	)

	result := reg.ByNames([]string{"b", "a", "c"})
	expected := []string{"b", "a", "c"}

	if len(result) != len(expected) {
		t.Fatalf("ByNames returned %d providers, want %d", len(result), len(expected))
	}
	for i, p := range result {
		if p.Name() != expected[i] {
			t.Errorf("ByNames[%d].Name() = %q, want %q", i, p.Name(), expected[i])
		}
	}
}
