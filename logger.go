package scim

type Logger interface {
	Error(args ...interface{})
}

type noopLogger struct{}

func (noopLogger) Error(...interface{}) {}
