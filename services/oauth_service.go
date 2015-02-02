package services

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/app"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

// OAuthService service for performing oauth tasks.
type OAuthService struct {
	Name        string
	StateFlag   string
	UserInfoURL string
	Conf        oauth2.Config
}

// NewOAuthService creates an instance of the oauth service.
func NewOAuthService(name string) (*OAuthService, error) {
	conf, userInfoURL, err := loadOauthConfig(strings.ToUpper(name))
	rb := make([]byte, 20)
	rand.Read(rb)
	stateFlag := "lcldy" + base64.URLEncoding.EncodeToString(rb)

	service := &OAuthService{
		Name:        strings.ToLower(name),
		StateFlag:   stateFlag,
		UserInfoURL: userInfoURL,
		Conf:        conf,
	}
	return service, err
}

// GenerateRedirectURL generates the URL that should be sent to the browser
// to send the user to the providers correct login screen.
func (o *OAuthService) GenerateRedirectURL() string {
	return o.Conf.AuthCodeURL(o.StateFlag)
}

// ProcessResponse processes the OAuth response.
func (o *OAuthService) ProcessResponse(ctx *web.Context) error {
	var err error
	request := ctx.Request
	if code := request.FormValue("code"); len(code) > 0 {
		app.Log(app.Debug, "Found auth code: %v.", code)
		app.Log(app.Debug, "Using client secret: %v.", o.Conf.ClientSecret)
		tok, err := o.Conf.Exchange(oauth2.NoContext, code)
		if err == nil {
			client := o.Conf.Client(oauth2.NoContext, tok)
			resp, _ := client.Get(o.UserInfoURL)
			m := make(map[string]interface{})
			dec := json.NewDecoder(resp.Body)
			dec.Decode(&m)

			fmt.Println(m)
		}
	} else {
		err = errors.New("No token has been submitted to this endpoint.")
	}
	return err
}

func loadOauthConfig(name string) (oauth2.Config, string, error) {
	var conf oauth2.Config
	var m map[string]map[string]interface{}
	var userInfoURL string
	sysConfig := app.LoadConfiguration()
	data, err := ioutil.ReadFile("app/oauth_config.yaml")
	if err == nil {
		err = yaml.Unmarshal(data, &m)
	}
	if t, found := m[name]; found {
		redirectURL := sysConfig.HostURL + "oauth/callback/" + strings.ToLower(name)
		userInfoURL = t["UserInfoURL"].(string)
		clientID := os.Getenv(name + "_CLIENT_ID")
		clientSecret := os.Getenv(name + "_CLIENT_SECRET")
		endpoint := oauth2.Endpoint{
			AuthURL:  t["AuthURL"].(string),
			TokenURL: t["TokenURL"].(string),
		}

		if scopes, found := t["Scopes"]; found {
			scope := app.ToStringSlice(scopes.([]interface{}))
			conf = oauth2.Config{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				Scopes:       scope,
				Endpoint:     endpoint,
				RedirectURL:  redirectURL,
			}
		} else {
			conf = oauth2.Config{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				Endpoint:     endpoint,
				RedirectURL:  redirectURL,
			}
		}
	}
	return conf, userInfoURL, err
}
