package server

import (
	"fmt"
	"log"
	"time"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/assets"
)

// AppServer the application server.
type AppServer struct {
	Port uint16
}

// Start initializes and starts the server.
func (a AppServer) Start() {
	startTime := time.Now()

	cssController := assets.CreateCSSController()
	jsController := assets.CreateJSController()

	web.Get("/css/localiday_(.*).css", cssController.RenderCSS)
	web.Get("/js/localiday_(.*).js", jsController.RenderJS)
	log.Printf("Started server on port %v in %v.\n", a.Port, time.Since(startTime))

	web.Run(fmt.Sprintf("0.0.0.0:%v", a.Port))
}
