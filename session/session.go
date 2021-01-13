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
    "time"

    "boolshit.net/kern/log"
    "boolshit.net/kern/module"
    "boolshit.net/kern/redis"
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
        destroy( session )
    }
}

// TODO: redis-stubs
func load( req *http.Request, session *Session ) {
    log.Info( "Loading Session: ", session.Id )
    session.active = true
    session.Values["archer"] = "well guess what?"
    session.Values["krieger"] = "yep yep yep yep"
}

func save( req *http.Request, session *Session ) {
    log.Info( "Saving Session: ", session.Id )
    rdb, err := redis.Of( req )
    if err {
        log.Error( "Saving session failed: redis not in http.Request context, is the module loaded?" )
    }
    rdb.Do( "SET", "kern.go", "session saved?" )
    rdb.Flush()
}

func destroy( session *Session ) {
    log.Info( "Destroying Session: ", session.Id )
}

type contextType int; const contextId = contextType(42) // internal context key

// implement module.Request interface (privately)
type sessionModule struct {}
func (m *sessionModule) StartRequest(res http.ResponseWriter, reqIn *http.Request) (reqOut *http.Request, ok bool) {
    session := &Session {}
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
