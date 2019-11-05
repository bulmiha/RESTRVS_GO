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
	c := new(conf)
	dbHost, lookup := os.LookupEnv("DB_HOST")
	if !lookup {
		dbHost = "localhost"
	}
	dbPort, lookup := os.LookupEnv("DB_PORT")
	if !lookup {
		dbPort = "6379"
	}

	dbName, lookup := os.LookupEnv("DB_NAME")
	if lookup {
		dbnumber, err := strconv.ParseInt(dbName, 10, 32)
		if err != nil {
			panic(err)
		}
		c.DbName = int(dbnumber)
	} else {
		c.DbName = 0
	}

	appHost, lookup := os.LookupEnv("APP_HOST")
	if !lookup {
		appHost = "0.0.0.0"
	}

	appPort, lookup := os.LookupEnv("APP_PORT")
	if !lookup {
		appPort = "8080"
	}

	c.AppHost = appHost + ":" + appPort
	c.DbHost = dbHost + ":" + dbPort
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

func ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
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
	r.HandleFunc("/ready", ready).Methods("GET")

	config := GetConfig()

	c, err = redis.Dial("tcp", config.DbHost, redis.DialDatabase(config.DbName))

	if err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(config.AppHost, r))
}
