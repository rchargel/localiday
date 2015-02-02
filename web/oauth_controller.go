package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/app"
	"github.com/rchargel/localiday/services"
)

var stateMap = make(map[string]*services.OAuthService)

// OAuthController the controller for OAuth2 authentication calls.
type OAuthController struct{}

// RedirectToAuthScreen redirects the user to the correct auth screen for their request.
func (c OAuthController) RedirectToAuthScreen(ctx *web.Context, provider string) {
	s, err := services.NewOAuthService(provider)
	if err != nil {
		NewResponseWriter(ctx).SendError(HTTPServerErrorCode, err)
	} else {
		redirectURL := s.GenerateRedirectURL()
		stateMap[s.StateFlag] = s
		app.Log(app.Debug, "Redirecting user to oauth endpoint %v.", redirectURL)
		http.Redirect(ctx.ResponseWriter, ctx.Request, redirectURL, 200)
	}
}

func (c OAuthController) ProcessAuthReply(ctx *web.Context, provider string) {
	if state := ctx.Request.FormValue("state"); len(state) > 0 {
		if s, found := stateMap[state]; found {
			s.ProcessResponse(ctx)
		} else if state == "localiday" {
			s, _ = services.NewOAuthService(provider)
			s.ProcessResponse(ctx)
		} else {
			NewResponseWriter(ctx).SendError(HTTPForbiddenCode, fmt.Errorf("State %v not expected value.", state))
		}
	} else {
		NewResponseWriter(ctx).SendError(HTTPBadRequestCode, errors.New("Not a valid request."))
	}
}
