/*
    Routers provide simple separation of mountable modules.
    This provides a reusability of code,
    i.e. a webshop can run standalone when mounted under "/",
    or by mounted under "/shop" of a full featured website.

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package router

import (
    gopath "path"
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"

    "boolshit.net/kern/log"
    "boolshit.net/kern/module"
)

// Callback function if a RouteHandler was a match but routing should continue
type RouteNext func()

// Extend the (res, req) handler interface with a resume-route callback
type RouteHandler func (res http.ResponseWriter, req *http.Request, next RouteNext )

type Route struct {
    Method string
    Path string
    Handler RouteHandler
}

type Router struct {
    MountPoint string
    Routes []Route
    NotFoundHandler RouteHandler
}

// New router with it's mountpoint fixed.
// Hint: use this function when creating mountable modules which return their own router.
func New( mountPoint string ) (router *Router) {
    router = &Router{
        MountPoint: mountPoint,
        Routes: make([]Route, 0),
        NotFoundHandler: nil,
    }
    return
}

// Gets called by `http`, not to be used by app
func (router *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
    req, ok := module.ExecuteStartRequest( res, req )
    if ok {
        router.serve( res, req )
        module.ExecuteEndRequest( res, req )
    }
}

func (router *Router) serve(res http.ResponseWriter, req *http.Request) {
    for _, route := range router.Routes {
        if (route.Method == "ALL" || req.Method == route.Method ) && strings.HasPrefix( req.URL.Path, route.Path ) {
            resume := false
            route.Handler( res, req, func() {
                resume = true
            })
            if resume == false {
                return
            }
        }
    }
    if router.NotFoundHandler != nil {
        router.NotFoundHandler( res, req, nil )
    } else {
        notFound( res, "kern.go: set Router.NotFoundHandler( res, req, next ) to display a custom response" )
    }
}

// Add RouteHandler with explicit method mounted at `path`. Use `All`, `Get` OR `Post` unless crazy methods are required
func (router *Router) Add( method string, path string, handler RouteHandler ) {
    mountPath := gopath.Join( router.MountPoint, path )
    log.Debugf( "Router %s handles %s (%s)", router.MountPoint, path, mountPath )

    route := Route{
        Method: method,
        // TODO: evaluate wether joined Path is cool or a express-based prefix removal is nicer
        Path: mountPath,
        Handler: handler,
    }
    router.Routes = append( router.Routes, route )
}
// Render an `error` as status code 500
func Err( res http.ResponseWriter, err error ) {
    res.WriteHeader(500)
    res.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprintf( res, "<h1>Error</h1><pre>" )
    fmt.Fprintf( res, err.Error() )
    fmt.Fprintf( res, "</pre>" )
    log.Error( err )
}
// Wrapper for Err, provides a RouteHandler for convenience
// see view/view.go for example usage
func ErrHandler( err error ) RouteHandler {
    return func( res http.ResponseWriter, req *http.Request, next RouteNext ) {
        Err( res, err )
    }
}
// Match all methods on `path`
func (router *Router) All( path string, handler RouteHandler ) {
    router.Add( "ALL", path, handler )
}
// Match all `GET` requests on `path`
func (router *Router) Get( path string, handler RouteHandler ) {
    router.Add( http.MethodGet, path, handler )
}
// Match all `POST` requests on `path`
func (router *Router) Post( path string, handler RouteHandler ) {
    router.Add( http.MethodPost, path, handler )
}
// Mount router created by `New` on existing router i.e. `app.Router`
func (router *Router) Mount( subRouter *Router ) {
    mountPoint := gopath.Join( router.MountPoint, subRouter.MountPoint )
    router.All( mountPoint, func( res http.ResponseWriter, req *http.Request, next RouteNext ) {
        subRouter.NotFoundHandler = func( _ http.ResponseWriter, _ *http.Request, _ RouteNext ) {
            next()
        }
        subRouter.serve( res, req )
    })
}
// Wrapper to create a mounted router.
// Hint: use when implementing a simple tree-navigation in an app
func (router *Router) NewMounted( mountPoint string ) (subRouter *Router) {
    subRouter = New( mountPoint )
    router.Mount( subRouter )
    return
}

// Provide a static file.
// Kern automatically provides a favicon via this function:
//     kern.Router.StaticFile( "/favicon.ico", "image/x-icon", "./default/images/favicon.ico" )
func (router *Router) StaticFile( path string, contentType string, filename string ) {
    router.Get( path, func( res http.ResponseWriter, req *http.Request, next RouteNext ) {
        content, err := ioutil.ReadFile( filename )
        if err != nil {
            Err( res, err )
            return
        }

        res.Header().Set( "Content-Type", contentType )
        res.Write( content )
    })
}
// `FileServer` wrapper for exposing the contents of `dir` under `path`
func (router *Router) StaticDir( path string, dir string ) {
    fileServer := http.FileServer( http.Dir(dir) )
    router.Get( path, func( res http.ResponseWriter, req *http.Request, next RouteNext ) {
        http.StripPrefix( path, fileServer ).ServeHTTP( res, req )
    })
}
// Send `text` with the correct mimetype
func (router *Router) StaticText( path string, text string ) {
    router.Get( path, func( res http.ResponseWriter, req *http.Request, next RouteNext ) {
        res.Header().Set("Content-Type", "text/plain; charset=utf-8")
        fmt.Fprintf( res, text )
    })
}
// Send `html` with the correct mimetype
// Example:
//     router.StaticHtml( '<html><body><h1>Oh noes!, something went terribly wrong</h1></body></html>' )
// Hint: usefull for static error messages which need a bit of formatting, use `view.View` for all else.
func (router *Router) StaticHtml( path string, html string ) {
    router.Get( path, func( res http.ResponseWriter, req *http.Request, next RouteNext ) {
        res.Header().Set("Content-Type", "text/html; charset=utf-8")
        fmt.Fprintf( res, html )
    })
}

// Helper functions
func notFound( res http.ResponseWriter, text string ) {
    res.WriteHeader(404)
    res.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprintf( res, `<html lang="en"><head><title>Not Found</title></head><body><h1>404 Not Found</h1><p>` + text + `</p></body>` )
}
