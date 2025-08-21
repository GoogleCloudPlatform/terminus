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

package widget

import (
	"time"

	"github.com/yourusername/terminusgo/pkg/terminus"
)

// SpinnerStyle represents different spinner animation styles
type SpinnerStyle int

const (
	SpinnerDots SpinnerStyle = iota
	SpinnerLine
	SpinnerCircle
	SpinnerArrow
	SpinnerBounce
	SpinnerPulse
	SpinnerClock
	SpinnerBraille
)

// spinnerChars contains the character sequences for different spinner styles
var spinnerChars = map[SpinnerStyle][]string{
	SpinnerDots:    {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	SpinnerLine:    {"|", "/", "-", "\\"},
	SpinnerCircle:  {"◐", "◓", "◑", "◒"},
	SpinnerArrow:   {"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
	SpinnerBounce:  {"⠁", "⠂", "⠄", "⠂"},
	SpinnerPulse:   {"●", "○", "●", "○"},
	SpinnerClock:   {"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"},
	SpinnerBraille: {"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
}

// Spinner is an animated loading indicator widget
type Spinner struct {
	Model

	// Animation state
	currentFrame int
	isSpinning   bool
	startTime    time.Time

	// Configuration
	spinnerStyle SpinnerStyle
	customChars  []string
	text         string
	textPosition TextPosition
	speed        time.Duration

	// Styling
	style          terminus.Style
	textStyle      terminus.Style
	spinnerColor   terminus.Style

	// Animation control
	ticker   *time.Ticker
	tickChan chan terminus.Msg
}

// TextPosition represents where the text appears relative to the spinner
type TextPosition int

const (
	TextLeft TextPosition = iota
	TextRight
	TextAbove
	TextBelow
)

// SpinnerTickMsg is sent to advance the spinner animation
type SpinnerTickMsg struct {
	ID string
}

// NewSpinner creates a new spinner widget
func NewSpinner() *Spinner {
	return &Spinner{
		Model:        NewModel(),
		currentFrame: 0,
		isSpinning:   false,
		spinnerStyle: SpinnerDots,
		text:         "",
		textPosition: TextRight,
		speed:        100 * time.Millisecond,
		style:        terminus.NewStyle(),
		textStyle:    terminus.NewStyle(),
		spinnerColor: terminus.NewStyle().Foreground(terminus.Cyan),
	}
}

// SetSpinnerStyle sets the spinner animation style
func (s *Spinner) SetSpinnerStyle(style SpinnerStyle) *Spinner {
	s.spinnerStyle = style
	return s
}

// SetCustomChars sets custom characters for the spinner animation
func (s *Spinner) SetCustomChars(chars []string) *Spinner {
	s.customChars = chars
	return s
}

// SetText sets the loading text
func (s *Spinner) SetText(text string) *Spinner {
	s.text = text
	return s
}

// SetTextPosition sets where the text appears relative to the spinner
func (s *Spinner) SetTextPosition(position TextPosition) *Spinner {
	s.textPosition = position
	return s
}

// SetSpeed sets the animation speed
func (s *Spinner) SetSpeed(speed time.Duration) *Spinner {
	s.speed = speed
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = time.NewTicker(s.speed)
	}
	return s
}

// SetStyle sets the default style
func (s *Spinner) SetStyle(style terminus.Style) *Spinner {
	s.style = style
	return s
}

// SetTextStyle sets the text style
func (s *Spinner) SetTextStyle(style terminus.Style) *Spinner {
	s.textStyle = style
	return s
}

// SetSpinnerColor sets the spinner character style
func (s *Spinner) SetSpinnerColor(style terminus.Style) *Spinner {
	s.spinnerColor = style
	return s
}

// Start starts the spinner animation
func (s *Spinner) Start() *Spinner {
	if !s.isSpinning {
		s.isSpinning = true
		s.startTime = time.Now()
		s.currentFrame = 0
		
		// Start the ticker for animation
		if s.ticker != nil {
			s.ticker.Stop()
		}
		s.ticker = time.NewTicker(s.speed)
	}
	return s
}

// Stop stops the spinner animation
func (s *Spinner) Stop() *Spinner {
	if s.isSpinning {
		s.isSpinning = false
		if s.ticker != nil {
			s.ticker.Stop()
			s.ticker = nil
		}
	}
	return s
}

// IsSpinning returns whether the spinner is currently animating
func (s *Spinner) IsSpinning() bool {
	return s.isSpinning
}

// Text returns the current text
func (s *Spinner) Text() string {
	return s.text
}

// ElapsedTime returns how long the spinner has been running
func (s *Spinner) ElapsedTime() time.Duration {
	if !s.isSpinning {
		return 0
	}
	return time.Since(s.startTime)
}

// getChars returns the character sequence for the current spinner style
func (s *Spinner) getChars() []string {
	if len(s.customChars) > 0 {
		return s.customChars
	}
	if chars, ok := spinnerChars[s.spinnerStyle]; ok {
		return chars
	}
	return spinnerChars[SpinnerDots] // fallback
}

// getCurrentChar returns the current spinner character
func (s *Spinner) getCurrentChar() string {
	chars := s.getChars()
	if len(chars) == 0 {
		return "●"
	}
	return chars[s.currentFrame%len(chars)]
}

// Init implements the Component interface
func (s *Spinner) Init() terminus.Cmd {
	return nil
}

// Update implements the Component interface
func (s *Spinner) Update(msg terminus.Msg) (terminus.Component, terminus.Cmd) {
	switch msg.(type) {
	case SpinnerTickMsg:
		if s.isSpinning {
			s.currentFrame++
			// Return a new tick command to continue animation
			return s, s.tick()
		}
	}

	// Check if we need to start the animation based on ticker
	if s.isSpinning && s.ticker != nil {
		select {
		case <-s.ticker.C:
			s.currentFrame++
			return s, s.tick()
		default:
			// No tick available
		}
	}

	return s, nil
}

// tick creates a tick command for animation
func (s *Spinner) tick() terminus.Cmd {
	if !s.isSpinning {
		return nil
	}
	
	return func() terminus.Msg {
		time.Sleep(s.speed)
		return SpinnerTickMsg{ID: "spinner"}
	}
}

// View implements the Component interface
func (s *Spinner) View() string {
	if !s.isSpinning && s.text == "" {
		return ""
	}

	spinnerChar := ""
	if s.isSpinning {
		spinnerChar = s.spinnerColor.Render(s.getCurrentChar())
	}

	textContent := ""
	if s.text != "" {
		textContent = s.textStyle.Render(s.text)
	}

	// Arrange spinner and text based on position
	switch s.textPosition {
	case TextLeft:
		if textContent != "" && spinnerChar != "" {
			return textContent + " " + spinnerChar
		} else if textContent != "" {
			return textContent
		} else {
			return spinnerChar
		}

	case TextRight:
		if spinnerChar != "" && textContent != "" {
			return spinnerChar + " " + textContent
		} else if textContent != "" {
			return textContent
		} else {
			return spinnerChar
		}

	case TextAbove:
		if textContent != "" && spinnerChar != "" {
			return textContent + "\n" + spinnerChar
		} else if textContent != "" {
			return textContent
		} else {
			return spinnerChar
		}

	case TextBelow:
		if spinnerChar != "" && textContent != "" {
			return spinnerChar + "\n" + textContent
		} else if textContent != "" {
			return textContent
		} else {
			return spinnerChar
		}

	default:
		return s.style.Render(spinnerChar + " " + textContent)
	}
}

// WithLoadingText is a convenience method that sets text and starts the spinner
func (s *Spinner) WithLoadingText(text string) *Spinner {
	return s.SetText(text).Start()
}

// Reset resets the spinner to its initial state
func (s *Spinner) Reset() *Spinner {
	s.Stop()
	s.currentFrame = 0
	return s
}

// Frame returns the current animation frame number
func (s *Spinner) Frame() int {
	return s.currentFrame
}

// SetFrame sets the current animation frame (useful for testing)
func (s *Spinner) SetFrame(frame int) *Spinner {
	s.currentFrame = frame
	return s
}

// Spinner presets for common use cases

// NewLoadingSpinner creates a spinner with "Loading..." text
func NewLoadingSpinner() *Spinner {
	return NewSpinner().
		SetText("Loading...").
		SetSpinnerStyle(SpinnerDots).
		SetSpeed(100 * time.Millisecond)
}

// NewProcessingSpinner creates a spinner with "Processing..." text
func NewProcessingSpinner() *Spinner {
	return NewSpinner().
		SetText("Processing...").
		SetSpinnerStyle(SpinnerBraille).
		SetSpeed(80 * time.Millisecond)
}

// NewSavingSpinner creates a spinner with "Saving..." text
func NewSavingSpinner() *Spinner {
	return NewSpinner().
		SetText("Saving...").
		SetSpinnerStyle(SpinnerCircle).
		SetSpeed(150 * time.Millisecond)
}

// NewMinimalSpinner creates a minimal spinner with just the animation
func NewMinimalSpinner() *Spinner {
	return NewSpinner().
		SetSpinnerStyle(SpinnerLine).
		SetSpeed(200 * time.Millisecond)
}