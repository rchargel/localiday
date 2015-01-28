package controllers

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fatih/structs"
	"github.com/hoisie/web"
	"github.com/rchargel/localiday/domain"
	"github.com/rchargel/localiday/util"
)

// UserController controller for user rest calls
type UserController struct{}

// ProcessRequest processes a user request.
func (u UserController) ProcessRequest(ctx *web.Context, request string) {
	w := util.NewResponseWriter(ctx)
	method := util.MakeFirstLetterUpperCase(request)

	args := make([]reflect.Value, 1)
	args[0] = reflect.ValueOf(w)
	ru := reflect.ValueOf(u)
	rm := ru.MethodByName(method)
	if rm.IsValid() {
		rm.Call(args)
	} else {
		w.SendError(util.HTTPInvalidMethodCode, fmt.Errorf("No method %v found", request))
	}
}

// Login processes a login request from the user.
func (u UserController) Login(w *util.ResponseWriter) {
	r := w.Request
	var cred struct {
		Username string
		Password string
	}
	d := json.NewDecoder(r.Body)
	err := d.Decode(&cred)
	if err == nil {
		u, err := domain.User{}.FindByUsernameAndPassword(cred.Username, cred.Password)
		if err != nil {
			w.SendError(util.HTTPUnauthorizedCode, err)
		} else {
			s := domain.CreateNewSession(u.ID)
			output := toMap(s, u)
			w.SendJSON(output)
		}
	} else {
		w.SendError(util.HTTPBadRequestCode, err)
	}
}

// Logout logs the user out of the session.
func (u UserController) Logout(w *util.ResponseWriter) {
	if sessionID, err := w.GetSessionIDAuthorization(); err == nil {
		domain.DeleteSession(sessionID)
		w.SendSuccess()
	} else {
		w.SendError(util.HTTPUnauthorizedCode, err)
		util.Log(util.Error, "There was no authorization in the request.")
	}
}

func toMap(s *domain.Session, u *domain.User) map[string]interface{} {
	m := structs.Map(u)
	m["SessionID"] = s.SessionID
	m["TokenType"] = "Bearer"
	m["Authorites"] = u.GetAuthoritiesStrings()
	delete(m, "Password")

	return m
}
