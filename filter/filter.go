/*
    sanitize input - all unwanted characters are removed

    __HINT:__ always use theese functions for receiving user input.
    Only access the input directly if you are certain there is no other way.

    (c)copyright 2014-2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package filter

import (
    "net/http"
    "net/url"
    "regexp"
    "strings"
    "runtime"
    "reflect"

    "github.com/GeraldWodni/kern.go/log"
)

type Filter func( text string ) string

// most latin characters, use in regex with `LC`
const localChars = "¢£ªºÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþŸÿĀāĐđĒēĲĳĿŀŒœŠšŽžſ€";

func f( text string, regex string ) string {
    regex = strings.ReplaceAll( regex, "LC", localChars )
    re := regexp.MustCompile( regex )
    return re.ReplaceAllString( text, "" )
}

func Address    ( t string ) string { return f( t, "[^-,.\\/ a-zA-Z0-9LC]"      ) }
func Allocnum   ( t string ) string { return f( t, "[^a-zA-Z0-9LC]"             ) }
func Alpha      ( t string ) string { return f( t, "[^a-zA-Z]"                  ) }
func Alnum      ( t string ) string { return f( t, "[^a-zA-Z0-9]"               ) }
func AlnumList  ( t string ) string { return f( t, "[^,a-zA-Z0-9]"              ) }
func Boolean    ( t string ) string { return f( t, "[^01]"                      ) }
func Color      ( t string ) string { return f( t, "[^#a-fA-F0-9]"              ) }
func DateTime   ( t string ) string { return strings.ReplaceAll( f( t, "[^-\\/_.: 0-9]" ), "_", " " ) }
func Decimal    ( t string ) string { return strings.ReplaceAll( f( t, "[^-.,0-9]"      ), ",", "." ) }
func Email      ( t string ) string { return f( t, "[^-@+_.0-9a-zA-Z]"          ) }
func EscapedLink( t string ) string {
    if t, err := url.QueryUnescape( t ); err != nil {
        log.Error( "filter.EscapedLink:", err )
        return ""
    } else {
        return f( t, "[^-_.a-zA-Z0-9\\/]" )
    }
}
func Filename   ( t string ) string { return f( t, "[^-_.0-9a-zA-Z]"            ) }
func Filepath   ( t string ) string { return f( t, "[^-\\/_.0-9a-zA-Z]"         ) }
func Hex        ( t string ) string { return f( t, "[^-0-9a-f]"                 ) }
func Id         ( t string ) string { return f( t, "[^-_.:a-zA-Z0-9]"           ) }
func Int        ( t string ) string { return f( t, "[^-0-9]"                    ) }
func Link       ( t string ) string { return f( t, "[^-_.:a-zA-Z0-9\\/]"        ) }
func LinkItem   ( t string ) string { return f( t, "[^-_.:a-zA-Z0-9]"           ) }
func LinkList   ( t string ) string { return f( t, "[^-,_.:a-zA-Z0-9]"          ) }
func Password   ( t string ) string { return t }
func Raw        ( t string ) string { return t }
func SingleLine ( t string ) string { return f( t, "[^-_\\/ a-zA-Z0-9LC]"       ) }
func Telephone  ( t string ) string { return f( t, "[^-+ 0-9]"                  ) }
func Text       ( t string ) string { return t }
func Uint       ( t string ) string { return f( t, "[^0-9]"                     ) }
func Url        ( t string ) string { return f( t, "[^-?#@&,+_.:\\/a-zA-Z0-9]"  ) }
func Username   ( t string ) string { return f( t, "[^-@_.a-zA-Z0-9]"           ) }

// Reflection based lookup
func filterName( filter Filter ) string {
    name := runtime.FuncForPC( reflect.ValueOf(filter).Pointer() ).Name()
    return strings.ReplaceAll( name, "github.com/GeraldWodni/kern.go/filter.", "" )
}

func fieldName( filter Filter ) string {
    return strings.ToLower( filterName( filter ) )
}

// Get PostValue from request and sanizie it )
func PostName( req *http.Request, name string, filter Filter ) string {
    return filter( req.PostFormValue( name ) )
}

// Get PostValue from request, using the lowecase filter-name as name
func Post( req *http.Request, filter Filter ) string {
    name := fieldName( filter )
    return PostName( req, name, filter )
}
