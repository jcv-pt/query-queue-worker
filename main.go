package main

import (
	"flag"
	"query-queue-worker/config"
	"query-queue-worker/database"
	"query-queue-worker/engine"
	"query-queue-worker/keys"
	"query-queue-worker/log"
	"query-queue-worker/os"
)

// Initializes app components and starts worker
func main() {
	/**************** ARGS ****************/
	var silentMode = flag.Bool("silent", false, "Weather to display stdout")
	flag.Parse()
	/**************** INIT ****************/
	// Load config
	config.Init()
	// Load log
	log.Init(&config.Settings, *silentMode)
	// Load MYSQL
	database.Load()
	// Init OS package (handle OS sigterms)
	os.Init()
	// Init keys package (handle keyboard bindings)
	keys.Init()
	// Init engine
	engine.Init()
	/**************** BANNER ****************/
	log.Writer.Infof("Query-Queue-Worker : V0.1")
	/**************** START ****************/
	// Start worker engine (start processing)
	engine.Start()
	// Start app
	for keys.IsShuttingDown() == false && os.IsShuttingDown() == false {
		// Endless loop until OS signal or user input to Quit
	}
	/**************** SHUTDOWN ****************/
	// Stop keys
	keys.Shutdown()
	// Stop engine
	engine.Stop()
}
