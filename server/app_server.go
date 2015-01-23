package server

import (
	"fmt"
	"log"
	"time"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/controllers"
)

// AppServer the application server.
type AppServer struct {
	Port uint16
}

// Start initializes and starts the server.
func (a AppServer) Start() {
	startTime := time.Now()

	cssController := controllers.CreateCSSController()
	jsController := controllers.CreateJSController()
	htmlController := controllers.CreateHTMLController()

	web.Get("/css/localiday_(.*).css", cssController.RenderCSS)
	web.Get("/js/localiday_(.*).js", jsController.RenderJS)
	web.Get("/js/(.*)", jsController.RenderJSFile)
	web.Get("/", htmlController.RenderRoot)
	log.Printf("Started server on port %v in %v.\n", a.Port, time.Since(startTime))

	web.Run(fmt.Sprintf("0.0.0.0:%v", a.Port))
}
