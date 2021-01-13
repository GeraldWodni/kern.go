# kern.go

Web-framework for web-applications. The [biggest implementation of kern is currently written for node.js](https://github.com/GeraldWodni/kern.js).

As I needed a framework in go for developing kubernetes controllers, I decided to start this minimalistic port.


## Demo

A simple demo application can be found in the [demo-repository](https://github.com/GeraldWodni/kern.go-demo)

### tl;dr demo
```go
package main
import (
    "boolshit.net/kern"
    "boolshit.net/kern/view"
)
func main() {
    app := kern.New(":5000")
    view.Globals["AppName"] = "kern.go demo app"
    app.Router.Get("/", view.NewHandler( "index.gohtml" ) )
    app.Run()
}
```

---
# Documentation
---
# kern

kern.go main include, see the [demo
repository](https://github.com/GeraldWodni/kern.go-demo) for a full demo.


## Usage

#### type Kern

```go
type Kern struct {
	Router   *router.Router
	BindAddr string
}
```


#### func  New

```go
func New(bindAddr string) (kern *Kern)
```
Kern instance hosted on `bindAddr` Hint: mounts `/favicon.ico`, `/css`, `/js`,
`/images`, `/files` from `/default/*`

#### func (*Kern) Run

```go
func (kern *Kern) Run()
```
Run `http.ListenAndServe` for `Kern` instance

---

# router

Routers provide simple separation of mountable modules. This provides a
reusability of code, i.e. a webshop can run standalone when mounted under "/",
or by mounted under "/shop" of a full featured website.



## Usage

#### func  Err

```go
func Err(res http.ResponseWriter, err error)
```
Render an `error` as status code 500

#### type Route

```go
type Route struct {
	Method  string
	Path    string
	Handler RouteHandler
}
```


#### type RouteHandler

```go
type RouteHandler func(res http.ResponseWriter, req *http.Request, next RouteNext)
```

Extend the (res, req) handler interface with a resume-route callback

#### func  ErrHandler

```go
func ErrHandler(err error) RouteHandler
```
Wrapper for Err, provides a RouteHandler for convenience see view/view.go for
example usage

#### type RouteNext

```go
type RouteNext func()
```

Callback function if a RouteHandler was a match but routing should continue

#### type Router

```go
type Router struct {
	// optionally identify router per name
	Name            string
	MountPoint      string
	Routes          []Route
	NotFoundHandler RouteHandler
}
```


#### func  New

```go
func New(mountPoint string) (router *Router)
```
New router with it's mountpoint fixed. Hint: use this function when creating
mountable modules which return their own router.

#### func (*Router) Add

```go
func (router *Router) Add(method string, path string, handler RouteHandler)
```
Add RouteHandler with explicit method mounted at `path`. Use `All`, `Get` OR
`Post` unless crazy methods are required

#### func (*Router) All

```go
func (router *Router) All(path string, handler RouteHandler)
```
Match all methods on `path`

#### func (*Router) Get

```go
func (router *Router) Get(path string, handler RouteHandler)
```
Match all `GET` requests on `path`

#### func (*Router) Mount

```go
func (router *Router) Mount(subRouter *Router)
```
Mount router created by `New` on existing router i.e. `app.Router`

#### func (*Router) NewMounted

```go
func (router *Router) NewMounted(mountPoint string) (subRouter *Router)
```
Wrapper to create a mounted router. Hint: use when implementing a simple
tree-navigation in an app

#### func (*Router) Post

```go
func (router *Router) Post(path string, handler RouteHandler)
```
Match all `POST` requests on `path`

#### func (*Router) ServeHTTP

```go
func (router *Router) ServeHTTP(res http.ResponseWriter, req *http.Request)
```
Gets called by `http`, not to be used by app

#### func (*Router) StaticDir

```go
func (router *Router) StaticDir(path string, dir string)
```
`FileServer` wrapper for exposing the contents of `dir` under `path`

#### func (*Router) StaticFile

```go
func (router *Router) StaticFile(path string, contentType string, filename string)
```
Provide a static file. Kern automatically provides a favicon via this function:

    kern.Router.StaticFile( "/favicon.ico", "image/x-icon", "./default/images/favicon.ico" )

#### func (*Router) StaticHtml

```go
func (router *Router) StaticHtml(path string, html string)
```
Send `html` with the correct mimetype Example:

    router.StaticHtml( '<html><body><h1>Oh noes!, something went terribly wrong</h1></body></html>' )

Hint: usefull for static error messages which need a bit of formatting, use
`view.View` for all else.

#### func (*Router) StaticText

```go
func (router *Router) StaticText(path string, text string)
```
Send `text` with the correct mimetype

---

# view

Provides a wrapper class around `html.template`. Loaded templates are kept in
cache but watched with `fsnotify` which invalidates the cache and forces a read
on the next `Render`



## Usage

```go
var Globals = make(StringMap)
```
Available to all templates, i.e. `{{.Globals.FooBar}}

#### func  Handler

```go
func Handler(view *View) (routeHandler router.RouteHandler)
```
Wrapper for Render using a `router.RouteHandler`

#### func  NewHandler

```go
func NewHandler(filename string) (routeHandler router.RouteHandler)
```
Load a view and directly return `router.RouteHandler` Hint: useful for views
without `locals`

#### type Message

```go
type Message struct {
	Type  string // TODO: make enum?
	Title string
	Text  string
}
```


#### type StringMap

```go
type StringMap map[string]string
```


#### type View

```go
type View struct {
	Template       *template.Template
	Filename       string
	ReloadRequired bool
}
```


#### func  New

```go
func New(filename string) (view *View, err error)
```
Creates a new `View` which is immidiatly loaded and watched for file changes

#### func (*View) Render

```go
func (view *View) Render(res http.ResponseWriter, req *http.Request, next router.RouteNext, locals interface{})
```
Render view using `Globals` as well as values passed via `locals`

---

# redis


## Usage

#### func  Of

```go
func Of(req *http.Request) (rdb redis.Conn, ok bool)
```
get redis connection from request-context i.e. `redis.Of( req ).Do( "SET",
"Lana", "aaaaaaaaa" )`

---

# session

session management - via a single cookie



## Usage

#### func  Destroy

```go
func Destroy(res http.ResponseWriter, req *http.Request)
```
Destroy existing session

#### func  NewSessionId

```go
func NewSessionId() (sessionId string)
```

#### type Session

```go
type Session struct {
	Id     string
	Values map[string]string

	// logged in username
	Username    string
	LoggedIn    bool
	Permissions string
}
```


#### func  New

```go
func New(res http.ResponseWriter, req *http.Request) (session *Session)
```
Start a new session

#### func  Of

```go
func Of(req *http.Request) (session *Session, ok bool)
```
get session for request-context i.e. `session.Of( req ).Id`

---

# module

module interfaces - implement proper module for easy extension of kern.go

see `session.sessionModule` for a simple example



## Usage

#### func  ExecuteEndRequest

```go
func ExecuteEndRequest(res http.ResponseWriter, req *http.Request)
```
Called internally by Router

#### func  ExecuteStartRequest

```go
func ExecuteStartRequest(res http.ResponseWriter, reqIn *http.Request) (reqOut *http.Request, ok bool)
```
Called internally by Router

#### func  RegisterRequest

```go
func RegisterRequest(requestModule Request)
```

#### type Request

```go
type Request interface {
	// Executed upon request start. Returns a new `http.request` - usually `reqIn` wrapped in a new `Context`.
	// if `ok` is false, all further request handling will be stopped, handler needs to write `res` himself
	StartRequest(res http.ResponseWriter, reqIn *http.Request) (reqOut *http.Request, ok bool)
	// Executed upon exit of request
	EndRequest(res http.ResponseWriter, req *http.Request)
}
```

Request-modules are invoked upon every request

---

# login

login - handler rejects further routing and displays login form



## Usage

#### func  PermissionReqired

```go
func PermissionReqired(path string, permission string) (loginRouter *router.Router)
```
Stops all further routing when `permission` is not held by current session.
Displays `loginView` (`login.gohtml`) when no session is found

---

# log

Colorful logger interface which UTC timestampts and multi-level severity



## Usage

```go
var Colors = map[string]string{
	"Reset": "\x1b[0m",

	"Black":   "\x1b[30m",
	"Red":     "\x1b[31m",
	"Green":   "\x1b[32m",
	"Yellow":  "\x1b[33m",
	"Blue":    "\x1b[34m",
	"Magenta": "\x1b[35m",
	"Cyan":    "\x1b[36m",
	"White":   "\x1b[37m",

	"BrightBlack":   "\x1b[1;30m",
	"BrightRed":     "\x1b[1;31m",
	"BrightGreen":   "\x1b[1;32m",
	"BrightYellow":  "\x1b[1;33m",
	"BrightBlue":    "\x1b[1;34m",
	"BrightMagenta": "\x1b[1;35m",
	"BrightCyan":    "\x1b[1;36m",
	"BrightWhite":   "\x1b[1;37m",
}
```

```go
var LevelDebug = Colors["BrightBlack"]
```

```go
var LevelError = Colors["BrightRed"]
```

```go
var LevelFatal = Colors["Red"]
```

```go
var LevelInfo = Colors["BrightBlue"]
```

```go
var LevelSection = Colors["BrightMagenta"]
```

```go
var LevelSubSection = Colors["BrightCyan"]
```

```go
var LevelSuccess = Colors["BrightGreen"]
```

```go
var LevelWarning = Colors["BrightYellow"]
```

#### func  Debug

```go
func Debug(a ...interface{})
```

#### func  Debugf

```go
func Debugf(format string, a ...interface{})
```

#### func  Error

```go
func Error(a ...interface{})
```

#### func  Errorf

```go
func Errorf(format string, a ...interface{})
```

#### func  Fatal

```go
func Fatal(a ...interface{})
```

#### func  Info

```go
func Info(a ...interface{})
```

#### func  Infof

```go
func Infof(format string, a ...interface{})
```

#### func  Log

```go
func Log(level string, a ...interface{})
```

#### func  Logf

```go
func Logf(level string, format string, a ...interface{})
```

#### func  Section

```go
func Section(a ...interface{})
```

#### func  Sectionf

```go
func Sectionf(format string, a ...interface{})
```

#### func  SubSection

```go
func SubSection(a ...interface{})
```

#### func  SubSectionf

```go
func SubSectionf(format string, a ...interface{})
```

#### func  Success

```go
func Success(a ...interface{})
```

#### func  Successf

```go
func Successf(format string, a ...interface{})
```

#### func  Warning

```go
func Warning(a ...interface{})
```

#### func  Warningf

```go
func Warningf(format string, a ...interface{})
```
