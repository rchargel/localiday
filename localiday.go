package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/rchargel/localiday/app"
	"github.com/rchargel/localiday/db"
	"github.com/rchargel/localiday/web"
)

func main() {
	var rebuildDB bool
	flag.BoolVar(&rebuildDB, "recreate-db", false, "Rebuilds the database from the ground up.")
	flag.Parse()

	start := time.Now()
	config := app.LoadConfiguration()
	fmt.Println(config.ToString())

	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)
	app.Log(app.Info, "Running on %v cores.", cores)
	sport := os.Getenv("PORT")

	port, err := strconv.ParseUint(sport, 10, 16)
	if err != nil {
		app.Log(app.Fatal, "Could not read port", err)
	}
	err = db.NewDatabase("postgres", "postgres", "localhost", "localiday", rebuildDB)
	if err != nil {
		app.Log(app.Fatal, "Could not connect to database.", err)
	}
	err = db.BootStrap()
	if err != nil {
		app.Log(app.Fatal, "Could not bootstrap database.", err)
	}
	app.Log(app.Info, "Application started in %v.", time.Since(start))

	appServer := web.AppServer{Port: uint16(port)}
	appServer.Start()
}
