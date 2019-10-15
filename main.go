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

func increment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var number req
	json.NewDecoder(r.Body).Decode(&number)
	c.Send("GET", fmt.Sprint(number.Number))
	c.Flush()
	data, _ := c.Receive()
	if data != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(myerror{
			Error: "Already exists",
			Type:  1,
		})
		return
	}
	c.Send("GET", fmt.Sprint(number.Number+1))
	c.Flush()
	data, _ = c.Receive()
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
	var err error
	var db_host, db_port, app_host, app_port, db_name string
	db_host = os.Getenv("DB_HOST")
	if len(db_host) == 0 {
		db_host = "redis"
	}
	db_port = os.Getenv("DB_PORT")
	if len(db_port) == 0 {
		db_port = "6379"
	}

	db_name = os.Getenv("DB_NAME")
	if len(db_name) == 0 {
		db_name = "0"
	}

	app_host = os.Getenv("APP_HOST")
	if len(app_host) == 0 {
		app_host = "0.0.0.0"
	}

	app_port = os.Getenv("APP_PORT")
	if len(app_port) == 0 {
		app_port = "8080"
	}

	dbnumber, _ := strconv.ParseInt(db_name, 10, 32)

	c, err = redis.Dial("tcp", db_host+":"+db_port, redis.DialDatabase(int(dbnumber)))
	if err != nil {
		panic(err)
	}
	log.Fatal(http.ListenAndServe(app_host+":"+app_port, r))
}
