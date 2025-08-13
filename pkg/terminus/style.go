package terminus

import "github.com/yourusername/terminusgo/pkg/terminus/style"

// Style exports
type (
	Style = style.Style
	Color = style.Color
)

// Style constructors
var (
	NewStyle = style.New
	
	// Color constructors
	ColorFromString = style.ColorFromString
	ANSI256         = style.ANSI256
	RGB             = style.RGB
	
	// Predefined colors
	Black         = style.Black
	Red           = style.Red
	Green         = style.Green
	Yellow        = style.Yellow
	Blue          = style.Blue
	Magenta       = style.Magenta
	Cyan          = style.Cyan
	White         = style.White
	BrightBlack   = style.BrightBlack
	BrightRed     = style.BrightRed
	BrightGreen   = style.BrightGreen
	BrightYellow  = style.BrightYellow
	BrightBlue    = style.BrightBlue
	BrightMagenta = style.BrightMagenta
	BrightCyan    = style.BrightCyan
	BrightWhite   = style.BrightWhite
)