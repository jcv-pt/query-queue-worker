// Package os listens for OS exit signals and provides this flag as a public method
package os

import (
	"os"
	"os/signal"
	"syscall"
)

// Flag that indicates if a signal as been received by the OS
var stop = false

// Initializes package
func Init() {
	// Bind OS signals
	handleOSSignals()
}

// Returns shutdown status, other libs can rely on this method to know if the OS is being shutdown
func IsShuttingDown() bool {
	return stop
}

// Set shutdown flag which will stop go routine to check for OS signals
func Shutdown() {
	// Set shutdown flag
	stop = true
}

// Intercepts OS shutdown signals in order to attempt to stop engine gracefully
func handleOSSignals() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
		for stop == false {
			s := <-c
			switch s {
			case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
				Shutdown()
			case syscall.SIGHUP:
			case syscall.SIGSEGV:
			default:
				return
			}
		}
	}()
}
