/*
    Provides a wrapper class around `html.template`.
    Loaded templates are kept in cache but watched with `fsnotify` which invalidates the cache and forces a read on the next `Render`

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package view

import (
    "errors"
    "path"
    "sync"
    "net/http"
    "html/template"

    "github.com/fsnotify/fsnotify"

    "boolshit.net/kern/log"
    "boolshit.net/kern/router"
)

type StringMap map[string]string

type Message struct {
    Type string // TODO: make enum?
    Title string
    Text string
}

// Available to all templates, i.e. `{{.Globals.FooBar}}
var Globals = make(StringMap)

type View struct {
    Template *template.Template
    Filename string
    ReloadRequired bool
    reloadRequiredMutex *sync.Mutex
}

// Creates a new `View` which is immidiatly loaded and watched for file changes
func New( filename string ) (view *View, err error) {
    view = &View {
        Filename: filename,
        ReloadRequired: false,
        reloadRequiredMutex: &sync.Mutex{},
    }
    err = view.load()
    return
}

// Wrapper for Render using a `router.RouteHandler`
func Handler( view *View ) ( routeHandler router.RouteHandler ) {
    return func (res http.ResponseWriter, req *http.Request, next router.RouteNext ) {
        view.Render( res, req, next, nil )
    }
}

// Load a view and directly return `router.RouteHandler`
// Hint: useful for views without `locals`
func NewHandler( filename string ) ( routeHandler router.RouteHandler ) {
    view, err := New( filename )
    if err != nil {
        log.Error( err )
        return router.ErrHandler( err )
    }
    return Handler( view )
}

func (view *View)load() (err error) {
    filename := path.Join( "./default/views", view.Filename )
    view.Template, err = template.ParseFiles( filename )

    // watch for file changes
    if err == nil {
        var watcher *fsnotify.Watcher
        watcher, err = fsnotify.NewWatcher()
        if err != nil {
            return
        }
        go func() {
            for {
                select {
                    case _, ok := <-watcher.Events:
                        if !ok {
                            return
                        }
                        view.reloadRequiredMutex.Lock()
                        if !view.ReloadRequired {
                            log.Infof( "Change detected, reload scheduled for view: %s", view.Filename )
                        }
                        view.ReloadRequired = true
                        view.reloadRequiredMutex.Unlock()
                    case err, ok := <-watcher.Errors:
                        if !ok {
                            return
                        }
                        log.Error( "view.Load->watcher", err )
                }
            }
        }()
        err = watcher.Add( filename )
    }

    return
}

// Render view using `Globals` as well as values passed via `locals`
func (view *View)Render( res http.ResponseWriter, req *http.Request, next router.RouteNext, locals interface{} ) {
    if view.Template == nil {
        router.Err( res, errors.New( "View.Template is nil, check log for previous Errors" ) )
        return
    }

    // reload on template change
    view.reloadRequiredMutex.Lock()
    defer view.reloadRequiredMutex.Unlock()
    if view.ReloadRequired {
        view.ReloadRequired = false
        log.Infof( "Reloading View: %s", view.Filename )
        err := view.load()
        if err != nil {
            router.Err( res, err )
            return
        }
    }

    data := struct {
        Globals StringMap
        Locals interface{}
    }{
        Globals: Globals,
        Locals: locals,
    }

    view.Template.Execute( res, data )
}
