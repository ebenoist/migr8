.PHONY: all test build linux

all: test build
clean:
	@rm -rf bin/*

build:
	@gb build

test:
	@gb test

linux:
	@GOOS=linux gb build
