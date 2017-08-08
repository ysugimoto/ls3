package main

import (
	"github.com/nsf/termbox-go"
)

// Object action type
type ObjectAction int

const (
	Back ObjectAction = iota
	Download
	None = 999
)

// Action command struct
type ActionCommand struct {

	// Operation type
	op ObjectAction

	// Acrion name
	name string

	Writer
}

// Writer::String implementation
func (a ActionCommand) String() string {
	return a.name
}

// Writer::Writer implementation
func (a ActionCommand) Write(y int, filter string) {
	for i, r := range []rune(a.name) {
		termbox.SetCell(i, y, r, termbox.ColorWhite, termbox.ColorDefault)
	}
}

// Define ActionCommand slice type
type ActionList []ActionCommand

// Transform to Selectable type
func (a ActionList) Selectable() Selectable {
	s := Selectable{}
	for _, v := range a {
		s = append(s, v)
	}

	return s
}
