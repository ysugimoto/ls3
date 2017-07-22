package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
)

type Selectable []fmt.Stringer

type Selector struct {
	offset    int
	guard     chan struct{}
	keyEvent  chan termbox.Event
	interrupt chan error
}

func NewSelector(rowOffset int) *Selector {
	return &Selector{
		offset: rowOffset,

		guard:     make(chan struct{}, 1),
		keyEvent:  make(chan termbox.Event, 1),
		interrupt: make(chan error, 1),
	}
}

func (s *Selector) SetOffset(offset int) {
	s.offset = offset
}

func (s *Selector) Choose(list Selectable) (int, error) {
	s.guard <- struct{}{}

	defer func() {
		<-s.guard
	}()

	pointer := 0
	selected := make(chan int, 1)
	errChan := make(chan error, 1)
	s.display(list)
	go s.pollEvent()
	go func() {
		for {
			select {
			case evt := <-s.keyEvent:
				switch {
				case evt.Ch == 106, evt.Key == termbox.KeyArrowDown:
					if pointer+1 < len(list) {
						s.inactive(pointer)
						pointer++
						s.active(pointer)
						termbox.Flush()
					}
				case evt.Ch == 107, evt.Key == termbox.KeyArrowUp:
					if pointer-1 >= 0 {
						s.inactive(pointer)
						pointer--
						s.active(pointer)
						termbox.Flush()
					}
				case evt.Key == termbox.KeyEnter:
					selected <- pointer
					errChan <- nil
					return
				}
			case err := <-s.interrupt:
				selected <- 0
				errChan <- err
				return
			}
		}
	}()

	return <-selected, <-errChan
}

func (s *Selector) display(lines Selectable) {
	_, height := termbox.Size()
	for i, line := range lines[0:height] {
		for j, r := range []rune(line.String()) {
			termbox.SetCell(j, i+s.offset, r, termbox.ColorDefault, termbox.ColorDefault)
		}
	}
	s.active(0)
	termbox.Flush()
}

func (s *Selector) inactive(pointer int) {
	width, _ := termbox.Size()
	index := (pointer + s.offset) * width
	cb := termbox.CellBuffer()
	for i := 0; i < width; i++ {
		cell := cb[index+i]
		cell.Fg = termbox.ColorWhite
		cell.Bg = termbox.ColorDefault
		cb[index+i] = cell
	}
}

func (s *Selector) active(pointer int) {
	width, _ := termbox.Size()
	index := (pointer + s.offset) * width
	cb := termbox.CellBuffer()
	for i := 0; i < width; i++ {
		cell := cb[index+i]
		cell.Fg = termbox.ColorWhite
		cell.Bg = termbox.ColorMagenta
		cb[index+i] = cell
	}
}

func (s *Selector) pollEvent() {
	for {
		switch evt := termbox.PollEvent(); evt.Type {
		case termbox.EventKey:
			if evt.Key == termbox.KeyCtrlC || evt.Key == termbox.KeyEsc {
				if len(s.guard) > 0 {
					s.interrupt <- fmt.Errorf("interrupted")
				}
				return
			}
			if len(s.guard) > 0 {
				s.keyEvent <- evt
			}
		}
	}
}
