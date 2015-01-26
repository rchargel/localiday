package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rchargel/localiday/db"
)

const (
	sessionStringSize     = 32
	sessionTimeoutSeconds = 300
)

var sessionLock sync.Mutex

// Session an active user session.
type Session struct {
	ID             int64
	UserID         int64  `db:"user_id"`
	SessionID      string `db:"session_id"`
	LastAccessed   int64  `db:"last_accessed"`
	SessionCreated int64  `db:"session_created"`
}

// CreateNewSession creates a new session and inserts it into the database.
func CreateNewSession(userID int64) *Session {
	if session, err := getSessionByUserID(userID); err != nil {
		return session
	}

	session := &Session{
		UserID:    userID,
		SessionID: createSessionString(),
	}

	insert(session)

	return session
}

// GetSessionBySessionID gets an active session, if it exists, and updates its last accessed value.
func GetSessionBySessionID(sessionID string) (Session, error) {
	sessionLock.Lock()
	var s Session
	err := db.DB.SelectOne(&s,
		fmt.Sprintf("select * from sessions where session_id = '%v' and last_accessed > now() - interval '%v seconds'",
			sessionID, sessionTimeoutSeconds))

	if err == nil {
		_, err = db.DB.Update("update sessions set last_accessed = now() where id = ?", s.ID)
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
		log.Printf("Purged %v expired sessions in %v.", count, time.Since(s))
	} else {
		log.Println("Failed purge: ", err)
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
	err := db.DB.SelectOne(s, "select * from sesions where user_id = ?", userID)
	if err == nil {
		_, err = db.DB.Update("update sessions set last_accessed = new() where id = ?", s.ID)
	}
	sessionLock.Unlock()
	return s, err
}

func createSessionString() string {
	rb := make([]byte, sessionStringSize)
	rand.Read(rb)
	return base64.URLEncoding.EncodeToString(rb)
}
