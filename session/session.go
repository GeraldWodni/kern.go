package session

import (
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

func Start( res http.ResponseWriter, req *http.Request ) {
    sessionId := newSessionId()
    setCookie( res, sessionId )
}

func load( res http.ResponseWriter, sessionId string ) {
    log.Info( "Loading Session", sessionId )
    session := Session{
        Id: sessionId,
    }
    log.Info( session )
}

func Handle( res http.ResponseWriter, req *http.Request ) {
    if cookie, err := req.Cookie( cookieName ); err == nil {
        sessionId := cookie.Value
        load( res, sessionId )
        setCookie( res, sessionId )
    }
}
