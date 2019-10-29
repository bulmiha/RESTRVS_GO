package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

var c redis.Conn

var schemaLoader gojsonschema.JSONLoader

type req struct {
	Number int `json:"number"`
}
type res struct {
	Result int `json:"result"`
}

type myerror struct {
	Error string `json:"error"`
	Type  int    `json:"type"`
}

type conf struct {
	DbHost  string
	DbName  int
	AppHost string
}

//GetConfig returns configuration from ENV
func GetConfig() *conf {
	var dbHost, dbPort, appHost, appPort, dbName string
	dbHost = os.Getenv("DB_HOST")
	if len(dbHost) == 0 {
		dbHost = "localhost"
	}
	dbPort = os.Getenv("DB_PORT")
	if len(dbPort) == 0 {
		dbPort = "6379"
	}

	dbName = os.Getenv("DB_NAME")
	if len(dbName) == 0 {
		dbName = "0"
	}

	appHost = os.Getenv("APP_HOST")
	if len(appHost) == 0 {
		appHost = "0.0.0.0"
	}

	appPort = os.Getenv("APP_PORT")
	if len(appPort) == 0 {
		appPort = "8080"
	}

	dbnumber, _ := strconv.ParseInt(dbName, 10, 32)
	c := new(conf)
	c.AppHost = appHost + ":" + appPort
	c.DbHost = dbHost + ":" + dbPort
	c.DbName = int(dbnumber)
	return c
}

func increment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var number req
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(myerror{
			Error: "Wrong request format",
			Type:  3,
		})
		return
	}
	docLoader := gojsonschema.NewBytesLoader(body)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)

	if err != nil {
		panic(err.Error())
	}

	if !result.Valid() {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(myerror{
			Error: "Wrong request format",
			Type:  3,
		})
		return
	}

	json.Unmarshal(body, &number)

	data, _ := c.Do("GET", fmt.Sprint(number.Number))
	if data != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(myerror{
			Error: "Already exists",
			Type:  1,
		})
		return
	}
	data, _ = c.Do("GET", fmt.Sprint(number.Number+1))
	if data != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(myerror{
			Error: "One less than existing",
			Type:  2,
		})
		return
	}
	c.Do("SET", fmt.Sprint(number.Number), "true")
	json.NewEncoder(w).Encode(res{Result: number.Number})

}

func main() {
	r := mux.NewRouter()
	file, err := os.Open("request_schema.json")
	if err != nil {
		panic(err)
	}
	schema, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	schemaLoader = gojsonschema.NewBytesLoader(schema)
	r.HandleFunc("/increment", increment).Methods("POST")

	config := GetConfig()

	c, err = redis.Dial("tcp", config.DbHost, redis.DialDatabase(config.DbName))

	if err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(config.AppHost, r))
}
