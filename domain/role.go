package domain

import (
	"log"

	"github.com/rchargel/localiday/db"
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
	log.Printf("Added role %v to user %v", authority.Authority, user.Username)
}

// CreateAuthority creates a new role in the database.
func CreateAuthority(authority string) *Role {
	role := &Role{Authority: authority}

	db.DB.Insert(role)
	log.Printf("Added role %v", role.Authority)
	return role
}

// GetAuthorities get the list of user authorities.
func (u *User) GetAuthorities() []Role {
	var roles []Role
	db.DB.Select(&roles, `select r.* from users u inner join user_roles ur on u.id = ur.user_id
  inner join roles r on ur.role_id = r.id where u.id = ?`, u.ID)
	return roles
}
