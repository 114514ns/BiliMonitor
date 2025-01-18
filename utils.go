package main

import (
	"fmt"
	"time"
)

func substr(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}
func formatTime(input string) string {
	if input == "0000-00-00 00:00:00" {
		return "Invalid Date"
	}

	// Define layout compatible with the input
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, input)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return "Parsing Error"
	}
	return t.Format(layout)
}

func abs(a int) int {
	if a < 0 {
		return -a
	} else {
		return a
	}
}
