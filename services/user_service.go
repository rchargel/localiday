package services

import (
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/rchargel/localiday/db"
)

// UserService defines a set of functions to simplify working with user data.
type UserService struct{}

// NewUserService creates a pointer to the user service.
func NewUserService() *UserService {
	return &UserService{}
}

// CreateSessionForOAuthUser creates a session for an oauth user.
func (u *UserService) CreateSessionForOAuthUser(username, name, screenName, email, provider string) (*db.Session, error) {
	user, err := db.User{}.FindByUsername(username)
	if err != nil {
		rb := make([]byte, 20)
		rand.Read(rb)
		user, err = db.CreateNewUser(username, base64.StdEncoding.EncodeToString(rb), name, screenName, email)
		if err == nil {
			userRole, err := db.Role{}.FindByAuthority(db.RoleUser)
			if err == nil {
				db.AddAuthorityToUser(user, userRole)
			}
			oauthRole, err := db.Role{}.FindByAuthority(db.RoleOAuthUser)
			if err == nil {
				db.AddAuthorityToUser(user, oauthRole)
			}
			providerRole, err := db.Role{}.FindByAuthority(strings.ToUpper(provider) + "_USER")
			if err == nil {
				db.AddAuthorityToUser(user, providerRole)
			}
		}
	}
	if err == nil {
		return db.CreateNewSession(user.ID), nil
	}
	return nil, err
}
