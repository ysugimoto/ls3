package main

import (
	"time"
)

// Writer interface for selectable
type Writer interface {
	Write(y int, filters []rune)
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

// Check rune contains in rune slice
func isRuneContains(haystack []rune, needle rune) (contains bool) {
	for _, r := range haystack {
		if r == needle {
			contains = true
			break
		}
	}
	return
}
