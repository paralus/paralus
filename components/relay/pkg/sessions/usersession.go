package sessions

import (
	"hash"
	"net/http"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/julienschmidt/httprouter"
	"github.com/twmb/murmur3"
)

//UserSession session cache
type UserSession struct {
	// Type of the server. Relay means user-facing
	// Dialin means cluster-facing
	Type string

	// ServerName of the server which accepted the connection
	ServerName string

	// CertSNI derived from client certificate
	CertSNI string

	// server block of this connection
	//server *Server

	// DialinCachedKey already stiched to a dialin
	DialinCachedKey string

	// ErrorFlag indicate the session got 401/502
	ErrorFlag bool

	// UserName used in login
	UserName string

	// RoleName Role
	RoleName string

	// IsReadrole read/write
	IsReadrole bool

	// IsReadrole OrgAdmin
	IsOrgAdmin bool

	// EnforceOrgAdminOnlySecret access
	EnforceOrgAdminOnlySecret bool
}

var (
	// userSessions cache
	userSessions *ristretto.Cache

	// for URL matching
	roleCheck *httprouter.Router

	// for URL matching
	roleCheckSecret *httprouter.Router

	_hashPool = sync.Pool{
		New: func() interface{} {
			// The Pool's New function should generally only return pointer
			// types, since a pointer can be put into the return interface
			// value without an allocation:
			return murmur3.New64()
		},
	}
)

var _dummyHandler = func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {}

//InitUserSessionCache init user cache
func InitUserSessionCache() error {
	var err error
	userSessions, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // Num keys to track frequency of (10k).
		MaxCost:     1 << 25, // Maximum cost of cache (32MB).
		BufferItems: 64,      // Number of keys per Get buffer.
	})

	roleCheck = httprouter.New()
	roleCheck.Handle("POST", "/api/:version/namespaces/:namespace/pods/:pod/exec", _dummyHandler)
	roleCheck.Handle("POST", "/api/:version/namespaces/:namespace/secrets", _dummyHandler)
	roleCheck.Handle("GET", "/api/:version/namespaces/:namespace/secrets", _dummyHandler)
	roleCheck.Handle("GET", "/api/:version/secrets", _dummyHandler)

	roleCheckSecret = httprouter.New()
	roleCheckSecret.Handle("POST", "/api/:version/namespaces/:namespace/secrets", _dummyHandler)
	roleCheckSecret.Handle("GET", "/api/:version/namespaces/:namespace/secrets", _dummyHandler)
	roleCheckSecret.Handle("GET", "/api/:version/secrets", _dummyHandler)

	return err
}

// GetRoleCheck check the given method and URL matches
func GetRoleCheck(method, url string) bool {
	h, _, _ := roleCheck.Lookup(method, url)
	return h != nil
}

// GetRoleCheck check the given method and URL matches
func GetSecretRoleCheck(method, url string) bool {
	h, _, _ := roleCheckSecret.Lookup(method, url)
	return h != nil
}

//GetUserSession get the user session
func GetUserSession(skey string) (*UserSession, bool) {
	hkey := getUserCacheKey(skey)
	if val, ok := userSessions.Get(hkey); ok {
		return val.(*UserSession), true
	}
	return nil, false
}

//AddUserSession add user session
func AddUserSession(s *UserSession, skey string) {
	hkey := getUserCacheKey(skey)
	userSessions.SetWithTTL(hkey, s, 100, time.Minute*15)
}

//DeleteUserSession add user session
func DeleteUserSession(skey string) {
	hkey := getUserCacheKey(skey)
	userSessions.Del(hkey)
}

//UpdateUserSessionExpiry set a short TTL when error happens
func UpdateUserSessionExpiry(skey string, secs int) {
	hkey := getUserCacheKey(skey)
	if val, ok := userSessions.Get(hkey); ok {
		s := val.(*UserSession)
		if !s.ErrorFlag {
			s.ErrorFlag = true
			userSessions.Del(hkey)
			userSessions.SetWithTTL(hkey, s, 100, time.Second*time.Duration(secs))
		}
	}
}

//SetSessionErrorFlag ser session errflg
func SetSessionErrorFlag(skey string) {
	hkey := getUserCacheKey(skey)
	if val, ok := userSessions.Get(hkey); ok {
		s := val.(*UserSession)
		if !s.ErrorFlag {
			s.ErrorFlag = true
		}
	}
}

//getUserCacheKey get cache key
func getUserCacheKey(skey string) (key uint64) {
	hasher := _hashPool.Get().(hash.Hash64)
	hasher.Reset()
	hasher.Write([]byte(skey))
	key = hasher.Sum64()
	_hashPool.Put(hasher)
	return
}
