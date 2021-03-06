package db

import (
	"errors"
	"fmt"

	"github.com/coopernurse/gorp"
	"github.com/rchargel/localiday/app"
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
func CreateNewUser(username, password, fullname, nickname, email string) (*User, error) {
	user := &User{
		Username:        username,
		Password:        password,
		FullName:        fullname,
		NickName:        nickname,
		Email:           email,
		PasswordExpired: false,
		Active:          true,
	}

	err := insert(user)
	return user, err
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

// Get gets the user by the ID.
func (u User) Get(userID int64) (*User, error) {
	err := DB.SelectOne(&u, fmt.Sprintf("select * from users where id = %v", userID))
	return &u, err
}

// GetUserBySession gets a user by the session ID, also updates the sessions last accessed value.
func GetUserBySession(sessionID string) (*User, error) {
	var u User
	s, err := GetSessionBySessionID(sessionID)
	if err != nil {
		err = DB.SelectOne(&u, fmt.Sprintf("select * from users where id = %v", s.UserID))
	}
	return &u, err
}

// FindByUsername used to find a user by their username.
func (u User) FindByUsername(username string) (*User, error) {
	var found User
	err := DB.SelectOne(&found, fmt.Sprintf("select * from users where username = '%v'", username))
	if err != nil || len(found.Username) == 0 {
		app.Log(app.Debug, "Could not find user: "+username, err)
		return nil, fmt.Errorf("Could not find a user with the supplied username: %v.", username)
	}
	return &found, nil
}

// FindByUsernameAndPassword used to find a user in order to perform a login.
func (u User) FindByUsernameAndPassword(username, password string) (*User, error) {
	var found User
	err := DB.SelectOne(&found, fmt.Sprintf("select * from users where username = '%v'", username))
	if err != nil {
		app.Log(app.Debug, "Could not find user: "+username, err)
		return nil, fmt.Errorf("Could not find a user with the supplied username: %v.", username)
	}
	err = bcrypt.CompareHashAndPassword([]byte(found.Password), []byte(password))
	if err != nil {
		app.Log(app.Debug, "Passwords did not match", err)
		return nil, errors.New("Username and password do not match.")
	}
	return &found, nil
}

// CountActive counts the number of active users in the system.
func (u User) CountActive() uint32 {
	return count("select count(*) from users where active = 't'")
}

// Count counts the number of users in the system.
func (u User) Count() uint32 {
	return count("select count(*) from users")
}
