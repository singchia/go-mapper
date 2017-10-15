package proceduer

import (
	"testing"

	"github.com/garyburd/redigo/redis"
)

var conn redis.Conn

func Init() (err error) {
	conn, err = redis.Dial("tcp", "127.0.0.1:6379")
	return
}

func Fini() (err error) {
	return conn.Close()
}

func concreteEVAL(lua_scritp string, t *testing.T, args ...interface{}) {
	err := Init()
	if err != nil {
		t.Error(err)
		return
	}
	defer Fini()

	var ret interface{}
	switch len(args) {
	case 0:
		ret, err = conn.Do("EVAL", lua_scritp, 0)
	case 1:
		ret, err = conn.Do("EVAL", lua_scritp, 1, args[0])
	case 2:
		ret, err = conn.Do("EVAL", lua_scritp, 1, args[0], args[1])
	case 3:
		ret, err = conn.Do("EVAL", lua_scritp, 1, args[0], args[1], args[2])
	default:
		t.Error("wrong count of args")
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	retInt64, ok := ret.(int64)
	//some move operations return succeed count
	if ok && retInt64 != 0 && retInt64 != 40021 && retInt64 > 1000 {
		t.Errorf("redis operation error number: %d", retInt64)
	} else if ok && (retInt64 == 0 || retInt64 == 40021 || retInt64 <= 1000) {
		//do nothing
	} else {
		retSlice, ok := ret.([]uint8)
		if ok {
			t.Log(string(retSlice))
		} else {
			t.Errorf("unexpected type, ret: %v", ret)
			return
		}
	}

	topo, err := conn.Do("EVAL", LuaRetrieveTopo, 0)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(topo.([]uint8)))
	Fini()
}

//go test -v -test.run Test_RetrieveTopo
func Test_RetrieveTopo(t *testing.T) {
	err := Init()
	if err != nil {
		t.Error(err)
		return
	}
	defer Fini()

	topo, err := conn.Do("EVAL", LuaRetrieveTopo, 0)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(topo.([]uint8)))
}

//go test -v -test.run Test_CreateMap$
func Test_CreateMap(t *testing.T) {
	concreteEVAL(LuaCreateMap, t, "map2")
}

//go test -v -test.run TestCreateM$
func Test_CreateM(t *testing.T) {
	concreteEVAL(LuaCreateM, t, "m1")
}

//go test -v -test.run Test_CreatePer$
func Test_CreatePer(t *testing.T) {
	concreteEVAL(LuaCreatePer, t, "per2")
}

//go test -v -test.run TestCreateP$
func Test_CreateP(t *testing.T) {
	concreteEVAL(LuaCreateP, t, "p1")
}

//test UPDATE
func Test_AssignM2Map(t *testing.T) {
	concreteEVAL(LuaAssignM2Map, t, "m1", "map2")
}

func Test_AssignMultiM2Map(t *testing.T) {
	concreteEVAL(LuaAssignMultiM2Map, t, "m2 m1", "map2")
}

func Test_AssignP2Per(t *testing.T) {
	concreteEVAL(LuaAssignP2Per, t, "p1", "per1")
}

func Test_AssignMultiP2Per(t *testing.T) {
	concreteEVAL(LuaAssignMultiP2Per, t, "p1 p2", "per2")
}

func Test_LuaAssignMap2Per(t *testing.T) {
	concreteEVAL(LuaAssignMap2Per, t, "map2", "per1")
}

func Test_LuaMovePs(t *testing.T) {
	concreteEVAL(LuaMovePs, t, "per2", 1, "per1")
}

func Test_LuaMoveMs(t *testing.T) {
	concreteEVAL(LuaMoveMs, t, "map1", 1, "map2")
}

func Test_LuaMoveFixedPs(t *testing.T) {
	concreteEVAL(LuaMoveFixedPs, t, "per1", "p1", "per2")
}

func Test_LuaMoveFixedMs(t *testing.T) {
	concreteEVAL(LuaMoveFixedMs, t, "map2", "m1", "map1")
}

func Test_LuaBookPs(t *testing.T) {
	concreteEVAL(LuaBookPs, t, "per2", 1)
}

func Test_LuaTakePs(t *testing.T) {
	concreteEVAL(LuaTakePs, t, "per2", "per1")
}

func Test_LuaBookFixedPs(t *testing.T) {
	concreteEVAL(LuaBookFixedPs, t, "per2", "p1")
}

func Test_LuaBookMs(t *testing.T) {
	concreteEVAL(LuaBookMs, t, "map1", 1)
}

func Test_LuaTakeMs(t *testing.T) {
	concreteEVAL(LuaTakeMs, t, "map2", "map2")
}

func Test_LuaBookFixedMs(t *testing.T) {
	concreteEVAL(LuaBookFixedMs, t, "map2", "m1")
}
