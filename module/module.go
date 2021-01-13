/*
    module interfaces - implement proper module for easy extension of kern.go

    see `session.sessionModule` for a simple example

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package module

import (
    "net/http"
)

// Request-modules are invoked upon every request
type Request interface {
    // Executed upon request start. Returns a new `http.request` - usually `reqIn` wrapped in a new `Context`.
    // if `ok` is false, all further request handling will be stopped, handler needs to write `res` himself
    StartRequest(res http.ResponseWriter, reqIn *http.Request) (reqOut *http.Request, ok bool)
    // Executed upon exit of request
    EndRequest(res http.ResponseWriter, req *http.Request)
}

var requestModules []Request
func RegisterRequest( requestModule Request ) {
    requestModules = append( requestModules, requestModule )
}

// Called internally by Router
func ExecuteStartRequest( res http.ResponseWriter, reqIn *http.Request) (reqOut *http.Request, ok bool) {
    reqOut = reqIn
    ok = true
    for _, requestModule := range( requestModules ) {
        reqOut, ok = requestModule.StartRequest( res, reqOut )
        if !ok {
            return
        }
    }
    return
}
// Called internally by Router
func ExecuteEndRequest( res http.ResponseWriter, req *http.Request) {
    for i := len(requestModules)-1; i >= 0; i-- {
        requestModule := requestModules[i]
        requestModule.EndRequest( res, req )
    }
}

