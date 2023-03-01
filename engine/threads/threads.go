// Package threads handles thread allocation logic for the engine
//
// Responsible for the following tasks:
//  - Keep tracking of current number of active and idle threads per job type (pending, update, maintenance)
//  - Allocate new threads according to the config settings and job type
//  - Distribute thread allocation according to the job type and demand for that particular job type
//  - If defined on the settings, wait for all thread completion before allowing application to shutdown
package threads

import (
	"github.com/creasty/defaults"
	_ "github.com/creasty/defaults"
	"log"
	"query-queue-worker/config"
	"query-queue-worker/types"
	"query-queue-worker/util"
	"sync"
)

var stats = types.EngineThreads{}
var wg = sync.WaitGroup{}
var mu = sync.Mutex{}

var Type = types.EngineThreadsProcessType{} // Returns the thread process types

// Initializes package
func Init() {
	// Initialize values
	stats.Max = config.Settings.Threads.Max
	// Initialize Type
	defaults.Set(&Type)
	// Initialize max counts: minimum of thread count per task is 1
	if stats.Max < 3 {
		util.Die("Error: Thread max count is too low, at least 3 threads are required")
	}
	stats.Pending.Max = 0
	stats.Update.Max = 0
	stats.Maintenance.Max = 0
}

// Returns the remaining count of available threads to allocate
func GetAllocationCount() int {
	return stats.Max - stats.Used
}

// Returns the available threads per processing type
//
// Parameters:
//   - processType string : Reference to the types.EngineThreadsProcessType as string
func GetUsedCount(processType string) int {
	var available = 0
	switch processType {
	case Type.Pending:
		available = stats.Pending.Used
		break
	case Type.Update:
		available = stats.Update.Used
		break
	case Type.Maintenance:
		available = stats.Maintenance.Used
		break
	}
	return available
}

// Returns the available threads per processing type
//
// Parameters:
//   - processType string : Reference to the types.EngineThreadsProcessType as string
func GetAvailableCount(processType string) int {
	var available = 0
	switch processType {
	case Type.Pending:
		available = stats.Pending.Max - stats.Pending.Used
		break
	case Type.Update:
		available = stats.Update.Max - stats.Update.Used
		break
	case Type.Maintenance:
		available = stats.Maintenance.Max - stats.Maintenance.Used
		break
	}
	if available < 0 {
		return 0
	}
	return available
}

// Allocate thread distribution dependent on current load and processing types
//
// Parameters:
//   - allocateToPending (int) : How many threads needed to be allocated to pending type jobs
//   - allocateToUpdate (int) : How many threads needed to be allocated to update type jobs
//   - allocateToMaintenance (int) : How many threads needed to be allocated to maintenance type jobs
func Allocate(allocateToPending int, allocateToUpdate int, allocateToMaintenance int) {
	var totalAvailable = stats.Max - stats.Used
	// Check if we can proceed with allocation
	if totalAvailable < 1 {
		return
	}
	// For Maintenance: this is a small task and only run on intervals, furthermore only one process is needed
	if allocateToMaintenance >= 1 {
		stats.Maintenance.Max = 1
		totalAvailable--
	}
	// Check if we can proceed with allocation
	if totalAvailable < 1 {
		return
	}
	/********* Allocation Scenarios *********/
	// Check if we can accommodate all allocations
	if allocateToPending+allocateToUpdate <= totalAvailable {
		stats.Pending.Max = stats.Pending.Used + allocateToPending
		stats.Update.Max = stats.Update.Used + allocateToUpdate
		return
	}
	// If we cant accommodate all allocations, then try to distribute with 50% each
	if allocateToPending > 0 {
		// For Pending: we want to focus our main allocation on pending
		stats.Pending.Max = stats.Pending.Used + (totalAvailable / 2)
		stats.Update.Max = stats.Update.Used + (totalAvailable % 2)
	} else if allocateToPending == 0 && allocateToUpdate >= 1 {
		// For Update: we want to focus our main allocation here
		stats.Pending.Max = stats.Pending.Used
		stats.Update.Max = stats.Pending.Used + totalAvailable
	}
}

// Add thread count
//
// Parameters:
//   - processType string : Reference to the types.EngineThreadsProcessType as string
func Add(processType string) {
	// Lock sync
	mu.Lock()
	// Add to the pool and threadCount
	wg.Add(1)
	// Increment stats
	switch processType {
	case Type.Pending:
		stats.Pending.Used++
		stats.Used++
		break
	case Type.Update:
		stats.Update.Used++
		stats.Used++
		break
	case Type.Maintenance:
		stats.Maintenance.Used++
		stats.Used++
		break
	}
	// Unlock sync
	mu.Unlock()
}

// Remove thread count
//
// Parameters:
//   - processType string : Reference to the types.EngineThreadsProcessType as string
func Remove(processType string) {
	// Lock sync
	mu.Lock()
	// Remove from pool and threadCount
	wg.Done()
	// Decrement stats
	switch processType {
	case Type.Pending:
		stats.Pending.Used--
		stats.Used--
		break
	case Type.Update:
		stats.Update.Used--
		stats.Used--
		break
	case Type.Maintenance:
		stats.Maintenance.Used--
		stats.Used--
		break
	}
	// Unlock sync
	mu.Unlock()
}

// Wait for WorkGroup threads to finish
func Wait() {
	// Wait for threads to finish
	if config.Settings.Threads.WaitToFinish && stats.Used >= 1 {
		log.Printf("Waiting for %d threads to finish...", stats.Used)
		wg.Wait()
	}
}
