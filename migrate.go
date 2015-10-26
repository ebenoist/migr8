package main

import (
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"sync"
  "time"
)

var (
	source     = flag.String("source", "127.0.0.1:6379", "Redis server to pull data from")
	dest       = flag.String("dest", "127.0.0.1:6379", "Redis server to send data to")
	batch      = flag.Int("batch", 10, "Batch size")
  workers    = flag.Int("workers", 2, "Count of workers to spin up")
  key_prefix = flag.String("key_prefix", "", "The keys that you are interested in migrating")
  clear_dest = flag.Bool("clear_dest", false, "Clear the destination of all of it's keys and values")
)

type Task struct {
	list []string
}

var keys_processed uint64 = 0
var started_at time.Time

func key_processed() {
  // there is no mutex here, but I don't care as this is just information and does not need
  // to be accurate
  keys_processed += 1
  var duration time.Duration = time.Now().Sub(started_at)
  kps := float64(keys_processed) / float64(duration.Seconds())
  fmt.Printf("\r%v keys processd in %v KPS",keys_processed, kps)
}

func dump_and_restore(source_conn redis.Conn, dest_conn redis.Conn, key string) {
	dumped_key, err := redis.String(source_conn.Do("dump", key))
	dumped_key_ttl, err := redis.Int64(source_conn.Do("pttl", key))
	if err != nil {
		fmt.Println(err)
	}
	dest_conn.Do("restore", key, dumped_key_ttl, dumped_key)
  key_processed()
}

func source_connection(source string) redis.Conn {
	// attempt to connect to source server
	source_conn, err := redis.Dial("tcp", source)
	if err != nil {
		fmt.Println(err)
	}
	return source_conn
}

func dest_connection(dest string) redis.Conn {
	// attempt to connect to source server
	dest_conn, err := redis.Dial("tcp", dest)
	if err != nil {
		fmt.Println(err)
	}
	return dest_conn
}

func migrate_keys(queue chan Task, wg *sync.WaitGroup) {
  source_conn := source_connection(*source)
  dest_conn := dest_connection(*dest)
  for task := range queue {
    for _, key := range task.list {
      dump_and_restore(source_conn, dest_conn, key)
    }
  }
	wg.Done()
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

func clear_destination() {
  fmt.Println("Are you sure you want to delete all keys at", *dest, "? . Please type Y or N.")
  var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
    fmt.Println(err)
	}

  if response == "Y" {
    dest_conn := dest_connection(*dest)
    fmt.Println("Deleting all keys of destination")
    _, err := dest_conn.Do("flushall")
    if err != nil {
      fmt.Println(err)
    }
  } else {
    fmt.Println("Skipping key deletion")
  }

}

func main() {
	// parse the cli flags
	flag.Parse()

  if *clear_dest {
    clear_destination()
  }

  // wait for threads to finish
  wg := &sync.WaitGroup{}

  // make a buffered channel as a queue
	work_queue := make(chan Task, *workers)
  started_at = time.Now()

  // one more thing to wait for
  wg.Add(1)
  go scan_keys(work_queue, wg)

  // make the workers
	for i := 0; i <= *workers; i++ {
	  wg.Add(1)
    go migrate_keys(work_queue, wg)
  }
  wg.Wait()
}
