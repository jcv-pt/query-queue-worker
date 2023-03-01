// Package database connects and provides MYSQL database connection
package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"query-queue-worker/config"
	"query-queue-worker/util"
)

// SQL Connection to the server
var Con *sql.DB

// Opens a new connection to MYSQL server
func Load() {
	// Build conn string
	var dsn = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		config.Settings.Mysql.Username,
		config.Settings.Mysql.Password,
		config.Settings.Mysql.Hostname,
		config.Settings.Mysql.Port,
		config.Settings.Mysql.Database,
	)
	db, err := sql.Open("mysql", dsn)
	Con = db
	if err != nil {
		util.Die("Error: cannot connect to MYSQL\n %v\n", err.Error())
	}
	// Set max idle conns on db in order to prevent packages.go mysql eof error
	db.SetMaxIdleConns(0)
}
