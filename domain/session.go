package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/rchargel/localiday/db"
	"github.com/rchargel/localiday/util"
)

const (
	sessionStringSize     = 32
	sessionTimeoutSeconds = 300
)

var sessionLock sync.Mutex

// Session an active user session.
type Session struct {
	ID        int64
	UserID    int64  `db:"user_id"`
	SessionID string `db:"session_id"`
}

// CreateNewSession creates a new session and inserts it into the database.
func CreateNewSession(userID int64) *Session {
	if session, err := getSessionByUserID(userID); err == nil {
		util.Log(util.Debug, "Found Existing session for user %v.", session)
		return session
	}

	sessionID := createSessionString()
	util.Log(util.Debug, "Creating new session %v.", sessionID)

	session := &Session{
		UserID:    userID,
		SessionID: sessionID,
	}

	insert(session)

	return session
}

// GetSessionBySessionID gets an active session, if it exists, and updates its last accessed value.
func GetSessionBySessionID(sessionID string) (Session, error) {
	sessionLock.Lock()
	var s Session
	err := db.DB.SelectOne(&s,
		fmt.Sprintf("select id, user_id, session_id from sessions where session_id = '%v' and last_accessed > now() - interval '%v seconds'",
			sessionID, sessionTimeoutSeconds))

	if err == nil {
		updateLastAccessedSessionTime(s.ID)
	}
	sessionLock.Unlock()
	return s, err
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
	result, err := db.DB.Exec(fmt.Sprintf("delete from sessions where last_accessed < now() - interval '%v seconds'", sessionTimeoutSeconds))
	if err == nil {
		count, _ := result.RowsAffected()
		util.Log(util.Warn, "Purged %v expired sessions in %v.", count, time.Since(s))
	} else {
		util.Log(util.Error, "Failed purge: ", err)
	}

	sessionLock.Unlock()
	return err
}

// DeleteSession used when the user logs out of their session.
func DeleteSession(sessionID string) {
	sessionLock.Lock()
	db.DB.Exec("delete from sessions where session_id = ?", sessionID)
	sessionLock.Unlock()
}

func getSessionByUserID(userID int64) (*Session, error) {
	sessionLock.Lock()
	s := &Session{}
	err := db.DB.SelectOne(s, fmt.Sprintf("select id, user_id, session_id from sessions where user_id = %v", userID))
	if err == nil {
		updateLastAccessedSessionTime(s.ID)
	} else {
		util.Log(util.Error, "Error finding session with user id %v: %v", userID, err)
	}
	sessionLock.Unlock()
	return s, err
}

func createSessionString() string {
	rb := make([]byte, sessionStringSize)
	rand.Read(rb)
	return base64.URLEncoding.EncodeToString(rb)
}

func updateLastAccessedSessionTime(sessionID int64) error {
	_, err := db.DB.Exec(fmt.Sprintf("update sessions set last_accessed = now() where id = %v", sessionID))
	if err != nil {
		util.Log(util.Error, "Error updating session access time.", err)
	}
	return err
}
