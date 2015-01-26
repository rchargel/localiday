package domain

import (
	"fmt"

	"github.com/rchargel/localiday/db"
	"github.com/rchargel/localiday/util"
)

// UserRole mapping between User and Role.
type UserRole struct {
	ID     int64
	UserID int64 `db:"user_id"`
	RoleID int64 `db:"role_id"`
}

// Role the user's authorities.
type Role struct {
	ID        int64
	Authority string
}

// AddAuthorityToUser adds an existing authority/role to an existing user.
func AddAuthorityToUser(user *User, authority *Role) {
	userRole := &UserRole{
		UserID: user.ID,
		RoleID: authority.ID,
	}

	db.DB.Insert(userRole)
	util.Log(util.Debug, "Added role %v to user %v", authority.Authority, user.Username)
}

// CreateAuthority creates a new role in the database.
func CreateAuthority(authority string) *Role {
	role := &Role{Authority: authority}

	db.DB.Insert(role)
	util.Log(util.Debug, "Added role %v", role.Authority)
	return role
}

// GetAuthorities get the list of user authorities.
func (u *User) GetAuthorities() []Role {
	var roles []Role
	db.DB.Select(&roles, fmt.Sprintf(`select r.* from users u inner join user_roles ur on u.id = ur.user_id
  inner join roles r on ur.role_id = r.id where u.id = %v`, u.ID))
	return roles
}

// GetAuthoritiesStrings gets the list of authorities as a string.
func (u *User) GetAuthoritiesStrings() []string {
	r := u.GetAuthorities()
	s := make([]string, len(r))
	for i, a := range r {
		s[i] = a.Authority
	}
	return s
}
