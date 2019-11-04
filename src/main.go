package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

const DATABASE = "senai"
const COLLECTION = "people"

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	firstname string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	lastname  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
	contact   *Contact           `json:"contact,omitempty"`
}

var persons []Person

type Contact struct {
	address *Address `json:"address,omitempty"`
	phone   *Phone   `json:"phone,omitempty"`
}

type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

type Phone struct {
	ddd    string `json:"ddd,omitempty"`
	number string `json:"number,omitempty"`
}

func createPerson(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var person Person
	_ = json.NewDecoder(request.Body).Decode(&person)
	collection := client.Database(DATABASE).Collection(COLLECTION)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, person)
	json.NewEncoder(response).Encode(result)
}

func readPerson(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var people []Person
	collection := client.Database(DATABASE).Collection(COLLECTION)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	personID := mux.Vars(request)["id"]
	if len(personID) == 0 {
		retrivePerson(ctx, collection, response, request)
	} else {
		retriveOnePerson(personID, response, request)
	}

	json.NewEncoder(response).Encode(people)
}

func retriveOnePerson(personID string, response http.ResponseWriter, request *http.Request) {

	id, _ := primitive.ObjectIDFromHex(personID)
	var person Person
	collection := client.Database(DATABASE).Collection(COLLECTION)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
}

func retrivePerson(ctx context.Context, collection *mongo.Collection,
	response http.ResponseWriter, request *http.Request) {
	var people []Person
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
}
func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func deletePerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	personID := mux.Vars(request)["_id"]
	for i, singlePerson := range persons {
		if singlePerson.ID == personID {
			persons = append(persons[:i], persons[i+1:]...)
			fmt.Fprintf(response, "Pessoa com ID %v foi deletodado com sucesso." personID)
		}
	}
}

func updatePerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	personID := mux.Vars(request)["_id"]
	var updatePerson Person
	
	reqBody, err := ioutil.ReadAll(request.Body)
	if err != nil{
		fmt.Fprintf(response, "Informe os dados do evento")
	}
	json.Unmarshal(reqBody, &updatePerson)
	for i, singlePerson := range persons{
		if singlePerson.ID == personID{
			singlePerson.contact = updatePerson.contact
			persons = append(persons[:i], singlePerson)
			json,newEnconder(response).Encode(singlePerson)
		}
	}


	params := mux.Vars(r)
	for index, item := range persons {
		if item.ID == params["id"] {
			persons = append(persons[:index], persons[index+1:]...)
			var person Person
			_ = json.NewDecoder(r.Body).Decode(persons)
			person.ID = params["id"]
			persons = append(persons, persons)
			json.NewEncoder(w).Encode(&persons)
			return
		}
	}
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/person", createPerson).Methods("POST")
	router.HandleFunc("/person", readPerson).Methods("GET")
	router.HandleFunc("/person/{id}", readPerson).Methods("GET")
	router.HandleFunc("/person/{id}", deletePerson).Methods("DELETE")
	router.HandleFunc("/person/{id}", updatePerson).Methods("PATCH")
	http.ListenAndServe("localhost:8081", router)
}
