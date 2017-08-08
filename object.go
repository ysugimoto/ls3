package main

import (
	"fmt"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// S3 object struct
type Object struct {

	// File size
	size int64

	// File name
	key string

	// Last modified time
	lastModified time.Time

	// directroy flag
	dir bool

	// parent flag
	parent bool

	Writer
}

// Create new object pointer
func NewObject(key string, size int64, lastModified time.Time, dir bool) *Object {
	return &Object{
		size:         size,
		key:          key,
		lastModified: lastModified,
		dir:          dir,
	}
}

// Create new object pointer as parent
func NewParentObject() *Object {
	return &Object{
		key:    "../",
		parent: true,
	}
}

// Writer::String implementation
func (o *Object) String() string {
	if o.parent {
		return ""
	} else if o.dir {
		return fmt.Sprintf("%s %10s  %s/", utcToJst(o.lastModified), "-", o.key)
	} else {
		return fmt.Sprintf("%s %10d  %s", utcToJst(o.lastModified), o.size, o.key)
	}
}

// Writer::Write implementation
func (o *Object) Write(y int, filter string) {
	i := 0
	// parent directory, write "../"
	if o.parent {
		for _, r := range []rune(o.key) {
			termbox.SetCell(i, y, r, termbox.ColorWhite, termbox.ColorBlue)
			i++
		}
		// Write as directory
	} else if o.dir {
		for _, r := range []rune(utcToJst(o.lastModified)) {
			termbox.SetCell(i, y, r, termbox.ColorWhite, termbox.ColorDefault)
			i++
		}
		for _, r := range []rune(fmt.Sprintf(" %12s    ", "-")) {
			termbox.SetCell(i, y, r, termbox.ColorCyan, termbox.ColorDefault)
			i++
		}

		first, last := findHighlightRange(o.key, filter)
		for j, r := range []rune(fmt.Sprintf("%s/", o.key)) {
			color := termbox.ColorGreen
			if j >= first && j < last {
				color = termbox.ColorYellow
			}
			termbox.SetCell(i, y, r, color|termbox.AttrBold, termbox.ColorDefault)
			i += runewidth.RuneWidth(r)
		}
		// Write as object
	} else {
		for _, r := range []rune(utcToJst(o.lastModified)) {
			termbox.SetCell(i, y, r, termbox.ColorWhite, termbox.ColorDefault)
			i++
		}
		for _, r := range []rune(fmt.Sprintf(" %12d    ", o.size)) {
			termbox.SetCell(i, y, r, termbox.ColorCyan, termbox.ColorDefault)
			i++
		}

		first, last := findHighlightRange(o.key, filter)
		for j, r := range []rune(fmt.Sprintf("%s", o.key)) {
			color := termbox.ColorWhite
			if j >= first && j < last {
				color = termbox.ColorYellow
			}
			termbox.SetCell(i, y, r, color, termbox.ColorDefault)
			i += runewidth.RuneWidth(r)
		}
	}
}

// Define Object list type
type Objects []*Object

// Transform to Selectable type
func (o Objects) Selectable() Selectable {
	s := Selectable{}
	for _, v := range o {
		s = append(s, v)
	}

	return s
}
