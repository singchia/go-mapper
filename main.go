package main

import (
	"flag"
	"net/http"

	"./redis"

	"github.com/gorilla/mux"
)

const keyPrefix = "mapper:"

var addr string
var entityMap map[string]interface{}

func main() {

	flag.StringVar(&addr, "addr", ":1202", "mapper listen addr")
	flag.StringVar(&redisCli.RedisAddr, "redisAddr", ":6379", "redis listen addr")
	flag.IntVar(&redisCli.MaxIdle, "redisMaxIdle", 10, "redis max idle connections")
	flag.IntVar(&redisCli.MaxActive, "redisMaxActive", 500, "redis max active connections")
	flag.Int64Var(&redisCli.IdleTimeout, "redisIdleTimeout", 30, "redis connection idle timeout")
	flag.Parse()

	entityMap = make(map[string]interface{}, 4)
	entityMap["ms"] = &M{}
	entityMap["ps"] = &P{}
	entityMap["maps"] = &Map{}
	entityMap["pers"] = &Per{}

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
			status, data = retriever.RetrieveP(mux.Vars(req)["value1"])
		} else {
			status, data = retriever.RetrieveCoEntity(mux.Vars(req)["value1"])
		}
		w.WriteHeader(status)
		w.Write(data)
	}).Methods("GET")
}

func PUT(r *mux.Router) {
	r.HandleFunc("/{entity1}/{value1}/{entity2}/{value2}", func(w http.ResponseWriter, req *http.Request) {
		entity1, ok := mux.Vars(req)["entity1"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		entity2, ok := mux.Vars(req)["entity2"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		value1, ok := mux.Vars(req)["value1"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value2, ok := mux.Vars(req)["value2"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if isEntity(entity1) || isEntity(entity2) {
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
			status, data = updater.MoveFixedEntities(value1, elements, value2)
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
		entity1, ok := mux.Vars(req)["entity1"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		queries := req.URL.Query()
		count := queries.Get("count")
		elements := queries.Get("elements")

		updater, _ := entityMap[entity1].(Updater)
		var status int
		var data []byte
		if count != "" && canBooked(entity1) {
			status, data = updater.BookEntities(mux.Vars(req)["value1"], mux.Vars(req)["count"])
		} else if elements != "" && canBooked(entity1) {
			status, data = updater.BookFixedEntities(mux.Vars(req)["value1"], mux.Vars(req)["count"])
		} else {
			status, data = http.StatusNotFound, nil
		}

		w.WriteHeader(status)
		w.Write(data)
	}).Methods("PUT")
}

func DELETE(r *mux.Router) {
	r.HandleFunc("/{entity}/{value}", func(w http.ResponseWriter, req *http.Request) {
		entity, ok := mux.Vars(req)["entity"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		//TODO
		deleter, _ := entityMap[entity].(Deleter)
		status, data := deleter.Delete(mux.Vars(req)["value1"])
		w.WriteHeader(status)
		w.Write(data)
	}).Methods("DELETE")

	r.HandleFunc("/{entity1}/{value1}/{entity2}", func(w http.ResponseWriter, req *http.Request) {
		entity, ok := mux.Vars(req)["entity1"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		//TODO
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
	if field1 == "ps" && field2 == "pers" {
		return true
	} else if field1 == "pers" && field2 == "maps" {
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
	if field1 == "ps" && field2 == "pers" {
		return true
	} else if field1 == "pers" && field2 == "maps" {
		return true
	} else if field1 == "ps" && field2 == "pers" {
		return true
	} else if field1 == "ps" && field2 == "ms" {
		return true
	} else {
		return false
	}

}
