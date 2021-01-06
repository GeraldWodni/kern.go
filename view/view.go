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
var Globals = make(StringMap)

type View struct {
    Template *template.Template
    Filename string
    ReloadRequired bool
    reloadRequiredMutex *sync.Mutex
}

func New( filename string ) (view *View, err error) {
    view = &View {
        Filename: filename,
        ReloadRequired: false,
        reloadRequiredMutex: &sync.Mutex{},
    }
    err = view.Load()
    return
}

func Handler( view *View ) ( routeHandler router.RouteHandler ) {
    return func (res http.ResponseWriter, req *http.Request, next router.RouteNext ) {
        view.Render( res, req, next, nil )
    }
}

func NewHandler( filename string ) ( routeHandler router.RouteHandler ) {
    view, err := New( filename )
    if err != nil {
        log.Error( err )
        return router.ErrHandler( err )
    }
    return Handler( view )
}

func (view *View)Load() (err error) {
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

func (view *View)Render( res http.ResponseWriter, req *http.Request, next router.RouteNext, locales interface{} ) {
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
        err := view.Load()
        if err != nil {
            router.Err( res, err )
            return
        }
    }

    data := struct {
        Globals StringMap
    }{
        Globals: Globals,
    }

    view.Template.Execute( res, data )
}
