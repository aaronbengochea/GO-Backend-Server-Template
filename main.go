package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type testStruct struct {
	Name   string
	Number int
	Time   int `bson:"time,omitempty" json:"time,omitempty"`
}

type userPlate struct {
	Name     string `json:"name"`
	Email    string `bson:"email,omitempty" json:"email,omitempty"`
	Password string `bson:"password,omitempty" json:"password,omitempty"`
	Text     string `bson:"text,omitempty" json:"text,omitempty"`
}

var dbClient *mongo.Client

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got request: / \n")
	io.WriteString(w, "This is the root of the server")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got request: /hello \n")
	io.WriteString(w, "go server, hello!")
}

func getJson(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Got request: /PostJSON \n")
	payload := map[string]string{
		"Name": "Aaron",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Error marshalling json data", http.StatusInternalServerError)
		fmt.Printf("error marshalling json data: %v \n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

	fmt.Printf("got request: /JSON \n")
}

func postJson(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Got request: /PostJSON \n")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error parsing json body")
	}

	//fmt.Println(string(body))

	var test testStruct
	err = json.Unmarshal(body, &test)
	if err != nil {
		fmt.Printf("Error unmarshalling data into prebuilt struct")
	}

	fmt.Println(test.Name, test.Number, test.Time)
}

// have to pass dbclient pointer
func getOneFromDB(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Got request: /getOneFromDB \n")

	//read in body from the request

	//body is a byte stream [] not actually a json object
	//that is why we need to unmarshall and marshall when receiving and responding
	//that is also why its important to have struct templates so that you can easily parse data
	//from the received byte streams [] into the structs for more functional use
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error parsing json body")
		return
	}

	//create template users for unmarshalling
	//
	var user userPlate
	err = json.Unmarshal(body, &user)
	if err != nil {
		fmt.Printf("Error unmarshalling data into prebuilt struct")
		return
	}

	fmt.Println(user.Name, user.Email, user.Password)

	coll := dbClient.Database("sample_mflix").Collection("users")

	var result userPlate
	filter := bson.D{{"name", user.Name}}

	err = coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Printf("No data found in given db/collection for filter: %v \n", filter)
			return
		}
		fmt.Printf("Error querying db: %v \n", err)
		return
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Error marshalling json data", http.StatusInternalServerError)
		fmt.Printf("error marshalling json data: %v \n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
	fmt.Printf("User found: %s \n", result.Name)

}

func getManyFromDB(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Got request: /getManyfromDB \n")

	//read in body from the request

	//body is a byte stream [] not actually a json object
	//that is why we need to unmarshall and marshall when receiving and responding
	//that is also why its important to have struct templates so that you can easily parse data
	//from the received byte streams [] into the structs for more functional use
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error parsing json body")
		return
	}

	//create template users for unmarshalling
	//
	var user userPlate
	err = json.Unmarshal(body, &user)
	if err != nil {
		fmt.Printf("Error unmarshalling data into prebuilt struct")
		return
	}

	fmt.Println(user.Name, user.Email, user.Password)

	coll := dbClient.Database("sample_mflix").Collection("comments")

	filter := bson.D{{"name", user.Name}}

	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Printf("No data found in given db/collection for filter: %v \n", filter)
			return
		}
		fmt.Printf("Error querying db: %v \n", err)
		return
	}

	var results []userPlate
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Printf("Failed to append db response to userPlate array")
	}

	for _, result := range results {
		res, _ := bson.MarshalExtJSON(result, false, false)
		fmt.Println("comment------------- \n", string(res))
	}

	/*
		jsonData, err := json.Marshal(result)
		if err != nil {
			http.Error(w, "Error marshalling json data", http.StatusInternalServerError)
			fmt.Printf("error marshalling json data: %v \n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
		fmt.Printf("User found: %s \n", result.Name)
	*/

}

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found at root")
		os.Exit(1)
	}

	var uri string
	if uri = os.Getenv("MONGODB_URI"); uri == "" {
		fmt.Println("Error with mongodb uri setting in .env file at root")
		os.Exit(1)
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	var err error
	dbClient, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		fmt.Printf("Error connrecting to db with credentials")
		os.Exit(1)
	}

	defer func() {
		if err = dbClient.Disconnect(context.TODO()); err != nil {
			fmt.Printf("Error occured when disconnecting from the db")
			os.Exit(1)
		}
	}()

	if err := dbClient.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		fmt.Printf("Error pinging db as admin when connecting")
		os.Exit(1)
	}

	fmt.Printf("Successfully Connected to mongoDB deployment \n")

	r := mux.NewRouter()

	r.HandleFunc("/", getRoot)
	r.HandleFunc("/hello", getHello)
	r.HandleFunc("/getJSON", getJson).Methods("GET")
	r.HandleFunc("/postJSON", postJson).Methods("POST")
	r.HandleFunc("/getOneFromDB", getOneFromDB).Methods("GET")
	r.HandleFunc("/getManyFromDB", getManyFromDB).Methods("GET")
	http.Handle("/", r)
	//http.HandleFunc("/", getRoot)
	//http.HandleFunc("/hello", getHello)

	go func() {
		err := http.ListenAndServe(":3333", nil)
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed \n")
			os.Exit(1)
		} else if err != nil {
			fmt.Printf("error starting server: %s \n", err)
			os.Exit(1)
		}
	}()

	fmt.Printf("Server is listening on port 3333 \n")
	select {}
}
