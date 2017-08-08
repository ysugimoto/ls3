package main

// Store selecting paramter struct
type SelectorState struct {

	// Cussor pointer row
	pointer int

	// Displaying page number
	page int

	// Displaying list amount
	listSize int

	// Max page number
	maxPage int

	// Filtering query
	filters []rune

	// List items
	items Selectable
}

// Make new state pointer struct
func NewSelectorState(list Selectable) *SelectorState {
	return &SelectorState{
		page:    1,
		filters: []rune{},
		items:   list,
	}
}

// Down cursor
func (s *SelectorState) DownCursor(step int) (old, updated int, paging bool) {
	old = s.pointer
	if s.pointer+step < s.listSize {
		s.pointer += step
	} else if s.page < s.maxPage {
		s.pointer = 0
		s.page++
		paging = true
	} else { // go to first page from last
		s.pointer = 0
		s.page = 1
		paging = true
	}
	updated = s.pointer
	return
}

// Up cursor
func (s *SelectorState) UpCursor(step int) (old, updated int, paging bool) {
	old = s.pointer
	if s.pointer-step >= 0 {
		s.pointer -= step
	} else if s.page == 1 { // go to last page from first
		// lazy update pointer
		s.pointer = -1
		s.page = s.maxPage
		paging = true
	} else {
		// lazy update pointer
		s.pointer = -2
		s.page--
		paging = true
	}
	updated = s.pointer
	return
}

// Update page state
func (s *SelectorState) updatePage(maxPage int) {
	if s.page > maxPage {
		s.page = 1
	} else if s.page < 1 {
		s.page = maxPage
	}
	s.maxPage = maxPage
}

// Pop filter word
func (s *SelectorState) popFilter() bool {
	if len(s.filters) == 0 {
		return false
	}
	s.filters = s.filters[0 : len(s.filters)-1]
	return true
}

// Add filter word
func (s *SelectorState) addFilter(f rune) {
	s.filters = append(s.filters, f)
}
