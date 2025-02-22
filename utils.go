package main

import (
	"fmt"
	"os"
	"sort"
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
func Last(dir string) (fileName string, modTime time.Time, err error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return "", time.Time{}, err
	}
	var onlyFiles []os.DirEntry
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			onlyFiles = append(onlyFiles, entry)
		}
	}
	if len(onlyFiles) == 0 {
		return "", time.Time{}, fmt.Errorf("no files found in the directory: %s", dir)
	}
	sort.Slice(onlyFiles, func(i, j int) bool {
		infoI, _ := onlyFiles[i].Info()
		infoJ, _ := onlyFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})
	latestFile := onlyFiles[0]
	info, err := latestFile.Info()
	if err != nil {
		return "", time.Time{}, err
	}
	return latestFile.Name(), info.ModTime(), nil
}

func abs(a int) int {
	if a < 0 {
		return -a
	} else {
		return a
	}
}
