package main

import (
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"sync"
)

var (
	mode     string
	source   = flag.String("source", "127.0.0.1:6379", "Redis server to pull data from")
	batch    = flag.Int("batch", 10, "Batch size")
	scanners = flag.Int("scanners", 1, "Count of scanners to spin up")
	workers  = flag.Int("workers", 2, "Count of workers to spin up")
	match    = flag.String("match", "12345", "regex of keys to delete")
)

type Task struct {
	list []string
}

func delete(source_conn redis.Conn, key string) {
	redis.String(source_conn.Do("del", key))
	fmt.Printf("Deleted %s \n", key)
}

func source_connection(source string) redis.Conn {
	// attempt to connect to source server
	source_conn, err := redis.Dial("tcp", source)
	if err != nil {
		fmt.Println(err)
	}
	return source_conn
}

func main() {
	// parse the cli flags
	flag.Parse()

	// grab all source keys
	var wg sync.WaitGroup
	work_queue := make(chan Task, *workers)

	// Start the scanner
	wg.Add(1)
	go func(wg *sync.WaitGroup, work_queue chan Task) {
		defer close(work_queue)
		var returned_keys []string
		var cursor int64
		source_conn := source_connection(*source)

		// Initial redis scan
		source_keys, _ := redis.Values(source_conn.Do("scan", "0", "count", *batch, "match", *match))
		redis.Scan(source_keys, &cursor, &returned_keys)
		work_queue <- Task{list: returned_keys}

		for {
			reply, _ := redis.Values(source_conn.Do("scan", cursor, "count", *batch, "match", *match))
			redis.Scan(reply, &cursor, &returned_keys)
			// Set the cursor to the next page
			if len(returned_keys) == 0 && cursor == 0 {
				wg.Done()
				break
			}
			work_queue <- Task{list: returned_keys}
		}

	}(&wg, work_queue)

	// Start the delete workers
	for i := 0; i <= *workers; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			source_conn := source_connection(*source)
			for job := range work_queue {
				for _, key := range job.list {
					delete(source_conn, key)
				}
			}
		}(&wg)

	}

	// Wait for all goroutines to complete
	wg.Wait()
}
