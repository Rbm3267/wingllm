package ui

import (
	"fmt"
	"strings"
	"time"
)

// PlayPlaneBanner renders a clean, single-line aerodynamic sweep.
// A minimal supersonic jet (✈) glides smoothly across the line,
// revealing "WINGLLM" in its contrail with zero vertical jitter.
func PlayPlaneBanner() {
	width := 36
	brand := "W I N G L L M"
	brandStart := 10

	// 1-line sleek sweep
	for i := 0; i <= width+4; i++ {
		var line strings.Builder
		line.WriteString("\r  ")

		for col := 0; col < width; col++ {
			if col == i {
				// The aircraft
				line.WriteString("\033[1;36m✈\033[0m")
			} else if col < i {
				// Revealed zone behind the aircraft
				if col >= brandStart && col < brandStart+len(brand) {
					charIdx := col - brandStart
					line.WriteString(fmt.Sprintf("\033[1;37m%c\033[0m", brand[charIdx]))
				} else {
					line.WriteString("\033[90m─\033[0m")
				}
			} else {
				// Ahead of the aircraft
				line.WriteString(" ")
			}
		}

		fmt.Print(line.String())
		time.Sleep(16 * time.Millisecond)
	}

	// Final polished resting state
	fmt.Print("\r\033[K  \033[1;36m✈  WINGLLM\033[0m  \033[90m·  2x Speculative Turbo for Local AI\033[0m\n\n")
}
