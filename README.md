# migr8
Redis Migration Utility written in Go

## Running and Building

Steps on how to build and run the programs

**Migrate:**

Install dependencies:
```
  go get github.com/garyburd/redigo/redis
```

Run inline:
```
  go run migrate.go -source=old.com:6379 -dest=new.com:6379 -key_prefix=reverb:service:bump -batch=1000 -workers=50 -clear_dest=true
```

Compile for Linux:
```
  GOOS=linux go build migrate.go
```
