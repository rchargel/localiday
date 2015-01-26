package server

import (
	"fmt"
	"time"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/controllers"
	"github.com/rchargel/localiday/util"
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
	imagesController := controllers.CreateImagesController()

	web.Get("/r/user/(.*)", controllers.UserController{}.ProcessRequest)
	web.Post("/r/user/(.*)", controllers.UserController{}.ProcessRequest)

	web.Get("/css/localiday_(.*).css", cssController.RenderCSS)
	web.Get("/js/localiday_(.*).js", jsController.RenderJS)
	web.Get("/js/(.*)", jsController.RenderJSFile)
	web.Get("/images/bg.jpg", imagesController.RenderBGImage)
	web.Get("/images/(.*)", imagesController.RenderImage)
	web.Get("/templates/(.*)", htmlController.Render)
	web.Get("/(.*)", htmlController.RenderRoot)
	util.Log(util.Info, "Started server on port %v in %v.", a.Port, time.Since(startTime))

	web.Run(fmt.Sprintf("0.0.0.0:%v", a.Port))
}
