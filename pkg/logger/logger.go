package logger

// Logger :
// Describes a common interface used for logging purposes.
// A single method is needed to allow the logging of some
// messages based on a content and a severity.
//
// The `Trace` allows to log a message with the specified
// level.
type Logger interface {
	Trace(level Severity, module string, message string)
}
