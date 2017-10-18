package redisCli

import (
	"fmt"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	redisCli   *Redis
	redisMutex sync.Mutex

	RedisAddr   string
	MaxIdle     int
	MaxActive   int
	IdleTimeout int64
)

func GetRedisInstance() *Redis {
	if redisCli == nil {
		redisMutex.Lock()
		if redisCli == nil {
			redisCli = NewRedis(RedisAddr, MaxIdle, MaxActive, time.Duration(IdleTimeout)*time.Second)
		}
		redisMutex.Unlock()
	}
	return redisCli
}

type Redis struct {
	Pool        *redis.Pool
	addr        string
	maxIdle     int
	maxActive   int
	idleTimeout time.Duration
}

func NewRedis(addr string, maxIdle, maxActive int, idleTimeout time.Duration) *Redis {
	redisCli := &Redis{addr: addr, maxIdle: maxIdle, maxActive: maxActive, idleTimeout: idleTimeout}
	redisCli.initRedisPool()
	return redisCli
}
func (r *Redis) initRedisPool() {
	r.Pool = &redis.Pool{
		MaxIdle:     r.maxIdle,
		MaxActive:   r.maxActive,
		IdleTimeout: r.idleTimeout,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", r.addr)
		},
	}
}

// GetConn 从redis pool中获取一个redis Conn
func (r *Redis) GetConn() redis.Conn {
	if r.Pool == nil {
		return nil
	}
	return r.Pool.Get()
}

// Do ...
func (r *Redis) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	conn := r.GetConn()
	if conn == nil {
		return nil, fmt.Errorf("redis.Pool GetConn failed")
	}
	defer conn.Close()

	return conn.Do(commandName, args...)
}

// Close ...
func (r *Redis) Close() {
	if r.Pool != nil {
		r.Pool.Close()
		r.Pool = nil
	}
}
