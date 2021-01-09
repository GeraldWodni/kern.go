// login - handler rejects further routing and displays login form
// (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
package login

import (
    "net/http"

    "boolshit.net/kern/log"
    "boolshit.net/kern/router"
    "boolshit.net/kern/view"
)

var loginView *view.View

func init() {
    view, err := view.New( "login.gohtml" )
    if err != nil {
        log.Fatal( "login: cannot load view" )
    }
    loginView = view
}

func PermissionReqired( permission string ) router.RouteHandler {
    return func (res http.ResponseWriter, req *http.Request, next router.RouteNext ) {
        if true {
            loginView.Render( res, req, next, nil )
        } else {
            next()
        }
    }
}
