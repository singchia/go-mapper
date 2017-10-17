package main

import (
	"net/http"

	"./lua-procedure"
	"./redis"
)

type Creator interface {
	Create(e string) (int, []byte)
}

type Retriever interface {
	//retrieve all entities
	Retrieve(e string) (int, []byte)

	//retrieve corresponding entity
	RetrieveCoEntity(e1 string, e2 string) (int, []byte)
}

type Deleter interface {
	//delete entity
	Delete(e string) (int, []byte)
}

type Updater interface {
	Assign(e1, e2 string) (int, []byte)
	MoveEntities(e1 string, count int, e2 string) (int, []byte)
	MoveFixedEntities(e1, elements, e2 string) (int, []byte)
	BookEntities(e string, count int) (int, []byte)
	TakeEntities(e string, count int) (int, []byte)
	BookFixedEntities(e, elements string) (int []byte)
	TakeFixedEntities(e, elements string) (int []byte)
}

type CURD interface {
	Creator
	Retriever
	Deleter
	Updater
}

type BaseEntity struct{}

func (t *BaseEntity) Create(e string) (int, []byte) {
	return http.StatusNotFound, nil
}

func (t *BaseEntity) Retrieve(e string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) RetrieveCoEntity(e1 string, e2 string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) Assign(e1, e2 string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) MoveEntities(e1 string, count int, e2 string) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) MoveFixedEntities(e1 string, elements string, e2 string) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) BookEntities(e string, count int) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) TakeEntities(e string, count int) (int, []byte) {
	return http.StatusNotFound, nil
}

func (t *BaseEntity) BookFixedEntities(e string, elements string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) TakeFixedEntities(e string, elements string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) Delete(e string) (int, []byte) {
	return http.StatusNotFound, nil
}

type Topology struct {
	BaseEntity
}

func (t *Topology) Retrieve() (int, []byte) {
	ret, err := redis.GetRedisInstance().Do("EVAL", procedure.LuaRetrieveTopo, 0)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retSlice, ok := ret.([]uint8)
	if ok {
		return http.StatusOK, retSlice
	}
	return http.StatusInternalServerError, nil
}
