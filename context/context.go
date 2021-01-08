// request context, holds kern's request specific types
// (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
package context

import (
    gocontext "context"
    "net/http"
)

// this type is only to uniquly identify our context
type internalContextType int
const internalContextId = internalContextType(42)

// the Context type grants access to all kern request specific types
type Context struct {
    Stuff int
    I map[string] interface{}
}

// create decorated response
func New(res http.ResponseWriter, req *http.Request) *http.Request {
    ctx := req.Context()

    // initialize kern's Context
    kContext := &Context{
        Stuff: 42,
    }

    kContext.Stuff = 43

    ctx = gocontext.WithValue( ctx, internalContextId, kContext )
    return req.WithContext( ctx );
}

func Get(req *http.Request) *Context {
    return req.Context().Value( internalContextId ).(*Context)
}
