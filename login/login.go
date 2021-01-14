/*
    login - handler rejects further routing and displays login form

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package login

import (
    "net/http"

    "boolshit.net/kern/filter"
    "boolshit.net/kern/log"
    "boolshit.net/kern/router"
    "boolshit.net/kern/session"
    "boolshit.net/kern/view"
)

var loginView *view.View
// this field must be present for kern.go to recognize the request as a valid login request
// TODO: replace this by a redis-based CSRF
var loginField string
var loginValue string

func init() {
    view, err := view.New( "login.gohtml" )
    if err != nil {
        log.Fatal( "login: cannot load view", err )
    }
    loginView = view
    loginField = session.NewSessionId()
    loginValue = session.NewSessionId()
}

func checkCredentials( username string, password string ) bool {
    // TODO: implement file-based credentials
    log.Info( "username:", username, "password:", password )
    return username == "tester" && password == "mc testface"
}

// check if login is correct
func loginOk( res http.ResponseWriter, req *http.Request, messages *[]view.Message ) bool {

    // check
    if req.PostFormValue(loginField) != loginValue  {
        *messages = append( *messages, view.Message{
            Type: "error",
            Title: "No Login Field",
            Text: "Your POST request does not contain a correct Login key. Please login (again)",
        })
        return false
    }

    username := filter.Post( req, filter.Username )
    password := filter.Post( req, filter.Password )
    if checkCredentials( username, password ) {
        log.Successf( "login: '%s'", username )
        s := session.New( res, req )
        s.Username = username
        s.LoggedIn = true
        s.Values["customId"] = "customValue"
        return true
    }

    *messages = append( *messages, view.Message{
        Type: "error",
        Title: "Wrong credentials",
        Text: "Please provide a correct username and password",
    })

    return false
}

// Check if current session has sufficient rights
func sessionOk( req *http.Request, permission string ) bool {
    s, ok := session.Of( req )
    // TODO: implement permissions
    if ok && s.LoggedIn {
        return true
    }
    return false
}

func renderForm( res http.ResponseWriter, req *http.Request, next router.RouteNext, messages []view.Message ) {
    locals := struct{
        LoginField string
        LoginValue string
        Username string
        Messages []view.Message
    }{
        LoginField: loginField,
        LoginValue: loginValue,
        Username: filter.Post( req, filter.Username ),
        Messages: messages,
    }
    loginView.Render( res, req, next, locals )
}

// Stops all further routing when `permission` is not held by current session.
// Displays `loginView` (`login.gohtml`) when no session is found
func PermissionReqired( path string, permission string ) (loginRouter *router.Router) {
    loginRouter = router.New( path )
    loginRouter.Name = "Login"
    loginRouter.Post("/", func (res http.ResponseWriter, req *http.Request, next router.RouteNext ) {
        messages := []view.Message{}
        if sessionOk( req, permission ) {
            next() // keep on routing
            return
        }
        if loginOk( res, req, &messages ) {
            req.Method = "GET" // re-write method (login successfull)
            next() // keep on routing
            return
        }
        renderForm( res, req, next, messages )
    })
    loginRouter.Get("/", func (res http.ResponseWriter, req *http.Request, next router.RouteNext ) {
        if sessionOk( req, permission ) {
            next() // keep on routing
            return
        }
        renderForm( res, req, next, []view.Message{} )
    })
    return
}
