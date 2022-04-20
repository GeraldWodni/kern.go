/*
    kern.go main include, see the [demo repository](https://github.com/GeraldWodni/kern.go-demo) for a full demo.

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@hmail.com>
*/
package kern

import (
    "net/http"

    "boolshit.net/kern/hierarchy"
    "boolshit.net/kern/log"
    "boolshit.net/kern/router"
    "boolshit.net/kern/view"

    // import modules
    // Hint: use `_` when no direct interface is needed, so they are correctly registers as `module`
    _ "boolshit.net/kern/session"
)

type Kern struct {
    Router *router.Router
    Hierarchy *hierarchy.Hierarchy
    BindAddr string
}

// Kern instance hosted on `bindAddr`
// Hint: mounts `/favicon.ico`, `/css`, `/js`, `/images`, `/files` from `/default/*`
func New( bindAddr string, hierarchyPrefixes []string ) (kern *Kern) {

    hierarchyInstance, err := hierarchy.New( hierarchyPrefixes )
    if err != nil {
        log.Fatal( err )
    }

    kern = &Kern {
        Router: router.New("/"),
        Hierarchy: hierarchyInstance,
        BindAddr: bindAddr,
    }

    // Set router name for debugging
    kern.Router.Name = "kern"

    // Set default globals
    view.Globals[ "AppPrefix" ] = "kern.go:"
    view.Globals[ "TitleSuffix" ] = " <- kern.go"

    // activate modules via generic route
    kern.Router.All( "/", func( res http.ResponseWriter, req *http.Request, next router.RouteNext ) {
        // Log every call
        log.SubSection( req.Method, req.URL )
        next()
    })

    // static routes go first
    kern.Router.StaticFile( "/favicon.ico", "image/x-icon", "./default/images/favicon.ico" )
    kern.Router.HierarchyDir( hierarchyInstance, "/css" )
    kern.Router.HierarchyDir( hierarchyInstance, "/js" )
    kern.Router.HierarchyDir( hierarchyInstance, "/images" )
    kern.Router.HierarchyDir( hierarchyInstance, "./default/files"  )

    return
}

// Run `http.ListenAndServe` for `Kern` instance
func (kern *Kern) Run() {
    log.Section("Starting kern.go")

    // mount main router
    http.Handle( "/", kern.Router )

    // Catchall 404 at the end of routing
    notFound, err := view.NewHtml( kern.Hierarchy.LookupFatal( "views", "errors/404.gohtml" ) )
    if err != nil {
        log.Error( err )
    }
    kern.Router.NotFoundHandler = view.Handler( notFound )

    // run server
    if err := http.ListenAndServe(kern.BindAddr, nil); err != nil {
        log.Fatal( err )
    }

    log.Section("I'll be back")
}
