package main

import (
	"testing"
)

func TestFindHighlightRange(t *testing.T) {
	first, last := findHighlightRange("Lorem ipsum", "ipsum")
	if first != 6 {
		t.Errorf("first expected 6, actual %d", first)
	}
	if last != 11 {
		t.Errorf("last expected 11, actual %d", last)
	}
}
