package redis

import (
	"fmt"
	"sync"
	"time"
)

var (
	redis      *Redis
	redisMutex sync.Mutex

	redisAddr   string
	maxIdle     int
	maxActive   int
	idleTimeout int64
)

func GetRedisInstance() *Redis {
	if redis == nil {
		redisMutex.Lock()
		if redis == nil {
			redis = NewRedis(redisAddr, maxIdle, maxActive, time.Duration*idleTimeout)
		}
		redisMutex.Unlock()
	}
	return redis
}

type Redis struct {
	Pool        *redis.Pool
	addr        string
	maxIdle     int
	maxActive   int
	idleTimeout time.Duration
}

func NewRedis(addr string, maxIdle, maxActive int, idleTimeout time.Duration) *Redis {
	redis := &Redis{addr: addr, maxIdle: maxIdle, maxActive: maxActive, idleTimeout: idleTimeout}
	redis.initRedisPool()
	return redis
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
