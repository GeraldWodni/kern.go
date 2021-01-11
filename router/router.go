package router

import (
    gopath "path"
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"

    "boolshit.net/kern/context"
    "boolshit.net/kern/log"
    . "boolshit.net/kern/K"
)

type RouteNext func()
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
    ExecutePrePost bool
}
func New( mountPoint string ) (router *Router) {
    router = &Router{
        MountPoint: mountPoint,
        Routes: make([]Route, 0),
        NotFoundHandler: nil,
        ExecutePrePost: false,
    }
    return
}

func (router *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
    kernReq := context.New( res, req )
    router.serve( res, kernReq )
}

// handlers for start and end of routing
func (router *Router) preServe(res http.ResponseWriter, req *http.Request) {
    k := context.Get(req)
    log.Debug( "STUFF (sta):", k.Stuff )
    k.Stuff = 44
    K(req).Stuff = 1337
}
func (router *Router) postServe(res http.ResponseWriter, req *http.Request) {
    k := context.Get(req)
    log.Debug( "STUFF (end):", k.Stuff, k )
}

func (router *Router) serve(res http.ResponseWriter, req *http.Request) {
    router.preServe( res, req )
    defer router.postServe( res, req )

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
func Err( res http.ResponseWriter, err error ) {
    res.WriteHeader(500)
    res.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprintf( res, "<h1>Error</h1><pre>" )
    fmt.Fprintf( res, err.Error() )
    fmt.Fprintf( res, "</pre>" )
    log.Error( err )
}
func ErrHandler( err error ) RouteHandler {
    return func( res http.ResponseWriter, req *http.Request, next RouteNext ) {
        Err( res, err )
    }
}
func (router *Router) All( path string, handler RouteHandler ) {
    router.Add( "ALL", path, handler )
}
func (router *Router) Get( path string, handler RouteHandler ) {
    router.Add( http.MethodGet, path, handler )
}
func (router *Router) Post( path string, handler RouteHandler ) {
    router.Add( http.MethodPost, path, handler )
}

func (router *Router) Mount( subRouter *Router ) {
    mountPoint := gopath.Join( router.MountPoint, subRouter.MountPoint )
    router.All( mountPoint, func( res http.ResponseWriter, req *http.Request, next RouteNext ) {
        subRouter.NotFoundHandler = func( _ http.ResponseWriter, _ *http.Request, _ RouteNext ) {
            next()
        }
        subRouter.serve( res, req )
    })
}
func (router *Router) NewMounted( mountPoint string ) (subRouter *Router) {
    subRouter = New( mountPoint )
    router.Mount( subRouter )
    return
}

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
func (router *Router) StaticDir( path string, dir string ) {
    fileServer := http.FileServer( http.Dir(dir) )
    router.Get( path, func( res http.ResponseWriter, req *http.Request, next RouteNext ) {
        http.StripPrefix( path, fileServer ).ServeHTTP( res, req )
    })
}
func (router *Router) StaticText( path string, text string ) {
    router.Get( path, func( res http.ResponseWriter, req *http.Request, next RouteNext ) {
        res.Header().Set("Content-Type", "text/plain; charset=utf-8")
        fmt.Fprintf( res, text )
    })
}
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
