/*
    Provides a wrapper class around `html.template`.
    Loaded templates are kept in cache but watched with `fsnotify` which invalidates the cache and forces a read on the next `Render`

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package view

import (
    "errors"
    "html/template"
    "net/http"
    "os"
    "sync"
    "strings"
    "time"

    "github.com/fsnotify/fsnotify"

    "boolshit.net/kern/log"
    "boolshit.net/kern/router"
)

type InterfaceMap map[string]interface{}

type Message struct {
    Type string // TODO: make enum?
    Title string
    Text string
}

// Available to all templates, i.e. `{{.Globals.FooBar}}
var Globals = make(InterfaceMap)

// Environment
const envViewPrefix = "KERN_VIEW_"
var envValues = make(InterfaceMap)

// Functions exposed to template
var funcs = template.FuncMap{
    "Hallo": func () string {
        return "HALLO FUNC"
    },
    "Hallos": func () []string {
        return []string {
            "H1",
            "h2",
            "h3",
        }
    },
    "ToUpper": strings.ToUpper,
    "Extra": func(text string) string {
        return text + "EXTRA"
    },
}

type View struct {
    Template *template.Template
    Filename string
    ReloadRequired bool
    reloadRequiredMutex *sync.Mutex
}

// Load environment
func envGetName( name string ) string {
    return strings.TrimPrefix( name, envViewPrefix )
}
func init() {
    for _, env := range os.Environ() {
        parts := strings.SplitN( env, "=", 2 )
        name  := parts[0]
        value := parts[1]
        viewName := envGetName( name )
        if strings.HasPrefix( name, envViewPrefix ) {
            if strings.HasSuffix( name, "_HTML" ) {
                envValues[ viewName ] = template.HTML( value )
            } else {
                envValues[ viewName ] = value
            }
        }
    }
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
    view.Template, err = template.ParseFiles( view.Filename )

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
        err = watcher.Add( view.Filename )
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

    hostname, _, _ := strings.Cut( req.Host, ":" )
    now := time.Now().UTC()
    data := struct {
        Globals InterfaceMap
        Env InterfaceMap
        Funcs template.FuncMap
        Locals interface{}
        Hostname string
        Now time.Time
        NowISO string
    }{
        Globals: Globals,
        Env: envValues,
        Funcs: funcs,
        Locals: locals,
        Hostname: hostname,
        Now: now,
        NowISO: now.Format("2006-01-02 15:04:05"),
    }

    err := view.Template.Execute( res, data )
    if err != nil {
        log.Error( "view.Render", err );
    }
}
