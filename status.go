package main

import (
	"github.com/nsf/termbox-go"
	"time"
)

type Status struct {
	row     int
	bgColor termbox.Attribute
	fgColor termbox.Attribute
	message string

	width  int
	height int
}

func NewStatus(row int) *Status {
	width, height := termbox.Size()
	return &Status{
		row:     row,
		bgColor: termbox.ColorDefault,
		fgColor: termbox.ColorDefault,
		width:   width,
		height:  height,
	}
}

func (s *Status) Clear() {
	for i := 0; i < s.width; i++ {
		termbox.SetCell(i, s.row, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
	}
}

func (s *Status) Message(message string, delay int64) chan struct{} {
	s.message = message
	s.bgColor = termbox.ColorDefault
	s.fgColor = termbox.ColorWhite
	return s.display([]rune(message), delay)
}

func (s *Status) Info(message string, delay int64) chan struct{} {
	s.message = message
	s.bgColor = termbox.ColorCyan
	s.fgColor = termbox.ColorBlack
	return s.display([]rune(message), delay)
}

func (s *Status) Warn(message string, delay int64) chan struct{} {
	s.message = message
	s.bgColor = termbox.ColorYellow
	s.fgColor = termbox.ColorBlack
	return s.display([]rune(message), delay)
}

func (s *Status) Error(message string, delay int64) chan struct{} {
	s.message = message
	s.bgColor = termbox.ColorRed
	s.fgColor = termbox.ColorWhite
	return s.display([]rune(message), delay)
}

func (s *Status) display(message []rune, delay int64) chan struct{} {
	s.Clear()
	w, _ := termbox.Size()
	for i, r := range message {
		termbox.SetCell(i, s.row, r, s.fgColor, s.bgColor)
	}
	for i := len(message); i < w; i++ {
		termbox.SetCell(i, s.row, rune(' '), termbox.ColorDefault, s.bgColor)
	}
	termbox.Flush()

	wait := make(chan struct{}, 1)
	go func() {
		if delay > 0 {
			<-time.After(time.Duration(delay) * time.Second)
		}
		wait <- struct{}{}
	}()
	return wait
}

func (s *Status) resize(width, height int) {
	s.width = width
	s.height = height
	if s.message != "" {
		s.display([]rune(s.message), 0)
	}
}
