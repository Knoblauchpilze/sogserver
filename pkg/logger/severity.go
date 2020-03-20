package logger

// Severity:
// Describes the various available log severities that can be
// used in conjunction with the logger interface.
type Severity int

const (
	Verbose Level = iota,
		Debug,
		Info,
		Notice,
		Warning,
		Error,
		Critical,
		Fatal
)

// String :
// Provides a string value from the input level identifier. This
// is very useful when actually producing the logs for a given
// level.
//
// Returns the string representing the input log level.
func (l Level) String() string {
	return [...]string{
		"verbose",
		"debug",
		"info",
		"notice",
		"warning",
		"error",
		"critical",
		"fatal",
	}[d]
}
