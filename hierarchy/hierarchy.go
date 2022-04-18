/*
    Hierarchical lookup for pathes.
    Instead of a hardcoded path, a list of directories is traversed to allow for easy extension and subvolume-mounting.

    (c)copyright 2022 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package hierarchy

import (
    "path"
    "os"

    "boolshit.net/kern/log"
)

type Hierarchy struct {
    Prefixes []string
}

// Creates a new `Hierarchy` with a list of prefixes; hint: `default` is automatically appended.
func New( prefixes []string ) (hierarchy *Hierarchy, err error) {
    prefixes = append( prefixes, "default" )

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
        filename = path.Join( prefix, suffix )
        file, err := os.Open( filename )

        if err != nil && os.IsNotExist( err ) {
            continue
        }
        file.Close()
        ok = true
        return
    }

    ok = false
    return
}
