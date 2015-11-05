migr8
---

Redis Migration Utility written in Go

## Build
migr8 uses [gb](http://getgb.io) to vendor dependencies. To install it run, `go get github.com/constabulary/gb/...`

`make build`

## Usage
```
NAME:
   migr8 - It's time to move some redis

USAGE:
   bin/migr8 [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
   migrate	Migrate one redis to a new redis
   delete
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --source, -s "127.0.0.1:6379"	The redis server to pull data from
   --dest, -d "127.0.0.1:6379"		The destination redis server
   --workers, -w "2"			The count of workers to spin up
   --batch, -b "10"			The batch size
   --prefix, -p 			The key prefix to act on
   --clear-dest, -c			Clear the destination of all it's keys and values
   --help, -h				show help
   --version, -v			print the version
```

## Compile for Linux:
`make release`
