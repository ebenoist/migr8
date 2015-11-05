package main

import (
	"log"
	"sync"

	"github.com/garyburd/redigo/redis"
)

func deleteKey(source_conn redis.Conn, key string) {
	redis.String(source_conn.Do("del", key))
	log.Printf("Deleted %s \n", key)
}

func deleteKeys(queue chan Task, wg *sync.WaitGroup) {
	sourceConn := sourceConnection(config.Source)
	for task := range queue {
		for _, key := range task.list {
			deleteKey(sourceConn, key)
		}
	}

	wg.Done()
}
