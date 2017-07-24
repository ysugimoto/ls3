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
func (o *Object) Write(y int) {
	i := 0
	if o.parent {
		for _, r := range []rune(o.key) {
			termbox.SetCell(i, y, r, termbox.ColorWhite, termbox.ColorBlue)
			i++
		}
	} else if o.dir {
		for _, r := range []rune(utcToJst(o.lastModified)) {
			termbox.SetCell(i, y, r, termbox.ColorWhite, termbox.ColorDefault)
			i++
		}
		for _, r := range []rune(fmt.Sprintf(" %12s    ", "-")) {
			termbox.SetCell(i, y, r, termbox.ColorCyan, termbox.ColorDefault)
			i++
		}
		for _, r := range []rune(fmt.Sprintf("%s/", o.key)) {
			size := runewidth.RuneWidth(r)
			termbox.SetCell(i, y, r, termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault)
			i += size
		}
	} else {
		for _, r := range []rune(utcToJst(o.lastModified)) {
			termbox.SetCell(i, y, r, termbox.ColorWhite, termbox.ColorDefault)
			i++
		}
		for _, r := range []rune(fmt.Sprintf(" %12d    ", o.size)) {
			termbox.SetCell(i, y, r, termbox.ColorCyan, termbox.ColorDefault)
			i++
		}
		for _, r := range []rune(fmt.Sprintf("%s", o.key)) {
			size := runewidth.RuneWidth(r)
			termbox.SetCell(i, y, r, termbox.ColorWhite, termbox.ColorDefault)
			i += size
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
