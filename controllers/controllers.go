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
