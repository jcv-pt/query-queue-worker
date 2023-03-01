// Package stats gathers and presents real time engine statistic data
package stats

import (
	"github.com/olekukonko/tablewriter"
	"os"
	"query-queue-worker/engine"
	"strconv"
)

// Shows an ASCII table with engine stat data
func ShowTable() {
	// Get engine data
	var engineData = engine.GetData()
	// Create table data
	data := [][]string{
		[]string{
			"Pending",
			engineData.Processes.Pending.LastRun.Format("15:04:05"),
			strconv.Itoa(engineData.Processes.Pending.Count.Total),
			strconv.Itoa(engineData.Processes.Pending.Count.Successful),
			strconv.Itoa(engineData.Processes.Pending.Count.Failed),
			strconv.Itoa(len(engineData.Processes.Pending.Count.Blacklist)),
		},
		[]string{
			"Update",
			engineData.Processes.Update.LastRun.Format("15:04:05"),
			strconv.Itoa(engineData.Processes.Update.Count.Total),
			strconv.Itoa(engineData.Processes.Update.Count.Successful),
			strconv.Itoa(engineData.Processes.Update.Count.Failed),
			strconv.Itoa(len(engineData.Processes.Update.Count.Blacklist)),
		},
		[]string{
			"Maintenance",
			engineData.Processes.Update.LastRun.Format("15:04:05"),
			strconv.Itoa(engineData.Processes.Update.Count.Total),
			"--",
			strconv.Itoa(engineData.Processes.Update.Count.Failed),
			"--",
		},
	}
	// Init new table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Process Type", "Last Run", "Total", "Successful", "Failed", "Blacklist"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}
