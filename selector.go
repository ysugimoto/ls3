package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"math"
	"strings"
)

type Writer interface {
	Write(y int)
	String() string
}
type Selectable []Writer

type Selector struct {
	offset       int
	enableFilter bool
	guard        chan struct{}
	eventQueue   chan termbox.Event
	keyEvent     chan termbox.Event
	resizeEvent  chan termbox.Event
	interrupt    chan error
}

func NewSelector(rowOffset int) *Selector {
	return &Selector{
		offset:       rowOffset,
		enableFilter: true,

		guard:       make(chan struct{}, 1),
		eventQueue:  make(chan termbox.Event, 1),
		keyEvent:    make(chan termbox.Event, 1),
		resizeEvent: make(chan termbox.Event, 1),
		interrupt:   make(chan error, 1),
	}
}

func (s *Selector) WithNoFilter() *Selector {
	s.enableFilter = false
	return s
}

func (s *Selector) SetOffset(offset int) {
	s.offset = offset
}

func (s *Selector) Choose(list Selectable) (int, error) {
	s.guard <- struct{}{}

	defer func() {
		<-s.guard
	}()

	selected := make(chan int, 1)
	errChan := make(chan error, 1)
	stopPoll := make(chan struct{}, 1)
	go s.control(list, selected, errChan)
	go func() {
		for {
			select {
			case <-stopPoll:
				return
			default:
				s.eventQueue <- termbox.PollEvent()
			}
		}
	}()

	index := <-selected
	err := <-errChan
	stopPoll <- struct{}{}
	return index, err
}

func (s *Selector) getSize() (width, height int) {
	width, height = termbox.Size()
	height -= s.offset

	return
}

func (s *Selector) control(list Selectable, selected chan int, errChan chan error) {
	var (
		pointer  int = 0
		page     int = 1
		listSize int
		maxPage  int
		filters  []rune = []rune{}
	)
	listSize, page, maxPage, pointer = s.display(list, filters, page, pointer)

	for {
		evt := <-s.eventQueue
		switch evt.Type {
		case termbox.EventKey:
			switch {
			case evt.Key == termbox.KeyCtrlC:
				fallthrough
			case evt.Key == termbox.KeyEsc:
				selected <- 0
				errChan <- fmt.Errorf("interrupted")

			case evt.Key == termbox.KeyArrowDown:
				if pointer+1 < listSize {
					s.inactive(pointer)
					pointer++
					s.active(pointer)
					termbox.Flush()
				} else if maxPage > 1 {
					listSize, page, maxPage, pointer = s.display(list, filters, page+1, pointer)
				}
			case evt.Key == termbox.KeyArrowUp:
				if pointer-1 >= 0 {
					s.inactive(pointer)
					pointer--
					s.active(pointer)
					termbox.Flush()
				} else if maxPage > 1 {
					page--
					listSize, page, maxPage, pointer = s.display(list, filters, page-1, pointer)
				}
			case evt.Key == termbox.KeyEnter:
				index, err := s.getFilteredIndex(list, filters, page, pointer)
				selected <- index
				errChan <- err
				return
			case s.enableFilter && evt.Key == termbox.KeyBackspace2:
				if len(filters) > 0 {
					filters = filters[0 : len(filters)-1]
					listSize, page, maxPage, pointer = s.display(list, filters, page, pointer)
				}
			case s.enableFilter && evt.Ch > 0:
				filters = append(filters, evt.Ch)
				listSize, page, maxPage, pointer = s.display(list, filters, page, pointer)
			}
		case termbox.EventResize:
			listSize, page, maxPage, pointer = s.display(list, filters, page, pointer)
		}
	}
}

func (s *Selector) getFilteredIndex(list Selectable, filters []rune, page, pointer int) (int, error) {
	_, indexMap := s.filterList(list, filters)
	_, height := s.getSize()
	index := (page-1)*height + pointer

	if indexMap == nil {
		return index, nil
	}
	if v, ok := indexMap[index]; !ok {
		return 0, fmt.Errorf("Unexpected index")
	} else {
		return v, nil
	}
}

func (s *Selector) filterList(list Selectable, filters []rune) (Selectable, map[int]int) {
	if len(filters) == 0 {
		return list, nil
	}
	filter := string(filters)
	filtered := Selectable{}
	indexMap := make(map[int]int)
	index := 0
	for i, v := range list {
		if strings.Contains(v.String(), filter) {
			filtered = append(filtered, v)
			indexMap[index] = i
			index++
		}
	}
	return filtered, indexMap
}

func (s *Selector) display(lines Selectable, filters []rune, page, pointer int) (int, int, int, int) {
	filtered, _ := s.filterList(lines, filters)
	if len(filtered) == 0 {
		return 0, 1, 0, 0
	}
	_, height := s.getSize()
	maxPage := int(math.Ceil(float64(len(filtered)) / float64(height)))
	if page > maxPage {
		page = maxPage
	} else if page < 1 {
		page = 1
	}
	start := (page - 1) * height
	end := start + height
	if end > len(filtered) {
		end = len(filtered)
	}
	displayList := filtered[start:end]
	s.Clear()
	pointerFound := 0
	for i, line := range displayList {
		line.Write(i + s.offset)
		// for j, r := range []rune(line.String()) {
		// 	termbox.SetCell(j, i+s.offset, r, termbox.ColorDefault, termbox.ColorDefault)
		// }
		if pointer == i {
			s.active(pointer)
			pointerFound = pointer
		}
	}
	termbox.Flush()
	if s.enableFilter {
		status := NewStatus(1)
		status.Message(fmt.Sprintf("Filter query> %s", string(filters)), 0)
	}
	return len(displayList), page, maxPage, pointerFound
}

func (s *Selector) Clear() {
	width, height := s.getSize()
	for i := s.offset; i < height; i++ {
		for j := 0; j < width; j++ {
			termbox.SetCell(j, i, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func (s *Selector) inactive(pointer int) {
	width, _ := termbox.Size()
	index := (pointer + s.offset) * width
	cb := termbox.CellBuffer()
	for i := 0; i < width; i++ {
		if len(cb) < index+i {
			continue
		}
		cell := cb[index+i]
		cell.Bg = termbox.ColorDefault
		cb[index+i] = cell
	}
}

func (s *Selector) active(pointer int) {
	width, _ := termbox.Size()
	index := (pointer + s.offset) * width
	cb := termbox.CellBuffer()
	for i := 0; i < width; i++ {
		if len(cb) < index+i {
			continue
		}
		cell := cb[index+i]
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
		case termbox.EventResize:
			if len(s.guard) > 0 {
				s.resizeEvent <- evt
			}
		}
	}
}
