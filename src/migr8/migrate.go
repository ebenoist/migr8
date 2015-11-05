package main

import (
	"fmt"
	"sync"

	"github.com/garyburd/redigo/redis"
)

func dumpAndRestore(source_conn redis.Conn, dest_conn redis.Conn, key string) {
	dumped_key, err := redis.String(source_conn.Do("dump", key))
	if err != nil {
		fmt.Println(err)
	}
	dumped_key_ttl, err := redis.Int64(source_conn.Do("pttl", key))
	if err != nil {
		fmt.Println(err)
	}

	// when doing pttl, -1 means no expiration
	// when doing restore, 0 means no expiration
	if dumped_key_ttl == -1 {
		dumped_key_ttl = 0
	}

	dest_conn.Do("restore", key, dumped_key_ttl, dumped_key)
	keyProcessed()
}

func migrateKeys(queue chan Task, wg *sync.WaitGroup) {
	sourceConn := sourceConnection(config.Source)
	destConn := destConnection(config.Dest)

	for task := range queue {
		for _, key := range task.list {
			dumpAndRestore(sourceConn, destConn, key)
		}
	}

	wg.Done()
}

func clearDestination(dest string) {
	fmt.Println("Are you sure you want to delete all keys at", dest, "? . Please type Y or N.")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println(err)
	}

	if response == "Y" {
		dest_conn := destConnection(dest)
		fmt.Println("Deleting all keys of destination")
		_, err := dest_conn.Do("flushall")
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("Skipping key deletion")
	}
}
