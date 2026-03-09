package scim

// Logger defines an interface for logging errors.
type Logger interface {
	Error(args ...any)
}

type noopLogger struct{}

func (noopLogger) Error(...any) {}
