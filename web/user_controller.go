package web

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fatih/structs"
	"github.com/hoisie/web"
	"github.com/rchargel/localiday/app"
	"github.com/rchargel/localiday/db"
)

// UserController controller for user rest calls
type UserController struct{}

// ProcessRequest processes a user request.
func (u UserController) ProcessRequest(ctx *web.Context, request string) {
	w := NewResponseWriter(ctx)
	method := app.MakeFirstLetterUpperCase(request)

	args := make([]reflect.Value, 1)
	args[0] = reflect.ValueOf(w)
	ru := reflect.ValueOf(u)
	rm := ru.MethodByName(method)
	if rm.IsValid() {
		rm.Call(args)
	} else {
		w.SendError(HTTPInvalidMethodCode, fmt.Errorf("No method %v found", request))
	}
}

// Login processes a login request from the user.
func (u UserController) Login(w *ResponseWriter) {
	r := w.Request
	var cred struct {
		Username string
		Password string
	}
	d := json.NewDecoder(r.Body)
	err := d.Decode(&cred)
	if err == nil {
		u, err := db.User{}.FindByUsernameAndPassword(cred.Username, cred.Password)
		if err != nil {
			w.SendError(HTTPUnauthorizedCode, err)
		} else {
			s := db.CreateNewSession(u.ID)
			output := toUserMap(s, u)
			w.SendJSON(output)
		}
	} else {
		w.SendError(HTTPBadRequestCode, err)
	}
}

// Logout logs the user out of the session.
func (u UserController) Logout(w *ResponseWriter) {
	if sessionID, err := w.GetSessionIDAuthorization(); err == nil {
		db.DeleteSession(sessionID)
		w.SendSuccess()
	} else {
		w.SendError(HTTPUnauthorizedCode, err)
		app.Log(app.Error, "There was no authorization in the request.")
	}
}

// Validate validates the user session.
func (u UserController) Validate(w *ResponseWriter) {
	var err error
	if sessionID, err := w.GetSessionIDAuthorization(); err == nil {
		if sess, err := db.GetSessionBySessionID(sessionID); err == nil {
			user, err := db.User{}.Get(sess.UserID)
			if err == nil {
				output := toUserMap(sess, user)
				w.SendJSON(output)
			}
		}
	}
	if err != nil {
		w.SendError(HTTPUnauthorizedCode, err)
	}
}

func toUserMap(s *db.Session, u *db.User) map[string]interface{} {
	m := structs.Map(u)
	m["SessionID"] = s.SessionID
	m["TokenType"] = "Bearer"
	m["Authorities"] = u.GetAuthoritiesStrings()
	m["LastAccessed"] = s.LastAccessed.Unix()
	delete(m, "Password")

	return m
}
