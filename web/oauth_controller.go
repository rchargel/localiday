package web

import (
	"errors"
	"fmt"

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
		if len(s.RequestTokenURL) == 0 {
			stateMap[s.StateFlag] = s
		}
		app.Log(app.Debug, "Redirecting user to oauth endpoint %v.", redirectURL)
		ctx.Redirect(HTTPFoundRedirectCode, redirectURL)
	}
}

// ProcessAuthReply called by the redirect code.
func (c OAuthController) ProcessAuthReply(ctx *web.Context, provider string) {
	if state := ctx.Request.FormValue("state"); len(state) > 0 {
		if s, found := stateMap[state]; found {
			sessionID, err := s.ProcessResponse(ctx)
			delete(stateMap, state)
			if err == nil && len(sessionID) > 0 {
				ctx.Redirect(HTTPFoundRedirectCode, fmt.Sprintf("/?token=%v", sessionID))
			} else {
				NewResponseWriter(ctx).SendError(HTTPServerErrorCode, err)
			}
		} else {
			NewResponseWriter(ctx).SendError(HTTPForbiddenCode, fmt.Errorf("State %v not expected value.", state))
		}
	} else if token := ctx.Request.FormValue("oauth_token"); len(token) > 0 {
		verifier := ctx.Request.FormValue("oauth_verifier")
		s, _ := services.NewOAuthService(provider)
		sessionID, err := s.ProcessOAuthTokenResponse(ctx, token, verifier)
		if err == nil && len(sessionID) > 0 {
			ctx.Redirect(HTTPFoundRedirectCode, fmt.Sprintf("/?token=%v", sessionID))
		} else {
			NewResponseWriter(ctx).SendError(HTTPServerErrorCode, err)
		}
	} else {
		NewResponseWriter(ctx).SendError(HTTPBadRequestCode, errors.New("Not a valid request."))
	}
}
