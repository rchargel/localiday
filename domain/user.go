package domain

import (
	"fmt"

	"github.com/coopernurse/gorp"
	"github.com/rchargel/localiday/db"
	"golang.org/x/crypto/bcrypt"
)

// User the system user.
type User struct {
	ID              int64
	Username        string
	Password        string
	FullName        string `db:"full_name"`
	NickName        string
	Email           string
	PasswordExpired bool `db:"password_expired"`
	Active          bool
}

// CreateNewUser creates a new user with default configuration.
func CreateNewUser(username, password, fullname, nickname, email string) *User {
	user := &User{
		Username:        username,
		Password:        password,
		FullName:        fullname,
		NickName:        nickname,
		Email:           email,
		PasswordExpired: false,
		Active:          true,
	}

	insert(user)
	return user
}

// PreInsert called before the user is inserted into the database.
func (u *User) PreInsert(s gorp.SqlExecutor) error {
	return u.encryptPassword(u.Password)
}

// PreDelete called before the user is deleted.
func (u *User) PreDelete(s gorp.SqlExecutor) error {
	query := fmt.Sprintf("delete from user_roles where user_id = %v", u.ID)
	_, err := s.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

// SetPassword sets an encrypted version of the password.
func (u *User) SetPassword(password string) error {
	return u.encryptPassword(password)
}

func (u *User) encryptPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u.Password = string(hashed)
	return err
}

// GetUserBySession gets a user by the session ID, also updates the sessions last accessed value.
func GetUserBySession(sessionID string) (User, error) {
	var u User
	s, err := GetSessionBySessionID(sessionID)
	if err != nil {
		err = db.DB.SelectOne(&u, "select * from user where id = ?", s.UserID)
	}
	return u, err
}

// CountActive counts the number of active users in the system.
func (u User) CountActive() uint32 {
	return count("select count(*) from users where active = 't'")
}

// Count counts the number of users in the system.
func (u User) Count() uint32 {
	return count("select count(*) from users")
}
