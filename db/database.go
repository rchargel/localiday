package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coopernurse/gorp"
	// The postgresql driver.
	_ "github.com/lib/pq"
	"github.com/rchargel/localiday/app"
)

// Database a connection to the database.
type Database struct {
	Username string
	Password string
	Hostname string
	Database string

	*gorp.DbMap
}

type rollback struct {
	version uint16
	script  string
	name    string
}

type update struct {
	version uint16
	script  string
	name    string
}

type rollbackSorter []rollback
type updateSorter []update

// DB the root object for database call
var DB *Database

// NewDatabase creates a new connection to the database.
func NewDatabase(username, password, hostname, database string, rebuild bool) error {
	start := time.Now()
	db := &Database{Username: username, Password: password, Hostname: hostname, Database: database}
	if rebuild {
		db.recreateDatabase()
	}
	dbMap, error := db.init()

	app.Log(app.Info, "Data initialized in %v.", time.Since(start))

	DB = &Database{username, password, hostname, database, dbMap}
	return error
}

func (db *Database) recreateDatabase() error {
	config := app.LoadConfiguration()
	conn, err := sql.Open("postgres", fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", db.Username, db.Password, db.Hostname, db.Database))
	defer conn.Close()
	if err == nil {
		rows, err := conn.Query("select version from application")
		var version uint16
		version = 0
		if err == nil {
			if rows.Next() {
				if inrerr := rows.Scan(&version); inrerr != nil {
					app.Log(app.Fatal, "Could not read row count.", inrerr)
				}
				rows.Close()
			}
		}

		err = recreateTables(version, config.DBVersion, conn)
	}
	return err
}

func (db *Database) init() (*gorp.DbMap, error) {
	conn, err := sql.Open("postgres", fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", db.Username, db.Password, db.Hostname, db.Database))

	config := app.LoadConfiguration()

	rows, err := conn.Query("select version from application")
	var version uint16
	version = 0
	if err != nil {
		err = createTables(version, config.DBVersion, conn)
	} else {
		if rows.Next() {
			if inrerr := rows.Scan(&version); inrerr != nil {
				app.Log(app.Fatal, "Could not read row count.", inrerr)
			}
			rows.Close()
		}
		if version != config.DBVersion {
			err = createTables(version, config.DBVersion, conn)
		}
	}

	dbmap := &gorp.DbMap{Db: conn, Dialect: gorp.PostgresDialect{}}

	return dbmap, err
}

func processUpdates(currVersion, appVersion uint16, updates []update, conn *sql.DB) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	for _, update := range updates {
		if update.version > currVersion && update.version <= appVersion {
			err = runSqlFile(conn, update.name, update.script)
			if err != nil {
				return err
			}
		}
	}
	err = tx.Commit()
	return err
}

func processRollbacks(currVersion, appVersion uint16, rollbacks []rollback, conn *sql.DB) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	for _, rollback := range rollbacks {
		if rollback.version > appVersion && rollback.version <= currVersion {
			err = runSqlFile(conn, rollback.name, rollback.script)
			if err != nil {
				return err
			}
		}
	}
	err = tx.Commit()
	return err
}

func runSqlFile(conn *sql.DB, filename, script string) error {
	app.Log(app.Info, "Running sql script %v.", filename)
	_, err := conn.Exec(script)

	return err
}

func createTables(currVersion, appVersion uint16, conn *sql.DB) error {
	app.Log(app.Debug, "Creating database tables")
	return runSqlFiles(currVersion, appVersion, conn)
}

func recreateTables(currVersion, appVersion uint16, conn *sql.DB) error {
	app.Log(app.Debug, "Recreating database tables")
	start := time.Now()
	rollbackFiles, updateFiles, err := getAllSqlFiles()
	if err == nil {
		app.Log(app.Info, "REBUILDING DATABASE")
		processRollbacks(65535, 0, rollbackFiles, conn)
		processUpdates(currVersion, appVersion, updateFiles, conn)
		app.Log(app.Info, "DATABASE REBUILD COMPLETE (%v)", time.Since(start))
	}
	return err
}

func runSqlFiles(currVersion, appVersion uint16, conn *sql.DB) error {
	start := time.Now()
	rollbackFiles, updateFiles, err := getAllSqlFiles()

	if err == nil {
		app.Log(app.Info, "RUNNING DATABASE MIGRATION")
		processRollbacks(currVersion, appVersion, rollbackFiles, conn)
		processUpdates(currVersion, appVersion, updateFiles, conn)
		app.Log(app.Info, "DATABASE MIGRATION COMPLETE (%v)", time.Since(start))
	}

	return err
}

func getAllSqlFiles() ([]rollback, []update, error) {
	rollbackFiles := make([]rollback, 0, 10)
	updateFiles := make([]update, 0, 10)

	files, err := ioutil.ReadDir("sql")
	if err == nil {
		for _, file := range files {
			if strings.Contains(file.Name(), "update_") {
				update, err := createUpdate(file)
				if err != nil {
					return rollbackFiles, updateFiles, err
				}
				updateFiles = append(updateFiles, update)
			} else if strings.Contains(file.Name(), "rollback_") {
				rollback, err := createRollback(file)
				if err != nil {
					return rollbackFiles, updateFiles, err
				}
				rollbackFiles = append(rollbackFiles, rollback)
			}
		}
		sort.Sort(updateSorter(updateFiles))
		sort.Sort(rollbackSorter(rollbackFiles))
	}
	return rollbackFiles, updateFiles, err
}

func createRollback(file os.FileInfo) (rollback, error) {
	var r rollback
	version, script, err := createContentAndVersion(file)

	if err == nil {
		r = rollback{version: version, script: script, name: file.Name()}
	}
	return r, err
}

func createUpdate(file os.FileInfo) (update, error) {
	var u update
	version, script, err := createContentAndVersion(file)

	if err == nil {
		u = update{version: version, script: script, name: file.Name()}
	}
	return u, err
}

func createContentAndVersion(file os.FileInfo) (uint16, string, error) {
	name := file.Name()
	si := strings.Index(name, "_") + 1
	ei := strings.Index(name, ".sql")
	v := name[si:ei]
	version, err := strconv.ParseUint(v, 10, 16)

	if err != nil {
		return 0, "", err
	}
	data, err := ioutil.ReadFile("sql/" + name)
	if err != nil {
		return 0, "", err
	}

	return uint16(version), string(data), err
}

func (l updateSorter) Len() int           { return len(l) }
func (l updateSorter) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l updateSorter) Less(i, j int) bool { return l[i].version < l[j].version }

func (l rollbackSorter) Len() int           { return len(l) }
func (l rollbackSorter) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l rollbackSorter) Less(i, j int) bool { return l[i].version > l[j].version }
