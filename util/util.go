// Package util has methods to help debugging and json handling
package util

import (
	"encoding/json"
	"fmt"
	debugger "log"
	"os"
	"os/exec"
	"query-queue-worker/log"
)

// Reads JSON file into variable
//
// Parameters:
//   - fileName (string) : Absolute path of the json file to read
//   - v (interface{}) : Object on which the json value will be stored
func ReadJson(fileName string, v interface{}) (err error) {
	configFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	// Decode
	err = json.NewDecoder(configFile).Decode(&v)
	// Close file
	closeErr := configFile.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}

// Prints a Printf message and exists app
//
// Parameters:
//   - msg (string) : An exit message shown that will be printed on stdout
//   - params (...strings) : Collection of strings to be used in the SprintF
func Die(msg string, params ...string) {
	var error = msg
	for _, param := range params {
		error = fmt.Sprintf(error, param)
	}
	log.Writer.Error(error)
	Exit(1)
}

// Prints variables and exists app
//
// Parameters:
//   - data (...interface{}) : Any type data that will be printed on stdout
func Debug(data ...interface{}) {
	debugger.Print(data)
	Exit(1)
}

// Exits app
//
// Parameters:
//   - code (int) : Exit code for the app
func Exit(code int) {
	// Restore terminal settings
	exec.Command("stty", "-F", "/dev/tty", "sane").Run()
	// Exit
	os.Exit(code)
}
