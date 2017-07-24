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
func (b *Bucket) Write(y int, filters []rune) {
	i := 0
	for _, r := range []rune("[Bucket] ") {
		termbox.SetCell(i, y, r, termbox.ColorCyan, termbox.ColorDefault)
		i++
	}
	for _, r := range []rune(b.name) {
		size := runewidth.RuneWidth(r)
		color := termbox.ColorWhite
		if isRuneContains(filters, r) {
			color = termbox.ColorYellow
		}
		termbox.SetCell(i, y, r, color, termbox.ColorDefault)
		i += size
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
