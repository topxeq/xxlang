// Script to update the build number in main.go
// Usage: go run scripts/update_build.go [sequence]
//   - sequence: optional daily sequence number (default: 01)
//
// The build number format is: YYYYMMDDNN (year month day + 2-digit sequence)
// Example: 2026032401 means March 24, 2026, first build of the day

package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

func main() {
	// Get current date
	now := time.Now()
	datePrefix := now.Format("20060102") // YYYYMMDD

	// Get sequence number from argument or default to 01
	seq := "01"
	if len(os.Args) > 1 {
		seq = os.Args[1]
		if len(seq) == 1 {
			seq = "0" + seq
		}
	}

	buildNumber := datePrefix + seq

	// Read main.go
	mainPath := "cmd/xxl/main.go"
	content, err := os.ReadFile(mainPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", mainPath, err)
		os.Exit(1)
	}

	// Find and replace BuildNumber
	re := regexp.MustCompile(`BuildNumber\s*=\s*"\d+"`)
	newContent := re.ReplaceAllString(string(content), fmt.Sprintf(`BuildNumber = "%s"`, buildNumber))

	if newContent == string(content) {
		fmt.Println("BuildNumber pattern not found in main.go")
		os.Exit(1)
	}

	// Write back
	err = os.WriteFile(mainPath, []byte(newContent), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", mainPath, err)
		os.Exit(1)
	}

	// Extract current sequence and suggest next
	seqInt, _ := strconv.Atoi(seq)
	nextSeq := fmt.Sprintf("%02d", seqInt+1)

	fmt.Printf("Build number updated to: %s\n", buildNumber)
	fmt.Printf("For next build, run: go run scripts/update_build.go %s\n", nextSeq)
}
