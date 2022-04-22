/*
    Hierarchical lookup for pathes.
    Instead of a hardcoded path, a list of directories is traversed to allow for easy extension and subvolume-mounting.

    (c)copyright 2022 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package hierarchy

import (
    "path"
    "io/ioutil"
    "os"

    "github.com/GeraldWodni/kern.go/log"
)

type Hierarchy struct {
    Prefixes []string
}

// Creates a new `Hierarchy` with a list of prefixes; hint: `default` is automatically appended.
func New( prefixes []string ) (hierarchy *Hierarchy, err error) {
    prefixes = append( prefixes, "./default" )

    hierarchy = &Hierarchy {
        Prefixes: prefixes,
    }
    err = hierarchy.init()

    return
}

// check if all prefixes are readable directories to avoid later confusion
func (hierarchy *Hierarchy)init() (err error) {
    for _, prefix := range hierarchy.Prefixes {
        _, err = os.ReadDir( prefix )
        if err != nil {
            return
        }
    }

    return
}

// lookup with fatal fail
func (hierarchy *Hierarchy)LookupFatal( suffixes ...string ) (filename string) {
    filename, ok := hierarchy.Lookup( suffixes... )
    if ! ok {
        log.Fatal( "Hierarchy cannot LookupFatal:", filename )
    }

    return
}

// lookup with optional fail
func (hierarchy *Hierarchy)Lookup( suffixes ...string ) (filename string, ok bool) {
    suffix := path.Join( suffixes... )
    for _, prefix := range hierarchy.Prefixes {
        filename, ok = hierarchy.LookupFile( prefix, suffix )
        if ok {
            return
        }
    }

    return
}

func (hierarchy *Hierarchy) Exists( prefix string, suffix string ) (ok bool) {
    _, ok = hierarchy.LookupFile( prefix, suffix )
    return
}

func (hierarchy *Hierarchy) LookupFile( prefix string, suffix string ) (filename string, ok bool) {
    filename = path.Join( prefix, suffix )
    file, err := os.Open( filename )

    if err != nil && os.IsNotExist( err ) {
        return
    }
    file.Close()
    ok = true
    return
}

// load contents of folder and allow hierarchical overwriting
func (hierarchy *Hierarchy) LookupDirectory( suffixes ...string ) (filenames []string, ok bool) {
    suffix := path.Join( suffixes... )
    filenames = []string{}
    foundFilenames := []string{}
    for _, prefix := range hierarchy.Prefixes {
        if newFilenames, newOk := lookupDirectory( prefix, suffix ); newOk {
            for _, newFilename := range newFilenames {
                if contains( foundFilenames, newFilename ) {
                    continue
                }
                foundFilenames = append( foundFilenames, newFilename )
                filename := path.Join( prefix, suffix, newFilename )
                filenames = append( filenames, filename )
            }
            ok = true
        }
    }
    return
}

func contains[ T comparable ]( items []T, needle T ) bool {
    for _, item := range items {
        if item == needle {
            return true
        }
    }
    return false
}

func lookupDirectory( prefix string, suffix string ) (filenames []string, ok bool) {
    filenames = []string{}
    dirname := path.Join( prefix, suffix )
    files, err := ioutil.ReadDir( dirname )

    if err != nil {
        return
    }

    for _, file := range files {
        if !file.IsDir() {
            filenames = append( filenames, file.Name() )
        }
    }
    ok = true
    return
}
