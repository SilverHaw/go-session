package session

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type SessionMgr struct {
	CookieName  string
	Lock        sync.RWMutex
	MaxFifeTime int64
	Sessions    map[string]*Session
}

type Session struct {
	SessionID        string
	LastTimeAccessed time.Time
	Values           map[interface{}]interface{}
}

// Create a session manager
func NewSessionMgr(cookieName string, maxFileTime int64) (mgr *SessionMgr) {
	mgr = &SessionMgr{
		CookieName:  cookieName,
		MaxFifeTime: maxFileTime,
		Sessions:    make(map[string]*Session),
	}
	go mgr.GC()
	return
}

// Timeout recovery
func (mgr *SessionMgr) GC() {
	mgr.Lock.Lock()
	defer mgr.Lock.Unlock()
	for sessionID, session := range mgr.Sessions {
		if session.LastTimeAccessed.Unix()+mgr.MaxFifeTime < time.Now().Unix() {
			delete(mgr.Sessions, sessionID)
		}
	}
	time.AfterFunc(time.Duration(mgr.MaxFifeTime)*time.Second, func() {
		mgr.GC()
	})
}

//
func (mgr *SessionMgr) InitSession(w http.ResponseWriter, r *http.Request) string {
	mgr.Lock.Lock()
	defer mgr.Lock.Unlock()
	newSessionID := url.QueryEscape(mgr.NewSessionID())
	var session *Session = &Session{SessionID: newSessionID, LastTimeAccessed: time.Now(), Values: make(map[interface{}]interface{})}
	mgr.Sessions[newSessionID] = session
	cookie := http.Cookie{Name: mgr.CookieName, Value: newSessionID, Path: "/", HttpOnly: true, MaxAge: int(mgr.MaxFifeTime)}
	http.SetCookie(w, &cookie)
	return newSessionID
}

// Manually delete session
// Delete the cookie
func (mgr *SessionMgr) Destroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(mgr.CookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {
		mgr.Lock.Lock()
		defer mgr.Lock.Unlock()
		delete(mgr.Sessions, cookie.Value)
		expiration := time.Now()
		cookie := http.Cookie{Name: mgr.CookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}

// Deletes the specified session
func (mgr *SessionMgr) Delete(sessionID string) {
	mgr.Lock.Lock()
	defer mgr.Lock.Unlock()
	delete(mgr.Sessions, sessionID)
}

// Sets the specified session
func (mgr *SessionMgr) Set(sessionID string, key interface{}, value interface{}) {
	mgr.Lock.Lock()
	defer mgr.Lock.Unlock()

	if session, ok := mgr.Sessions[sessionID]; ok {
		session.Values[key] = value
	}
}

// Gets the specified session
func (mgr *SessionMgr) Get(sessionID string, key interface{}) (val interface{}, ok bool) {
	mgr.Lock.Lock()
	defer mgr.Lock.Unlock()

	if session, ok := mgr.Sessions[sessionID]; ok {
		if val, ok = session.Values[key]; ok {
			return
		}
	}
	return
}

// Gets the sessionID list
func (mgr *SessionMgr) GetSessionIDlist() (sessionIDList []string) {
	mgr.Lock.Lock()
	defer mgr.Lock.Unlock()
	for k, _ := range mgr.Sessions {
		sessionIDList = append(sessionIDList, k)
	}
	return
}

// Gets the specified sessionid from the cookie
func (mgr *SessionMgr) GetCookie(w http.ResponseWriter, r *http.Request) (sessionID string) {
	cookie, err := r.Cookie(mgr.CookieName)

	if cookie == nil || err != nil {
		return ""
	}
	mgr.Lock.Lock()
	defer mgr.Lock.Unlock()
	sessionID = cookie.Value
	if session, ok := mgr.Sessions[sessionID]; ok {
		session.LastTimeAccessed = time.Now()
		return sessionID
	}
	return
}

// Gets the last login time
func (mgr *SessionMgr) GetLastAccessTime(sessionID string) time.Time {
	mgr.Lock.Lock()
	defer mgr.Lock.Unlock()

	if session, ok := mgr.Sessions[sessionID]; ok {
		return session.LastTimeAccessed
	}

	return time.Now()
}

// Create the sessionID
func (mgr *SessionMgr) NewSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {

	}
	return base64.URLEncoding.EncodeToString(b)
}
