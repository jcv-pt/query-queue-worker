// Package types defines all app type data
package types

import "time"

/************ AppConfig ************/

type AppConfig struct {
	Debug   bool             `json:"debug"`
	Logs    AppConfigLogs    `json:"logs"`
	Threads AppConfigThreads `json:"threads"`
	Mysql   AppConfigMysql   `json:"mysql"`
	Worker  AppConfigWorker  `json:"worker"`
}

type AppConfigLogs struct {
	Enabled  bool   `json:"enabled"`
	Path     string `json:"path"`
	MaxSize  uint32 `json:"maxSize"`
	MaxCount int    `json:"maxCount"`
}

type AppConfigMysql struct {
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AppConfigWorker struct {
	Idle       int                      `json:"idle"`
	Executable string                   `json:"executable"`
	Commands   AppConfigWorkerCommands  `json:"commands"`
	Processes  AppConfigWorkerProcesses `json:"processes"`
}

type AppConfigWorkerCommands struct {
	Single      string `json:"single"`
	Update      string `json:"update"`
	Maintenance string `json:"maintenance"`
}

type AppConfigWorkerProcesses struct {
	Maintenance AppConfigWorkerProcessesMaintenance `json:"maintenance"`
}

type AppConfigWorkerProcessesMaintenance struct {
	Idle int `json:"idle"`
}

type AppConfigThreads struct {
	Max          int  `json:"max"`
	WaitToFinish bool `json:"waitToFinish"`
}

/************ Engine ************/

type Engine struct {
	Status    string `default:"stopped"`
	Cycles    int    `default:"-1"`
	Processes EngineProcesses
}

type EngineProcesses struct {
	Pending     EngineProcessType `default:"{\"name\": \"Pending\"}"`
	Update      EngineProcessType `default:"{\"name\": \"Update\"}"`
	Maintenance EngineProcessType `default:"{\"name\": \"Maintenance\"}"`
}

type EngineProcessType struct {
	Name    string    `default:"Unknown"`
	LastRun time.Time `default:"time.Now()"`
	Count   EngineProcessTypeCounts
}

type EngineProcessTypeCounts struct {
	Failed     int `default:"0"`
	Successful int `default:"0"`
	Total      int `default:"0"`
	Blacklist  []string
}

/************ Engine Threads ************/

type EngineThreads struct {
	Max         int `default:"0"`
	Used        int `default:"0"`
	Pending     EngineThreadsProcessTypeStat
	Update      EngineThreadsProcessTypeStat
	Maintenance EngineThreadsProcessTypeStat
}

type EngineThreadsProcessType struct {
	Pending     string `default:"Pending"`
	Update      string `default:"Update"`
	Maintenance string `default:"Maintenance"`
}

type EngineThreadsProcessTypeStat struct {
	Max        int `default:"0"`
	Used       int `default:"0"`
	Percentage int `default:"0"`
}

/************ SQL Tables ************/

type TblCRQueryQueue struct {
	PkQueryQueueID int    `TbField:"pkQueryQueueID"`
	RunStatus      string `TbField:"runStatus"`
	RunError       string `TbField:"runError"`
	RunTime        int    `TbField:"runTime"`
	RunRepeat      string `TbField:"runRepeat"`
	RunFirst       string `TbField:"runFirst"`
	RunLast        string `TbField:"runLast"`
	RunNext        string `TbField:"runNext"`
	QueryName      string `TbField:"queryName"`
	QuerySignature string `TbField:"querySignature"`
}
