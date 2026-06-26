package stream

import "context"

// Meta identifies who is streaming and in which meeting phase.
type Meta struct {
	ParticipantID string
	Phase         string
	Detail        string
}

// Handlers receives streaming lifecycle callbacks for one LLM completion.
type Handlers struct {
	Meta    Meta
	OnStart func(Meta)
	OnDelta func(string)
	OnEnd   func()
}

type ctxKey struct{}

// WithHandlers attaches stream callbacks to ctx for the duration of one Respond call.
func WithHandlers(ctx context.Context, h Handlers) context.Context {
	return context.WithValue(ctx, ctxKey{}, h)
}

// HandlersFrom returns stream callbacks previously attached to ctx.
func HandlersFrom(ctx context.Context) (Handlers, bool) {
	h, ok := ctx.Value(ctxKey{}).(Handlers)
	return h, ok
}
