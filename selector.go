package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"math"
	"strings"
)

type Selector struct {
	offset       int
	enableFilter bool
	guard        chan struct{}
	interrupt    chan error
	width        int
	height       int

	status *Status

	onResize   chan struct{}
	onKeyPress chan termbox.Event
}

func NewSelector(rowOffset int, status *Status) *Selector {
	width, height := termbox.Size()
	return &Selector{
		offset:       rowOffset,
		enableFilter: true,
		guard:        make(chan struct{}, 1),
		interrupt:    make(chan error, 1),
		width:        width,
		height:       height,

		status: status,

		onResize:   make(chan struct{}, 1),
		onKeyPress: make(chan termbox.Event, 1),
	}
}

func (s *Selector) WithOutFilter() *Selector {
	s.enableFilter = false
	return s
}

func (s *Selector) WithFilter() *Selector {
	s.enableFilter = false
	return s
}

func (s *Selector) keyPress(evt termbox.Event) {
	if len(s.guard) > 0 {
		s.onKeyPress <- evt
	}
}

func (s *Selector) resize(width, height int) {
	s.width = width
	s.height = height

	if len(s.guard) > 0 {
		s.onResize <- struct{}{}
	}
}

func (s *Selector) SetOffset(offset int) *Selector {
	s.offset = offset
	return s
}

func (s *Selector) Choose(list Selectable) (int, error) {
	s.guard <- struct{}{}

	defer func() {
		<-s.guard
	}()

	selected := make(chan int, 1)
	errChan := make(chan error, 1)
	go s.control(list, selected, errChan)

	index := <-selected
	err := <-errChan
	return index, err
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
		select {
		case evt := <-s.onKeyPress:
			switch {
			case evt.Key == termbox.KeyCtrlC:
				fallthrough
			case evt.Key == termbox.KeyEsc:
				selected <- 0
				errChan <- fmt.Errorf("interrupted")

			case evt.Key == termbox.KeyArrowDown:
				if pointer+1 < listSize {
					logger.log("Down cursor")
					s.inactive(pointer)
					pointer++
					s.active(pointer)
					termbox.Flush()
				} else if maxPage > 1 {
					logger.log(fmt.Sprintf("Paging. poiner: %d", pointer))
					s.inactive(pointer)
					pointer = 0
					s.active(pointer)
					listSize, page, maxPage, pointer = s.display(list, filters, page+1, pointer)
				}
			case evt.Key == termbox.KeyArrowUp:
				if pointer-1 >= 0 {
					logger.log("Up cursor")
					s.inactive(pointer)
					pointer--
					s.active(pointer)
					termbox.Flush()
				} else if maxPage > 1 {
					if page == 1 { // back from first to last
						listSize, page, maxPage, pointer = s.display(list, filters, page-1, pointer)
						s.inactive(pointer)
						pointer = listSize - 1
						s.active(pointer)
						termbox.Flush()
					} else {
						s.inactive(pointer)
						_, h := termbox.Size()
						pointer = h - s.offset - 1
						logger.log(fmt.Sprintf("Paging. poiner: %d", pointer))
						s.active(pointer)
						listSize, page, maxPage, pointer = s.display(list, filters, page-1, pointer)
					}
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
		case <-s.onResize:
			listSize, page, maxPage, pointer = s.display(list, filters, page, pointer)
		}
	}
}

func (s *Selector) getFilteredIndex(list Selectable, filters []rune, page, pointer int) (int, error) {
	_, indexMap := s.filterList(list, filters)
	_, height := termbox.Size()
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
	maxPage := int(math.Ceil(float64(len(filtered)) / float64(s.height)))
	if page > maxPage {
		page = 1
	} else if page < 1 {
		page = maxPage
	}
	start := (page - 1) * (s.height - s.offset)
	end := start + (s.height - s.offset)
	if end > len(filtered) {
		end = len(filtered)
	}
	displayList := filtered[start:end]
	s.Clear()
	pointerFound := 0
	strFilter := string(filters)
	for i, line := range displayList {
		line.Write(i+s.offset, strFilter)
		if pointer == i {
			s.active(pointer)
			pointerFound = pointer
		}
	}
	if s.enableFilter {
		s.displayInfo(len(filtered), page, maxPage)
		s.status.Message(fmt.Sprintf("Filter query> %s", string(filters)), 0)
	}
	termbox.Flush()

	return len(displayList), page, maxPage, pointerFound
}

func (s *Selector) displayInfo(listLen, page, maxPage int) {
	info := []rune(fmt.Sprintf("(Total %d: %d of %d)", listLen, page, maxPage))
	x := s.width - len(info)
	for _, r := range info {
		termbox.SetCell(x, 0, r, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}

func (s *Selector) Clear() {
	for i := s.offset; i < s.height; i++ {
		for j := 0; j < s.width; j++ {
			termbox.SetCell(j, i, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func (s *Selector) inactive(pointer int) {
	index := (pointer + s.offset) * s.width
	cb := termbox.CellBuffer()
	for i := 0; i < s.width; i++ {
		cell := cb[index+i]
		cell.Bg = termbox.ColorDefault
		cb[index+i] = cell
	}
}

func (s *Selector) active(pointer int) {
	index := (pointer + s.offset) * s.width
	cb := termbox.CellBuffer()
	for i := 0; i < s.width; i++ {
		cell := cb[index+i]
		cell.Bg = termbox.ColorMagenta
		cb[index+i] = cell
	}
}
