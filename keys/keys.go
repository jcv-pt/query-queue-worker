// Package keys binds keyboard pressings to events
package keys

import (
	"os"
	"os/exec"
	"query-queue-worker/engine/stats"
	"strings"
)

// Flag that indicates if keypressing handler routine should be shutdown
var stop = false

// Initializes package
func Init() {
	// Create a routine that constantly checks for key inputs
	go func() {
		// Disable input buffering on terminal
		exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
		// Disable entered characters on the screen
		exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
		var b []byte = make([]byte, 1)
		for stop == false {
			os.Stdin.Read(b)
			handleKeyPress(string(b))
		}
	}()
}

// Returns shutdown status, other libs can rely on this method to know if the the "Q" (quit) key has been pressed
func IsShuttingDown() bool {
	return stop
}

// Set shutdown flag which will stop go routine to check for key pressings and reset terminal
func Shutdown() {
	// Set stopping flag
	stop = true
	// Restore terminal settings
	exec.Command("stty", "-F", "/dev/tty", "sane").Run()
}

// Handler that processes events based on the key pressed
//
// Parameters:
//   - key (string) : String representation of the key that was pressed
func handleKeyPress(key string) {
	// Apply lowercase transformation
	key = strings.ToLower(key)
	switch key {
	// Quit case : Provides shutdown flag to app
	case "q":
		Shutdown()
		break
	// Stats case : Shows stats table
	case "s":
		stats.ShowTable()
		break
	}
}
