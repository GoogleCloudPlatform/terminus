// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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