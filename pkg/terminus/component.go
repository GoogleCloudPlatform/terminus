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

// Msg is a marker interface for messages that can be sent to components
type Msg interface{}

// Cmd represents a command that performs side effects and returns a message
type Cmd func() Msg

// Component is the core interface that all UI components must implement
type Component interface {
	// Init is called once when the component is first created
	// It returns an optional command to execute
	Init() Cmd

	// Update handles incoming messages and updates the component state
	// It returns the new component state and an optional command
	Update(msg Msg) (Component, Cmd)

	// View renders the component's current state as a string
	View() string
}