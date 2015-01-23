package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/rchargel/localiday/conf"
	"github.com/rchargel/localiday/db"
	"github.com/rchargel/localiday/domain"
	"github.com/rchargel/localiday/server"
)

func main() {
	start := time.Now()
	config := conf.LoadConfiguration()
	fmt.Println(config.ToString())

	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)
	log.Printf("Running on %v cores.\n", cores)
	sport := os.Getenv("PORT")

	port, err := strconv.ParseUint(sport, 10, 16)
	if err != nil {
		log.Panic(err)
	}
	err = db.NewDatabase("postgres", "postgres", "localhost", "localiday")
	if err != nil {
		log.Fatal("Could not connect to database.", err)
	}
	err = domain.BootStrap()
	if err != nil {
		log.Panic("Could not bootstrap database.", err)
	}
	appServer := server.AppServer{Port: uint16(port)}
	appServer.Start()

	log.Printf("Application started in %v.", time.Since(start))
}
