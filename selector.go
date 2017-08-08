package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"math"
	"strings"
	"sync"
)

// Selectable items struct
type Selector struct {

	// Row offset
	offset int

	// Flag of incremental search
	enableFilter bool

	// Duplicate guard
	guard chan struct{}

	// Screen width
	width int

	// Screen height
	height int

	// Key handling mutex
	mutex *sync.Mutex

	// DI: status struct
	status *Status

	// Resize channel
	onResize chan struct{}

	// Key event channel
	onKeyPress chan termbox.Event
}

// Struct pointer maker
func NewSelector(rowOffset int, status *Status) *Selector {
	width, height := termbox.Size()
	return &Selector{
		offset:       rowOffset,
		enableFilter: true,
		guard:        make(chan struct{}, 1),
		width:        width,
		height:       height,
		mutex:        new(sync.Mutex),
		status:       status,
		onResize:     make(chan struct{}, 1),
		onKeyPress:   make(chan termbox.Event, 1),
	}
}

// Switch disabling filter
func (s *Selector) WithOutFilter() *Selector {
	s.enableFilter = false
	return s
}

// Switch enabling filter
func (s *Selector) WithFilter() *Selector {
	s.enableFilter = true
	return s
}

// Pre handle keyPress event from App
func (s *Selector) keyPress(evt termbox.Event) {
	if len(s.guard) > 0 {
		s.onKeyPress <- evt
	}
}

// Pre handle resize event from App
func (s *Selector) resize(width, height int) {
	s.width = width
	s.height = height

	if len(s.guard) > 0 {
		s.onResize <- struct{}{}
	}
}

// Change row offset
func (s *Selector) SetOffset(offset int) *Selector {
	s.offset = offset
	return s
}

// Choose item from selectable list
func (s *Selector) Choose(list Selectable) (int, error) {
	s.guard <- struct{}{}

	defer func() {
		<-s.guard
	}()

	// start select
	selected := make(chan int, 1)
	errChan := make(chan error, 1)
	go s.doSelect(list, selected, errChan)

	return <-selected, <-errChan
}

func (s *Selector) doSelect(list Selectable, selected chan int, errChan chan error) {
	state := NewSelectorState(list)
	s.display(state)

	for {
		select {

		// Handle resize event
		case <-s.onResize:
			s.display(state)

		// Handle key event
		case evt := <-s.onKeyPress:
			logger.log("Handle keypress")
			s.mutex.Lock()
			switch {

			// Pressed Ctrl+C or Esc
			case evt.Key == termbox.KeyCtrlC || evt.Key == termbox.KeyEsc:
				selected <- 0
				errChan <- fmt.Errorf("interrupted")

			// Pressed Arrow-Down key
			case evt.Key == termbox.KeyArrowDown:
				old, updated, paging := state.DownCursor(1)
				logger.log("Down cursor")
				s.inactive(old)
				s.active(updated)
				if paging {
					s.display(state)
				}
				termbox.Flush()

			// Pressed Arrow-Up key
			case evt.Key == termbox.KeyArrowUp:
				old, updated, paging := state.UpCursor(1)
				logger.log("Up cursor")
				s.inactive(old)
				if paging {
					if updated == -1 {
						s.display(state)
						state.pointer = state.listSize - 1
						s.active(state.pointer)
					} else {
						state.pointer = s.height - s.offset - 1
						s.active(state.pointer)
						s.display(state)
					}
				} else {
					s.active(updated)
				}
				termbox.Flush()

			// Pressed Enter key
			case evt.Key == termbox.KeyEnter:
				logger.log("Press Enter")
				index, err := s.getFilteredIndex(state)
				selected <- index
				errChan <- err
				s.mutex.Unlock()
				return

			// Pressed Backspace
			case s.enableFilter && evt.Key == termbox.KeyBackspace2:
				logger.log("Press Backspace")
				if update := state.popFilter(); update {
					s.display(state)
				}

			// Other character key
			case s.enableFilter && evt.Ch > 0:
				logger.log("Press " + string(evt.Ch))
				state.addFilter(evt.Ch)
				s.display(state)
			}
			s.mutex.Unlock()
		}
	}
}

// Get selected item considering with filter query
func (s *Selector) getFilteredIndex(state *SelectorState) (int, error) {
	_, indexMap := s.filterList(state)
	index := (state.page-1)*s.height + state.pointer

	if indexMap == nil {
		return index, nil
	}
	if v, ok := indexMap[index]; !ok {
		return 0, fmt.Errorf("Unexpected index")
	} else {
		return v, nil
	}
}

// Filter list items by input query, and returns list and indexed map
func (s *Selector) filterList(state *SelectorState) (Selectable, map[int]int) {
	if len(state.filters) == 0 {
		return state.items, nil
	}
	filter := string(state.filters)
	filtered := Selectable{}
	// key is filtered index, value is real item index
	indexMap := make(map[int]int)
	index := 0
	for i, v := range state.items {
		if strings.Contains(v.String(), filter) {
			filtered = append(filtered, v)
			indexMap[index] = i
			index++
		}
	}
	return filtered, indexMap
}

// Display selectable UI
func (s *Selector) display(state *SelectorState) {
	s.Clear()
	// Get filtered list items
	filtered, _ := s.filterList(state)
	// Cauclaute max page
	state.updatePage(int(math.Ceil(float64(len(filtered)) / float64(s.height))))
	// Calcualte start and end index
	start := (state.page - 1) * (s.height - s.offset)
	end := start + (s.height - s.offset)
	if end > len(filtered) {
		end = len(filtered)
	}

	// Slice list per page and write to term
	displayList := filtered[start:end]
	strFilter := string(state.filters)
	state.listSize = 0
	pointer := 0
	for i, line := range displayList {
		line.Write(i+s.offset, strFilter)
		if state.pointer == i {
			s.active(state.pointer)
			pointer = state.pointer
		}
		state.listSize++
	}
	state.pointer = pointer
	if s.enableFilter {
		s.displayInfo(len(filtered), state.page, state.maxPage)
		s.status.Message(fmt.Sprintf("Filter query> %s", string(state.filters)), 0)
	}
	termbox.Flush()
}

// Display filtered total item amounts and page / maxPage
func (s *Selector) displayInfo(listLen, page, maxPage int) {
	info := []rune(fmt.Sprintf("(Total %d: %d of %d)", listLen, page, maxPage))
	x := s.width - len(info)

	// FIXME: This is fragile...
	for i := x - 10; i < x; i++ {
		termbox.SetCell(i, 0, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
	for _, r := range info {
		termbox.SetCell(x, 0, r, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}

// Clear the termbox buffer only selector drawable indexes
func (s *Selector) Clear() {
	for i := s.offset; i < s.height; i++ {
		for j := 0; j < s.width; j++ {
			termbox.SetCell(j, i, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

// Inactive cursor
func (s *Selector) inactive(pointer int) {
	index := (pointer + s.offset) * s.width
	cb := termbox.CellBuffer()
	for i := 0; i < s.width; i++ {
		cell := cb[index+i]
		cell.Bg = termbox.ColorDefault
		cb[index+i] = cell
	}
}

// Activate cursor
func (s *Selector) active(pointer int) {
	index := (pointer + s.offset) * s.width
	cb := termbox.CellBuffer()
	for i := 0; i < s.width; i++ {
		cell := cb[index+i]
		cell.Bg = termbox.ColorMagenta
		cb[index+i] = cell
	}
}
