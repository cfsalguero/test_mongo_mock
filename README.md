# Testing a Go web service with gorilla/mux and MongoDB

I wrote a small web service that uses gorilla/mux as a router and MongoDB to store items information.
As I saw many persons asking for testing/moking examples, I've decided to write one.

Let's take a look at the program first. (View on [GitHub](https://github.com/cfsalguero/test_mongo_mock)) 

The main file: `test.go`

```go
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
```

We also have a singleton for storing a MongoDB session in the file  `database/database.go` and the controller that handles the request in the file: `controllers/controllers.go`

```go
package controllers

import (
    "net/http"
    "encoding/json"
    "github.com/cfsalguero/test/database"
    "github.com/gorilla/mux"
    "labix.org/v2/mgo/bson"
)

type Item struct {
    Id          bson.ObjectId `bson:"_id"`
    Description string        `bson:"description"`
}

func DefaultGet(w http.ResponseWriter, r *http.Request) {
    var item *Item
    vars := mux.Vars(r)
    id, _ := vars["id"]
    db := database.New()
    collection := db.C("items")
    _ = collection.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&item)
    if item != nil {
        outData, _ := json.Marshal(item)
        w.Write(outData)
    } else {
        http.Error(w, "Not found", 404)
    }
}
```

The program defines only one route, `/` that handles `GET` requests with an `id`

<a name="steps"></a>In the controller we follow these steps to get a document from MongoDB:

1. Get the `id` from the muxer: `id, _ := vars["id"]`
2. Connect to MongoDB: `db := database.New()`
3. Instantiate the collection to read: `collection := db.C("items")`
4. Execute the query: `    _ = collection.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&item)`
5. Check if the returned item has a value or if it is null to write the json representation of item to the response or send a 404 status.

## Testing our controller

### Prerequisites

We are going to test 3rd party libraries that don't use interfaces, they return structs, and in many cases they act over pointers to structs directly, like __Find().One(&result)__ in [Labix's mgo library](http://labix.org/mgo) so, to be able to test our program, we need to mock up the library calls.

To do that, there is an excellent package, [withmock](https://github.com/qur/withmock), that let us to mock 3rd party libraries like this one.
Withmock relies on [gomock](http://godoc.org/code.google.com/p/gomock/gomock) and it uses [goimports](http://godoc.org/code.google.com/p/go.tools/cmd/goimports) so we need to install all of them before we can start testing our package.

#### Installing prerequisites

```bash
go install code.google.com/p/gomock/gomock
go install code.google.com/p/gomock/mockgen
go install code.google.com/p/go.tools/cmd/goimports
go get github.com/qur/withmock
go get github.com/qur/withmock/mocktest
``` 

### Writing tests

I'm going to skip testing basics, but if you want to refresh those concepts, you can check [this](http://golangtutorials.blogspot.com.ar/2011/10/gotest-unit-testing-and-benchmarking-go.html) page.

First,  we need to create a gomock controller and setup a mock on the MongoDB library:

```go
    // Create the controller
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    // Setup a mock on mongo driver
    mgo.MOCK().SetController(ctrl)
```
Previously I listed [5 steps](#steps) outlining the service. In steps 2 thru 5 we are calling the database, so we need to mock those calls.  To do that, gomock provides a function named EXPECT that allows us to tell our mock that it has to wait for a call to a function (the one we want to mock) and then execute an action.  

To mock step 2, we need to tell our mock that it has to wait until we call mgo's `Dial("localhost")` and then it has to return a fake mongo session.  
```go
    // Create a fake session
    session := &mgo.Session{}
    // We expect Dial against localhost and return the fake session
    mgo.EXPECT().Dial("localhost").Return(session, nil)
```

For steps 3 and 4 we need to do the same, mocking the DB, the collection and all queries.  
The complete mockup looks like this: 
```go
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
    
    // We then expect a query to be created against the collection
    collection.EXPECT().Find(bson.M{"_id": bson.ObjectIdHex("52f6aef226f149b7048b4567")}).Return(query)
    
    // Finally we expect the query to be asked for ONE record
    query.EXPECT().One(gomock.Any()).SetArg(0, &Item{Id: bson.ObjectIdHex("52f6aef226f149b7048b4567"), Description: "Nones"}).Return(nil)
    
```
In the mockup for the __One__ function, we set the parameter to a special matcher `gomock.Any()` that matches any parameter.  
In our controller, the line to read a document from the db is:  

` _ = collection.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&item) `  

where __id__ is the id we received in the url, so we are telling gomock to wait for a call to One with any parameter and then it must set **that** parameter (position 0=first parameter in the call to the One function) to `&Item{Id: bson.ObjectIdHex("52f6aef226f149b7048b4567"), Description: "Nones"}` and it must always return `nil` as the value for the query's error.

At this point, the mock service has been set up, and it's time to call the **DefaultGet** function in our controller to test it.

First, we are going to create a request and a response recorder to track the function's response.
```go
    request, _ := http.NewRequest("GET", "/52f6aef226f149b7048b4567", nil)
    response := httptest.NewRecorder()
```
Now we could call the DefaultGet function like this:  
```go
    DefaultFunction(response, request)
```
but there is a problem with that: the function reads the `id` from the muxer and if we call the function directly, the muxer is being bypassed.

To solve that, we need to create a mux router, set a HandlerFunc pointing to the DefeultGet function, and call mux's ServeHTTP method:
```go
    m := mux.NewRouter()
    m.HandleFunc("/{id}", DefaultGet).Methods("GET") 
    m.ServeHTTP(response, request)
```
After that, **response** will have the function's response and we can check the http status, read the response body, etc, and throw errors if we found something went wrong.

### Runing our tests
To run the tests simply execute:  
```go
withmock go test
```
And the output will be something like this: 
```
PASS
ok 0.005s
```

### Final note  

There is a potential problem in my code: what happens if we change the id in the request from `52f6aef226f149b7048b4567` to `qwe`?  
As we are asuming the id will have a valid Mongo ObjectId, the program will fail because `bson.ObjectIdHex(id)` will return an error since `qwe` is not a valid ObjectId.

I'll let that fix up to you...

