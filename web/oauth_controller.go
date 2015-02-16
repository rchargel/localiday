package web

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/hoisie/web"
	"github.com/rchargel/goauth"
	"github.com/rchargel/localiday/app"
	"github.com/rchargel/localiday/services"
	"gopkg.in/yaml.v2"
)

// CreateOAuthController creates the OAuth controller.
func CreateOAuthController() *OAuthController {
	m := make(map[string]map[string]interface{})
	data, err := ioutil.ReadFile("app/oauth_config.yaml")
	if err != nil {
		app.Log(app.Fatal, "Could not initialize OAuth Controller.", err)
	}
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		app.Log(app.Fatal, "Could not initialize OAuth Controller.", err)
	}
	appConfig := app.LoadConfiguration()
	serviceProviders := make(map[string]goauth.OAuthServiceProvider, 3)

	for provider, conf := range m {
		provName := strings.ToLower(provider)
		conf["ProviderName"] = provName
		conf["RedirectURL"] = fmt.Sprintf("%v/oauth/callback/%v", appConfig.HostURL, provName)
		conf["ClientID"] = os.Getenv(strings.ToUpper(provider) + "_CLIENT_ID")
		conf["ClientSecret"] = os.Getenv(strings.ToUpper(provider) + "_CLIENT_SECRET")
		oauthVersion, found := conf["OAuthVersion"]
		if !found {
			app.Log(app.Fatal, "No oauth version for provider %v.", provider)
		}
		oauthVersionString := strconv.FormatFloat(oauthVersion.(float64), 'f', 1, 32)
		if oauthVersionString == goauth.OAuthVersion1 {
			// build version 1.0
			oauthConfiguration := goauth.OAuth1ServiceProviderConfig{}
			configureOAuthServiceProvider(&oauthConfiguration, conf)
			serviceProviders[provName] = goauth.NewOAuth1ServiceProvider(oauthConfiguration)
		} else if oauthVersionString == goauth.OAuthVersion2 {
			// build version 2.0
			oauthConfiguration := goauth.OAuth2ServiceProviderConfig{}
			configureOAuthServiceProvider(&oauthConfiguration, conf)
			serviceProviders[provName] = goauth.NewOAuth2ServiceProvider(oauthConfiguration)
		} else {
			app.Log(app.Fatal, "Invalid oauth version: %v.", oauthVersionString)
		}
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

func configureOAuthServiceProvider(configPtr interface{}, conf map[string]interface{}) {
	v := reflect.ValueOf(configPtr)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		app.Log(app.Fatal, "Type \"%v\" is not a struct.", v.Kind())
	}
	t := reflect.TypeOf(v.Interface())

	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		fieldVal, found := conf[fieldName]
		if found {
			field := v.FieldByName(fieldName)
			switch field.Kind() {
			case reflect.String:
				field.SetString(fieldVal.(string))
			case reflect.Int:
				val, _ := strconv.Atoi(fieldVal.(string))
				field.SetInt(int64(val))
			case reflect.Slice:
				vals := fieldVal.([]interface{})
				strs := make([]string, len(vals))
				for i, strVal := range vals {
					strs[i] = strVal.(string)
				}
				field.Set(reflect.ValueOf(strs))
			}
		}
	}
}
