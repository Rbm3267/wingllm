package ui

import (
	"fmt"
	"strings"
	"time"
)

// PlayPlaneBanner plays an animated sweep of a minimal jet flying left-to-right,
// revealing the word "W I N G L L M" in its contrail.
func PlayPlaneBanner() {
	targetWord := "W I N G L L M"
	totalWidth := 52
	startReveal := 18 // Column where targetWord begins

	// Jet representation (minimal sleek outline)
	// We use a clean 3-line aerodynamic silhouette
	// Line 0:      \
	// Line 1:  ─────\───►
	// Line 2:      /

	fmt.Println()

	steps := totalWidth + 10
	for step := 0; step <= steps; step++ {
		// Construct the middle line (the contrail + revealed letters + plane)
		var midLine strings.Builder
		var topLine strings.Builder
		var botLine strings.Builder

		planeHead := step
		planeTail := step - 7

		for col := 0; col < totalWidth; col++ {
			if col >= planeTail && col <= planeHead {
				// We are in the plane section
				offset := col - planeTail
				switch offset {
				case 0, 1, 2:
					midLine.WriteString("─")
					topLine.WriteString(" ")
					botLine.WriteString(" ")
				case 3:
					midLine.WriteString("─")
					topLine.WriteString("\\")
					botLine.WriteString("/")
				case 4, 5, 6:
					midLine.WriteString("─")
					topLine.WriteString(" ")
					botLine.WriteString(" ")
				case 7:
					midLine.WriteString("►")
					topLine.WriteString(" ")
					botLine.WriteString(" ")
				default:
					midLine.WriteString("─")
					topLine.WriteString(" ")
					botLine.WriteString(" ")
				}
			} else if col < planeTail {
				// Behind the plane: reveal the target word if within bounds, else contrail
				if col >= startReveal && col < startReveal+len(targetWord) {
					charIndex := col - startReveal
					char := targetWord[charIndex]
					midLine.WriteByte(char)
				} else if col%2 == 0 {
					midLine.WriteString("·")
				} else {
					midLine.WriteString(" ")
				}
				topLine.WriteString(" ")
				botLine.WriteString(" ")
			} else {
				// Ahead of the plane: blank space
				midLine.WriteString(" ")
				topLine.WriteString(" ")
				botLine.WriteString(" ")
			}
		}

		// Print frame (using ANSI cursor up to overwrite in place)
		if step == 0 {
			fmt.Println(topLine.String())
			fmt.Println(midLine.String())
			fmt.Println(botLine.String())
		} else {
			// Move up 3 lines and rewrite
			fmt.Printf("\033[3A\033[2K%s\n\033[2K%s\n\033[2K%s\n", topLine.String(), midLine.String(), botLine.String())
		}

		time.Sleep(22 * time.Millisecond)
	}

	// Final rest state: clean, modern title
	fmt.Printf("\033[3A\033[2K\n\033[2K  \033[1;36mW I N G L L M\033[0m   \033[90m|   2x Speculative Turbo for Local AI\033[0m\n\033[2K\n\n")
}
