package main

import (
    "fmt"
    "github.com/gorilla/mux"
    "github.com/cfsalguero/test/database"
    "github.com/cfsalguero/test/controllers"
    "net/http"
)

func main() {
    _ = database.New("localhost")
    r := mux.NewRouter()
    r.StrictSlash(true)
    r.HandleFunc("/{id}", controllers.DefaultGet).Methods("GET") 
    http.Handle("/", r)
    fmt.Printf("Listening on port %s\n", "8080")
    http.ListenAndServe(":8080", nil)
}

