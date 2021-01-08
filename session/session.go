// session management - via a single cookie
// use Of( http.Request ) to access
// (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
package session

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "fmt"
    "net/http"
    "time"

    //"boolshit.net/kern/context"
    "boolshit.net/kern/log"
)

type Session struct {
    Id string
    Values map[string]string
    active bool
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

func newSessionId() (sessionId string) {
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
func New( res http.ResponseWriter, req *http.Request ) {
    session, _ := Of( req )
    if session.active {
        log.Fatal( "session.New: session already exists" )
        return
    }

    session.Id = newSessionId()
    session.active = true
    setCookie( res, session.Id )
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
func load( session *Session ) {
    log.Info( "Loading Session: ", session.Id )
    session.active = true
    session.Values["archer"] = "well guess what?"
    session.Values["krieger"] = "yep yep yep yep"
}

func save( session *Session ) {
    log.Info( "Saving Session: ", session.Id )
}

func destroy( session *Session ) {
    log.Info( "Destroying Session: ", session.Id )
}

// Install-Of interface: augment request with contexts
type contextType int; const contextId = contextType(42) // internal context key

func Install( res http.ResponseWriter, req *http.Request ) *http.Request {
    session := &Session {}

    if cookie, err := req.Cookie( cookieName ); err == nil {
        session.Id = cookie.Value
        load( session )
        setCookie( res, session.Id )
    }

    ctx := context.WithValue( req.Context(), contextId, session )
    return req.WithContext( ctx )
}

func Uninstall( res http.ResponseWriter, req *http.Request ) {
    if session, active := Of( req ); active {
        save( session )
    }
}

// get session for request-context
func Of( req *http.Request ) (session *Session, ok bool) {
    session = req.Context().Value( contextId ).(*Session)
    ok = session.active
    return
}
