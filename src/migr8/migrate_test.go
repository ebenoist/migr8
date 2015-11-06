package main

import (
	"fmt"
	"testing"

	"github.com/garyburd/redigo/redis"
)

func Test_MigrateAllKeysWithAPrefix(t *testing.T) {
	config = Config{
		Source:  sourceServer.url,
		Dest:    destServer.url,
		Workers: 1,
		Batch:   10,
		Prefix:  "bar",
	}

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("bar:%d", i)
		sourceServer.conn.Do("SET", key, i)
	}

	sourceServer.conn.Do("SET", "baz:foo", "yolo")

	RunAction(migrateKeys)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("bar:%d", i)
		exists, _ := redis.Bool(destServer.conn.Do("EXISTS", key))

		if !exists {
			t.Errorf("Could not find a key %d that should have been migrated", key)
		}
	}

	exists, _ := redis.Bool(destServer.conn.Do("EXISTS", "baz:foo"))

	if exists {
		t.Errorf("Found a key %s that should not have been migrated", "baz:foo")
	}
}
