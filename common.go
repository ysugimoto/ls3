package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// We're living in Asia/Tokyo location :)
var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

// Transform from UTC to JST
func utcToJst(utc time.Time) string {
	jst := utc.In(JST)
	return jst.Format("2006-01-02 15:03:04")
}

type Bucket struct {
	name string

	Writer
}

func NewBucket(b *s3.Bucket) *Bucket {
	return &Bucket{
		name: *b.Name,
	}
}

func (b *Bucket) String() string {
	return fmt.Sprintf("[Bucket] %s", b.name)
}

func (b *Bucket) Write(y int) {
	i := 0
	for _, r := range []rune("[Bucket] ") {
		termbox.SetCell(i, y, r, termbox.ColorCyan, termbox.ColorDefault)
		i++
	}
	for _, r := range []rune(b.name) {
		size := runewidth.RuneWidth(r)
		termbox.SetCell(i, y, r, termbox.ColorWhite, termbox.ColorDefault)
		i += size
	}
}

type Buckets []*Bucket

func (b Buckets) Selectable() Selectable {
	s := Selectable{}
	for _, v := range b {
		s = append(s, v)
	}

	return s
}

type Object struct {
	size         int64
	key          string
	lastModified time.Time
	dir          bool
	parent       bool

	Writer
}

func NewObject(key string, size int64, lastModified time.Time, dir bool) *Object {
	return &Object{
		size:         size,
		key:          key,
		lastModified: lastModified,
		dir:          dir,
	}
}

func NewParentObject() *Object {
	return &Object{
		key:    "../",
		parent: true,
	}
}

func (o *Object) String() string {
	if o.parent {
		return ""
	} else if o.dir {
		return fmt.Sprintf("%s %10s  %s/", utcToJst(o.lastModified), "-", o.key)
	} else {
		return fmt.Sprintf("%s %10d  %s", utcToJst(o.lastModified), o.size, o.key)
	}
}

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

type Objects []*Object

func (o Objects) Selectable() Selectable {
	s := Selectable{}
	for _, v := range o {
		s = append(s, v)
	}

	return s
}

type ObjectAction int

const (
	Back ObjectAction = iota
	View
	Download
	None = 999
)

type ActionCommand struct {
	op   ObjectAction
	name string

	Writer
}

func (a ActionCommand) String() string {
	return a.name
}

func (a ActionCommand) Write(y int) {
	for i, r := range []rune(a.name) {
		termbox.SetCell(i, y, r, termbox.ColorWhite, termbox.ColorDefault)
	}
}

type ActionList []ActionCommand

func (a ActionList) Selectable() Selectable {
	s := Selectable{}
	for _, v := range a {
		s = append(s, v)
	}

	return s
}
