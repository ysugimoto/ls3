package main

import (
	"strings"
	"time"
)

// Writer interface for selectable
type Writer interface {
	Write(y int, filter string)
	String() string
}

// Selectable slice type
type Selectable []Writer

// We're living in Asia/Tokyo location :)
var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

// Transform from UTC to JST
func utcToJst(utc time.Time) string {
	jst := utc.In(JST)
	return jst.Format("2006-01-02 15:03:04")
}

// Find and get highlight range
func findHighlightRange(haystack, needle string) (first, last int) {
	if needle == "" {
		first = -1
		last = -1
		return
	}

	first = strings.Index(haystack, needle)
	if first == -1 {
		last = -1
		return
	}
	last = first + len(needle)

	return
}
