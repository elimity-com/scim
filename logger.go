package scim

// Logger defines an interface for logging errors.
type Logger interface {
	Error(args ...interface{})
}

type noopLogger struct{}

func (noopLogger) Error(...interface{}) {}
