package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// Bucket struct
type Bucket struct {

	// Bucket name
	name string

	Writer
}

// Create new bucket pointer
func NewBucket(b *s3.Bucket) *Bucket {
	return &Bucket{
		name: *b.Name,
	}
}

// Writer::String implementation
func (b *Bucket) String() string {
	return fmt.Sprintf("[Bucket] %s", b.name)
}

// Writer::Write implementation
func (b *Bucket) Write(y int, filter string) {
	i := 0
	for _, r := range []rune("[Bucket] ") {
		termbox.SetCell(i, y, r, termbox.ColorCyan, termbox.ColorDefault)
		i++
	}

	first, last := findHighlightRange(b.name, filter)
	for j, r := range []rune(b.name) {
		color := termbox.ColorWhite
		if j >= first && j < last {
			color = termbox.ColorYellow
		}
		termbox.SetCell(i, y, r, color, termbox.ColorDefault)
		i += runewidth.RuneWidth(r)
	}
}

// Define Bucket slice type
type Buckets []*Bucket

// Transform to Selectable type
func (b Buckets) Selectable() Selectable {
	s := Selectable{}
	for _, v := range b {
		s = append(s, v)
	}

	return s
}
