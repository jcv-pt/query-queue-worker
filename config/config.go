// Package config loads configuration file "query-queue-config.json" into a public Settings variable to be accessed publicly
//
// It relies on types.go package to parse JSON structure
package config

import (
	"query-queue-worker/types"
	"query-queue-worker/util"
)

var Settings = types.AppConfig{} // Holds configuration from the JSON config file

// Loads config from JSON config file
func Init() {
	err := util.ReadJson("query-queue-config.json", &Settings)
	if err != nil {
		util.Die("Error: cannot load config\n %v\n", err.Error())
	}
}
