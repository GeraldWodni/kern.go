/*
    Colorful logger interface which UTC timestampts and multi-level severity

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package log

import (
    "io"
    "log"
    "time"
    "fmt"
)

var Colors = map[string]string{
    "Reset":    "\x1b[0m",

    "Black":    "\x1b[30m",
    "Red":      "\x1b[31m",
    "Green":    "\x1b[32m",
    "Yellow":   "\x1b[33m",
    "Blue":     "\x1b[34m",
    "Magenta":  "\x1b[35m",
    "Cyan":     "\x1b[36m",
    "White":    "\x1b[37m",

    "BrightBlack":    "\x1b[1;30m",
    "BrightRed":      "\x1b[1;31m",
    "BrightGreen":    "\x1b[1;32m",
    "BrightYellow":   "\x1b[1;33m",
    "BrightBlue":     "\x1b[1;34m",
    "BrightMagenta":  "\x1b[1;35m",
    "BrightCyan":     "\x1b[1;36m",
    "BrightWhite":    "\x1b[1;37m",
}

var LevelFatal      = Colors["Red"]
var LevelError      = Colors["BrightRed"]
var LevelWarning    = Colors["BrightYellow"]
var LevelInfo       = Colors["BrightBlue"]
var LevelSuccess    = Colors["BrightGreen"]
var LevelSection    = Colors["BrightMagenta"]
var LevelSubSection = Colors["BrightCyan"]
var LevelDebug      = Colors["BrightBlack"]

type prefixWriter struct {
    f func() string
    w io.Writer
}

func (p prefixWriter) Write(b []byte) (size int, err error) {
    if size, err = p.w.Write( []byte(p.f()) ); err != nil {
        return
    }
    sizeText, err := p.w.Write(b)
    return size+sizeText, err
}

func init() {
    log.SetFlags(0)
    log.SetOutput( prefixWriter{
        f: func() string {
            return Colors["White"] + time.Now().UTC().Format( "2006-01-02 15:04:05" ) + Colors["Reset"] + " "
        },
        w: log.Writer(),
    })
}

func Log( level string, a ...interface{} ) {
    out := level
    switch level {
        case LevelFatal:        out += "FATAL      " + Colors["BrightRed"]
        case LevelError:        out += "ERROR      " + Colors["Red"]
        case LevelWarning:      out += "WARNING    " + Colors["Yellow"]
        case LevelInfo:         out += "INFO       " + Colors["Blue"]
        case LevelSuccess:      out += "SUCCESS    " + Colors["Green"]
        case LevelSection:      out += "SECTION    " + Colors["Magenta"]
        case LevelSubSection:   out += "SUBSECTION " + Colors["Cyan"]
        case LevelDebug:        out += "DEBUG      "
    }
    out += " ";

    separator := ""
    for _, argument := range a {
        out += separator
        out += fmt.Sprintf( "%v", argument )
        separator = " "
    }
    out += Colors["Reset"]
    if level == LevelFatal {
        log.Fatal( out )
    } else {
        log.Println( out )
    }
}

func Logf( level string, format string, a ...interface{} ) {
    out := fmt.Sprintf( format, a... )
    Log( level, out )
}

func Fatal(     a ...interface{} ) { Log( LevelFatal       , a... ) }
func Error(     a ...interface{} ) { Log( LevelError       , a... ) }
func Warning(   a ...interface{} ) { Log( LevelWarning     , a... ) }
func Info(      a ...interface{} ) { Log( LevelInfo        , a... ) }
func Success(   a ...interface{} ) { Log( LevelSuccess     , a... ) }
func Section(   a ...interface{} ) { Log( LevelSection     , a... ) }
func SubSection(a ...interface{} ) { Log( LevelSubSection  , a... ) }
func Debug(     a ...interface{} ) { Log( LevelDebug       , a... ) }

func Errorf(     format string, a ...interface{} ) { Logf( LevelError      , format , a... ) }
func Warningf(   format string, a ...interface{} ) { Logf( LevelWarning    , format , a... ) }
func Infof(      format string, a ...interface{} ) { Logf( LevelInfo       , format , a... ) }
func Successf(   format string, a ...interface{} ) { Logf( LevelSuccess    , format , a... ) }
func Sectionf(   format string, a ...interface{} ) { Logf( LevelSection    , format , a... ) }
func SubSectionf(format string, a ...interface{} ) { Logf( LevelSubSection , format , a... ) }
func Debugf(     format string, a ...interface{} ) { Logf( LevelDebug      , format , a... ) }
