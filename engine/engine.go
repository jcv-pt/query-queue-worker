// Package engine contains all the logic to for the worker: lookups database and creates processes for jobs
package engine

import (
	"fmt"
	"github.com/creasty/defaults"
	"os/exec"
	"query-queue-worker/config"
	"query-queue-worker/database"
	"query-queue-worker/engine/threads"
	"query-queue-worker/log"
	"query-queue-worker/types"
	"query-queue-worker/util"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var engine = types.Engine{}

// Initializes package
func Init() {
	// Initialize engine data
	defaults.Set(&engine)
	// Initialize threads
	threads.Init()
}

// Starts worker thread
func Start() {
	// Starting engine
	log.Writer.Info("Starting worker engine...")
	// Start
	engine.Status = "started"
	// Start engine cycle
	go func() {
		for engine.Status == "started" {
			// Re-check only after idle timeout
			if engine.Cycles == -1 || engine.Cycles >= config.Settings.Worker.Idle {
				// Reset elapsed so that we count another idle timeout
				engine.Cycles = 0
				// Check for availability to allocate new jobs
				if threads.GetAllocationCount() > 0 {
					// Process new jobs lookup and allocation
					pendingCount, updateCount, maintenanceCount := processLookup()
					// Notify
					log.Writer.Infof("Lookup for pending jobs: Pending(%d) ; Update(%d) ; Maintenance(%d);", pendingCount, updateCount, maintenanceCount)
					// Process pending queries
					if pendingCount > 0 {
						processPending()
					}
					// Process update on current queries
					if updateCount > 0 {
						processUpdate()
					}
					// Process maintenance
					if maintenanceCount > 0 {
						processMaintenance()
					}
				}
			}
			// Increment elapsed and sleep for one second
			engine.Cycles++
			time.Sleep(time.Second)
		}
		// Wait for threads to complete
		threads.Wait()
	}()
}

// Stops worker thread
func Stop() {
	// Notify
	log.Writer.Info("Stopping worker engine...")
	// Set worker status
	engine.Status = "stopped"
}

// Gets engine data
//
// Return:
//   - types.Engine : Engine struct with worker statistics data
func GetData() types.Engine {
	return engine
}

// Processes database lookup for pending jobs (pending, update, maintenance) and tries to allocate threads based on the number of jobs required
//
// Returns:
//   - totalPending (int) : Total number of jobs of "pending" type
//   - totalUpdate (int) : Total number of jobs of "update" type
//   - totalMaintenance (int) : Total number of jobs of "maintenance" type
func processLookup() (totalPending int, totalUpdate int, totalMaintenance int) {
	// Check for pending and update jobs count
	var query = `
		SELECT
			COUNT(IF(runStatus = 'pending',1,NULL)) AS TotalPending,
			COUNT(IF(runStatus = 'completed',1,NULL)) AS TotalUpdate
		FROM tblCRQueryQueue
		WHERE
			(runStatus = 'pending') OR
			(runStatus = 'completed' AND runRepeat IS NOT NULL AND (runNext IS NULL OR runNext <= NOW()))`
	result := database.Con.QueryRow(query)
	// Get allocation counts
	var err = result.Scan(&totalPending, &totalUpdate)
	if err != nil {
		util.Die("Error: cannot select allocation from CrQueryQueue table \n %v\n", err.Error())
	}
	// Check for next run on maintenance job
	totalMaintenance = 0
	var lastRun = engine.Processes.Maintenance.LastRun.Unix()
	var idle = time.Duration(config.Settings.Worker.Processes.Maintenance.Idle)
	var nextRun = time.Unix(lastRun, 0).Add(time.Second * idle)
	if nextRun.Before(time.Now()) {
		totalMaintenance = 1
	}
	// Allocate
	threads.Allocate(totalPending, totalUpdate, totalMaintenance)
	return
}

// Lookup for queries with pending status db and starts new threads based on jobs that it finds
func processPending() {
	// Get current available threads
	var availableThreads = threads.GetAvailableCount(threads.Type.Pending)
	if availableThreads <= 0 {
		log.Writer.Info("Skipping pending process, no threads available")
		return
	}
	// Check for pending jobs with no more than available threads
	var query = `
		SELECT
			querySignature,
			queryName
		FROM tblCRQueryQueue 
		WHERE 
		    runStatus = 'pending'
		ORDER BY 
		    runFirst IS NULL DESC,
		    pkQueryQueueID ASC
		LIMIT ` + strconv.Itoa(availableThreads)
	results, err := database.Con.Query(query)
	if err != nil {
		util.Die("Error: cannot select pending tasks from CrQueryQueue table \n %v\n", err.Error())
	}
	// Create new workers for each query
	var counter = 0
	for results.Next() {
		var row = new(types.TblCRQueryQueue)
		// For each row, scan the result into our tag composite object
		err = results.Scan(&row.QuerySignature, &row.QueryName)
		if err != nil {
			util.Die("Error: cannot scan pending tasks from CrQueryQueue table \n %v\n", err.Error())
		}
		// Process query
		go processJob(row.QuerySignature, row.QueryName, threads.Type.Pending, row.QuerySignature)
		counter++
	}
	// Report if no queries are pending
	if counter <= 0 {
		log.Writer.Info("No pending queries to be processed...")
	}
}

// Lookup for queries that require update from db and starts new threads based on jobs that it finds
func processUpdate() {
	// Get current available threads
	var availableThreads = threads.GetAvailableCount(threads.Type.Update)
	if availableThreads <= 0 {
		log.Writer.Info("Skipping update process, no threads available")
		return
	}
	// Check for pending jobs with no more than available threads
	var query = `
		SELECT
			querySignature,
			queryName
		FROM tblCRQueryQueue 
		WHERE 
		    runStatus = 'completed' AND
		    runRepeat IS NOT NULL AND 
		    (runNext IS NULL OR runNext <= NOW())
		ORDER BY 
		    runFirst IS NULL DESC,
		    runLast IS NULL DESC,
		    pkQueryQueueID ASC
		LIMIT ` + strconv.Itoa(availableThreads)
	results, err := database.Con.Query(query)
	if err != nil {
		util.Die("Error: cannot select update tasks from CrQueryQueue table \n %v\n", err.Error())
	}
	// Create new workers for each query
	var counter = 0
	for results.Next() {
		var row = new(types.TblCRQueryQueue)
		// For each row, scan the result into our tag composite object
		err = results.Scan(&row.QuerySignature, &row.QueryName)
		if err != nil {
			util.Die("Error: cannot scan update tasks from CrQueryQueue table \n %v\n", err.Error())
		}
		// Process query
		go processJob(row.QuerySignature, row.QueryName, threads.Type.Update, row.QuerySignature)
		counter++
	}
	// Report if no queries are pending
	if counter <= 0 {
		log.Writer.Info("No queries to be updated...")
	}
}

// Processes maintenance job type
func processMaintenance() {
	// Get current available threads
	var availableThreads = threads.GetAvailableCount(threads.Type.Maintenance)
	if availableThreads <= 0 {
		log.Writer.Info("Skipping maintenance process, no threads available")
		return
	}
	// Process maintenance
	go processJob("MAINT", "System Maintenance", threads.Type.Maintenance)
}

// Creates a new threaded process for running a job
// Returns:
//   - jobId (string) : The unique identifier of this job, should be set to the query signature apart from the maintenance job type
//   - jobName (string) : The job name, usually column "name" of the query to be processed
//   - threadType (string) : String representation of the threadType to run (pending, update, maintenance)
//   - cmdArgs (...interface{}) : Arguments passed to the shell cmd command defined in the Settings config
func processJob(jobId string, jobName string, threadType string, cmdArgs ...string) {
	// Initialize new thread
	threads.Add(threadType)
	var threadId = strconv.Itoa(threads.GetUsedCount(threadType))
	// Get command based on thread type
	var cmd = "echo 1"
	switch threadType {
	case threads.Type.Pending:
		cmd = config.Settings.Worker.Commands.Single
		break
	case threads.Type.Update:
		cmd = config.Settings.Worker.Commands.Update
		break
	case threads.Type.Maintenance:
		cmd = config.Settings.Worker.Commands.Maintenance
		break
	}
	// Build command args
	var command = cmd
	for _, arg := range cmdArgs {
		command = fmt.Sprintf(command, arg)
	}
	args := strings.Split(command, " ")
	// Build identifier
	var jobIdentifier = threadType + " | Thread" + threadId + " : "
	// Run command
	log.Writer.Info(jobIdentifier + "Running new job with ID #" + jobId + ": " + jobName)
	exec := exec.Command(config.Settings.Worker.Executable, args...)
	// Prevent CMD from stopping execution when syscall.SIGINT is issued
	exec.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	out, err := exec.CombinedOutput()
	var output = string(out)
	if err != nil {
		util.Die(jobIdentifier+"Error: cannot execute command \n %s\n", output)
	}
	var lines = strings.Split(output, "\n")
	for _, line := range lines {
		if line != "" {
			log.Writer.Info(jobIdentifier + line)
		}
	}
	// Finalize thread count
	threads.Remove(threadType)
	// Add to Engine stats
	addStats(jobId, threadType, err != nil)
	// Notify
	log.Writer.Info(jobIdentifier + "Finalized job")
}

// Adds statistical data relevant to a job (pending, update, maintenance) into the engine statistics struct
// Returns:
//   - identifier (string) : The unique identifier of this job, should be set to the query signature apart from the maintenance job type
//   - threadType (string) : String representation of the threadType to run (pending, update, maintenance)
//   - successful (bool) : Weather this job was or not processed successfully
func addStats(identifier string, threadType string, successful bool) {
	switch threadType {
	case threads.Type.Pending:
		engine.Processes.Pending.Count.Total++
		engine.Processes.Pending.LastRun = time.Now()
		if successful {
			engine.Processes.Pending.Count.Successful++
		} else {
			engine.Processes.Pending.Count.Failed++
			// TODO Maybe create a statistical table with this info at some point
			//engine.Processes.Pending.Count.Blacklist = append(engine.Processes.Pending.Count.Blacklist, identifier)
		}
		break
	case threads.Type.Update:
		engine.Processes.Update.Count.Total++
		engine.Processes.Update.LastRun = time.Now()
		if successful {
			engine.Processes.Update.Count.Successful++
		} else {
			engine.Processes.Update.Count.Failed++
			// TODO Maybe create a statistical table with this info at some point
			//engine.Processes.Update.Count.Blacklist = append(engine.Processes.Update.Count.Blacklist, identifier)
		}
		break
	case threads.Type.Maintenance:
		engine.Processes.Maintenance.Count.Total++
		engine.Processes.Maintenance.LastRun = time.Now()
		if successful {
			engine.Processes.Maintenance.Count.Successful++
		} else {
			engine.Processes.Maintenance.Count.Failed++
		}
		break
	}
}
