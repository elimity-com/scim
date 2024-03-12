package scim

import "log/slog"

var log *slog.Logger = slog.Default().WithGroup("scim")

// SetLogger sets the logger for the scim package.
func SetLogger(l *slog.Logger) {
	log = l
}
