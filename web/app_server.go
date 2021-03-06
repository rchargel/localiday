package web

import (
	"fmt"
	"time"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/app"
)

// AppServer the application server.
type AppServer struct {
	Port uint16
}

// Start initializes and starts the server.
func (a AppServer) Start() {
	startTime := time.Now()

	cssController := CreateCSSController()
	jsController := CreateJSController()
	htmlController := CreateHTMLController()
	imagesController := CreateImagesController()

	userController := UserController{}
	oauthController := CreateOAuthController()
	//var oauthController OAuthController

	web.Post("/r/user/(.*)", userController.ProcessRequest)

	web.Get("/css/localiday_(.*).css", cssController.RenderCSS)
	web.Get("/js/localiday_(.*).js", jsController.RenderJS)
	web.Get("/js/(.*)", jsController.RenderJSFile)
	web.Get("/images/bg.jpg", imagesController.RenderBGImage)
	web.Get("/images/(.*)", imagesController.RenderImage)
	web.Get("/templates/(.*)", htmlController.Render)
	web.Get("/oauth/authenticate/(.*)", oauthController.RedirectToAuthScreen)
	web.Get("/oauth/callback/(.*)", oauthController.ProcessOAuthReply)
	web.Get("/(.*)", htmlController.RenderRoot)
	app.Log(app.Info, "Started server on port %v in %v.", a.Port, time.Since(startTime))

	web.Run(fmt.Sprintf("0.0.0.0:%v", a.Port))
}
