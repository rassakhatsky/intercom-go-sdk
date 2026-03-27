package intercom

// Logger defines the interface for SDK logging.
// *[log/slog.Logger] satisfies this interface.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}
