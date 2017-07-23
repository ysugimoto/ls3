package main

import (
	"github.com/nsf/termbox-go"
	"time"
)

type Status struct {
	row     int
	bgColor termbox.Attribute
	fgColor termbox.Attribute
}

func NewStatus(row int) *Status {
	return &Status{
		row:     row,
		bgColor: termbox.ColorDefault,
		fgColor: termbox.ColorDefault,
	}
}

func (s *Status) Clear() {
	w, _ := termbox.Size()
	for i := 0; i < w; i++ {
		termbox.SetCell(i, s.row, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
	}
}

func (s *Status) Message(message string, delay int64) chan struct{} {
	s.bgColor = termbox.ColorDefault
	s.fgColor = termbox.ColorWhite
	return s.display([]rune(message), delay)
}

func (s *Status) Info(message string, delay int64) chan struct{} {
	s.bgColor = termbox.ColorCyan
	s.fgColor = termbox.ColorBlack
	return s.display([]rune(message), delay)
}

func (s *Status) Warn(message string, delay int64) chan struct{} {
	s.bgColor = termbox.ColorYellow
	s.fgColor = termbox.ColorBlack
	return s.display([]rune(message), delay)
}

func (s *Status) Error(message string, delay int64) chan struct{} {
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
