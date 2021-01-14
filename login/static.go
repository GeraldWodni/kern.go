/*
    static credentials, recommended only for developement purposes

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package login

type staticCredentials struct {
    user User
}

// Static credentials (recommended for developement purposes only)
// Example: ``
func NewStaticCredentials( username string, password string, permissions string ) *staticCredentials {
    return &staticCredentials{
        user: User {
            Username: username,
            Password: password,
            Permissions: permissions,
        },
    }
}

func (credentialChecker *staticCredentials) Check( username string, password string ) (permissions string, ok bool) {
    if credentialChecker.user.Username == username && credentialChecker.user.Password == password {
        permissions = credentialChecker.user.Permissions
        ok = true
        return
    }
    permissions = ""
    ok = false
    return
}
