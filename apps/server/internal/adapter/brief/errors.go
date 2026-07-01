package brief

import "errors"

var (
	ErrNotFound      = errors.New("brief template not found")
	ErrInvalidID     = errors.New("invalid brief template id")
	ErrInvalidYAML   = errors.New("invalid BRIEF.yaml")
	ErrBuiltinReadonly = errors.New("builtin brief template is read-only")
)
