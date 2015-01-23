package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coopernurse/gorp"
	// The postgresql driver.
	_ "github.com/lib/pq"
	"github.com/rchargel/localiday/server"
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
func NewDatabase(username, password, hostname, database string) error {
	start := time.Now()
	db := &Database{Username: username, Password: password, Hostname: hostname, Database: database}
	dbMap, error := db.init()

	log.Printf("Database initialized in %v", time.Since(start))

	DB = &Database{username, password, hostname, database, dbMap}
	return error
}

func (db *Database) init() (*gorp.DbMap, error) {
	conn, err := sql.Open("postgres", fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", db.Username, db.Password, db.Hostname, db.Database))

	config := server.LoadConfiguration()

	rows, err := conn.Query("select version from application")
	var version uint16
	version = 0
	if err != nil {
		err = createTables(version, conn)
	} else {
		if rows.Next() {
			if inrerr := rows.Scan(&version); inrerr != nil {
				log.Fatal("Could not read row count", inrerr)
			}
			rows.Close()
		}
		if version != config.DBVersion {
			err = createTables(version, conn)
		}
	}

	dbmap := &gorp.DbMap{Db: conn, Dialect: gorp.PostgresDialect{}}

	return dbmap, err
}

func runSqlFiles(version uint16, conn *sql.DB) error {
	start := time.Now()
	files, err := ioutil.ReadDir("sql")
	if err != nil {
		return err
	}

	rollbackFiles := make([]rollback, 0, 10)
	updateFiles := make([]update, 0, 10)

	for _, file := range files {
		if strings.Contains(file.Name(), "update_") {
			update, err := createUpdate(file)
			if err != nil {
				return err
			}
			updateFiles = append(updateFiles, update)
		} else if strings.Contains(file.Name(), "rollback_") {
			rollback, err := createRollback(file)
			if err != nil {
				return err
			}
			rollbackFiles = append(rollbackFiles, rollback)
		}
	}
	sort.Sort(updateSorter(updateFiles))
	sort.Sort(rollbackSorter(rollbackFiles))

	log.Println("RUNNING DATABASE MIGRATION")
	processRollbacks(version, rollbackFiles, conn)
	processUpdates(version, updateFiles, conn)
	log.Printf("DATABASE MIGRATION COMPLETE: %v", time.Since(start))

	return err
}

func processUpdates(version uint16, updates []update, conn *sql.DB) error {
	config := server.LoadConfiguration()

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	for _, update := range updates {
		if update.version > version && update.version <= config.DBVersion {
			err = runSqlFile(conn, update.name, update.script)
			if err != nil {
				return err
			}
		}
	}
	err = tx.Commit()
	return err
}

func processRollbacks(version uint16, rollbacks []rollback, conn *sql.DB) error {
	config := server.LoadConfiguration()

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	for _, rollback := range rollbacks {
		if rollback.version > config.DBVersion && rollback.version <= version {
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
	log.Println(filename)
	_, err := conn.Exec(script)

	return err
}

func createTables(version uint16, conn *sql.DB) error {
	log.Print("Creating database tables")
	return runSqlFiles(version, conn)
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
