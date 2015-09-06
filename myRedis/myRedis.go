package myRedis

import (
	"github.com/mediocregopher/radix.v2/pool"
	"log"
)

const DefaultConnectionCount = 10
const DefaultArrayKey = "markov"
const MinStoredStrings = 1000
const MaxStoredStrings = 5000

type RedisState struct {
	ConnectionPool *pool.Pool
	lineCount      int
}

// NewConnection establishes a new connection to a Redis instance
func InitRedis(host, port string) *RedisState {

	pool, err := pool.New("tcp", host+":"+port, DefaultConnectionCount)
	if err != nil {
		log.Println("Error while creating connection pool:" + err.Error())
		return nil
	}

	initialLines, err := pool.Cmd("LLEN", DefaultArrayKey).Int()
	if err != nil {
		log.Println("Error initializing connection pool:" + err.Error())
		return nil
	}
	log.Printf("Initialized connection pool with count:%v", initialLines)
	return &RedisState{pool, initialLines}
}

func (redis *RedisState) Write(p []byte) (n int, err error) {
	reply, err := redis.ConnectionPool.Cmd("RPUSH", DefaultArrayKey, string(p)).Int()
	if err != nil {
		log.Printf("Redis Write error:%v", err.Error())
		return 0, err
	}
	redis.lineCount++
	return reply, err

}

func (redis *RedisState) GetString() (s string, err error) {
	reply, err := redis.ConnectionPool.Cmd("LPOP", DefaultArrayKey).Str()
	if err != nil {
		log.Printf("Redis Read error:%v", err.Error())
		return "", err
	}
	redis.lineCount--
	return reply, err

}

func (redis *RedisState) Full() bool {
	return redis.lineCount >= MaxStoredStrings
}

func (redis *RedisState) Low() bool {
	return redis.lineCount <= MinStoredStrings
}
