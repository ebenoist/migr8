package main

import (
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"sync"
)

var (
	mode    string
	source  = flag.String("source", "127.0.0.1:6379", "Redis server to pull data from")
	batch   = flag.Int("batch", 10, "Batch size")
	workers = flag.Int("workers", 2, "Count of workers to spin up")
	match   = flag.String("match", "12345", "prefix of keys to delete")
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

func scan_keys(queue chan Task, wg *sync.WaitGroup) {
	var cursor int
	cursor = 0
	// this will store the keys of each iteration
	var tmp_keys []string
	conn := source_connection(*source)

	key_search := fmt.Sprintf("%s*", *key_prefix)
	fmt.Println("Starting Scan with keys", key_search)

	for {
		// we scan with our cursor offset, starting at 0
		reply, _ := redis.Values(conn.Do("scan", cursor, "match", key_search, "count", *batch))

		// this func name is confusing...it actually just converts array returns to Go values
		redis.Scan(reply, &cursor, &tmp_keys)

		// put this thing in the queue
		queue <- Task{list: tmp_keys}
		// check if we need to stop...
		if cursor == 0 {
			fmt.Println("Finished!")

			// close the channel
			close(queue)
			wg.Done()
			break
		}
	}
}

func delete_keys(queue chan Task, wg *sync.WaitGroup) {
	source_conn := source_connection(*source)
	for task := range queue {
		for _, key := range task.list {
			delete(source_conn, key)
		}
	}
	wg.Done()
}

func main() {
	// parse the cli flags
	flag.Parse()

	// grab all source keys
	wg := &sync.WaitGroup{}
	work_queue := make(chan Task, *workers)

	go scan_keys(work_queue, wg)

	// Start the scanner
	wg.Add(1)
	// Start the delete workers
	for i := 0; i <= *workers; i++ {
		wg.Add(1)
		go delete_keys(work_queue, wg)
	}

	// Wait for all goroutines to complete
	wg.Wait()
}
