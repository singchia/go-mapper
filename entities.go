package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"./lua-procedure"
	"./redis"
)

type Creator interface {
	Create(e string) (int, []byte)
}

type Retriever interface {
	//retrieve all entities
	Retrieve() (int, []byte)

	//only M need to implement
	RetrieveP(e string) (int, []byte)

	//retrieve included entities
	RetrieveEntities(e string) (int, []byte)

	//retrieve corresponding entity
	RetrieveCoEntity(e string) (int, []byte)
}

type Deleter interface {
	//delete entity
	Delete(e string) (int, []byte)
	DeAssign(e string) (int, []byte)
	MultiDeAssign(e string) (int, []byte)
}

type Updater interface {
	Assign(e1, e2 string) (int, []byte)
	MultiAssign(e1, e2 string) (int, []byte)

	MoveEntities(e1 string, count string, e2 string) (int, []byte)
	MoveFixedEntities(e1, elements, e2 string) (int, []byte)

	BookEntities(e string, count string) (int, []byte)
	TakeEntities(e1 string, e2 string) (int, []byte)
	BookFixedEntities(e, elements string) (int, []byte)
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

func (t *BaseEntity) Retrieve() (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) RetrieveP() (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) RetrieveEntities(e string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) RetrieveCoEntity(e string) (int, []byte) {
	return http.StatusNotFound, nil
}

func (t *BaseEntity) Assign(e1, e2 string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) MultiAssign(e1, e2 string) (int, []byte) {
	return http.StatusNotFound, nil
}

func (t *BaseEntity) MoveEntities(e1 string, count string, e2 string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) MoveFixedEntities(e1 string, elements string, e2 string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) BookEntities(e string, count string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) TakeEntities(e1 string, e2 string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) BookFixedEntities(e string, elements string) (int, []byte) {
	return http.StatusNotFound, nil
}

func (t *BaseEntity) Delete(e string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) DeAssign(e string) (int, []byte) {
	return http.StatusNotFound, nil
}
func (t *BaseEntity) MultiDeAssign(e string) (int, []byte) {
	return http.StatusNotFound, nil
}

type Topology struct {
	BaseEntity
}

func (t *Topology) Retrieve() (int, []byte) {
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrieveTopo, 0)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retSlice, ok := ret.([]uint8)
	if ok {
		return http.StatusOK, retSlice
	}
	return http.StatusInternalServerError, nil
}

type M struct {
	BaseEntity
}

func (m *M) Retrieve() (int, []byte) {
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrieveAllMs)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retSliceSlice, ok := ret.([]interface{})
	if ok {
		list := make([]string, 0, len(retSliceSlice))
		for _, v := range retSliceSlice {
			vSlice, ok := v.([]uint8)
			if ok {
				list = append(list, string(vSlice))
			}
		}
		data, _ := json.Marshal(struct {
			Ms []string `json:"ms"`
		}{Ms: list})

		return http.StatusOK, data
	} else {
		return http.StatusInternalServerError, nil
	}
}

func (m *M) Create(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaCreateM, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusCreated, nil
		case AlreayExists:
			return http.StatusOK, nil
		case ContradictionExists:
			return http.StatusConflict, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil

}

func (m *M) Delete(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaDeleteM, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusAccepted, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (m *M) RetrieveCoEntity(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrieveMap, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retStr, ok := ret.([]uint8)
	if ok {
		data, _ := json.Marshal(struct {
			Map string `json:"map"`
		}{Map: string(retStr)})

		return http.StatusOK, data
	}

	_, ok = ret.(int64)
	if ok {
		return http.StatusNoContent, nil
	}
	return http.StatusInternalServerError, nil
}

func (m *M) RetrieveP() (int, []byte) {
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrievePsByM)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retSlice, ok := ret.([]uint8)
	if ok {
		data, _ := json.Marshal(struct {
			P string `json:"p"`
		}{P: string(retSlice)})

		return http.StatusOK, data
	}
	_, ok = ret.(int64)
	if ok {
		return http.StatusNoContent, nil
	}
	return http.StatusInternalServerError, nil
}

func (m *M) Assign(e1, e2 string) (int, []byte) {
	if e1 == "" || e2 == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaAssignM2Map, 1, e1, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusOK, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (m *M) MultiAssign(e1, e2 string) (int, []byte) {
	if e1 == "" || e2 == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaAssignMultiM2Map, 1, e1, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusOK, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (m *M) DeAssign(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaDeleteMAssignment, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusOK, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (m *M) MultiDeAssign(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaDeleteMultiMAssignment, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusOK, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

type Map struct {
	BaseEntity
}

func (m *Map) Retrieve() (int, []byte) {
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrieveAllMaps, 0)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retSliceSlice, ok := ret.([]interface{})
	if ok {
		list := make([]string, 0, len(retSliceSlice))
		for _, v := range retSliceSlice {
			vSlice, ok := v.([]uint8)
			if ok {
				list = append(list, string(vSlice))
			}
		}
		data, _ := json.Marshal(struct {
			Maps []string `json:"maps"`
		}{Maps: list})

		return http.StatusOK, data
	} else {
		return http.StatusInternalServerError, nil
	}
}

func (m *Map) RetrieveEntities(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrieveMs, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retSliceSlice, ok := ret.([]interface{})
	if ok {
		list := make([]string, 0, len(retSliceSlice))
		for _, v := range retSliceSlice {
			vSlice, ok := v.([]uint8)
			if ok {
				list = append(list, string(vSlice))
			}
		}
		data, _ := json.Marshal(struct {
			Ms []string `json:"ms"`
		}{Ms: list})

		return http.StatusOK, data
	} else {
		return http.StatusInternalServerError, nil
	}
}

func (m *Map) RetrieveCoEntity(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrievePer, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retStr, ok := ret.([]uint8)
	if ok {
		data, _ := json.Marshal(struct {
			Per string `json:"per"`
		}{Per: string(retStr)})

		return http.StatusOK, data
	}

	_, ok = ret.(int64)
	if ok {
		return http.StatusNoContent, nil
	}
	return http.StatusInternalServerError, nil
}

func (m *Map) Create(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaCreateMap, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusCreated, nil
		case AlreayExists:
			return http.StatusOK, nil
		case ContradictionExists:
			return http.StatusConflict, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (m *Map) Delete(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaDeleteMap, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusAccepted, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (m *Map) Assign(e1, e2 string) (int, []byte) {
	if e1 == "" || e2 == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaAssignMap2Per, 1, e1, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusOK, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (m *Map) DeAssign(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaDeleteMapAssignment, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusOK, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (m *Map) MoveEntities(e1 string, countStr string, e2 string) (int, []byte) {
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		return http.StatusBadRequest, nil
	}
	if e1 == "" || e2 == "" {
		return http.StatusBadRequest, nil
	}
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaMoveMs, 1, e1, countStr, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		if retInt64 == EntityNotExists {
			return http.StatusNoContent, nil
		}
		data, _ := json.Marshal(struct {
			Count int64 `json:"count"`
		}{Count: retInt64})

		return http.StatusOK, data
	}
	return http.StatusInternalServerError, nil
}

func (m *Map) MoveFixedEntities(e1 string, multiE string, e2 string) (int, []byte) {
	if e1 == "" || e2 == "" || multiE == "" {
		return http.StatusBadRequest, nil
	}
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaMoveFixedMs, 1, e1, multiE, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		if retInt64 == EntityNotExists {
			return http.StatusNoContent, nil
		}
		data, _ := json.Marshal(struct {
			Count int64 `json:"count"`
		}{Count: retInt64})

		return http.StatusOK, data
	}
	return http.StatusInternalServerError, nil
}

func (m *Map) BookEntities(e string, countStr string) (int, []byte) {
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 || e == "" {
		return http.StatusBadRequest, nil
	}
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaBookMs, 1, e, countStr)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case EntitiesLocked:
			return http.StatusLocked, nil
		case EntitiesNotEnough:
			return http.StatusNoContent, nil
		case SADDOperError:
			return http.StatusInternalServerError, nil
		default:
			data, _ := json.Marshal(struct {
				Count int64 `json:"count"`
			}{Count: retInt64})

			return http.StatusOK, data
		}
	}
	return http.StatusInternalServerError, nil
}

func (m *Map) BookFixedEntities(e string, multiE string) (int, []byte) {
	if e == "" || multiE == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaBookMs, 1, e, multiE)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case EntitiesLocked:
			return http.StatusLocked, nil
		case EntitiesNotEnough:
			return http.StatusNoContent, nil
		case SADDOperError:
			return http.StatusInternalServerError, nil
		default:
			data, _ := json.Marshal(struct {
				Count int64 `json:"count"`
			}{Count: retInt64})

			return http.StatusOK, data
		}
	}
	return http.StatusInternalServerError, nil
}

func (m *Map) TakeEntities(e1 string, e2 string) (int, []byte) {
	if e1 == "" || e2 == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaTakeMs, 1, e1, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case EntityNotExists:
			return http.StatusNoContent, nil
		case SADDOperError:
			return http.StatusInternalServerError, nil
		default:
			data, _ := json.Marshal(struct {
				Count int64 `json:"count"`
			}{Count: retInt64})

			return http.StatusOK, data
		}
	}
	return http.StatusInternalServerError, nil
}

type P struct {
	BaseEntity
}

func (p *P) Retrieve() (int, []byte) {
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrieveAllPs, 0)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retSliceSlice, ok := ret.([]interface{})
	if ok {
		list := make([]string, 0, len(retSliceSlice))
		for _, v := range retSliceSlice {
			vSlice, ok := v.([]uint8)
			if ok {
				list = append(list, string(vSlice))
			}
		}
		data, _ := json.Marshal(struct {
			Ps []string `json:"ps"`
		}{Ps: list})

		return http.StatusOK, data
	} else {
		return http.StatusInternalServerError, nil
	}
}

func (p *P) RetrieveCoEntity(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrievePer, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retStr, ok := ret.([]uint8)
	if ok {
		data, _ := json.Marshal(struct {
			Per string `json:"per"`
		}{Per: string(retStr)})

		return http.StatusOK, data
	}

	_, ok = ret.(int64)
	if ok {
		return http.StatusNoContent, nil
	}
	return http.StatusInternalServerError, nil

}

func (p *P) Create(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaCreateP, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusCreated, nil
		case AlreayExists:
			return http.StatusOK, nil
		case ContradictionExists:
			return http.StatusConflict, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (p *P) Delete(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaDeleteP, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusAccepted, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (p *P) Assign(e1, e2 string) (int, []byte) {
	if e1 == "" || e2 == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaAssignP2Per, 1, e1, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusOK, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (p *P) MultiAssign(e1, e2 string) (int, []byte) {
	if e1 == "" || e2 == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaAssignMultiP2Per, 1, e1, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusOK, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (p *P) DeAssign(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaDeletePAssignment, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusOK, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (p *P) MultiDeAssign(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaDeleteMultiPAssignment, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusOK, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

type Per struct {
	BaseEntity
}

func (p *Per) Retrieve() (int, []byte) {
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrieveAllPers, 0)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retSliceSlice, ok := ret.([]interface{})
	if ok {
		list := make([]string, 0, len(retSliceSlice))
		for _, v := range retSliceSlice {
			vSlice, ok := v.([]uint8)
			if ok {
				list = append(list, string(vSlice))
			}
		}
		data, _ := json.Marshal(struct {
			Pers []string `json:"pers"`
		}{Pers: list})

		return http.StatusOK, data
	} else {
		return http.StatusInternalServerError, nil
	}
}

func (p *Per) RetrieveEntities(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaRetrievePs, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retSliceSlice, ok := ret.([]interface{})
	if ok {
		list := make([]string, 0, len(retSliceSlice))
		for _, v := range retSliceSlice {
			vSlice, ok := v.([]uint8)
			if ok {
				list = append(list, string(vSlice))
			}
		}
		data, _ := json.Marshal(struct {
			Ps []string `json:"ps"`
		}{Ps: list})

		return http.StatusOK, data
	} else {
		return http.StatusInternalServerError, nil
	}
}

func (p *Per) Create(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaCreatePer, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusCreated, nil
		case AlreayExists:
			return http.StatusOK, nil
		case ContradictionExists:
			return http.StatusConflict, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (p *Per) Delete(e string) (int, []byte) {
	if e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaDeletePer, 1, e)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case Succeed:
			return http.StatusAccepted, nil
		case EntityNotExists:
			return http.StatusNoContent, nil
		default:
			return http.StatusInternalServerError, nil
		}
	}
	return http.StatusInternalServerError, nil
}

func (p *Per) MoveEntities(e1 string, countStr string, e2 string) (int, []byte) {
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 || e1 == "" || e2 == "" {
		return http.StatusBadRequest, nil
	}
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaMovePs, 1, e1, countStr, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		if retInt64 == EntityNotExists {
			return http.StatusNoContent, nil
		}
		data, _ := json.Marshal(struct {
			Count int64 `json:"count"`
		}{Count: retInt64})

		return http.StatusOK, data
	}
	return http.StatusInternalServerError, nil
}

func (p *Per) MoveFixedEntities(e1 string, multiE string, e2 string) (int, []byte) {
	if e1 == "" || e2 == "" || multiE == "" {
		return http.StatusBadRequest, nil
	}
	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaMoveFixedPs, 1, e1, multiE, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		if retInt64 == EntityNotExists {
			return http.StatusNoContent, nil
		}
		data, _ := json.Marshal(struct {
			Count int64 `json:"count"`
		}{Count: retInt64})

		return http.StatusOK, data
	}
	return http.StatusInternalServerError, nil
}

func (p *Per) BookEntities(e string, countStr string) (int, []byte) {
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 || e == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaBookPs, 1, e, countStr)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case EntitiesLocked:
			return http.StatusLocked, nil
		case EntitiesNotEnough:
			return http.StatusNoContent, nil
		case SADDOperError:
			return http.StatusInternalServerError, nil
		default:
			data, _ := json.Marshal(struct {
				Count int64 `json:"count"`
			}{Count: retInt64})

			return http.StatusOK, data
		}
	}
	return http.StatusInternalServerError, nil
}

func (p *Per) BookFixedEntities(e string, multiE string) (int, []byte) {
	if e == "" || multiE == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaBookPs, 1, e, multiE)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case EntitiesLocked:
			return http.StatusLocked, nil
		case EntitiesNotEnough:
			return http.StatusNoContent, nil
		case SADDOperError:
			return http.StatusInternalServerError, nil
		default:
			data, _ := json.Marshal(struct {
				Count int64 `json:"count"`
			}{Count: retInt64})

			return http.StatusOK, data
		}
	}
	return http.StatusInternalServerError, nil
}

func (p *Per) TakeEntities(e1 string, e2 string) (int, []byte) {
	if e1 == "" || e2 == "" {
		return http.StatusBadRequest, nil
	}

	ret, err := redisCli.GetRedisInstance().Do("EVAL", procedure.LuaTakePs, 1, e1, e2)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	retInt64, ok := ret.(int64)
	if ok {
		switch retInt64 {
		case EntityNotExists:
			return http.StatusNoContent, nil
		case SADDOperError:
			return http.StatusInternalServerError, nil
		default:
			data, _ := json.Marshal(struct {
				Count int64 `json:"count"`
			}{Count: retInt64})

			return http.StatusOK, data
		}
	}
	return http.StatusInternalServerError, nil
}
