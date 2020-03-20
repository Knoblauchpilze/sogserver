package logger

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
