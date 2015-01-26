package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/rchargel/localiday/conf"
	"github.com/rchargel/localiday/db"
	"github.com/rchargel/localiday/domain"
	"github.com/rchargel/localiday/server"
	"github.com/rchargel/localiday/util"
)

func main() {
	start := time.Now()
	config := conf.LoadConfiguration()
	fmt.Println(config.ToString())

	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)
	util.Log(util.Info, "Running on %v cores.", cores)
	sport := os.Getenv("PORT")

	port, err := strconv.ParseUint(sport, 10, 16)
	if err != nil {
		util.Log(util.Fatal, "Could not read port", err)
	}
	err = db.NewDatabase("postgres", "postgres", "localhost", "localiday")
	if err != nil {
		util.Log(util.Fatal, "Could not connect to database.", err)
	}
	err = domain.BootStrap()
	if err != nil {
		util.Log(util.Fatal, "Could not bootstrap database.", err)
	}
	appServer := server.AppServer{Port: uint16(port)}
	appServer.Start()

	util.Log(util.Info, "Application started in %v.", time.Since(start))
}
