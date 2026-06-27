package principal

// AllBindings returns every binding (for display name lookup).
func (r *Registry) AllBindings() []Binding {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Binding, 0, len(r.byScope))
	for _, b := range r.byScope {
		out = append(out, b)
	}
	return out
}
