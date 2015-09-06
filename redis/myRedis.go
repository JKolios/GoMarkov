package redis

import (
	"github.com/mediocregopher/radix.v2/redis"
	"log"
)

// NewConnection establishes a new connection to a Redis instance
func NewConnection(host, port string) *redis.Client {

	client, err := redis.Dial("tcp", host+":"+port)
	if err != nil {
		log.Println("Error while connecting:" + err.Error())
		return nil
	}
	return client
}
