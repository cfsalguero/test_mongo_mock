package database

import (
    "os"
    "fmt"
    "sync"
    "labix.org/v2/mgo"
)

type DB struct {
    Database    *mgo.Database
}

var _init_ctx sync.Once
var _instance *DB

func New() *mgo.Database {
    _init_ctx.Do(func() { 
        _instance = new(DB) 
        session, err := mgo.Dial("localhost")
        if err != nil {
            fmt.Printf("Error en mongo: %+v\n", err)
            os.Exit(1)
        }
        _instance.Database = session.DB("test")
    })
    return _instance.Database
}

