// Quick access to request kern-context
// HINT: not sure if this feature will stay, create an issue if you think this sucks or is good
// (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
package K

import (
    "net/http"
    "boolshit.net/kern/context"
)

// quick access to context
// example usage: `K(req).Stuff = 1337`
func K( req *http.Request ) *context.Context {
    return context.Get( req )
}
