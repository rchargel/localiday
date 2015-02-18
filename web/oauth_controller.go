package web

import (
	"fmt"
	"os"
	"strings"

	"github.com/hoisie/web"
	"github.com/rchargel/goauth"
	"github.com/rchargel/localiday/app"
	"github.com/rchargel/localiday/services"
)

// CreateOAuthController creates the OAuth controller.
func CreateOAuthController() *OAuthController {
	file, err := os.Open("app/oauth_config.yaml")
	defer file.Close()
	if err != nil {
		app.Log(app.Fatal, "Could not read oauth_config.yaml file", err)
	}
	appConfig := app.LoadConfiguration()
	serviceProviders, err := goauth.ConfigureProvidersFromYAML(file, appConfig.HostURL+"/oauth/callback/%v")
	if err != nil {
		app.Log(app.Fatal, "Could not initialize OAuth Controller", err)
	}
	return &OAuthController{serviceProviders}
}

// OAuthController the controller for OAuth2 authentication calls.
type OAuthController struct {
	serviceProviders map[string]goauth.OAuthServiceProvider
}

// RedirectToAuthScreen redirects the user to the correct auth screen for their request.
func (c *OAuthController) RedirectToAuthScreen(ctx *web.Context, providerName string) {
	provider, found := c.serviceProviders[providerName]
	if found {
		redirectURL, err := provider.GetRedirectURL()
		if err == nil {
			ctx.Redirect(HTTPFoundRedirectCode, redirectURL)
		} else {
			ctx.Abort(HTTPServerErrorCode, err.Error())
		}
	} else {
		ctx.Abort(HTTPBadRequestCode, fmt.Sprintf("%v is not a valid provider.", providerName))
	}
}

// ProcessOAuthReply called by the redirect code.
func (c *OAuthController) ProcessOAuthReply(ctx *web.Context, providerName string) {
	provider, found := c.serviceProviders[strings.ToLower(providerName)]
	if found {
		userData, err := provider.ProcessResponse(ctx.Request)
		if hasNoError(ctx, err) {
			app.Log(app.Debug, "Found user: %v", userData.String())
			session, err := services.NewUserService().CreateSessionForOAuthUser(userData)
			if hasNoError(ctx, err) {
				ctx.Redirect(HTTPFoundRedirectCode, fmt.Sprintf("/?token=%v", session.SessionID))
			}
		}
	} else {
		ctx.Abort(HTTPBadRequestCode, fmt.Sprintf("%v is not a valid provider.", providerName))
	}
}

func hasNoError(ctx *web.Context, err error) bool {
	if err != nil {
		app.Log(app.Error, err.Error())
		ctx.Abort(HTTPServerErrorCode, err.Error())
		return false
	}
	return true
}
