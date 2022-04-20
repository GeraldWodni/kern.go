/*
    Provides a wrapper class around `html.template`.
    Loaded templates are kept in cache but watched with `fsnotify` which invalidates the cache and forces a read on the next `Render`

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package view

import (
    "errors"
    "io"
    htmlTemplate "html/template"
    textTemplate "text/template"
    "net/http"
    "os"
    "path"
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

// Pipeline functions exposed to template
type FuncMap map[string] any
var htmlFuncMap = htmlTemplate.FuncMap{}
var textFuncMap = textTemplate.FuncMap{}
var funcs = FuncMap{
    "Hallos": func () []string {
        return []string { "H1", "h2", "h3", }
    },
    "Split": strings.Split,
    "Lines": func(text string) []string {
        return strings.Split( text, "\n" )
    },
    "NonEmpty": func(lines []string) []string {
        nonEmptyLines := make([]string, 0)
        for _, line := range lines {
            if len(line) > 0 {
                nonEmptyLines = append( nonEmptyLines, line )
            }
        }
        return nonEmptyLines
    },
    "ToUpper": strings.ToUpper,
    "ToLower": strings.ToLower,
}

type View struct {
    Filename string
    ReloadRequired bool
    reloadRequiredMutex *sync.Mutex
    ContentType string
}

type ViewTemplate interface {
    Execute(w io.Writer, data any) error
}
type ViewInterface interface {
    Render( http.ResponseWriter, *http.Request, router.RouteNext, interface{} )
    getTemplate() ViewTemplate
    getView() *View
    loadTemplate() error
}


type HtmlView struct {
    View
    Template *htmlTemplate.Template
}

type TextView struct {
    View
    Template *textTemplate.Template
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
                envValues[ viewName ] = htmlTemplate.HTML( value )
            } else {
                envValues[ viewName ] = value
            }
        }
    }
    for name, function := range funcs {
        htmlFuncMap[ name ] = function
        textFuncMap[ name ] = function
    }
}


// Creates a new `View` which is immidiatly loaded and watched for file changes
func NewHtml( filename string ) (view *HtmlView, err error) {
    view = &HtmlView {
        View: View {
            ContentType: "text/html; charset=utf-8",
            Filename: filename,
            ReloadRequired: false,
            reloadRequiredMutex: &sync.Mutex{},
        },
        Template: nil,
    }
    err = loadAndWatch( view )
    return
}
func NewText( filename string, contentType string ) (view *TextView, err error) {
    view = &TextView {
        View: View {
            ContentType: contentType,
            Filename: filename,
            ReloadRequired: false,
            reloadRequiredMutex: &sync.Mutex{},
        },
        Template: nil,
    }
    err = loadAndWatch( view )
    return
}

// Wrapper for Render using a `router.RouteHandler`
func Handler( view ViewInterface ) ( routeHandler router.RouteHandler ) {
    return func (res http.ResponseWriter, req *http.Request, next router.RouteNext ) {
        view.Render( res, req, next, nil )
    }
}

// Load a view and directly return `router.RouteHandler`
// Hint: useful for views without `locals`
func NewHtmlHandler( filename string ) ( routeHandler router.RouteHandler ) {
    var view ViewInterface
    var err error
    view, err = NewHtml( filename )
    if err != nil {
        log.Error( err )
        return router.ErrHandler( err )
    }
    return Handler( view )
}
func NewTextHandler( filename string, contentType string ) ( routeHandler router.RouteHandler ) {
    view, err := NewText( filename, contentType )
    if err != nil {
        log.Error( err )
        return router.ErrHandler( err )
    }
    return Handler( view )
}
func NewCssHandler( filename string ) ( routeHandler router.RouteHandler ) {
    return NewTextHandler( filename, "text/css; charset=utf-8" )
}

func (view *HtmlView) loadTemplate() (err error) {
    view.Template, err = htmlTemplate.New( path.Base(view.Filename) ).Funcs( htmlFuncMap ).ParseFiles( view.Filename )
    return
}
func (view *TextView) loadTemplate() (err error) {
    view.Template, err = textTemplate.New( path.Base(view.Filename) ).Funcs( textFuncMap ).ParseFiles( view.Filename )
    return
}
func (view *HtmlView) getView() *View {
    return &view.View
}
func (view *TextView) getView() *View {
    return &view.View
}
func (view *HtmlView) getTemplate() ViewTemplate {
    return view.Template
}
func (view *TextView) getTemplate() ViewTemplate {
    return view.Template
}



func loadAndWatch( viewInterface ViewInterface ) (err error) {
    err = viewInterface.loadTemplate()
    view := viewInterface.getView()

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
func (view *HtmlView) Render( res http.ResponseWriter, req *http.Request, next router.RouteNext, locals interface{} ) {
    render( view, res, req, next, locals )
}
func (view *TextView) Render( res http.ResponseWriter, req *http.Request, next router.RouteNext, locals interface{} ) {
    render( view, res, req, next, locals )
}
// TODO: figure out if there is a way to mount this function directly onto the type
func render( viewInterface ViewInterface, res http.ResponseWriter, req *http.Request, next router.RouteNext, locals interface{} ) {
    template := viewInterface.getTemplate()
    if template == nil {
        router.Err( res, errors.New( "View.Template is nil, check log for previous Errors" ) )
        return
    }

    // reload on template change
    view := viewInterface.getView()
    view.reloadRequiredMutex.Lock()
    defer view.reloadRequiredMutex.Unlock()
    if view.ReloadRequired {
        view.ReloadRequired = false
        log.Infof( "Reloading View: %s", view.Filename )
        err := loadAndWatch( viewInterface )
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
        Locals interface{}
        Hostname string
        Now time.Time
        NowISO string
    }{
        Globals: Globals,
        Env: envValues,
        Locals: locals,
        Hostname: hostname,
        Now: now,
        NowISO: now.Format("2006-01-02 15:04:05"),
    }

    res.Header().Set( "Content-Type", view.ContentType )
    err := template.Execute( res, data )
    if err != nil {
        log.Error( "View.Render", err )
    }
}

