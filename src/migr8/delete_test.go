package main

import (
	"fmt"
	"testing"

	"github.com/garyburd/redigo/redis"
)

func Test_DeleteAllKeysWithPrefix(t *testing.T) {
	config = Config{
		Source:  sourceServer.url,
		Workers: 1,
		Batch:   10,
		Prefix:  "bar",
	}

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("bar:%d", i)
		sourceServer.conn.Do("SET", key, i)
	}

	sourceServer.conn.Do("SET", "baz:foo", "yolo")

	RunAction(deleteKeys)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("bar:%d", i)
		exists, _ := redis.Bool(sourceServer.conn.Do("EXISTS", key))

		if exists {
			t.Errorf("Found a key %d that should have been deleted", key)
		}
	}

	exists, _ := redis.Bool(sourceServer.conn.Do("EXISTS", "baz:foo"))

	if !exists {
		t.Errorf("Deleted a key %s that should not have been deleted", "baz:foo")
	}
}
