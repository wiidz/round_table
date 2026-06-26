package principal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Binding links an external chat identity to a Principal for one scope (guild or DM).
type Binding struct {
	PrincipalID string    `json:"principal_id"`
	Platform    string    `json:"platform"`
	ExternalID  string    `json:"external_id"`
	DisplayName string    `json:"display_name,omitempty"`
	Scope       string    `json:"scope"`
	BoundAt     time.Time `json:"bound_at"`
}

// Registry stores Principal bindings by scope (adapter layer, not Meeting domain).
type Registry struct {
	path string
	mu   sync.RWMutex
	byScope map[string]Binding
}

// NewRegistry loads bindings from path (creates empty registry if file missing).
func NewRegistry(path string) (*Registry, error) {
	r := &Registry{
		path:    path,
		byScope: make(map[string]Binding),
	}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

// ScopeKey returns the binding scope for a guild message or DM.
func ScopeKey(platform, guildID, authorID string) string {
	if guildID != "" {
		return platform + ":guild:" + guildID
	}
	return platform + ":dm:" + authorID
}

// PrincipalIDForExternal builds a stable Principal identity from platform + external user id.
func PrincipalIDForExternal(platform, externalID string) string {
	return platform + ":" + externalID
}

// Get returns the Principal binding for scope, if any.
func (r *Registry) Get(scope string) (Binding, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.byScope[scope]
	return b, ok
}

// Bind sets the Principal for scope to the given external user.
func (r *Registry) Bind(scope, platform, externalID, displayName string) (Binding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.byScope[scope]; ok && existing.ExternalID != externalID {
		return Binding{}, fmt.Errorf("scope already bound to %s", existing.DisplayName)
	}

	b := Binding{
		PrincipalID: PrincipalIDForExternal(platform, externalID),
		Platform:    platform,
		ExternalID:  externalID,
		DisplayName: displayName,
		Scope:       scope,
		BoundAt:     time.Now().UTC(),
	}
	r.byScope[scope] = b
	return b, r.saveLocked()
}

// Unbind removes the binding if externalID matches the bound Principal.
func (r *Registry) Unbind(scope, externalID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.byScope[scope]
	if !ok {
		return fmt.Errorf("no principal bound in this scope")
	}
	if existing.ExternalID != externalID {
		return fmt.Errorf("only the bound principal can unbind")
	}
	delete(r.byScope, scope)
	return r.saveLocked()
}

func (r *Registry) load() error {
	if r.path == "" {
		return nil
	}
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	var items []Binding
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("load principal bindings: %w", err)
	}
	for _, b := range items {
		r.byScope[b.Scope] = b
	}
	return nil
}

func (r *Registry) saveLocked() error {
	if r.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}
	items := make([]Binding, 0, len(r.byScope))
	for _, b := range r.byScope {
		items = append(items, b)
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, 0o644)
}
