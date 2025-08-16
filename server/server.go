package server

import (
	"database/sql"
	"fmt"
	"net/http"

	u "hack/backend/utils"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DatabaseDriver int

const (
	POSTGRESQL DatabaseDriver = iota
)

var DatabaseDrivers = map[DatabaseDriver]string{
	POSTGRESQL: "pgx",
}

func Driver(driver DatabaseDriver) string {
	return DatabaseDrivers[driver]
}

type DatabaseConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Database string
	SSL      string
	Driver   string
}

type Database struct {
	db          *sql.DB
	IsConnected bool
	Config      DatabaseConfig
}

func (this *Database) Connect() {
	if this.IsConnected {
		return
	}

	var err error

	var connectionString = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s port=%d sslmode=%s", this.Config.User, this.Config.Password, this.Config.Database, this.Config.SSL, this.Config.Host, this.Config.Port, this.Config.SSL)

	this.db, err = sql.Open(this.Config.Driver, connectionString)
	if err != nil || this.db == nil {
		this.db = nil
		u.Error("Failed to Call Connect to Database!")
	} else {
		u.Info("Called Connect to Database")
	}

	err = this.db.Ping()
	if err != nil {
		this.db = nil
		this.IsConnected = false
		u.ErrorF("Failed to Connect To Database!:\t%s\n", err.Error())
		return
	}

	u.Info("Successfully Connected To Database!")
	this.IsConnected = true
}

func (this *Database) Close() {
	this.db.Close()
	this.IsConnected = false
	// u.Info("Closing Database Connection")
}

// / (this *Database) Query(query string, rows *sql.Rows, args ...any) (cols []string, error error)
func (this *Database) Query(query string, rows **sql.Rows, args ...any) (cols []string, err error) {
	var _err error

	*rows, _err = this.db.Query(query, args...)
	if _err != nil {
		return nil, err
	}

	var _cols []string
	_cols, _err = (*rows).Columns()
	if _err != nil {
		return nil, err
	}

	return _cols, nil
}

func IsURL_Index(r *http.Request) bool {
	switch r.RequestURI {
	case "/", "/index", "/index.html", "/index.htm", "/index.php":
		return true
	}

	return false
}
