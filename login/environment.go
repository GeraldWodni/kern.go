/*
    environment variable based login credentials (useful for docker images in i.e. kubernetes)

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package login

import (
    "os"
    "strings"

    "github.com/GeraldWodni/kern.go/log"
)

const envUserPrefix = "KERN_USER_"
const envPermissionPrefix = "KERN_PERMISSIONS_"

type envCredentials struct {
    users map[string]*User
}

func envGetUsername( name string ) string {
    return strings.TrimPrefix( name, envUserPrefix )
}

// Load users from environment variables, the prefixes are `KERN_USER_` and `KERN_PREMISSIONS_`.
// Example values: `KERN_USER_bob=soopersecret` `KERN_PERMISSIONS_bob=view,add,peel`
// For usage call: `login.Register( login.NewEnvironmentCredentialChecker() )`
func NewEnvironmentCredentialChecker() *envCredentials {
    credentialChecker := &envCredentials{
        users: make(map[string]*User),
    }
    for _, env := range os.Environ() {
        parts := strings.SplitN( env, "=", 2 )
        name  := parts[0]
        value := parts[1]
        username := envGetUsername( name )
        // username - password
        if strings.HasPrefix( name, envUserPrefix ) {
            if user, exists := credentialChecker.users[ username ]; exists {
                // update existing user
                user.Password = value
            } else {
                // create new user
                credentialChecker.users[ username ] = &User {
                    Username: username,
                    Password: value,
                }
            }
        }
        // username - permission
        if strings.HasPrefix( name, envPermissionPrefix ) {
            if user, exists := credentialChecker.users[ username ]; exists {
                // update existing user
                user.Permissions = value
            } else {
                // create new user
                credentialChecker.users[ username ] = &User {
                    Username: username,
                    Permissions: value,
                }
            }
        }
    }

    for name, user := range credentialChecker.users {
        if user.Password == "" {
            log.Errorf( "login.EnvironmentCredentialChecker: user '%s' has no password set, login disabled", name )
            delete( credentialChecker.users, name )
            continue
        }
        if user.Permissions == "" {
            log.Warningf( "login.EnvironmentCredentialChecker: user '%s' has no permissions set", name )
        }
        log.Successf( "login.EnvironmentCredentialChecker user added: %v", user )
    }
    return credentialChecker
}

func (credentialChecker *envCredentials) Check( username string, password string ) (permissions string, ok bool) {
    if user, exists := credentialChecker.users[ username ]; exists && user.Password == password {
        permissions = user.Permissions
        ok = true
        return
    }

    permissions = ""
    ok = false
    return
}
