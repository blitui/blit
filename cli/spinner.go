package cli

import (
	"fmt"
	"sync"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinnerResult is returned by Spin to stop the animation.
type SpinnerResult struct {
	// Stop halts the spinner and clears the line.
	Stop func()
}

// Spin shows an animated spinner with a message. Call Stop() on the returned
// SpinnerResult when done. The spinner runs in a background goroutine and
// clears the line on stop.
func Spin(message string) *SpinnerResult {
	done := make(chan struct{})
	var once sync.Once

	go func() {
		i := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				// Clear the line
				fmt.Printf("\r\033[K")
				return
			case <-ticker.C:
				fmt.Printf("\r%s %s", spinnerFrames[i%len(spinnerFrames)], message)
				i++
			}
		}
	}()

	return &SpinnerResult{
		Stop: func() {
			once.Do(func() { close(done) })
		},
	}
}
