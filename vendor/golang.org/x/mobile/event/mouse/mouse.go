// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mouse defines an event for mouse input.
//
// See the golang.org/x/mobile/app package for details on the event model.
package mouse // import "golang.org/x/mobile/event/mouse"

import (
	"fmt"

	"golang.org/x/mobile/event/key"
)

// Event is a mouse event.
type Event struct {
	// X and Y are the mouse location, in pixels.
	X, Y float32

	// Button is the mouse button being pressed or released. Its value may be
	// zero, for a mouse move or drag without any button change.
	Button Button

	// TODO: have a field to hold what other buttons are down, for detecting
	// drags or button-chords.

	// Modifiers is a bitmask representing a set of modifier keys:
	// key.ModShift, key.ModAlt, etc.
	Modifiers key.Modifiers

	// Direction is the direction of the mouse event: DirPress, DirRelease,
	// or DirNone (for mouse moves or drags).
	Direction Direction

	// TODO: add a Device ID, for multiple input devices?
	// TODO: add a time.Time?
}

type ScrollEvent struct {
	Event

	// Dx and Dy are horizontal and vertical scrolling deltas, in pixels.
	Dx, Dy  float32
	Precise bool
}

// Button is a mouse button.
type Button int32

const (
	ButtonNone        Button = +0
	ButtonLeft        Button = +1
	ButtonMiddle      Button = +2
	ButtonRight       Button = +3
	ButtonScrollWheel Button = -1
)

// Direction is the direction of the mouse event.
type Direction uint8

const (
	DirNone    Direction = 0
	DirPress   Direction = 1
	DirRelease Direction = 2
)

func (d Direction) String() string {
	switch d {
	case DirNone:
		return "None"
	case DirPress:
		return "Press"
	case DirRelease:
		return "Release"
	default:
		return fmt.Sprintf("mouse.Direction(%d)", d)
	}
}
