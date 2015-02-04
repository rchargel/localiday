package db

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rchargel/localiday/app"
)

const (
	sessionStringSize     = 20
	sessionTimeoutSeconds = 300
)

var sessionLock sync.Mutex

// Session an active user session.
type Session struct {
	ID             int64
	UserID         int64     `db:"user_id"`
	SessionID      string    `db:"session_id"`
	OAuthToken     string    `db:"oauth_token"`
	OAuthProvider  string    `db:"oauth_provider"`
	LastAccessed   time.Time `db:"last_accessed"`
	SessionCreated time.Time `db:"session_created"`
}

// CreateNewOAuthSession creates a new session from an oauth source and inserts it into the database.
func CreateNewOAuthSession(userID int64, oauthToken, oauthProvider string) *Session {
	if session, err := getSessionByUserID(userID); err == nil {
		app.Log(app.Debug, "Found existing session for user %v.", session.SessionID)
		return session
	}

	sessionID := createSessionString()
	app.Log(app.Debug, "Creating new session %v.", sessionID)

	session := &Session{
		UserID:         userID,
		SessionID:      sessionID,
		OAuthToken:     oauthToken,
		OAuthProvider:  strings.ToUpper(oauthProvider),
		SessionCreated: time.Now(),
	}

	if err := insert(session); err != nil {
		app.Log(app.Error, "Error creating session: ", err)
	}
	updateLastAccessedSessionTime(session.ID)

	return session
}

// CreateNewSession creates a new session and inserts it into the database.
func CreateNewSession(userID int64) *Session {
	if session, err := getSessionByUserID(userID); err == nil {
		app.Log(app.Debug, "Found existing session for user %v.", session.SessionID)
		return session
	}

	sessionID := createSessionString()
	app.Log(app.Debug, "Creating new session %v.", sessionID)

	session := &Session{
		UserID:         userID,
		SessionID:      sessionID,
		SessionCreated: time.Now(),
	}

	insert(session)

	updateLastAccessedSessionTime(session.ID)

	return session
}

// GetSessionBySessionID gets an active session, if it exists, and updates its last accessed value.
func GetSessionBySessionID(sessionID string) (*Session, error) {
	sessionLock.Lock()
	var s Session
	err := DB.SelectOne(&s,
		fmt.Sprintf("select * from sessions where session_id = '%v' and last_accessed > now() - interval '%v seconds'",
			sessionID, sessionTimeoutSeconds))

	if err == nil {
		updateLastAccessedSessionTime(s.ID)
	}
	sessionLock.Unlock()
	return &s, err
}

// ValidateSession validates the session ID and updates the last accessed value.
func ValidateSession(sessionID string) (bool, error) {
	if _, err := GetSessionBySessionID(sessionID); err != nil {
		return false, err
	}
	return true, nil
}

// CleanSessions cleans up the old session strings.
func CleanSessions() error {
	sessionLock.Lock()
	s := time.Now()
	result, err := DB.Exec(fmt.Sprintf("delete from sessions where last_accessed < now() - interval '%v seconds'", sessionTimeoutSeconds))
	if err == nil {
		count, _ := result.RowsAffected()
		if count == 0 {
			app.Log(app.Debug, "No expired sessions found (%v).", time.Since(s))
		} else {
			app.Log(app.Info, "Purged %v expired sessions in %v.", count, time.Since(s))
		}
	} else {
		app.Log(app.Error, "Failed purge: ", err)
	}

	sessionLock.Unlock()
	return err
}

// DeleteSession used when the user logs out of their session.
func DeleteSession(sessionID string) {
	sessionLock.Lock()
	DB.Exec("delete from sessions where session_id = ?", sessionID)
	sessionLock.Unlock()
}

// IsAuthorized determins if the user has any of the supplied roles.
func IsAuthorized(sessionID string, roles ...string) bool {
	if user, err := GetUserBySession(sessionID); err == nil {
		authorities := user.GetAuthoritiesStrings()
		for _, authority := range roles {
			if app.Contains(authorities, authority) {
				return true
			}
		}
	} else {
		app.Log(app.Error, "Could not find user for session %v.", sessionID)
	}
	return false
}

func getSessionByUserID(userID int64) (*Session, error) {
	sessionLock.Lock()
	s := &Session{}
	err := DB.SelectOne(s, fmt.Sprintf("select * from sessions where user_id = %v", userID))
	if err == nil {
		updateLastAccessedSessionTime(s.ID)
	} else {
		app.Log(app.Debug, "Error finding session with user id %v: %v", userID, err)
	}
	sessionLock.Unlock()
	return s, err
}

func createSessionString() string {
	rb := make([]byte, sessionStringSize)
	rand.Read(rb)
	return hex.EncodeToString(rb)
}

func updateLastAccessedSessionTime(sessionID int64) error {
	_, err := DB.Exec(fmt.Sprintf("update sessions set last_accessed = now() where id = %v", sessionID))
	if err != nil {
		app.Log(app.Error, "Error updating session access time.", err)
	}
	return err
}
