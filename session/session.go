/*
    session management - via a single cookie

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package session

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "fmt"
    "net/http"
    "strings"
    "time"

    redigo "github.com/gomodule/redigo/redis"

    "github.com/GeraldWodni/kern.go/log"
    "github.com/GeraldWodni/kern.go/module"
    "github.com/GeraldWodni/kern.go/redis"
)

type Session struct {
    Id string
    Values map[string]string
    active bool
    // logged in username
    Username string
    LoggedIn bool
    Permissions string
}

var cookieName string
var cookieTimeout time.Duration

func init() {
    var err error
    // TODO: make these configurable via kern/config module
    cookieName = "KERN_SESSION"
    cookieTimeout, err = time.ParseDuration( "1h" )
    if err != nil {
        log.Fatal( "session cannot parse cookieTimeout", err )
    }
}

func NewSessionId() (sessionId string) {
    hash := sha256.New()
    buffer := make([]byte, 256/8)
    rand.Read( buffer )
    hash.Write( buffer )
    sessionId = fmt.Sprintf( "%x", hash.Sum(nil) )
    return
}

func setCookie( res http.ResponseWriter, sessionId string ) {
    cookie := &http.Cookie {
        Name: cookieName,
        Value: sessionId,
        Path: "/",
        HttpOnly: false,
        Expires: time.Now().Add( cookieTimeout ),
    }

    http.SetCookie( res, cookie )
}
func deleteCookie( res http.ResponseWriter ) {
    cookie := &http.Cookie {
        Name: cookieName,
        Value: "",
        Path: "/",
        HttpOnly: false,
        Expires: time.Unix(0, 0),
    }

    http.SetCookie( res, cookie )
}

// Start a new session
func New( res http.ResponseWriter, req *http.Request ) (session *Session) {
    session, _ = Of( req )
    if session.active {
        log.Fatal( "session.New: session already exists" )
        return
    }

    session.Id = NewSessionId()
    session.active = true
    setCookie( res, session.Id )
    return
}

// Destroy existing session
func Destroy( res http.ResponseWriter, req *http.Request ) {
    session, active := Of( req )
    if active {
        session.active = false
        deleteCookie( res )
        destroy( req, session )
    }
}

const keyPrefix = "kern.go:session:"
const hashKeyPrefix = "usr_"

func (session *Session) keyName() string {
    return keyPrefix + session.Id
}

func load( req *http.Request, session *Session ) {
    rdb, ok := redis.Of( req )
    if !ok {
        log.Error( "Loading session failed: redis not in http.Request context, is the module loaded?" )
        return
    }

    // TODO: export StringMap in redis/redis.go
    hash, err := redigo.StringMap( rdb.Do("HGETALL", session.keyName() ) )
    if err != nil {
        log.Error( "Session load redis error:", err )
        return
    }

    for name, value := range( hash ) {
        if strings.HasPrefix( name, hashKeyPrefix ) {
            name = strings.TrimPrefix( name, hashKeyPrefix )
            session.Values[name] = value
        } else if name == "Username" {
            session.Username = value
            session.LoggedIn = value != ""
        } else {
            log.Warningf( "Session unknown hash-key: \"%s\" (=\"%s\")", name, value )
        }
    }

    session.active = true
    log.Infof( "Session loaded: %s (User: '%s')", session.Id, session.Username )
}

func save( req *http.Request, session *Session ) {
    rdb, ok := redis.Of( req )
    if !ok {
        log.Error( "Saving session failed: redis not in http.Request context, is the module loaded?" )
        return
    }
    args := make([]interface{}, 0)
    // add key name, username as argument
    args = append( args, session.keyName(), "Username", session.Username )

    // add session Values
    for name, value := range( session.Values ) {
        args = append( args, hashKeyPrefix + name, value )
    }

    rdb.Send( "HMSET", args... )
    rdb.Send( "EXPIRE", session.keyName(), int(cookieTimeout.Seconds()) )
    rdb.Flush()
    if _, err := rdb.Receive(); err != nil {
        log.Error( "Session save redis hash error:", err )
    }
    if _, err := rdb.Receive(); err != nil {
        log.Error( "Session save redis ttl error:", err )
    }
    log.Infof( "Session saved: %s (User: '%s')", session.Id, session.Username )
}

func destroy( req *http.Request, session *Session ) {
    log.Info( "Destroying Session: ", session.Id )
    rdb, ok := redis.Of( req )
    if !ok {
        log.Error( "Destroying session failed: redis is not in http.Request context, is the module loaded?" )
        return
    }
    if _, err := rdb.Do( "DEL", session.keyName() ); err != nil {
        log.Error( "Session delete redis error:", err )
    }
}

type contextType int; const contextId = contextType(42) // internal context key

// implement module.Request interface (privately)
type sessionModule struct {}
func (m *sessionModule) StartRequest(res http.ResponseWriter, reqIn *http.Request) (reqOut *http.Request, ok bool) {
    session := &Session {
        Values: make(map[string]string),
    }
    ok=true

    if cookie, err := reqIn.Cookie( cookieName ); err == nil {
        session.Id = cookie.Value
        load( reqIn, session )
        setCookie( res, session.Id )
    }

    ctx := context.WithValue( reqIn.Context(), contextId, session )
    reqOut = reqIn.WithContext( ctx )
    return
}
func (m *sessionModule) EndRequest(res http.ResponseWriter, req *http.Request) {
    if session, active := Of( req ); active {
        save( req, session )
    }
}

// privatly register this module upon import
func init() {
    module.RegisterRequest( module.Request(& sessionModule{}) )
    log.Info( "session module registered" )
}

// get session for request-context
// i.e. `session.Of( req ).Id`
func Of( req *http.Request ) (session *Session, ok bool) {
    session = req.Context().Value( contextId ).(*Session)
    ok = session.active
    return
}
