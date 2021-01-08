// kern.go main include
// (c)copyright 2021 by Gerald Wodni <gerald.wodni@hmail.com>
package kern

import (
    "net/http"

    "boolshit.net/kern/log"
    "boolshit.net/kern/router"
    "boolshit.net/kern/session"
    "boolshit.net/kern/view"
)

type Kern struct {
    Router *router.Router
    BindAddr string
}

func New( bindAddr string ) (kern *Kern) {
    kern = &Kern {
        Router: router.New("/"),
        BindAddr: bindAddr,
    }

    // Set default globals
    view.Globals[ "AppPrefix" ] = "kern.go:"
    view.Globals[ "TitleSuffix" ] = " <- kern.go"

    // activate modules via generic route
    kern.Router.All( "/", func( res http.ResponseWriter, req *http.Request, next router.RouteNext ) {
        // Log every call
        log.SubSection( req.Method, req.URL )
        // Handle session
        session.Install( res, req )
        next()
    })

    // static routes go first
    kern.Router.StaticFile( "/favicon.ico", "image/x-icon", "./default/images/favicon.ico" )
    kern.Router.StaticDir( "/css",    "./default/css"    )
    kern.Router.StaticDir( "/js",     "./default/js"     )
    kern.Router.StaticDir( "/images", "./default/images" )
    kern.Router.StaticDir( "/files",  "./default/files"  )

    return
}

func (kern *Kern) Run() {
    log.Section("Starting kern.go")

    // mount main router
    http.Handle( "/", kern.Router )

    // Catchall 404 at the end of routing
    notFound, err := view.New( "errors/404.gohtml" )
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
