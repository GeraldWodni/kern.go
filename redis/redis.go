/*
    provides a redis connection wrapper around each request

    __Hint:__ this module is implitly loaded by session.

    (c)copyright 2021 by Gerald Wodni <gerald.wodni@gmail.com>
*/
package redis

import (
    "context"
    "net/http"
    "time"

    "github.com/gomodule/redigo/redis"

    "github.com/GeraldWodni/kern.go/log"
    "github.com/GeraldWodni/kern.go/module"
)

// TODO: make configureable
const address = "localhost:6379"

var pool *redis.Pool

// statically initialize pool
func init() {
    pool = &redis.Pool{
        MaxIdle: 10,
        IdleTimeout: 240 * time.Second,
        Dial: func() (redis.Conn, error) {
            return redis.Dial("tcp", address)
        },
    }
}

type contextType int; const contextId = contextType(42) // internal context key

// implement module.Request interface (privately)
type redisModule struct {}
func (m *redisModule) StartRequest(res http.ResponseWriter, reqIn *http.Request) (reqOut *http.Request, ok bool) {
    ok=true
    rdb := pool.Get()
    ctx := context.WithValue( reqIn.Context(), contextId, rdb )
    reqOut = reqIn.WithContext( ctx )
    return
}
func (m *redisModule) EndRequest(res http.ResponseWriter, req *http.Request) {
    if rdb, active := Of( req ); active {
        rdb.Close()
    }
}

// privatly register this module upon import
func init() {
    module.RegisterRequest( module.Request(& redisModule{}) )
    log.Info( "redis module registered" )
}

// get redis connection from request-context
// i.e. `redis.Of( req ).Do( "SET", "Lana", "aaaaaaaaa" )`
func Of( req *http.Request ) (rdb redis.Conn, ok bool) {
    rdb, ok = req.Context().Value( contextId ).(redis.Conn)
    return
}
