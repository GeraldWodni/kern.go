/*
    logout - router destroys any session and warns otherwise

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package logout

import (
    "net/http"

    "boolshit.net/kern/log"
    "boolshit.net/kern/router"
    "boolshit.net/kern/session"
    "boolshit.net/kern/view"
)

var logoutView *view.View

func init() {
    view, err := view.New( "logout.gohtml" )
    if err != nil {
        log.Fatal( "logout: cannot load view", err )
    }
    logoutView = view
}

func renderView( res http.ResponseWriter, req *http.Request, next router.RouteNext, messages []view.Message ) {
    locals := struct{
        Messages []view.Message
    }{
        Messages: messages,
    }
    logoutView.Render( res, req, next, locals )
}

// Stops all further routing when `permission` is not held by current session.
// Displays `logoutView` (`logout.gohtml`) when no session is found
func Logout( path string ) (logoutRouter *router.Router) {
    logoutRouter = router.New( path )
    logoutRouter.Name = "Logout"
    logoutRouter.All("/", func (res http.ResponseWriter, req *http.Request, next router.RouteNext ) {
        messages := []view.Message{};
        if s, ok := session.Of( req ); ok {
            log.Info( "Logout: %s", s.Username )
            session.Destroy( res, req )
            messages = append( messages, view.Message{
                Type:  "success",
                Title: "Logout",
                Text:  "Have a nice day ;)",
            })
        } else {
            messages = append( messages, view.Message{
                Type:  "error",
                Title: "No session found",
                Text:  "You are not logged in, maybe the session already expired?",
            })
        }
        renderView( res, req, next, messages )
    })
    return
}
