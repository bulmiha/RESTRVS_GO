package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

var c redis.Conn

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
	DbPort  string
	DbName  int
	AppHost string
	AppPort string
}

//GetConfig returns configuration from ENV
func GetConfig() *conf {
	var dbHost, dbPort, appHost, appPort, dbName string
	dbHost = os.Getenv("DB_HOST")
	if len(dbHost) == 0 {
		dbHost = "redis"
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
	c.AppHost = appHost
	c.AppPort = appPort
	c.DbHost = dbHost
	c.DbPort = dbPort
	c.DbName = int(dbnumber)
	return c
}

func increment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var number req
	json.NewDecoder(r.Body).Decode(&number)
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
	r.HandleFunc("/increment", increment).Methods("POST")

	config := GetConfig()

	var err error

	c, err = redis.Dial("tcp", config.DbHost+":"+config.DbPort, redis.DialDatabase(config.DbName))

	if err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(config.AppHost+":"+config.AppPort, r))
}
