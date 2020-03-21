package logger

import "strings"

// Severity:
// Describes the various available log severities that can be
// used in conjunction with the logger interface.
type Severity int

const (
	Verbose Severity = iota
	Debug
	Info
	Notice
	Warning
	Error
	Critical
	Fatal
)

// Name :
// Provides a string value from the input level identifier. This
// is very useful when actually producing the logs for a given
// level.
//
// Returns the string representing the input log level.
func (s Severity) Name() string {
	return [...]string{
		"verbose",
		"debug",
		"info",
		"notice",
		"warning",
		"error",
		"critical",
		"fatal",
	}[s]
}

// Color :
// Provides a color value representing the severity. This is used
// as a visual way to distinguish between severity when displayed
// in a logging device.
//
// Returns a color value that can be used to change the display
// for the corresponding severity.
func (s Severity) Color() Color {
	return [...]Color{
		Grey,
		Blue,
		Green,
		Cyan,
		Yellow,
		Red,
		Red,
		Red,
	}[s]
}

// String :
// Provides a complete string representing the input severity and
// which includes some color formatting to display it with a color
// that matches its importance.
//
// Returns the string allowing to format the display device to
// print this severity.
func (s Severity) String() string {
	return FormatWithBrackets(s.Name(), s.Color())
}

// fromString :
// Converts the input string into the corresponding severity value.
// In case the input severity does not correspond to a known value
// a `verbose` severity is returned.
// Note that the case is not important (so `Debug`, `DeBug`, `debug`
// or any other variations will all be converted to a severity of
// `Debug`).
//
// The `level` represents the string to convert to a severity.
//
// Returns the severity associated to the input string.
func fromString(level string) Severity {
	// Lowercase the input string.
	lower := strings.ToLower(level)

	// Determine the severity associated to this string.
	switch lower {
	case "debug":
		return Debug
	case "info":
		return Info
	case "notice":
		return Notice
	case "warning":
		return Warning
	case "error":
		return Error
	case "critical":
		return Critical
	case "fatal":
		return Fatal
	case "verbose":
		fallthrough
	default:
		// Assume verbose by default.
		return Verbose
	}
}
