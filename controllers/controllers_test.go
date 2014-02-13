package controllers

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "code.google.com/p/gomock/gomock"
    "labix.org/v2/mgo" // mock
    "labix.org/v2/mgo/bson"
    "github.com/gorilla/mux"
)


func TestMongo(t *testing.T) {
    // Create a gomock controller
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    // Setup the mock mgo to use the controller
    mgo.MOCK().SetController(ctrl)
    
    // Create a mock *mgo.Session, *mgo.Database, *mgo.Collection, and
    // *mgo.Query
    session := &mgo.Session{}
    database := &mgo.Database{}
    collection := &mgo.Collection{}
    query := &mgo.Query{}
    
    // We expect Dial against localhost
    mgo.EXPECT().Dial("localhost").Return(session, nil)
    
    // We expect the named database to be opened
    session.EXPECT().DB("test").Return(database)
    
    // We also expect the named collection to be opened
    database.EXPECT().C("items").Return(collection)
    
    // We then expect a query to be created against the collection, using nil to
    // ask for all entries
    collection.EXPECT().Find(bson.M{"_id": bson.ObjectIdHex("52f6aef226f149b7048b4567")}).Return(query)
    
    // Finally we expect the query to be asked for all the matches
    query.EXPECT().One(gomock.Any()).SetArg(0, &Item{Id: bson.ObjectIdHex("52f6aef226f149b7048b4567"), Description: "Nones"}).Return(nil)
    
    request, _ := http.NewRequest("GET", "/52f6aef226f149b7048b4567", nil)
    response := httptest.NewRecorder()
    m := mux.NewRouter()
    m.HandleFunc("/{id}", DefaultGet).Methods("GET") 
    m.ServeHTTP(response, request)
    if response.Code != http.StatusOK {
        
        t.Fatalf("Response body did not contain expected %v:\n\tbody: %v", "200", response.Code)
    }
    
}
