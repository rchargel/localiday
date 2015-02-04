package services

import (
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/app"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

const (
	oauthAuthorization   = "Authorization"
	oauthPreamble        = "OAuth"
	oauthNonce           = "oauth_nonce"
	oauthSignature       = "oauth_signature"
	oauthSignatureMethod = "oauth_signature_method"
	oauthCallback        = "oauth_callback"
	oauthConsumerKey     = "oauth_consumer_key"
	oauthTimestamp       = "oauth_timestamp"
	oauthVersion         = "oauth_version"
	oauthToken           = "oauth_token"
	oauthSecretToken     = "oauth_token_secret"
	oauthVerifier        = "oauth_verifier"
)

var tokenSecrets = make(map[string]string)

// OAuthService service for performing oauth tasks.
type OAuthService struct {
	Name             string
	StateFlag        string
	UserInfoURL      string
	RequestTokenURL  string
	RequestTokenVerb string
	Provider         string
	Conf             oauth2.Config
}

type authConfig struct {
	provider         string
	userInfoURL      string
	requestTokenURL  string
	requestTokenVerb string
	conf             oauth2.Config
}

type oauthUser struct {
	id         string
	name       string
	screenName string
	email      string
}

type oauthPair struct {
	key   string
	value string
}

// NewOAuthService creates an instance of the oauth service.
func NewOAuthService(name string) (*OAuthService, error) {
	ac := authConfig{
		provider: strings.ToUpper(name),
	}
	err := loadOauthConfig(&ac)
	rb := make([]byte, 20)
	crand.Read(rb)
	stateFlag := "lcldy" + base64.URLEncoding.EncodeToString(rb)

	service := &OAuthService{
		Name:             strings.ToLower(name),
		StateFlag:        stateFlag,
		UserInfoURL:      ac.userInfoURL,
		RequestTokenURL:  ac.requestTokenURL,
		RequestTokenVerb: ac.requestTokenVerb,
		Provider:         ac.provider,
		Conf:             ac.conf,
	}
	return service, err
}

// GenerateRedirectURL generates the URL that should be sent to the browser
// to send the user to the providers correct login screen.
func (o *OAuthService) GenerateRedirectURL() string {
	url := o.Conf.AuthCodeURL(o.StateFlag)
	if len(o.RequestTokenURL) > 0 {
		token, err := o.fetchRequestToken()
		if err != nil {
			app.Log(app.Error, "Could not find oauth token.", err)
		}
		app.Log(app.Debug, "Request Token: %v", token)
		url = fmt.Sprintf("%v&%v=%v", url, oauthToken, token)
	}

	return url
}

// ProcessResponse processes the OAuth 2.0 response, and returns a new SessionID.
func (o *OAuthService) ProcessResponse(ctx *web.Context) (string, error) {
	var err error
	var sessionID string
	request := ctx.Request
	if code := request.FormValue("code"); len(code) > 0 {
		app.Log(app.Debug, "Found auth code: %v.", code)
		tok, err := o.Conf.Exchange(oauth2.NoContext, code)
		if err == nil {
			client := o.Conf.Client(oauth2.NoContext, tok)
			resp, _ := client.Get(o.UserInfoURL)
			m := make(map[string]interface{})
			dec := json.NewDecoder(resp.Body)
			dec.Decode(&m)
			user := convertMapToUser(m)

			session, err := NewUserService().CreateSessionForOAuthUser(user.id, user.name, user.screenName, user.email, code, o.Provider)
			if err == nil {
				sessionID = session.SessionID
			}
		}
	} else {
		err = errors.New("No token has been submitted to this endpoint.")
	}
	return sessionID, err
}

// ProcessOAuthTokenResponse process the OAuth 1.0 response and returns a new SessionID.
func (o *OAuthService) ProcessOAuthTokenResponse(ctx *web.Context, token, verifier string) (string, error) {
	var err error
	var sessionID string
	secretKey := tokenSecrets[token]
	delete(tokenSecrets, token)
	app.Log(app.Debug, "Token: %v / Verifier: %v / Secret: %v", token, verifier, secretKey)

	accessToken, accessTokenSecret, err := o.fetchAccessToken(token, secretKey, verifier)

	user, err := o.fetchOAuthUserInfo(accessToken, accessTokenSecret, verifier)
	if err == nil {
		session, err := NewUserService().CreateSessionForOAuthUser(user.id, user.name, user.screenName, user.email, accessToken, o.Provider)
		if err == nil {
			sessionID = session.SessionID
		}
	}

	return sessionID, err
}

func (o *OAuthService) fetchOAuthUserInfo(accessToken, accessTokenSecret, verifier string) (oauthUser, error) {
	verb := "GET"
	params := o.generateParams(accessToken, accessTokenSecret, verifier)

	baseStringParamOrder := []string{oauthConsumerKey, oauthNonce, oauthSignatureMethod, oauthTimestamp, oauthToken, oauthVersion}
	baseString := o.createBaseString(verb, o.UserInfoURL, toParamList(params, baseStringParamOrder))
	app.Log(app.Debug, "User-Info Base String: %v", baseString)

	methodSignature := o.createMethodSignature(baseString, accessToken, accessTokenSecret)
	params[oauthSignature] = methodSignature
	headerParamOrder := []string{oauthConsumerKey, oauthNonce, oauthSignature, oauthSignatureMethod, oauthTimestamp, oauthToken, oauthVersion}
	header := o.createHeader(toParamList(params, headerParamOrder))
	app.Log(app.Debug, "User-Info Authorization %v", header)

	client := &http.Client{}
	req, _ := http.NewRequest(verb, o.UserInfoURL, nil)
	req.Header.Add(oauthAuthorization, header)
	resp, err := client.Do(req)

	m := make(map[string]interface{})
	var user oauthUser
	if err == nil {
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&m); err == nil {
			user = convertMapToUser(m)
		}
	}
	return user, err
}

func (o *OAuthService) fetchAccessToken(token, secret, verifier string) (string, string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest(o.RequestTokenVerb, o.Conf.Endpoint.TokenURL, nil)

	params := o.generateParams(token, secret, verifier)
	baseStringParamOrder := []string{oauthConsumerKey, oauthNonce, oauthSignatureMethod, oauthTimestamp, oauthToken, oauthVerifier, oauthVersion}
	baseString := o.createBaseString(o.RequestTokenVerb, o.Conf.Endpoint.TokenURL, toParamList(params, baseStringParamOrder))
	app.Log(app.Debug, "Access Token Base String: %v", baseString)

	methodSignature := o.createMethodSignature(baseString, token, secret)
	params[oauthSignature] = methodSignature

	headerParamOrder := []string{oauthVerifier, oauthNonce, oauthSignature, oauthToken, oauthConsumerKey, oauthTimestamp, oauthSignatureMethod, oauthVersion}
	header := o.createHeader(toParamList(params, headerParamOrder))
	app.Log(app.Debug, "Access Token Authorization: %v", header)

	req.Header.Add(oauthAuthorization, header)
	var accessToken string
	var accessTokenSecret string
	resp, err := client.Do(req)
	if err == nil {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			if values, err := url.ParseQuery(string(body)); err == nil {
				accessToken = values.Get(oauthToken)
				accessTokenSecret = values.Get(oauthSecretToken)
				app.Log(app.Debug, "Access Token: %v / Secret: %v", accessToken, accessTokenSecret)
			}
		}
	}
	return accessToken, accessTokenSecret, err
}

func (o *OAuthService) fetchRequestToken() (string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest(o.RequestTokenVerb, o.RequestTokenURL, nil)

	params := o.generateParams("", "", "")

	baseStringParamOrder := []string{oauthCallback, oauthConsumerKey, oauthNonce, oauthSignatureMethod, oauthTimestamp, oauthVersion}
	baseStringParams := toParamList(params, baseStringParamOrder)
	baseString := o.createBaseString(o.RequestTokenVerb, o.RequestTokenURL, baseStringParams)
	app.Log(app.Debug, "Request Token Base String: %v", baseString)

	methodSignature := o.createMethodSignature(baseString, o.Conf.ClientSecret, "")
	params[oauthSignature] = methodSignature
	headerParamOrder := []string{oauthNonce, oauthSignature, oauthCallback, oauthConsumerKey, oauthTimestamp, oauthSignatureMethod, oauthVersion}
	headerParams := toParamList(params, headerParamOrder)
	header := o.createHeader(headerParams)
	app.Log(app.Debug, "Request Token Authorization %v", header)

	req.Header.Add(oauthAuthorization, header)

	var token string
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err == nil {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			if values, err := url.ParseQuery(string(body)); err == nil {
				token = values.Get(oauthToken)
				secret := values.Get(oauthSecretToken)
				tokenSecrets[token] = secret
				app.Log(app.Debug, "Token: %v / Secret: %v", token, secret)
			}
		}
	}

	return token, err
}

func (o *OAuthService) generateParams(token, secret, verifier string) map[string]string {
	params := make(map[string]string)

	params[oauthCallback] = o.Conf.RedirectURL
	params[oauthConsumerKey] = o.Conf.ClientID
	params[oauthNonce] = o.generateRandomToken()
	params[oauthSignatureMethod] = "HMAC-SHA1"
	params[oauthTimestamp] = fmt.Sprint(time.Now().Unix())
	params[oauthVersion] = "1.0"
	params[oauthToken] = token
	params[oauthSecretToken] = secret
	params[oauthVerifier] = verifier

	return params
}

func (o *OAuthService) generateRandomToken() string {
	return fmt.Sprint(time.Now().Unix()) + fmt.Sprint(rand.Intn(10))
}

func (o *OAuthService) createHeader(params []oauthPair) string {
	var header string
	for _, param := range params {
		if len(header) == 0 {
			header = fmt.Sprintf("%v=\"%v\"", param.key, url.QueryEscape(param.value))
		} else {
			header = fmt.Sprintf("%v, %v=\"%v\"", header, param.key, url.QueryEscape(param.value))
		}
	}
	return oauthPreamble + " " + header
}

func (o *OAuthService) createMethodSignature(baseString, clientSecret, oauthSecret string) string {
	secretKey := url.QueryEscape(o.Conf.ClientSecret) + "&"
	if len(oauthSecret) > 0 {
		secretKey = secretKey + url.QueryEscape(oauthSecret)
	}
	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write([]byte(baseString))
	encoded := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encoded)
}

func (o *OAuthService) createBaseString(verb, tourl string, params []oauthPair) string {
	paramString := ""
	for _, param := range params {
		if len(paramString) == 0 {
			paramString = param.key + "=" + url.QueryEscape(param.value)
		} else {
			paramString = paramString + "&" + param.key + "=" + url.QueryEscape(param.value)
		}
	}
	return fmt.Sprintf("%v&%v&%v", strings.ToUpper(verb),
		url.QueryEscape(tourl), url.QueryEscape(paramString))
}

func toParamList(params map[string]string, order []string) []oauthPair {
	paramList := make([]oauthPair, 0, len(order))
	for _, key := range order {
		if value, found := params[key]; found {
			paramList = append(paramList, oauthPair{key: key, value: value})
		}
	}
	return paramList
}

func convertMapToUser(data map[string]interface{}) oauthUser {
	user := oauthUser{id: app.ToStringValue(data["id"])}
	if name, found := data["name"]; found {
		user.name = name.(string)
	}
	if screenName, found := data["screen_name"]; found {
		user.screenName = screenName.(string)
	}
	if givenName, found := data["given_name"]; found {
		if familyName, found := data["family_name"]; found {
			user.screenName = fmt.Sprintf("%v %v", givenName, familyName)
		} else {
			user.screenName = givenName.(string)
		}
	}
	if firstName, found := data["first_name"]; found {
		if lastName, found := data["last_name"]; found {
			user.screenName = fmt.Sprintf("%v %v", firstName, lastName)
		} else {
			user.screenName = firstName.(string)
		}
	}
	if email, found := data["email"]; found {
		user.email = email.(string)
	}
	if len(user.name) == 0 {
		user.name = user.screenName
	}
	return user
}

func loadOauthConfig(ac *authConfig) error {
	var conf oauth2.Config
	var m map[string]map[string]interface{}
	var userInfoURL string
	var requestTokenURL string
	var requestTokenVerb string

	sysConfig := app.LoadConfiguration()
	data, err := ioutil.ReadFile("app/oauth_config.yaml")
	if err == nil {
		err = yaml.Unmarshal(data, &m)
	}
	if t, found := m[ac.provider]; found {
		redirectURL := sysConfig.HostURL + "oauth/callback/" + strings.ToLower(ac.provider)
		userInfoURL = t["UserInfoURL"].(string)
		if reqTokURL, found := t["RequestTokenURL"]; found {
			requestTokenURL = reqTokURL.(string)
		}
		if reqTokVerb, found := t["RequestTokenVerb"]; found {
			requestTokenVerb = reqTokVerb.(string)
		}
		clientID := os.Getenv(ac.provider + "_CLIENT_ID")
		clientSecret := os.Getenv(ac.provider + "_CLIENT_SECRET")
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
	ac.conf = conf
	ac.userInfoURL = userInfoURL
	ac.requestTokenURL = requestTokenURL
	ac.requestTokenVerb = requestTokenVerb
	return err
}
