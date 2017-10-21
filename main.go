package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"

	"github.com/singchia/go-mapper/redis"

	"github.com/gorilla/mux"
)

var addr string
var entityMap map[string]CURD

func main() {

	flag.StringVar(&addr, "addr", ":1202", "mapper listen addr")
	flag.StringVar(&redisCli.RedisAddr, "redisAddr", ":6379", "redis listen addr")
	flag.IntVar(&redisCli.MaxIdle, "redisMaxIdle", 10, "redis max idle connections")
	flag.IntVar(&redisCli.MaxActive, "redisMaxActive", 500, "redis max active connections")
	flag.Int64Var(&redisCli.IdleTimeout, "redisIdleTimeout", 30, "redis connection idle timeout")
	flag.Parse()

	entityMap = make(map[string]CURD, 4)
	entityMap["ms"] = &M{}
	entityMap["ps"] = &P{}
	entityMap["maps"] = &Map{}
	entityMap["pers"] = &Per{}
	entityMap["topology"] = &Topology{}

	r := mux.NewRouter()
	POST(r)
	GET(r)
	PUT(r)
	DELETE(r)
	http.ListenAndServe(addr, r)
}

func POST(r *mux.Router) {

	r.HandleFunc("/{entity}/{value}", func(w http.ResponseWriter, req *http.Request) {
		entity, ok := mux.Vars(req)["entity"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		creator, _ := entityMap[entity].(Creator)
		status, data := creator.Create(mux.Vars(req)["value"])
		w.WriteHeader(status)
		w.Write(data)
	}).Methods("POST")
}

func GET(r *mux.Router) {
	r.HandleFunc("/{entity}", func(w http.ResponseWriter, req *http.Request) {
		entity, ok := mux.Vars(req)["entity"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		retriever, _ := entityMap[entity].(Retriever)
		if retriever == nil {
			fmt.Printf("whtf, entity: %s", entity)
			return
		}
		status, data := retriever.Retrieve()
		w.WriteHeader(status)
		w.Write(data)
	}).Methods("GET")

	r.HandleFunc("/{entity}/{value}", func(w http.ResponseWriter, req *http.Request) {
		entity, ok := mux.Vars(req)["entity"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		retriever, _ := entityMap[entity].(Retriever)
		status, data := retriever.RetrieveEntities(mux.Vars(req)["value"])
		w.WriteHeader(status)
		w.Write(data)
	}).Methods("GET")

	r.HandleFunc("/{entity1}/{value1}/{entity2}", func(w http.ResponseWriter, req *http.Request) {
		entity1, _ := mux.Vars(req)["entity1"]
		entity2, _ := mux.Vars(req)["entity2"]

		if !canRetrieved(entity1, entity2) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		retriever, _ := entityMap[entity1].(Retriever)
		var status int
		var data []byte
		if entity2 == "ps" {
			var count = mux.Vars(req)["count"]
			if count == "" {
				count = "1"
			}
			status, data = retriever.RetrieveP(mux.Vars(req)["value1"], count)
		} else {
			status, data = retriever.RetrieveCoEntity(mux.Vars(req)["value1"])
		}
		w.WriteHeader(status)
		w.Write(data)
	}).Methods("GET")
}

func PUT(r *mux.Router) {
	r.HandleFunc("/{entity1}/{value1}/{entity2}/{value2}", func(w http.ResponseWriter, req *http.Request) {
		entity1, _ := mux.Vars(req)["entity1"]
		entity2, _ := mux.Vars(req)["entity2"]
		value1, _ := mux.Vars(req)["value1"]
		value2, _ := mux.Vars(req)["value2"]

		if !isEntity(entity1) || !isEntity(entity2) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		queries := req.URL.Query()
		count := queries.Get("count")
		elements := queries.Get("elements")

		updater, _ := entityMap[entity1].(Updater)
		var status int
		var data []byte
		if count != "" && canMoved(entity1, entity2) {
			//move with count
			status, data = updater.MoveEntities(value1, count, value2)
		} else if elements != "" && canMoved(entity1, entity2) {
			//move with elements
			es, err := base64.StdEncoding.DecodeString(elements)
			if err != nil {
				fmt.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			status, data = updater.MoveFixedEntities(value1, string(es), value2)

		} else if canTook(entity1, entity2) {
			//take the booked entities
			status, data = updater.TakeEntities(value1, value2)
		} else if canAssigned(entity1, entity2) {
			//assign the entities
			status, data = updater.MultiAssign(value1, value2)
		} else {
			status, data = http.StatusNotFound, nil
		}
		w.WriteHeader(status)
		w.Write(data)
	}).Methods("PUT")

	r.HandleFunc("/{entity1}/{value1}", func(w http.ResponseWriter, req *http.Request) {
		entity1, _ := mux.Vars(req)["entity1"]

		queries := req.URL.Query()
		count := queries.Get("count")
		elements := queries.Get("elements")

		updater, _ := entityMap[entity1].(Updater)
		var status int
		var data []byte
		if count != "" && canBooked(entity1) {
			status, data = updater.BookEntities(mux.Vars(req)["value1"], count)

		} else if elements != "" && canBooked(entity1) {
			es, err := base64.StdEncoding.DecodeString(elements)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}
			status, data = updater.BookFixedEntities(mux.Vars(req)["value1"], string(es))
		} else {
			status, data = http.StatusNotFound, nil
		}

		w.WriteHeader(status)
		w.Write(data)
	}).Methods("PUT")
}

func DELETE(r *mux.Router) {
	r.HandleFunc("/{entity}/{value}", func(w http.ResponseWriter, req *http.Request) {
		entity, _ := mux.Vars(req)["entity"]
		deleter, _ := entityMap[entity].(Deleter)
		status, data := deleter.Delete(mux.Vars(req)["value"])
		w.WriteHeader(status)
		w.Write(data)
	}).Methods("DELETE")

	r.HandleFunc("/{entity1}/{value1}/{entity2}", func(w http.ResponseWriter, req *http.Request) {
		entity, _ := mux.Vars(req)["entity1"]
		deleter, _ := entityMap[entity].(Deleter)
		status, data := deleter.MultiDeAssign(mux.Vars(req)["value1"])
		w.WriteHeader(status)
		w.Write(data)
	})

}

func isEntity(field string) bool {
	if field != "ms" && field != "ps" && field != "maps" && field != "pers" {
		return false
	}
	return true
}

func canBooked(field string) bool {
	if field == "pers" || field == "maps" {
		return true
	}
	return false
}

func canTook(field1, field2 string) bool {
	if field1 == field2 {
		return true
	}
	return false
}

func canAssigned(field1, field2 string) bool {
	if field1 == "ms" && field2 == "maps" {
		return true
	} else if field1 == "maps" && field2 == "pers" {
		return true
	} else if field1 == "ps" && field2 == "pers" {
		return true
	} else {
		return false
	}
}

func canMoved(field1, field2 string) bool {
	if field1 == field2 && (field1 == "pers" || field1 == "maps") {
		return true
	}
	return false
}

func canRetrieved(field1, field2 string) bool {
	if field1 == "ms" && field2 == "maps" {
		return true
	} else if field1 == "maps" && field2 == "pers" {
		return true
	} else if field1 == "ps" && field2 == "pers" {
		return true
	} else if field1 == "ms" && field2 == "ps" {
		return true
	} else {
		return false
	}

}
