package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"github.com/gomodule/redigo/redis"
)

var c redis.Conn

type req struct {
	Number     int  `json:"number"`
}
type res struct {
	Result	int	`json:"result"`
}

type myerror struct {
	Error	string	`json:"error"`
	Type	int	`json:"type"`
}

func increment(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "application/json")
	var number req
	json.NewDecoder(r.Body).Decode(&number)
	c.Send("GET",fmt.Sprint(number.Number))
	c.Flush()
	data,_:=c.Receive()
	if data!=nil{
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(myerror{
			Error: "Already exists",
			Type:  1,
		})
		return
	}
	c.Send("GET",fmt.Sprint(number.Number+1))
	c.Flush()
	data,_=c.Receive()
	if data!=nil{
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(myerror{
			Error: "One less than existing",
			Type:  2,
		})
		return
	}
	c.Do("SET",fmt.Sprint(number.Number),"true")
	json.NewEncoder(w).Encode(res{Result:number.Number})

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/increment", increment).Methods("POST")
	var err error
	var host, port string
	host = os.Getenv("DB_HOST")
	if len(host) == 0 {
		host = "redis"
	}
	port = os.Getenv("DB_PORT")
	if len(port) == 0 {
		port = "6379"
	}
	c, err = redis.Dial("tcp", host+":"+port)
	if err != nil {
		panic("Redis")
	}
	http.ListenAndServe(":8080", r)
}

