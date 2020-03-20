package logger

// Color :
// Defines the color that can be produced as valid standard
// output display.
type Color int

const (
	Black Color = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	Grey
)

// Defines the color code to use to produce the corresponding
// color display in the standard output.
//
// Returns a string allowing to switch the color display of
// the standard output to the desired color.
func GetColorCode(c Color) string {
	code := [...]string{
		"30",
		"31",
		"32",
		"33",
		"34",
		"35",
		"36",
		"37",
		"90",
	}[c]
	return "\033[1;" + code + "m"
}

// NoOp :
// Resets the color display of the standard output to the default
// color (i.e. white).
//
// Returns a string allowing to format the color display of the
// standard output to the initial format.
func NoOp() string {
	return "\033[0m"
}

// format :
// Used to format the input message with the input color and
// return the corresponding string to use to change the standard
// output logging device.
//
// The `msg` represents the content of the message that should
// be printed with the specified color.
//
// The `c` value represents the color with which the message is
// to be displayed.
//
// The `addBracket` allows to automatically add brackets to the
// content of the message.
//
// Returns the string to use to modify the standard output with
// the required color.
func format(msg string, c Color, addBracket bool) string {
	fMsg := ""
	if addBracket {
		fMsg += "["
	}
	fMsg += msg
	if addBracket {
		fMsg += "]"
	}
	return GetColorCode(c) + fMsg + NoOp()
}

// Wrapper around the `format` method assuming the user wants
// to add some brackets around the message.
//
// The `msg` represents the content of the message (which will
// be surrounded by brackets).
//
// The `c` represents the color with which the data should be
// displayed.
//
// Returns the string for the message displayed in the desired
// color.
func FormatWithBrackets(msg string, c Color) string {
	return format(msg, c, true)
}

// FormatWithNoBrackets :
// Similar to `FormatWithBrackets` but does not include some
// brackets around the message.
//
// Returns the string for the message displayed in the desired
// color.
func FormatWithNoBrackets(msg string, c Color) string {
	return format(msg, c, false)
}