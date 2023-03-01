// Package log initializes logger with settings and provides a Writter variable to be used with other packages
package log

import (
	"github.com/antigloss/go/logger"
	"query-queue-worker/types"
)

var Writer *logger.Logger

// Initializes package
// Parameters:
//   - settings (* types.AppConfig) : Pointer of the app settings
//   - silent (bool) : Weather logs should be output to stdout
func Init(settings *types.AppConfig, silent bool) {
	var destination = logger.LogDestBoth
	if settings.Logs.Enabled == false {
		destination = logger.LogDestConsole
	}
	if silent {
		if settings.Logs.Enabled == false {
			destination = logger.LogDestNone
		} else {
			destination = logger.LogDestFile
		}
	}
	//if config.Settings.Logs.Enabled {
	logger, _ := logger.New(&logger.Config{
		LogDir:          settings.Logs.Path,
		LogFileMaxSize:  settings.Logs.MaxSize,
		LogFileMaxNum:   settings.Logs.MaxCount,
		LogFileNumToDel: 1,
		LogLevel:        logger.LogLevelInfo,
		LogDest:         destination,
		Flag:            logger.ControlFlagLogLineNum,
	})
	// Assign writer variable
	Writer = logger
}
